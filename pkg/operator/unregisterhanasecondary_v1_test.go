package operator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/support/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

type UnregisterHANASecondaryOperatorTestSuite struct {
	suite.Suite
	mockExecutor *mocks.MockCmdExecutor
}

func TestUnregisterHANASecondaryOperator(t *testing.T) {
	suite.Run(t, new(UnregisterHANASecondaryOperatorTestSuite))
}

func (suite *UnregisterHANASecondaryOperatorTestSuite) SetupTest() {
	suite.mockExecutor = mocks.NewMockCmdExecutor(suite.T())
}

func (suite *UnregisterHANASecondaryOperatorTestSuite) TestUnregisterHANASecondaryPlanErrorMissingSid() {
	ctx := context.Background()

	unregisterHANASecondaryOperator := operator.NewUnregisterHANASecondary(
		operator.OperatorArguments{
			"foo": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.UnregisterHANASecondary]{},
	)

	report := unregisterHANASecondaryOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument sid not provided, could not use the operator", report.Error.Message)
}

func (suite *UnregisterHANASecondaryOperatorTestSuite) TestUnregisterHANASecondaryPlanErrorNonStringSid() {
	ctx := context.Background()

	unregisterHANASecondaryOperator := operator.NewUnregisterHANASecondary(
		operator.OperatorArguments{
			"sid": 42,
		},
		"test-op",
		operator.OperatorOptions[operator.UnregisterHANASecondary]{},
	)

	report := unregisterHANASecondaryOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse sid argument as string, argument provided: 42", report.Error.Message)
}

func (suite *UnregisterHANASecondaryOperatorTestSuite) TestUnregisterHANASecondaryPlanErrorEmptySid() {
	ctx := context.Background()

	unregisterHANASecondaryOperator := operator.NewUnregisterHANASecondary(
		operator.OperatorArguments{
			"sid": "",
		},
		"test-op",
		operator.OperatorOptions[operator.UnregisterHANASecondary]{},
	)

	report := unregisterHANASecondaryOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("sid argument is empty", report.Error.Message)
}
