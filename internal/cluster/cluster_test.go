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
