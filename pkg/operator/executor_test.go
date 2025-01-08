package operator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutorHappyFlow(t *testing.T) {
	executionContext := context.Background()
	phaser := NewMockphaser(t)
	emptyDiff := make(map[string]any)

	phaserCall := phaser.On("plan", executionContext).
		Return(nil)

	commitCall := phaser.On("commit", executionContext).
		Return(nil).
		NotBefore(phaserCall)

	verifyCall := phaser.On("verify", executionContext).
		Return(nil).
		NotBefore(commitCall)

	phaser.On("operationDiff", executionContext).
		Return(emptyDiff).
		NotBefore(verifyCall)

	executor := Executor{
		phaser:      phaser,
		operationID: "operation-id",
	}

	report := executor.Run(executionContext)

	assert.Equal(t, "operation-id", report.OperationID)
	assert.Equal(t, VERIFY, report.Success.LastPhase)
	assert.Equal(t, emptyDiff, report.Success.Diff)
	assert.Nil(t, report.Error)
}

func TestExecutorPlanError(t *testing.T) {
	executionContext := context.Background()
	phaser := NewMockphaser(t)
	planError := errors.New("error during plan phase")

	phaser.On("plan", executionContext).
		Return(planError)

	executor := Executor{
		phaser:      phaser,
		operationID: "operation-id",
	}

	report := executor.Run(executionContext)

	assert.Equal(t, planError.Error(), report.Error.Message)
	assert.Equal(t, PLAN, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorCommitErrorWithSuccessfulRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := NewMockphaser(t)
	commitError := errors.New("error during error phase")

	phaserCall := phaser.On("plan", executionContext).
		Return(nil)

	commitCall := phaser.On("commit", executionContext).
		Return(commitError).
		NotBefore(phaserCall)

	phaser.On("rollback", executionContext).
		Return(nil).
		NotBefore(commitCall)

	executor := Executor{
		phaser:      phaser,
		operationID: "operation-id",
	}

	report := executor.Run(executionContext)

	assert.Equal(t, commitError.Error(), report.Error.Message)
	assert.Equal(t, COMMIT, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorCommitErrorWithFailedRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := NewMockphaser(t)
	commitError := errors.New("error during error phase")
	rollbackError := errors.New("error during rollback phase")

	phaserCall := phaser.On("plan", executionContext).
		Return(nil)

	commitCall := phaser.On("commit", executionContext).
		Return(commitError).
		NotBefore(phaserCall)

	phaser.On("rollback", executionContext).
		Return(rollbackError).
		NotBefore(commitCall)

	executor := Executor{
		phaser:      phaser,
		operationID: "operation-id",
	}

	report := executor.Run(executionContext)

	assert.Equal(t, errors.Join(rollbackError, commitError).Error(), report.Error.Message)
	assert.Equal(t, ROLLBACK, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorVerifyErrorWithSuccessfulRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := NewMockphaser(t)
	verifyError := errors.New("error during verify phase")

	phaserCall := phaser.On("plan", executionContext).
		Return(nil)

	commitCall := phaser.On("commit", executionContext).
		Return(nil).
		NotBefore(phaserCall)

	verifyCall := phaser.On("verify", executionContext).
		Return(verifyError).
		NotBefore(commitCall)

	phaser.On("rollback", executionContext).
		Return(nil).
		NotBefore(verifyCall)

	executor := Executor{
		phaser:      phaser,
		operationID: "operation-id",
	}

	report := executor.Run(executionContext)

	assert.Equal(t, verifyError.Error(), report.Error.Message)
	assert.Equal(t, VERIFY, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorVerifyErrorWithFailedRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := NewMockphaser(t)
	verifyError := errors.New("error during verify phase")
	rollbackError := errors.New("error during rollback phase")

	phaserCall := phaser.On("plan", executionContext).
		Return(nil)

	commitCall := phaser.On("commit", executionContext).
		Return(nil).
		NotBefore(phaserCall)

	verifyCall := phaser.On("verify", executionContext).
		Return(verifyError).
		NotBefore(commitCall)

	phaser.On("rollback", executionContext).
		Return(rollbackError).
		NotBefore(verifyCall)

	executor := Executor{
		phaser:      phaser,
		operationID: "operation-id",
	}

	report := executor.Run(executionContext)

	assert.Equal(t, errors.Join(rollbackError, verifyError).Error(), report.Error.Message)
	assert.Equal(t, ROLLBACK, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}
