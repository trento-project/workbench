package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/trento-project/workbench/internal/support"
)

const (
	ClusterMaintenanceChangeOperatorName = "clustermaintenancechange"
	clusterIdlePattern                   = "S_IDLE"
	maintenanceOn                        = "on"
	maintenanceOff                       = "off"
)

var (
	clusterIdlePatternCompiled = regexp.MustCompile(clusterIdlePattern)
)

type ClusterMaintenanceChangeOption Option[ClusterMaintenanceChange]

type clusterMaintenanceChangeArguments struct {
	maintenance bool
	resourceID  string
}

type diffOutput struct {
	Maintenance bool   `json:"maintenance"`
	ResourceID  string `json:"resource_id,omitempty"`
}

// ClusterMaintenanceChange is an operator responsible for changing cluster maintenance
// and cluster resources managed state. `crmsh` is the used to tool to apply the write and most of
// the read operations in the cluster.
// The used commands differ if the the state to change is of the whole cluster or a particular resource.
//
// Find some helpful references about maintenance transitions and used commands:
// - https://www.suse.com/c/sles-for-sap-hana-maintenance-procedures-part-1-pre-maintenance-checks/
// - https://www.suse.com/c/sles-for-sap-hana-maintenance-procedures-part-2-manual-administrative-tasks-os-reboots-and-updation-of-os-and-hana/
// - https://crmsh.github.io/man-4.6/
// - https://crmsh.github.io/man-4.6/#cmdhelp_root_status
// - https://crmsh.github.io/man-4.6/#cmdhelp_maintenance
// - https://crmsh.github.io/man-4.6/#cmdhelp_resource
//
// The operator accepts the next arguments:
// - maintenance (bool): The desired maintenance state for the cluster or cluster resource.
//                       If true, the cluster or cluster resource are set in maintenance mode.
// - resource_id (string): If given, the operator changes the maintenance state of the resource.
//                         Otherwise, it changes the general maintenance state of the cluster.
//
// # Execution Phases
//
// - PLAN:
//   Check if a pacemaker cluster is present and store the current state.
//
// - COMMIT:
//   Change the cluster or cluster resource state if the cluster is in IDLE state.
//   If the maintenance state is removed, the cluster resource state is refreshed.
//
// - VERIFY:
//   Check if the cluster or cluster resource maintenance state has the expected value and
//   store the final state.
//
// - ROLLBACK:
//   Change the cluster or cluster resource state to the initial state if the cluster
//   is in IDLE state.

type ClusterMaintenanceChange struct {
	baseOperator
	parsedArguments *clusterMaintenanceChangeArguments
}

func NewClusterMaintenanceChange(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[ClusterMaintenanceChange],
) *Executor {
	clusterMaintenance := &ClusterMaintenanceChange{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
	}

	for _, opt := range options.OperatorOptions {
		opt(clusterMaintenance)
	}

	return &Executor{
		phaser:      clusterMaintenance,
		operationID: operationID,
	}
}

func (c *ClusterMaintenanceChange) plan(ctx context.Context) error {
	opArguments, err := parseClusterMaintenanceArguments(c.arguments)
	if err != nil {
		return err
	}
	c.parsedArguments = opArguments

	// check if a cluster is available and running
	_, err = c.executor.Exec(ctx, "crm", "status")
	if err != nil {
		return fmt.Errorf("error getting cluster status: %w", err)
	}

	currentState, err := getMaintenanceState(ctx, c.executor, c.parsedArguments.resourceID)
	if err != nil {
		return err
	}

	c.resources[beforeDiffField] = currentState

	return nil
}

func (c *ClusterMaintenanceChange) commit(ctx context.Context) error {
	if c.resources[beforeDiffField] == c.parsedArguments.maintenance {
		c.logger.Infof("maintenance state %v already set, skipping operation", c.parsedArguments.maintenance)
		return nil
	}

	err := isIdle(ctx, c.executor)
	if err != nil {
		return err
	}

	// refresh cluster or resource before removing maintenance state
	if !c.parsedArguments.maintenance {
		_, err = c.executor.Exec(ctx, "crm", "resource", "refresh", c.parsedArguments.resourceID)
		if err != nil {
			return fmt.Errorf("error updating maintenance state: %w", err)
		}
	}

	state := maintenanceStateString(c.parsedArguments.maintenance)

	_, err = c.executor.Exec(ctx, "crm", "maintenance", state, c.parsedArguments.resourceID)
	if err != nil {
		return fmt.Errorf("error updating maintenance state: %w", err)
	}

	return nil
}

func (c *ClusterMaintenanceChange) verify(ctx context.Context) error {
	currentState, err := getMaintenanceState(ctx, c.executor, c.parsedArguments.resourceID)
	if err != nil {
		return err
	}

	if c.parsedArguments.maintenance == currentState {
		c.resources[afterFieldDiff] = currentState
		return nil
	}

	return fmt.Errorf(
		"verify cluster maintenance failed, the maintenance value %v was not set in commit phase",
		c.parsedArguments.maintenance,
	)
}

