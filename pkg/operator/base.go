package operator

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

type BaseOption Option[base]

type errUnimplementedPhase struct {
	phase OPERATION_PHASES
}

func (e errUnimplementedPhase) Error() string {
	return fmt.Sprintf("phase function: %s not implemented", e.phase)
}

func WithLogger(logger *logrus.Logger) BaseOption {
	return func(b *base) {
		b.logger = logger
	}
}

type base struct {
	operationID   string
	arguments     OperatorArguments
	currentPhase  OPERATION_PHASES
	planResources map[string]any
	logger        *logrus.Logger
}

func newBaseOperator(operationID string, arguments OperatorArguments, options ...BaseOption) base {
	base := &base{
		operationID:   operationID,
		arguments:     arguments,
		planResources: make(map[string]any),
		logger:        logrus.StandardLogger(),
	}

	for _, opt := range options {
		opt(base)
	}

	return *base
}

func (sa *base) plan(_ context.Context) error {
	return errUnimplementedPhase{phase: PLAN}
}

func (sa *base) commit(_ context.Context) error {
	return errUnimplementedPhase{phase: COMMIT}
}

func (sa *base) verify(_ context.Context) error {
	return errUnimplementedPhase{phase: VERIFY}
}

func (sa *base) rollback(_ context.Context) error {
	return errUnimplementedPhase{phase: ROLLBACK}
}

func (sa *base) wrapRollbackError(phaseError error, rollbackError error) error {
	return errors.Join(rollbackError, phaseError)
}

func (sa *base) reportError(error error) *ExecutionReport {
	return &ExecutionReport{
		OperationID: sa.operationID,
		Error: &ExecutionError{
			Message:    error.Error(),
			ErrorPhase: sa.currentPhase,
		},
	}
}

func (sa *base) reportSuccess(diff map[string]string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: sa.operationID,
		Success: &ExecutionSuccess{
			Diff: diff,
		},
	}
}
