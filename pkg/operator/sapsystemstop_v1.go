package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trento-project/workbench/internal/sapcontrol"
)

const (
	SapSystemStopOperatorName = "sapsystemstop"
)

type sapSystemStopDiffOutput struct {
	Stopped bool `json:"stopped"`
}

type SAPSystemStopOption Option[SAPSystemStop]

// SAPSystemStop operator stops a SAP system.
//
// Arguments:
//  instance_number (required): String with the instance number of local instance to stop the whole system
//  timeout: Timeout in seconds to wait until the system is stopped
//  instance_type: Instance type to filter in the StopSystem process. Values: all|abap|j2ee|scs|enqrep.
//  Default value: all
//
// # Execution Phases
//
// - PLAN:
//   The operator gets the system current instances and stores the state.
//   The operation is skipped if the SAP system is already stopped.
//
// - COMMIT:
//   It stops the SAP system using the sapcontrol StopSystem command.
//
// - VERIFY:
//   Verify if the SAP system is stopped.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the system is started back again.

type SAPSystemStop struct {
	baseOperator
	parsedArguments     *sapSystemStateChangeArguments
	sapControlConnector sapcontrol.SAPControlConnector
	interval            time.Duration
}

func WithCustomStopSystemSapcontrol(sapControlConnector sapcontrol.SAPControlConnector) SAPSystemStopOption {
	return func(o *SAPSystemStop) {
		o.sapControlConnector = sapControlConnector
	}
}

func WithCustomStopSystemInterval(interval time.Duration) SAPSystemStopOption {
	return func(o *SAPSystemStop) {
		o.interval = interval
	}
}

func NewSAPSystemStop(
	arguments Arguments,
	operationID string,
	options Options[SAPSystemStop],
) *Executor {
	sapSystemStop := &SAPSystemStop{
		baseOperator: newBaseOperator(
			SapSystemStopOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		interval: defaultSapSystemStateInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(sapSystemStop)
	}

	return &Executor{
		phaser:      sapSystemStop,
		operationID: operationID,
		logger:      sapSystemStop.logger,
	}
}

func (s *SAPSystemStop) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSAPSystemStateChangeArguments(s.arguments)
	if err != nil {
		return false, err
	}
	s.parsedArguments = opArguments

	// Use custom sapControlConnector or create a new one based on the instance_number argument
	if s.sapControlConnector == nil {
		s.sapControlConnector = sapcontrol.NewSAPControlConnector(s.parsedArguments.instNumber)
	}

	stopped, err := allInstancesInState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrol.STATECOLORSAPControlGRAY,
	)
	if err != nil {
		return false, fmt.Errorf("error checking system state: %w", err)
	}

	s.resources[beforeDiffField] = stopped

	if stopped {
		s.logger.Info("system already stopped, skipping operation")
		s.resources[afterDiffField] = stopped
		return true, nil
	}

	return false, nil
}

func (s *SAPSystemStop) commit(ctx context.Context) error {
	request := new(sapcontrol.StopSystem)
	request.Options = &s.parsedArguments.instanceType
	_, err := s.sapControlConnector.StopSystemContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping system: %w", err)
	}

	return nil
}

func (s *SAPSystemStop) verify(ctx context.Context) error {
	err := waitUntilSapSystemState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrol.STATECOLORSAPControlGRAY,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("verify system stopped failed: %w", err)
	}

	s.resources[afterDiffField] = true
	return nil
}

func (s *SAPSystemStop) rollback(ctx context.Context) error {
	request := new(sapcontrol.StartSystem)
	request.Options = &s.parsedArguments.instanceType
	_, err := s.sapControlConnector.StartSystemContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting system: %w", err)
	}

	err = waitUntilSapSystemState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrol.STATECOLORSAPControlGREEN,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("rollback to started failed: %w", err)
	}

	return nil
}

func (s *SAPSystemStop) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := sapSystemStopDiffOutput{
		Stopped: s.resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff["before"] = string(before)

	afterDiffOutput := sapSystemStopDiffOutput{
		Stopped: s.resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff["after"] = string(after)

	return diff
}
