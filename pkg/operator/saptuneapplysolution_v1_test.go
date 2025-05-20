package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/saptune/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

type SaptuneApplySolutionOperatorTestSuite struct {
	suite.Suite
	mockSaptuneClient *mocks.MockSaptune
}

func TestSaptuneApplySolutionOperator(t *testing.T) {
	suite.Run(t, new(SaptuneApplySolutionOperatorTestSuite))
}

func (suite *SaptuneApplySolutionOperatorTestSuite) SetupTest() {
	suite.mockSaptuneClient = mocks.NewMockSaptune(suite.T())
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorBecauseFailingToParseArguments() {
	ctx := context.Background()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"foo": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument solution not provided, could not use the operator", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorBecauseFailingVersionCheck() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(errors.New("saptune version not supported")).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("saptune version not supported", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorBecauseFailingToDetermineInitiallyAppliedSolution() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine initially applied solution")).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("failed to determine initially applied solution", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionCommitFailingBecauseAnotherSolutionIsAlreadyApplied() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "S4HANA-DBSERVER",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("cannot apply solution S4HANA-DBSERVER because another solution HANA is already applied", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionCommitFailingBecauseApplyError() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		Once()

	suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to apply solution")).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("failed to apply solution", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionVerifyFailingBecauseUnableToDetermineAppliedSolution() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		Once()

	suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine applied solution during verify")).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("failed to determine applied solution during verify", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionVerifyFailingBecauseDetectedAppliedSolutionDiffersFromRequested() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		Once()

	suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("S4HANA-DBSERVER", nil).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify saptune apply failing, the solution HANA was not applied in commit phase", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionSuccess() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	solutionAppliedCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	solutionApplyCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(solutionAppliedCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(solutionApplyCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "",
		"after":  "HANA",
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionSuccessReapplyingAlreadyApplied() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		Times(2)

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClient(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "HANA",
		"after":  "HANA",
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
