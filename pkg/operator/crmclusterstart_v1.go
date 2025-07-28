package operator

import (
	"context"
	"errors"
	"time"
)

const (
	CrmClusterStartOperatorName = "crmclusterstart"
)

type CrmClusterStart struct {
	baseOperator
}

type CrmClusterStartOption Option[CrmClusterStart]

func NewCrmClusterStart(arguments OperatorArguments,
	operationID string,
	options OperatorOptions[CrmClusterStart]) *Executor {
	crmClusterStart := &CrmClusterStart{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		// interval:     defaultCrmClusterInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(crmClusterStart)
	}

	return &Executor{
		phaser:      crmClusterStart,
		operationID: operationID,
		logger:      crmClusterStart.logger,
	}
}

func (c *CrmClusterStart) plan(ctx context.Context) (bool, error) {
	return false, errors.New("not implemented yet")
}

func (c *CrmClusterStart) commit(ctx context.Context) error {
	return errors.New("not implemented yet")
}

func (c *CrmClusterStart) rollback(ctx context.Context) error {
	return errors.New("not implemented yet")
}

func (c *CrmClusterStart) verify(ctx context.Context) error {
	return errors.New("not implemented yet")
}

func (c *CrmClusterStart) operationDiff(ctx context.Context) map[string]any {
	return map[string]any{
		"error": errors.New("not implemented yet").Error(),
	}
}

func (c *CrmClusterStart) after(ctx context.Context) {
	// not implemented yet
}
