package operator

import (
	"context"
	"errors"
	"fmt"
)

type errUnimplementedPhase struct {
	phase OPERATION_PHASES
}

func (e errUnimplementedPhase) Error() string {
	return fmt.Sprintf("phase function: %s not implemented", e.phase)
}

type Base struct {
	operationID   string
	arguments     OperatorArguments
	currentPhase  OPERATION_PHASES
	planResources map[string]any
}

func (sa *Base) Plan(_ context.Context) error {
	return errUnimplementedPhase{phase: PLAN}
}

func (sa *Base) commit(_ context.Context) error {
	return errUnimplementedPhase{phase: COMMIT}
}

func (sa *Base) verify(_ context.Context) error {
	return errUnimplementedPhase{phase: VERIFY}
}

func (sa *Base) rollback(_ context.Context) error {
	return errUnimplementedPhase{phase: ROLLBACK}
}

func (sa *Base) wrapRollbackError(phaseError error, rollbackError error) error {
	return errors.Join(rollbackError, phaseError)
}

func (sa *Base) reportError(error error) *ExecutionReport {
	return &ExecutionReport{
		OperationID: sa.operationID,
		Error: &ExecutionError{
			Message:    error.Error(),
			ErrorPhase: sa.currentPhase,
		},
	}
}

func (sa *Base) reportSuccess(diff map[string]string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: sa.operationID,
		Success: &ExecutionSuccess{
			Diff: diff,
		},
	}
}
