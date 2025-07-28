package operator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/trento-project/workbench/internal/crm"
)

const (
	CrmClusterStartOperatorName = "crmclusterstart"
)

type CrmClusterStart struct {
	baseOperator
	parsedArguments *crmClusterStartArguments
	crmClient       crm.Crm
}

type CrmClusterStartOption Option[CrmClusterStart]

type crmClusterStartDiffOutput struct {
	Started bool `json:"started"`
}
func WithCustomCrmClient(clusterClient cluster.Cluster) CrmClusterStartOption {
func WithCustomCrmClient(crmClient cluster.Cluster) CrmClusterStartOption {
	return func(c *CrmClusterStart) {
		c.clusterClient = clusterClient
	}
}

func WithCustomCrmClient(crmClient crm.Crm) CrmClusterStartOption {
	return func(c *CrmClusterStart) {
		c.crmClient = crmClient
	}
}

func NewCrmClusterStart(arguments OperatorArguments,
	operationID string,
	options OperatorOptions[CrmClusterStart]) *Executor {
	crmClusterStart := &CrmClusterStart{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		crmClient:    crm.NewDefaultCrmClient(),
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
	opArguments, err := parseCrmClusterStartArguments(c.arguments)
	if err != nil {
		return false, fmt.Errorf("error parsing arguments: %w", err)
	}
	c.parsedArguments = opArguments

	// Check if the provided cluster ID matches the one found in the system.
	foundClusterID, err := c.crmClient.GetClusterId()
	if err != nil {
		return false, fmt.Errorf("error getting cluster ID: %w", err)
	}
	if foundClusterID != c.parsedArguments.clusterID {
		return false, fmt.Errorf("cluster ID mismatch: expected %s, found %s",
			c.parsedArguments.clusterID, foundClusterID)
	}

	// TODO: check if the cluster is already started.

	return true, nil
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

func parseCrmClusterStartArguments(rawArguments OperatorArguments) (*crmClusterStartArguments, error) {
	if rawArguments == nil {
		return nil, errors.New("arguments cannot be nil")
	}

	clusterID, ok := rawArguments["cluster_id"].(string)
	if !ok || clusterID == "" {
		return nil, errors.New("invalid or missing cluster_id argument")
	}

	return &crmClusterStartArguments{
		clusterID: clusterID,
	}, nil
}
