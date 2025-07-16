package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/trento-project/workbench/internal/sapcontrol"
)

const (
	SapSystemStartOperatorName        = "sapsystemstart"
	defaultSapSystemStateTimeout      = 5 * time.Minute
	defaultSapSystemStateInterval     = 10 * time.Second
	defaultSapSystemStateInstanceType = sapcontrol.StartStopOptionSAPControlALLINSTANCES

	instanceTypeALL    = "all"
	instanceTypeABAP   = "abap"
	instanceTypeJ2EE   = "j2ee"
	instanceTypeSCS    = "scs"
	instanceTypeENQREP = "enqrep"
)

type sapSystemStartDiffOutput struct {
	Started bool `json:"started"`
}

type sapSystemStateChangeArguments struct {
	instNumber   string
	timeout      time.Duration
	instanceType sapcontrol.StartStopOption
}

type SAPSystemStartOption Option[SAPSystemStart]

// SAPSystemStart operator starts a SAP system.
//
// Arguments:
//  instance_number (required): String with the instance number of local instance to start the whole system
//  timeout: Timeout in seconds to wait until the system is started
//  instance_type: Instance type to filter in the StartSystem process. Values: all|abap|j2ee|scs|enqrep. Default value: all
//
// # Execution Phases
//
// - PLAN:
//   The operator gets the system current instances and stores the state.
//   The operation is skipped if the SAP system is already started.
//
// - COMMIT:
//   It starts the SAP system using the sapcontrol StartSystem command.
//
// - VERIFY:
//   Verify if the SAP system is started.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the system is stopped back again.

type SAPSystemStart struct {
	baseOperator
	parsedArguments     *sapSystemStateChangeArguments
	sapControlConnector sapcontrol.SAPControlConnector
	interval            time.Duration
}

func WithCustomStartSystemSapcontrol(sapControlConnector sapcontrol.SAPControlConnector) SAPSystemStartOption {
	return func(o *SAPSystemStart) {
		o.sapControlConnector = sapControlConnector
	}
}

func WithCustomStartSystemInterval(interval time.Duration) SAPSystemStartOption {
	return func(o *SAPSystemStart) {
		o.interval = interval
	}
}

func NewSAPSystemStart(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[SAPSystemStart],
) *Executor {
	sapSystemStart := &SAPSystemStart{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		interval:     defaultSapSystemStateInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(sapSystemStart)
	}

	return &Executor{
		phaser:      sapSystemStart,
		operationID: operationID,
	}
}

func (s *SAPSystemStart) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSAPSystemStateChangeArguments(s.arguments)
	if err != nil {
		return false, err
	}
	s.parsedArguments = opArguments

	// Use custom sapControlConnector or create a new one based on the instance_number argument
	if s.sapControlConnector == nil {
		s.sapControlConnector = sapcontrol.NewSAPControlConnector(s.parsedArguments.instNumber)
	}

	started, err := allInstancesInState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrol.STATECOLORSAPControlGREEN,
	)
	if err != nil {
		return false, fmt.Errorf("error checking system state: %w", err)
	}

	s.resources[beforeDiffField] = started

	if started {
		s.logger.Info("system already started, skipping operation")
		s.resources[afterDiffField] = started
		return true, nil
	}

	return false, nil
}

func (s *SAPSystemStart) commit(ctx context.Context) error {
	request := new(sapcontrol.StartSystem)
	request.Options = &s.parsedArguments.instanceType
	_, err := s.sapControlConnector.StartSystemContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting system: %w", err)
	}

	return nil
}

func (s *SAPSystemStart) verify(ctx context.Context) error {
	err := waitUntilSapSystemState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrol.STATECOLORSAPControlGREEN,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("verify system started failed: %w", err)
	}

	s.resources[afterDiffField] = true
	return nil
}

