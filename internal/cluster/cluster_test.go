package cluster_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/cluster"
	"github.com/trento-project/workbench/internal/support/mocks"
)

type CrmTestSuite struct {
	suite.Suite
}

func TestCrm(t *testing.T) {
	suite.Run(t, new(CrmTestSuite))
}

func (suite *CrmTestSuite) TestIsHostOnlineTrue() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "status").Return([]byte("Online"), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	status := crmClient.IsHostOnline(ctx)
	suite.True(status, "Cluster should be online")
}

func (suite *CrmTestSuite) TestIsHostOnlineFalse() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "status").Return([]byte("Offline"), errors.New("cluster is not running"))

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	status := crmClient.IsHostOnline(ctx)
	suite.False(status, "Cluster should be offline")
}

func (suite *CrmTestSuite) TestIsIdle() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "cs_clusterstate", "-i").Return([]byte("Cluster state: S_IDLE"), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	isIdle, err := crmClient.IsIdle(ctx)
	suite.NoError(err, "IsIdle should not return an error")
	suite.True(isIdle, "Cluster should be idle")
}

func (suite *CrmTestSuite) TestIsIdleError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "cs_clusterstate", "-i").Return([]byte(""), errors.New("command failed"))

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	_, err := crmClient.IsIdle(ctx)

	suite.Error(err, "IsIdle should return an error")
}

func (suite *CrmTestSuite) TestIsIdleDifferentState() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "cs_clusterstate", "-i").Return([]byte("Cluster state: S_TRANSITION_ENGINE"), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	isIdle, err := crmClient.IsIdle(ctx)
	suite.NoError(err, "IsIdle should not return an error")
	suite.False(isIdle, "Cluster should not be idle")
}

func (suite *CrmTestSuite) TestResourceRefresh() {
	ctx := context.Background()
	commandOutput := `Waiting for 1 reply from the controller
... got reply (done)`

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "resource", "refresh").Return([]byte(commandOutput), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	err := crmClient.ResourceRefresh(ctx, "", "")
	suite.NoError(err)
}

func (suite *CrmTestSuite) TestResourceRefreshWithResource() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "resource", "refresh", "my-resource").Return([]byte("got reply (done)"), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	err := crmClient.ResourceRefresh(ctx, "my-resource", "")
	suite.NoError(err)
}

func (suite *CrmTestSuite) TestResourceRefreshWithResourceAndNode() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "resource", "refresh", "my-resource", "my-node").Return([]byte("got reply (done)"), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	err := crmClient.ResourceRefresh(ctx, "my-resource", "my-node")
	suite.NoError(err)
}

func (suite *CrmTestSuite) TestResourceRefreshWithNodeOnlyError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	err := crmClient.ResourceRefresh(ctx, "", "my-node")
	suite.Error(err)
	suite.EqualError(err, "nodeID cannot be provided without a resourceID")
}

func (suite *CrmTestSuite) TestResourceRefreshError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "resource", "refresh").Return([]byte("error output"), errors.New("some error"))

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	err := crmClient.ResourceRefresh(ctx, "", "")
	suite.Error(err)
	suite.Contains(err.Error(), "failed to refresh resource")
	suite.Contains(err.Error(), "some error")
	suite.Contains(err.Error(), "error output")
}

func (suite *CrmTestSuite) TestResourceRefreshUnexpectedOutputError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "crm", "resource", "refresh").Return([]byte("unexpected output"), nil)

	crmClient := cluster.NewClusterClient(mockExecutor, slog.Default())

	err := crmClient.ResourceRefresh(ctx, "", "")
	suite.Error(err)
	suite.Contains(err.Error(), "failed to refresh resource, unexpected output")
	suite.Contains(err.Error(), "unexpected output")
}
