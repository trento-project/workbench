package operator

import (
	"context"
	"errors"
)

type phaser interface {
	plan(ctx context.Context) (bool, error)
	commit(ctx context.Context) error
	rollback(ctx context.Context) error
	verify(ctx context.Context) error
	operationDiff(ctx context.Context) map[string]any
}

type Executor struct {
	currentPhase OPERATION_PHASES
	phaser       phaser
	operationID  string
}

func (e *Executor) Run(ctx context.Context) *ExecutionReport {
	e.currentPhase = PLAN
	alreadyApplied, err := e.phaser.plan(ctx)
	if err != nil {
		return executionReportWithError(err, e.currentPhase, e.operationID)
	}

	if alreadyApplied {
		diff := e.phaser.operationDiff(ctx)
		return executionReportWithSuccess(diff, e.currentPhase, e.operationID)
	}

	e.currentPhase = COMMIT

	err = e.phaser.commit(ctx)
	if err != nil {
		return e.handleRollback(ctx, err)
	}

	e.currentPhase = VERIFY
	err = e.phaser.verify(ctx)
	if err != nil {
		return e.handleRollback(ctx, err)
	}

	diff := e.phaser.operationDiff(ctx)

	return executionReportWithSuccess(diff, e.currentPhase, e.operationID)
}

func (e *Executor) handleRollback(ctx context.Context, err error) *ExecutionReport {
	rollbackError := e.phaser.rollback(ctx)
	if rollbackError != nil {
		e.currentPhase = ROLLBACK
		return executionReportWithError(
			wrapRollbackError(err, rollbackError),
			e.currentPhase,
			e.operationID,
		)
	}
	return executionReportWithError(err, e.currentPhase, e.operationID)
}

func wrapRollbackError(phaseError error, rollbackError error) error {
	return errors.Join(rollbackError, phaseError)
}
