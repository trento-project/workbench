package operator

import (
	"context"
	"encoding/json"
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

type crmClusterStartDiffOutput struct {
	Started bool `json:"started"`
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
	isOnline := c.clusterClient.IsHostOnline(ctx)
	isOnline := c.crmClient.IsHostOnline(ctx)
	c.resources[beforeDiffField] = isOnline

	if isOnline {
		c.logger.Info("CRM cluster is already online", "cluster_id", c.parsedArguments.clusterID)
		c.resources[afterDiffField] = true
		return true, nil
	}

	c.logger.Info("CRM cluster is offline, will attempt to start it", "cluster_id", c.parsedArguments.clusterID)
	return false, nil

}

func (c *CrmClusterStart) commit(ctx context.Context) error {
	err := c.crmClient.StartCluster(ctx)
	if err != nil {
		return fmt.Errorf("error starting CRM cluster: %w", err)
	}
	c.logger.Info("CRM cluster start operation committed", "cluster_id", c.parsedArguments.clusterID)
	return nil
}

func (c *CrmClusterStart) rollback(ctx context.Context) error {
	return errors.New("not implemented yet")
}

func (c *CrmClusterStart) verify(ctx context.Context) error {
	return errors.New("not implemented yet")
}

func (c *CrmClusterStart) operationDiff(ctx context.Context) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := crmClusterStartDiffOutput{
		Started: c.resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff["before"] = string(before)

	afterDiffOutput := crmClusterStartDiffOutput{
		Started: c.resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff["after"] = string(after)

	return diff
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