func (s *SAPSystemStart) rollback(ctx context.Context) error {
	request := new(sapcontrol.StopSystem)
	request.Options = &s.parsedArguments.instanceType
	_, err := s.sapControlConnector.StopSystemContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping system: %w", err)
	}

	err = waitUntilSapSystemState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrol.STATECOLORSAPControlGRAY,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("rollback to stopped failed: %w", err)
	}

	return nil
}

func (s *SAPSystemStart) operationDiff(ctx context.Context) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := sapSystemStartDiffOutput{
		Started: s.resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff["before"] = string(before)

	afterDiffOutput := sapSystemStartDiffOutput{
		Started: s.resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff["after"] = string(after)

	return diff
}

func allInstancesInState(
	ctx context.Context,
	connector sapcontrol.SAPControlConnector,
	instanceType sapcontrol.StartStopOption,
	expectedState sapcontrol.STATECOLOR,
) (bool, error) {
	request := new(sapcontrol.GetSystemInstanceList)
	response, err := connector.GetSystemInstanceListContext(ctx, request)
	if err != nil {
		return false, fmt.Errorf("error getting instance list: %w", err)
	}

	filteringMap := map[sapcontrol.StartStopOption]string{
		sapcontrol.StartStopOptionSAPControlALLINSTANCES:    "",
		sapcontrol.StartStopOptionSAPControlABAPINSTANCES:   "ABAP",
		sapcontrol.StartStopOptionSAPControlJ2EEINSTANCES:   "J2EE",
		sapcontrol.StartStopOptionSAPControlSCSINSTANCES:    "MESSAGESERVER",
		sapcontrol.StartStopOptionSAPControlENQREPINSTANCES: "ENQREP",
	}
	filteringValue := filteringMap[instanceType]

	for _, instance := range response.Instances {
		// filter out instances that are not part of the current instance type value
		if !strings.Contains(instance.Features, filteringValue) {
			continue
		}

		if *instance.Dispstatus != expectedState {
			return false, nil
		}
	}

	return true, nil
}

func waitUntilSapSystemState(
	ctx context.Context,
	connector sapcontrol.SAPControlConnector,
	instanceType sapcontrol.StartStopOption,
	expectedState sapcontrol.STATECOLOR,
	timeout time.Duration,
	interval time.Duration,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		inState, err := allInstancesInState(timeoutCtx, connector, instanceType, expectedState)
		if err != nil {
			return err
		}

		if timeoutCtx.Err() != nil {
			return fmt.Errorf("error waiting until system is in desired state")
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

func parseSAPSystemStateChangeArguments(rawArguments OperatorArguments) (*sapSystemStateChangeArguments, error) {
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

	timeout := defaultSapSystemStateTimeout
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

	instanceType := defaultSapSystemStateInstanceType
	if instanceTypeArgument, found := rawArguments["instance_type"]; found {
		instanceTypeStr, ok := instanceTypeArgument.(string)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse instance_type argument as a string, argument provided: %v",
				instanceTypeArgument,
			)
		}

		instancesMap := map[string]sapcontrol.StartStopOption{
			instanceTypeALL:    sapcontrol.StartStopOptionSAPControlALLINSTANCES,
			instanceTypeABAP:   sapcontrol.StartStopOptionSAPControlABAPINSTANCES,
			instanceTypeJ2EE:   sapcontrol.StartStopOptionSAPControlJ2EEINSTANCES,
			instanceTypeSCS:    sapcontrol.StartStopOptionSAPControlSCSINSTANCES,
			instanceTypeENQREP: sapcontrol.StartStopOptionSAPControlENQREPINSTANCES,
		}
		instanceType, ok = instancesMap[instanceTypeStr]
		if !ok {
			return nil, fmt.Errorf("invalid instance_type value: %s", instanceTypeStr)
		}
	}

	return &sapSystemStateChangeArguments{
		instNumber:   instNumber,
		timeout:      timeout,
		instanceType: instanceType,
	}, nil
}
