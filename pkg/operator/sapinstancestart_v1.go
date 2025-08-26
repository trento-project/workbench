package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trento-project/workbench/internal/sapcontrol"
)

const (
	SapInstanceStartOperatorName    = "sapinstancestart"
	defaultSapInstanceStateTimeout  = 5 * time.Minute
	defaultSapInstanceStateInterval = 10 * time.Second
)

type sapInstanceStartDiffOutput struct {
	Started bool `json:"started"`
}

type sapStateChangeArguments struct {
	instNumber string
	timeout    time.Duration
}

type SAPInstanceStartOption Option[SAPInstanceStart]

// SAPInstanceStart operator starts a SAP instance.
//
// Arguments:
//  instance_number (required): String with the instance number of the instance to start
//  timeout: Timeout in seconds to wait until the instance is started
//
// # Execution Phases
//
// - PLAN:
//   The operator gets the instance current processes and stores the state.
//   The operation is skipped if the SAP instances is already started.
//
// - COMMIT:
//   It starts the SAP instance using the sapcontrol Start command.
//
// - VERIFY:
//   Verify if the SAP instance is started.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the instance is stopped back again.

type SAPInstanceStart struct {
	baseOperator
	parsedArguments     *sapStateChangeArguments
	sapControlConnector sapcontrol.SAPControlConnector
	interval            time.Duration
}

func WithCustomStartSapcontrol(sapControlConnector sapcontrol.SAPControlConnector) SAPInstanceStartOption {
	return func(o *SAPInstanceStart) {
		o.sapControlConnector = sapControlConnector
	}
}

func WithCustomStartInterval(interval time.Duration) SAPInstanceStartOption {
	return func(o *SAPInstanceStart) {
		o.interval = interval
	}
}

func NewSAPInstanceStart(
	arguments Arguments,
	operationID string,
	options OperatorOptions[SAPInstanceStart],
) *Executor {
	sapInstanceStart := &SAPInstanceStart{
		baseOperator: newBaseOperator(
			SapInstanceStartOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		interval: defaultSapInstanceStateInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(sapInstanceStart)
	}

	return &Executor{
		phaser:      sapInstanceStart,
		operationID: operationID,
		logger:      sapInstanceStart.logger,
	}
}

func (s *SAPInstanceStart) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSAPStateChangeArguments(s.arguments)
	if err != nil {
		return false, err
	}
	s.parsedArguments = opArguments

	// Use custom sapControlConnector or create a new one based on the instance_number argument
	if s.sapControlConnector == nil {
		s.sapControlConnector = sapcontrol.NewSAPControlConnector(s.parsedArguments.instNumber)
	}

	started, err := allProcessesInState(ctx, s.sapControlConnector, sapcontrol.STATECOLORSAPControlGREEN)
	if err != nil {
		return false, fmt.Errorf("error checking processes state: %w", err)
	}

	s.resources[beforeDiffField] = started

	if started {
		s.logger.Info("instance already started, skipping operation")
		s.resources[afterDiffField] = started
		return true, nil
	}

	return false, nil
}

func (s *SAPInstanceStart) commit(ctx context.Context) error {
	request := new(sapcontrol.Start)
	_, err := s.sapControlConnector.StartContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting instance: %w", err)
	}

	return nil
}

func (s *SAPInstanceStart) verify(ctx context.Context) error {
	err := waitUntilSapInstanceState(
		ctx,
		s.sapControlConnector,
		sapcontrol.STATECOLORSAPControlGREEN,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("verify instance started failed: %w", err)
	}

	s.resources[afterDiffField] = true
	return nil
}

func (s *SAPInstanceStart) rollback(ctx context.Context) error {
	request := new(sapcontrol.Stop)
	_, err := s.sapControlConnector.StopContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping instance: %w", err)
	}

	err = waitUntilSapInstanceState(
		ctx,
		s.sapControlConnector,
		sapcontrol.STATECOLORSAPControlGRAY,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("rollback to stopped failed: %w", err)
	}

	return nil
}

func (s *SAPInstanceStart) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := sapInstanceStartDiffOutput{
		Started: s.resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff["before"] = string(before)

	afterDiffOutput := sapInstanceStartDiffOutput{
		Started: s.resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff["after"] = string(after)

	return diff
}

func allProcessesInState(
	ctx context.Context,
	connector sapcontrol.SAPControlConnector,
	expectedState sapcontrol.STATECOLOR,
) (bool, error) {
	request := new(sapcontrol.GetProcessList)
	response, err := connector.GetProcessListContext(ctx, request)
	if err != nil {
		return false, fmt.Errorf("error getting instance process list: %w", err)
	}

	// GetProcessList can return an empty list for some seconds when the instance
	// is started. Discard this scenario.
	if len(response.Processes) == 0 {
		return false, nil
	}

	for _, process := range response.Processes {
		if *process.Dispstatus != expectedState {
			return false, nil
		}
	}

	return true, nil
}

func waitUntilSapInstanceState(
	ctx context.Context,
	connector sapcontrol.SAPControlConnector,
	expectedState sapcontrol.STATECOLOR,
	timeout time.Duration,
	interval time.Duration,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		inState, err := allProcessesInState(timeoutCtx, connector, expectedState)
		if err != nil {
			return err
		}

		if timeoutCtx.Err() != nil {
			return fmt.Errorf("error waiting until instance is in desired state")
		}

		if inState {
			return nil
		}

		err = sleepContext(timeoutCtx, interval)
		if err != nil {
			return err
		}
	}

}

// sleepContext sleeps the running thread until the interval or the context are completed
func sleepContext(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func parseSAPStateChangeArguments(rawArguments Arguments) (*sapStateChangeArguments, error) {
	instNumberArgument, found := rawArguments["instance_number"]
	if !found {
		return nil, fmt.Errorf("argument instance_number not provided, could not use the operator")
	}

	instNumber, ok := instNumberArgument.(string)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse instance_number argument as string, argument provided: %v",
			instNumberArgument,
		)
	}

	timeout := defaultSapInstanceStateTimeout
	if timeoutArgument, found := rawArguments["timeout"]; found {
		timeoutFloat, ok := timeoutArgument.(float64)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse timeout argument as a number, argument provided: %v",
				timeoutArgument,
			)
		}

		timeout = time.Duration(timeoutFloat) * time.Second
	}

	return &sapStateChangeArguments{
		instNumber: instNumber,
		timeout:    timeout,
	}, nil
}
