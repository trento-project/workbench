package operator

import (
	"context"
)

type OPERATION_PHASES string

type OperationsArguments map[string]any
type Option[T any] func(*T)

const (
	PLAN     OPERATION_PHASES = "PLAN"
	COMMIT   OPERATION_PHASES = "COMMIT"
	VERIFY   OPERATION_PHASES = "VERIFY"
	ROLLBACK OPERATION_PHASES = "ROLLBACK"
)

type OperatorArguments OperationsArguments

type Runner interface {
	Run(ctx context.Context) *ExecutionReport
}

type OperatorOptions[T any] struct {
	BaseOperatorOptions []BaseOperationOption
	OperatorOptions     []Option[T]
}
