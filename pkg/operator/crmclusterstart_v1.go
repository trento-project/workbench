// CrmClusterStart operator starts a CRM cluster.
//
// Arguments:
//  cluster_id (required): String representing the expected cluster ID to start.
//
// # Execution Phases
//
// - PLAN:
//   Checks if the CRM cluster is already online. If it is, the operation is skipped.
//   If the cluster is offline, it checks if the cluster is idle before proceeding.
//   If the cluster is not idle, it returns an error.
//
// - COMMIT:
//   Starts the CRM cluster using the crmClient's StartCluster method.
//
// - VERIFY:
//   Verifies if the CRM cluster is online after the start operation, using exponential backoff retries.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the cluster is stopped using exponential backoff retries.
//
// # Details
//
// This operator is designed to safely start a CRM cluster, ensuring that the cluster is only started if it is offline.
// It uses a retry mechanism with exponential backoff for rollback and verification phases to handle transient failures.
// The operator provides detailed logging for each phase and maintains before/after state for diff reporting.

package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/trento-project/workbench/internal/crm"
	"github.com/trento-project/workbench/internal/support"
)

const (
	CrmClusterStartOperatorName = "crmclusterstart"
)

type CrmClusterStart struct {
	baseOperator
	parsedArguments *crmClusterStartArguments
	crmClient       crm.Crm
	retry           retry
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

func WithCustomRetry(maxRetries int, initialDelay, maxDelay time.Duration, factor int) CrmClusterStartOption {
	return func(c *CrmClusterStart) {
		c.retryOptions = support.BackoffOptions{
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			MaxRetries:   maxRetries,
			Factor:       factor,
		}
	}
}

func NewCrmClusterStart(arguments OperatorArguments,
	operationID string,
	options OperatorOptions[CrmClusterStart]) *Executor {
	crmClusterStart := &CrmClusterStart{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		crmClient:    crm.NewDefaultCrmClient(),
		retry: struct {
			initialDelay time.Duration
			maxDelay     time.Duration
			maxRetries   int
		}{
			maxDelay:   8 * time.Second,
			maxRetries: 5,
		},
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
	// check if the cluster is already started.
	isOnline := c.crmClient.IsHostOnline(ctx)
	c.resources[beforeDiffField] = isOnline

	if isOnline {
		c.logger.Info("CRM cluster is already online, skipping start operation", "cluster_id", c.parsedArguments.clusterID, "phase", PLAN)
		c.resources[afterDiffField] = true
		return true, nil
	}

	c.logger.Info("CRM cluster is offline, will attempt to start it", "cluster_id", c.parsedArguments.clusterID, "phase", PLAN)
	return false, nil

}

func (c *CrmClusterStart) commit(ctx context.Context) error {
	c.logger.Info("Begin", "phase", COMMIT)
	err := c.crmClient.StartCluster(ctx)
	if err != nil {
		return fmt.Errorf("error starting CRM cluster: %w", err)
	}
	c.logger.Info("Success", "phase", COMMIT, "cluster_id", c.parsedArguments.clusterID)

	return nil
}

func (c *CrmClusterStart) rollback(ctx context.Context) error {
	c.logger.Info("Begin", "phase", ROLLBACK)

	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retry.maxRetries,
		c.retry.initialDelay,
		c.retryOptions,
		},
	)

	if result.Err != nil {
		c.logger.Error("Failed to rollback CRM cluster start operation", "error", result.Err, "phase", ROLLBACK)
		return fmt.Errorf("error rolling back CRM cluster start: %w", result.Err)
	}

	return nil
}

func (c *CrmClusterStart) verify(ctx context.Context) error {
	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retry.maxRetries,
		c.retry.initialDelay,
		c.retryOptions,
			return isOnline, nil
		},
			if !isOnline {
				return false, fmt.Errorf("CRM cluster is not online, expected online state")
			}
			return true, nil

	if result.Err != nil {
		return fmt.Errorf("error verifying CRM cluster start: %w", result.Err)
	}
		return result.Err
	c.resources[afterDiffField] = true
	return nil
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
