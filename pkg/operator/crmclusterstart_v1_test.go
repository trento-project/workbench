package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/crm/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

type CrmClusterStartOperatorTestSuite struct {
	suite.Suite
}

func TestCrmClusterStartOperator(t *testing.T) {
	suite.Run(t, new(CrmClusterStartOperatorTestSuite))
}

func (suite *CrmClusterStartOperatorTestSuite) SetupTest() {
	// Setup code for the test suite can be added here
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartArgMissingClusterId() {
	ctx := context.Background()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error parsing arguments: invalid or missing cluster_id argument", report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterIdMismatch() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("different-cluster-id", nil).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomCrmClient(mockCrmClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)

}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterErrorGetClusterId() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("", errors.New("any error")).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomCrmClient(mockCrmClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)

}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterAlreadyOnline() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("test-cluster-id", nil).Once()
	mockCrmClient.On("IsHostOnline", ctx).Return(true).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCrmClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"started":true}`,
		"after":  `{"started":true}`,
	}, report.Success.Diff)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterRollbackFailure() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("test-cluster-id", nil).Once()
	mockCrmClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCrmClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster")).Once()
	mockCrmClient.On("StopCluster", ctx).Return(errors.New("failed to stop cluster")).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomCrmClient(mockCrmClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterRollbackSuccess() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("test-cluster-id", nil).Once()
	mockCrmClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCrmClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster")).Once()
	mockCrmClient.On("StopCluster", ctx).Return(nil).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCrmClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterStartVerifyFailure() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("test-cluster-id", nil)
	mockCrmClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCrmClient.On("StartCluster", ctx).Return(nil).Once()
	mockCrmClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCrmClient.On("StopCluster", ctx).Return(nil).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomCrmClient(mockCrmClient)),
				operator.Option[operator.CrmClusterStart](operator.WithCustomRetry(2, 100*time.Millisecond, 1*time.Second, 2)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterStartVerifySuccess() {
	ctx := context.Background()

	mockCrmClient := mocks.NewMockCrm(suite.T())
	mockCrmClient.On("GetClusterId").Return("test-cluster-id", nil)
	mockCrmClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCrmClient.On("StartCluster", ctx).Return(nil).Once()
	mockCrmClient.On("IsHostOnline", ctx).Return(true).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.OperatorArguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.OperatorOptions[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomCrmClient(mockCrmClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"started":false}`,
		"after":  `{"started":true}`,
	}, report.Success.Diff)
}
