package saptune_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/saptune"
	"github.com/trento-project/workbench/internal/support/mocks"
	"github.com/trento-project/workbench/test/helpers"
)

type SaptuneClientTestSuite struct {
	suite.Suite
	mockExecutor *mocks.MockCmdExecutor
	logger       *logrus.Entry
}

func TestSaptuneClient(t *testing.T) {
	suite.Run(t, new(SaptuneClientTestSuite))
}

func (suite *SaptuneClientTestSuite) SetupTest() {
	suite.mockExecutor = mocks.NewMockCmdExecutor(suite.T())
	suite.logger = logrus.NewEntry(logrus.StandardLogger())
}

func (suite *SaptuneClientTestSuite) TestVersionCheckFailureBecauseUnableToDetectVersion() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return(
		[]byte("package saptune is not installed"),
		errors.New("exit status 1"),
	)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.CheckVersionSupport(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "could not get the installed saptune version")
}

func (suite *SaptuneClientTestSuite) TestUnsupportedSaptuneVersionCheck() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.0.2"), nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.CheckVersionSupport(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "saptune version not supported")
}

func (suite *SaptuneClientTestSuite) TestSuccessfulSaptuneVersionCheck() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.CheckVersionSupport(ctx)

	suite.NoError(err)
}

func (suite *SaptuneClientTestSuite) TestGettingAppliedSolutionFailure() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	appliedSolution, err := saptuneClient.GetAppliedSolution(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "could not call saptune solution applied: error calling saptune")
	suite.Empty(appliedSolution)
}

func (suite *SaptuneClientTestSuite) TestGettingNoSolutionApplied() {
	ctx := context.Background()

	noSolutionApplied := helpers.ReadFixture("saptune/applied_no_solution.json")

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return(noSolutionApplied, nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	appliedSolution, err := saptuneClient.GetAppliedSolution(ctx)

	suite.NoError(err)
	suite.Empty(appliedSolution)
}

func (suite *SaptuneClientTestSuite) TestGettingAppliedSolution() {
	ctx := context.Background()

	hanaSolutionApplied := helpers.ReadFixture("saptune/applied_hana_solution.json")

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return(hanaSolutionApplied, nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	appliedSolution, err := saptuneClient.GetAppliedSolution(ctx)

	suite.NoError(err)
	suite.Equal("HANA", appliedSolution)
}

func (suite *SaptuneClientTestSuite) TestApplySolutionFailureBecauseCommandFails() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ApplySolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `could not perform saptune solution apply HANA`)
}

func (suite *SaptuneClientTestSuite) TestApplySolutionFailureBecauseAnAlreadyAppliedSolution() {
	ctx := context.Background()

	alreadyAppliedSolution := helpers.ReadFixture("saptune/apply_already_applied_solution.output")

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return(alreadyAppliedSolution, errors.New("exit status 1"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ApplySolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `could not perform saptune solution apply HANA`)
}

func (suite *SaptuneClientTestSuite) TestApplySolutionSuccess() {
	ctx := context.Background()

	applySolutionSuccess := helpers.ReadFixture("saptune/apply_solution_success.output")

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return(applySolutionSuccess, nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ApplySolution(ctx, "HANA")

	suite.NoError(err)
}

func (suite *SaptuneClientTestSuite) TestRevertSolutionFailureBecauseCommandFails() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"revert",
		"HANA",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.RevertSolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `could not perform saptune solution revert HANA, error: error calling saptune`)
}

func (suite *SaptuneClientTestSuite) TestRevertSolutionSuccess() {
	scenarios := []struct {
		name          string
		commandOutput []byte
	}{
		{
			name:          "reverting applied solution",
			commandOutput: helpers.ReadFixture("saptune/revert_solution_success.output"),
		},
		{
			name:          "reverting not applied solution",
			commandOutput: helpers.ReadFixture("saptune/revert_not_applied_solution.output"),
		},
	}

	for _, scenario := range scenarios {
		ctx := context.Background()

		suite.mockExecutor.On(
			"Exec",
			ctx,
			"saptune",
			"solution",
			"revert",
			"HANA",
		).Return(scenario.commandOutput, nil)

		saptuneClient := saptune.NewSaptuneClient(
			suite.mockExecutor,
			suite.logger,
		)
		err := saptuneClient.RevertSolution(ctx, "HANA")

		suite.NoError(err)
	}
}
