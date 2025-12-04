package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	clusterMocks "github.com/trento-project/workbench/internal/cluster/mocks"
	"github.com/trento-project/workbench/internal/support/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

const fakeID = "some-id"

func TestClusterMaintenanceChangeSuccessOn(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false}",
		"after":  "{\"maintenance\":true}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeSuccessOff(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("true"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"refresh",
		"",
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": false,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":true}",
		"after":  "{\"maintenance\":false}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeResourceSuccess(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	resourceID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("false"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
		resourceID,
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"resource_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"resource_id\":\"some-id\"}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeResourceWithIsManagedSuccess(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	resourceID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("not found"), nil)

	// is-managed has the reverse boolean logic than `maintenance`
	// so is-managed=true means maintenance=false
	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"is-managed",
	).Return([]byte("true"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
		resourceID,
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"is-managed",
	).Return([]byte("false"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"resource_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"resource_id\":\"some-id\"}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeResourceDefaultSuccess(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	resourceID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("not found"), nil).Once()

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"is-managed",
	).Return([]byte("not found"), nil)

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
		resourceID,
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"resource_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"resource_id\":\"some-id\"}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeNodeSuccessOn(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	nodeID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=off"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"--force",
		"node",
		"maintenance",
		nodeID,
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"node_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"node_id\":\"some-id\"}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeNodeSuccessOnWithoutPreviousState(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	nodeID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return(
		[]byte("scope=nodes  name=maintenance value=(null)"),
		errors.New("error getting node state"),
	).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"--force",
		"node",
		"maintenance",
		nodeID,
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"node_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"node_id\":\"some-id\"}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeNodeSuccessOff(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	nodeID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=true"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"resource",
		"refresh",
		"",
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"--force",
		"node",
		"ready",
		nodeID,
	).Return([]byte("ok"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=off"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": false,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":true,\"node_id\":\"some-id\"}",
		"after":  "{\"maintenance\":false,\"node_id\":\"some-id\"}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeMissingArgument(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "argument maintenance not provided, could not use the operator", report.Error.Message)
}

func TestClusterMaintenanceChangeInvalidArgument(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": "on",
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "could not parse maintenance argument as bool, argument provided: on", report.Error.Message)
}

func TestClusterMaintenanceChangeInvalidResourceIDArgument(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": 1,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "could not parse resource_id argument as string, argument provided: 1", report.Error.Message)
}

func TestClusterMaintenanceChangeInvalidNodeIDArgument(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     1,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "could not parse node_id argument as string, argument provided: 1", report.Error.Message)
}

func TestClusterMaintenanceChangeMutuallyExclusiveArgument(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": "some-resource",
			"node_id":     "some-node",
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "resource_id and node_id arguments are mutually exclusive, use only one of them", report.Error.Message)
}

func TestClusterMaintenanceChangePlanClusterNotFound(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(false)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "cluster is not runnint on host", report.Error.Message)
}

func TestClusterMaintenanceChangePlanGetMaintenanceError(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("error"), errors.New("cannot get state"))

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "error getting maintenance-mode: cannot get state", report.Error.Message)
}

func TestClusterMaintenanceChangePlanEmptyMaintenanceState(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte(""), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "error decoding maintenance-mode attribute: empty command output", report.Error.Message)
}

func TestClusterMaintenanceChangePlanNodeNotFound(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()
	nodeID := fakeID

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("Could not map name=some-id to a UUID"), errors.New("error getting node"))

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, "error getting node maintenance attribute: error getting node", report.Error.Message)
}

func TestClusterMaintenanceChangeCommitAlreadyApplied(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":true}",
		"after":  "{\"maintenance\":true}",
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.PLAN)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestClusterMaintenanceChangeCommitNotIdle(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	mockClusterClient.On("IsIdle", ctx).Return(false, nil).Once()
	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("ok"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.COMMIT)
	assert.EqualValues(t, "cluster is not in S_IDLE state", report.Error.Message)
}

func TestClusterMaintenanceChangeVerifyError(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil).Once()

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("ok"), nil).Once()

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("ok"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.VERIFY)
	assert.EqualValues(t, "verify cluster maintenance failed, the maintenance value true was not set in commit phase", report.Error.Message)
}

func TestClusterMaintenanceChangeRollbackNotIdle(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	mockClusterClient.On("IsIdle", ctx).Return(true, nil).Once()

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("error"), errors.New("error changing"))

	mockClusterClient.On("IsIdle", ctx).Return(false, nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.ROLLBACK)
	assert.EqualValues(t, "cluster is not in S_IDLE state\nerror updating maintenance state: error changing", report.Error.Message)
}

func TestClusterMaintenanceChangeRollbackErrorReverting(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	mockClusterClient := clusterMocks.NewMockCrm(t)
	ctx := context.Background()

	mockClusterClient.On("IsHostOnline", ctx).Return(true)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("error"), errors.New("error changing"))

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("error"), errors.New("error reverting"))

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.ROLLBACK)
	assert.EqualValues(t, "error rolling back maintenance state: error reverting\nerror updating maintenance state: error changing", report.Error.Message)
}
