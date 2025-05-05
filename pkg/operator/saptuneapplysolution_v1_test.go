package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trento-project/workbench/internal/support/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

const saptuneSolutionAppliedNoSolutionOutput = `
{"$schema":"file:///usr/share/saptune/schemas/1.0/saptune_solution_applied.schema.json","publish time":"2025-01-09 14:50:06.131","argv":"saptune --format json solution applied","pid":303,"command":"solution applied","exit code":0,"result":{"Solution applied":[]},"messages":[]}
`
const saptuneSolutionAppliedHanaSolutionOutput = `
{"$schema":"file:///usr/share/saptune/schemas/1.0/saptune_solution_applied.schema.json","publish time":"2025-01-09 14:52:39.641","argv":"saptune --format json solution applied","pid":826,"command":"solution applied","exit code":0,"result":{"Solution applied":[{"Solution ID":"HANA","applied partially":false}]},"messages":[]}
`

func TestSaptuneApplySolutionSuccess(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	checkSaptuneVersionCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	solutionAppliedCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil).
		NotBefore(checkSaptuneVersionCall).
		Twice() // it's called two times

	solutionApplyCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil).
		NotBefore(solutionAppliedCall)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedHanaSolutionOutput), nil).
		NotBefore(solutionApplyCall).
		Once() // We just need this output once in verify phase

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": saptuneSolutionAppliedNoSolutionOutput,
		"after":  saptuneSolutionAppliedHanaSolutionOutput,
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestSaptuneApplySolutionSuccessSolutionAlreadyApplied(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	checkSaptuneVersionCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedHanaSolutionOutput), nil).
		NotBefore(checkSaptuneVersionCall).
		Times(3) // it's called two times

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": saptuneSolutionAppliedHanaSolutionOutput,
		"after":  saptuneSolutionAppliedHanaSolutionOutput,
	}

	assert.Nil(t, report.Error)
	assert.Equal(t, report.Success.LastPhase, operator.VERIFY)
	assert.EqualValues(t, report.Success.Diff, expectedDiff)
}

func TestSaptuneApplySolutionPlanError(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("2.1.0"), nil)

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.PLAN)
	assert.EqualValues(t, report.Error.Message, "saptune version not supported, installed: 2.1.0, minimum supported: v3.1.0")
}

func TestSaptuneApplySolutionCommitErrorSuccessfulRollback(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	checkSaptuneVersionCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	solutionAppliedCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil).
		NotBefore(checkSaptuneVersionCall).
		Twice() // it's called two times

	solutionApplyCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return([]byte("error"), errors.New("error during solutionApply")).
		NotBefore(solutionAppliedCall)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"revert",
		"HANA",
	).Return([]byte(""), nil).
		NotBefore(solutionApplyCall)

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.COMMIT)
	assert.EqualValues(t, report.Error.Message, "could not perform the saptune apply solution HANA, error: error during solutionApply")
}

func TestSaptuneApplySolutionCommitErrorFailedRollback(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	checkSaptuneVersionCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	solutionAppliedCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil).
		NotBefore(checkSaptuneVersionCall).
		Twice() // it's called two times

	solutionApplyCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return([]byte("error"), errors.New("error during solutionApply")).
		NotBefore(solutionAppliedCall)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"revert",
		"HANA",
	).Return([]byte("rollback error"), errors.New("error")).
		NotBefore(solutionApplyCall)

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.ROLLBACK)
	assert.EqualValues(t, "could not revert saptune solution HANA during rollback, error: rollback error\ncould not perform the saptune apply solution HANA, error: error during solutionApply", report.Error.Message)
}

func TestSaptuneApplySolutionVerifyErrorSuccessfulRollback(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	checkSaptuneVersionCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil).
		NotBefore(checkSaptuneVersionCall).
		Times(3)

	solutionApplyCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"revert",
		"HANA",
	).Return([]byte(""), nil).
		NotBefore(solutionApplyCall)

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.VERIFY)
	assert.EqualValues(t, "verify saptune apply failing, the solution HANA was not applied in commit phase", report.Error.Message)
}

func TestSaptuneApplySolutionVerifyErrorFailedRollback(t *testing.T) {
	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
	ctx := context.Background()

	checkSaptuneVersionCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil).
		NotBefore(checkSaptuneVersionCall).
		Times(3)

	solutionApplyCall := mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return([]byte(saptuneSolutionAppliedNoSolutionOutput), nil)

	mockCmdExecutor.On(
		"Exec",
		ctx,
		"saptune",
		"solution",
		"revert",
		"HANA",
	).Return([]byte(""), errors.New("error during revert")).
		NotBefore(solutionApplyCall)

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.OperatorArguments{
			"solution": "HANA",
		},
		"test-op",
		operator.OperatorOptions[operator.SaptuneApplySolution]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomExecutor(mockCmdExecutor),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	assert.Nil(t, report.Success)
	assert.Equal(t, report.Error.ErrorPhase, operator.ROLLBACK)
	assert.EqualValues(t, "could not revert saptune solution HANA during rollback, error: \nverify saptune apply failing, the solution HANA was not applied in commit phase", report.Error.Message)
}