func (c *ClusterMaintenanceChange) rollback(ctx context.Context) error {
	err := isIdle(ctx, c.executor)
	if err != nil {
		return err
	}

	initialState, _ := c.resources[beforeDiffField].(bool)
	state := maintenanceStateString(initialState)

	_, err = c.executor.Exec(ctx, "crm", "maintenance", state, c.parsedArguments.resourceID)
	if err != nil {
		return fmt.Errorf("error rolling back maintenance state: %w", err)
	}

	return nil
}

func (c *ClusterMaintenanceChange) operationDiff(ctx context.Context) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := diffOutput{
		Maintenance: c.resources[beforeDiffField].(bool),
		ResourceID:  c.parsedArguments.resourceID,
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff["before"] = string(before)

	afterDiffOutput := diffOutput{
		Maintenance: c.resources[afterFieldDiff].(bool),
		ResourceID:  c.parsedArguments.resourceID,
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff["after"] = string(after)

	return diff
}

// getMaintanceState returns the current state of the cluster
// Find additional information here:
// https://clusterlabs.org/projects/pacemaker/doc/2.1/Pacemaker_Explained/html/resources.html#resource-meta-attributes
func getMaintenanceState(ctx context.Context, executor support.CmdExecutor, resourceID string) (bool, error) {
	// general cluster state
	if len(resourceID) == 0 {
		maintenanceMode, err := executor.Exec(ctx, "crm", "configure", "get_property", "-t", "maintenance-mode")
		if err != nil {
			return false, fmt.Errorf("error getting maintenance-mode: %w", err)
		}

		boolValue, err := parseStateOutput(maintenanceMode)
		if err != nil {
			return false, fmt.Errorf("error decoding maintenance-mode attribute: %w", err)
		}

		return boolValue, nil
	}

	// specific resource state
	// get "maintenance" attribute of the resource. This has preference over is-managed attribute
	output, err := executor.Exec(ctx, "crm", "resource", "meta", resourceID, "show", "maintenance")
	if err != nil {
		return false, fmt.Errorf("error getting maintenance attribute: %w", err)
	}

	if !strings.Contains(string(output), "not found") {
		boolValue, err := parseStateOutput(output)
		if err != nil {
			return false, fmt.Errorf("error decoding maintenance attribute: %w", err)
		}

		return boolValue, nil
	}

	// get "is-managed" attribute of the resource
	output, err = executor.Exec(ctx, "crm", "resource", "meta", resourceID, "show", "is-managed")
	if err != nil {
		return false, fmt.Errorf("error getting is-managed attribute: %w", err)
	}

	// none of maintenance or is-managed attributes found. Defaulting to not in maintenance
	if strings.Contains(string(output), "not found") {
		return false, nil
	}

	boolValue, err := parseStateOutput(output)
	if err != nil {
		return false, fmt.Errorf("error decoding is-managed attribute: %w", err)
	}

	// is-managed has the opposite logic than maintenance attribute
	return !boolValue, nil
}

func isIdle(ctx context.Context, executor support.CmdExecutor) error {
	idleOutput, err := executor.Exec(ctx, "cs_clusterstate", "-i")
	if err != nil {
		return fmt.Errorf("error running cs_clusterstate: %w", err)
	}

	if !clusterIdlePatternCompiled.Match(idleOutput) {
		return fmt.Errorf("cluster is not in S_IDLE state")
	}

	return nil
}

func maintenanceStateString(boolState bool) string {
	state := maintenanceOff
	if boolState {
		state = maintenanceOn
	}

	return state
}

// depending on the queried resource, the crm command might print some "debug" lines
// before returning the actual state of the attribute. That's why it needs some cleanup
// Example output:
// linux # crm resource meta msl_SAPHana_PRD_HDB00 show maintenance
// msl_SAPHana_PRD_HDB00 is active on more than one node, returning the default value for maintenance
// false
func parseStateOutput(output []byte) (bool, error) {
	trimmedString := strings.TrimSpace(string(output))
	if len(trimmedString) == 0 {
		return false, fmt.Errorf("empty command output")
	}

	lines := strings.Split(trimmedString, "\n")
	lastLine := lines[len(lines)-1]

	boolValue, err := strconv.ParseBool(lastLine)
	if err != nil {
		return false, err
	}
	return boolValue, nil
}

func parseClusterMaintenanceArguments(rawArguments OperatorArguments) (*clusterMaintenanceChangeArguments, error) {
	maintenanceArgument, found := rawArguments["maintenance"]
	if !found {
		return nil, errors.New("argument maintenance not provided, could not use the operator")
	}

	maintenance, ok := maintenanceArgument.(bool)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse maintenance argument as bool, argument provided: %v",
			maintenanceArgument,
		)
	}

	argument, found := rawArguments["resource_id"]
	if !found {
		return &clusterMaintenanceChangeArguments{maintenance: maintenance, resourceID: ""}, nil
	}

	resourceID, ok := argument.(string)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse resource_id argument as string, argument provided: %v",
			argument,
		)
	}

	return &clusterMaintenanceChangeArguments{maintenance: maintenance, resourceID: resourceID}, nil
}
