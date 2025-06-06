package operator

import (
	"context"
	"errors"
	"fmt"

	"github.com/trento-project/workbench/internal/hana"
	"github.com/trento-project/workbench/internal/support"
)

const UnregisterHANASecondaryOperatorName = "unregisterhanasecondary"

type unregisterHANASecondaryDiffOutput struct {
	Unregistered bool `json:"unregistered"`
}

type unregisterHanaSecondaryArguments struct {
	sid string
}

type UnregisterHANASecondaryOption Option[UnregisterHANASecondary]

// UnregisterHANASecondary is an operator responsible for unregistering a HANA secondary instance.
//
// Arguments:
// 		sid: The instance number of the HANA secondary instance to unregister.
//
// # Execution Phases
// - PLAN:
// - COMMIT:
// - VERIFY:
// - ROLLBACK:

type UnregisterHANASecondary struct {
	baseOperator
	parsedArguments *unregisterHanaSecondaryArguments
	hdbnsutil       hana.Hdbnsutil
}

func Withhdbnsutil(hdbnsutil hana.Hdbnsutil) UnregisterHANASecondaryOption {
	return func(o *UnregisterHANASecondary) {
		o.hdbnsutil = hdbnsutil
	}
}

func parseUnregisterHanaSecondaryArguments(rawArguments OperatorArguments) (*unregisterHanaSecondaryArguments, error) {
	argument, found := rawArguments["sid"]
	if !found {
		return nil, errors.New("argument sid not provided, could not use the operator")
	}

	instanceNumber, ok := argument.(string)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse sid argument as string, argument provided: %v",
			argument,
		)
	}

	if instanceNumber == "" {
		return nil, errors.New("sid argument is empty")
	}

	return &unregisterHanaSecondaryArguments{
		sid: instanceNumber,
	}, nil
}

func NewUnregisterHANASecondary(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[UnregisterHANASecondary],
) *Executor {
	unregisterHANASecondary := &UnregisterHANASecondary{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
	}

	unregisterHANASecondary.hdbnsutil = hana.NewHdbnsutil(
		support.CliExecutor{},
		unregisterHANASecondary.logger,
	)

	for _, opt := range options.OperatorOptions {
		opt(unregisterHANASecondary)
	}

	return &Executor{
		phaser:      unregisterHANASecondary,
		operationID: operationID,
	}
}

func (uhs *UnregisterHANASecondary) plan(ctx context.Context) (alreadyApplied bool, err error) {
	opArguments, err := parseUnregisterHanaSecondaryArguments(uhs.arguments)
	if err != nil {
		return false, err
	}
	uhs.parsedArguments = opArguments

	srState, err := uhs.hdbnsutil.SystemReplicationState(ctx, uhs.parsedArguments.sid)
	if err != nil {
		return false, err
	}

	if !srState.IsRegistered() {
		uhs.logger.Infof("HANA secondary instance %s is already unregistered, skipping operation", uhs.parsedArguments.sid)
		return true, nil
	}

	return true, nil
}

func (uhs *UnregisterHANASecondary) commit(ctx context.Context) error {
	return uhs.hdbnsutil.UnregisterHANASecondary(ctx, uhs.parsedArguments.sid)
}

func (uhs *UnregisterHANASecondary) verify(ctx context.Context) error {
	return nil
}

func (uhs *UnregisterHANASecondary) rollback(ctx context.Context) error {
	return nil
}

func (uhs *UnregisterHANASecondary) operationDiff(ctx context.Context) map[string]any {
	return map[string]any{
		"sid": uhs.parsedArguments.sid,
	}
}
