package operator

import (
	"context"
	"fmt"
)

type OPERATION_PHASES string

const (
	PLAN     OPERATION_PHASES = "PLAN"
	COMMIT   OPERATION_PHASES = "COMMIT"
	VERIFY   OPERATION_PHASES = "VERIFY"
	ROLLBACK OPERATION_PHASES = "ROLLBACK"
)

type ExecutionError struct {
	ErrorPhase OPERATION_PHASES
	Message    string
}

func (e ExecutionError) Error() string {
	return fmt.Sprintf(
		"error during operator exeuction in phase: %s, reason: %s",
		e.ErrorPhase,
		e.Message,
	)
}

type ExecutionSuccess struct {
	Diff map[string]string
}

type ExecutionReport struct {
	OperationID string
	Success     *ExecutionSuccess
	Error       *ExecutionError
}

type OperatorArguments map[string]any

type Operator interface {
	Run(ctx context.Context) *ExecutionReport
}

type Option func(Operator)
