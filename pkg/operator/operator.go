package operator

import (
	"context"
)

type OPERATION_PHASES string

type OperatorArguments map[string]any
type Option[T any] func(*T)

const (
	PLAN     OPERATION_PHASES = "PLAN"
	COMMIT   OPERATION_PHASES = "COMMIT"
	VERIFY   OPERATION_PHASES = "VERIFY"
	ROLLBACK OPERATION_PHASES = "ROLLBACK"
)

type Operator interface {
	Run(ctx context.Context) *ExecutionReport
}

type OperatorOptions[T any] struct {
	BaseOperatorOptions []BaseOperatorOption
	OperatorOptions     []Option[T]
}
