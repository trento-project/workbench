package operator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/cluster/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

type CrmClusterStopOperatorTestSuite struct {
	suite.Suite
}

func TestCrmClusterStopOperator(t *testing.T) {
	suite.Run(t, new(CrmClusterStopOperatorTestSuite))
}

func (suite *CrmClusterStopOperatorTestSuite) SetupTest() {
	// Setup code for the test suite can be added here
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterAlreadyOffline() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCluster(suite.T())
	mockCrmClient.On("IsHostOnline", ctx).Return(false).Once()

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCrmClient)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"stopped":true}`,
		"after":  `{"stopped":true}`,
	}, report.Success.Diff)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterRollbackFailure() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCluster(suite.T())
	mockCrmClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCrmClient.On("IsIdle", ctx).Return(false, nil)
	mockCrmClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster"))

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCrmClient)),
				operator.Option[operator.CrmClusterStop](operator.WithCustomRetryStop(2, 100*time.Millisecond, 1*time.Second, 1)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterRollbackSuccess() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCluster(suite.T())
	mockCrmClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCrmClient.On("IsIdle", ctx).Return(true, nil)
	mockCrmClient.On("StopCluster", ctx).Return(errors.New("failed to stop cluster"))
	mockCrmClient.On("StartCluster", ctx).Return(nil).Once()

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCrmClient)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterStartVerifyFailure() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCluster(suite.T())
	mockCrmClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCrmClient.On("IsIdle", ctx).Return(true, nil)
	mockCrmClient.On("StopCluster", ctx).Return(nil)
	mockCrmClient.On("IsHostOnline", ctx).Return(true)
	mockCrmClient.On("StartCluster", ctx).Return(nil).Once()

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCrmClient)),
				operator.Option[operator.CrmClusterStop](operator.WithCustomRetryStop(2, 100*time.Millisecond, 1*time.Second, 2)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopVerifySuccess() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCluster(suite.T())
	mockCrmClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCrmClient.On("IsIdle", ctx).Return(true, nil).Once()
	mockCrmClient.On("StopCluster", ctx).Return(nil).Once()
	mockCrmClient.On("IsHostOnline", ctx).Return(false)

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCrmClient)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"stopped":false}`,
		"after":  `{"stopped":true}`,
	}, report.Success.Diff)
}
