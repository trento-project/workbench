package crm_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/crm"
	"github.com/trento-project/workbench/internal/support/mocks"
)

type CrmTestSuite struct {
	suite.Suite
}

func TestCrm(t *testing.T) {
	suite.Run(t, new(CrmTestSuite))
}

func (suite *CrmTestSuite) SetupTest() {
	// Setup code for the test suite can be added here
}

func (suite *CrmTestSuite) IsIdleTest() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "cs_clusterstate", "-i").Return([]byte("Cluster state: S_IDLE"), nil)

	crmClient := crm.NewCrmClient(mockExecutor, slog.Default())

	isIdle, err := crmClient.IsIdle(ctx)
	suite.NoError(err, "IsIdle should not return an error")
	suite.True(isIdle, "Cluster should be idle")
}

func (suite *CrmTestSuite) IsIdleErrorTest() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "cs_clusterstate", "-i").Return([]byte(""), errors.New("command failed"))

	crmClient := crm.NewCrmClient(mockExecutor, slog.Default())

	_, err := crmClient.IsIdle(ctx)

	suite.Error(err, "IsIdle should return an error")
}

func (suite *CrmTestSuite) IsIdleDifferentStateTest() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCmdExecutor(suite.T())
	mockExecutor.On("Exec", ctx, "cs_clusterstate", "-i").Return([]byte("Cluster state: S_TRANSITION_ENGINE"), nil)

	crmClient := crm.NewCrmClient(mockExecutor, slog.Default())

	isIdle, err := crmClient.IsIdle(ctx)
	suite.NoError(err, "IsIdle should not return an error")
	suite.False(isIdle, "Cluster should not be idle")
}
