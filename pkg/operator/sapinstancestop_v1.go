package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trento-project/workbench/internal/sapcontrol"
)

const (
	SapInstanceStopOperatorName = "sapinstancestop"
)

type sapInstanceStopDiffOutput struct {
	Stopped bool `json:"stopped"`
}

type SAPInstanceStopOption Option[SAPInstanceStop]

type SAPInstanceStop struct {
	baseOperator
	parsedArguments     *sapStateChangeArguments
	sapControlConnector sapcontrol.SAPControlConnector
	interval            time.Duration
}

func WithCustomStopSapcontrol(sapControlConnector sapcontrol.SAPControlConnector) SAPInstanceStopOption {
	return func(o *SAPInstanceStop) {
		o.sapControlConnector = sapControlConnector
	}
}

func WithCustomStopInterval(interval time.Duration) SAPInstanceStopOption {
	return func(o *SAPInstanceStop) {
		o.interval = interval
	}
}

func NewSAPInstanceStop(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[SAPInstanceStop],
) *Executor {
	sapInstanceStop := &SAPInstanceStop{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		interval:     defaultSapInstanceStateInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(sapInstanceStop)
	}

	return &Executor{
		phaser:      sapInstanceStop,
		operationID: operationID,
	}
}

func (s *SAPInstanceStop) plan(ctx context.Context) error {
	opArguments, err := parseSAPStateChangeArguments(s.arguments)
	if err != nil {
		return err
	}
	s.parsedArguments = opArguments

	// Use custom sapControlConnector or create a new one based on the instance_number argument
	if s.sapControlConnector == nil {
		s.sapControlConnector = sapcontrol.NewSAPControlConnector(s.parsedArguments.instNumber)
	}

	stopped, err := allProcessesInState(ctx, s.sapControlConnector, sapcontrol.STATECOLORSAPControlGRAY)
	if err != nil {
		return fmt.Errorf("error checking processes state: %w", err)
	}

	s.resources[beforeDiffField] = stopped

	return nil
}

func (s *SAPInstanceStop) commit(ctx context.Context) error {
	if s.resources[beforeDiffField] == true {
		s.logger.Info("instance already stopped, skipping operation")
		return nil
	}

	request := new(sapcontrol.Stop)
	_, err := s.sapControlConnector.StopContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping instance: %w", err)
	}

	return waitUntilSapInstanceState(
		ctx,
		s.sapControlConnector,
		sapcontrol.STATECOLORSAPControlGRAY,
		s.parsedArguments.timeout,
		s.interval,
	)
}

func (s *SAPInstanceStop) verify(ctx context.Context) error {
	stopped, err := allProcessesInState(ctx, s.sapControlConnector, sapcontrol.STATECOLORSAPControlGRAY)
	if err != nil {
		return fmt.Errorf("error checking processes state: %w", err)
	}

	if stopped {
		s.resources[afterDiffField] = stopped
		return nil
	}

	return fmt.Errorf(
		"verify instance stopped failed, instance was not stopped in commit phase",
	)
}

func (s *SAPInstanceStop) rollback(ctx context.Context) error {
	request := new(sapcontrol.Start)
	_, err := s.sapControlConnector.StartContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting instance: %w", err)
	}

	return waitUntilSapInstanceState(
		ctx,
		s.sapControlConnector,
		sapcontrol.STATECOLORSAPControlGREEN,
		s.parsedArguments.timeout,
		s.interval,
	)
}

func (s *SAPInstanceStop) operationDiff(ctx context.Context) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := sapInstanceStopDiffOutput{
		Stopped: s.resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff["before"] = string(before)

	afterDiffOutput := sapInstanceStopDiffOutput{
		Stopped: s.resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff["after"] = string(after)

	return diff
}
