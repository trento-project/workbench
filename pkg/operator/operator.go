package operator

import (
	"context"
)

type OPERATION_PHASES string

const (
	PLAN     OPERATION_PHASES = "PLAN"
	COMMIT   OPERATION_PHASES = "COMMIT"
	VERIFY   OPERATION_PHASES = "VERIFY"
	ROLLBACK OPERATION_PHASES = "ROLLBACK"
)

type OperatorArguments map[string]any

type Operator interface {
	Run(ctx context.Context) *ExecutionReport
}

type Option[T any] func(*T)

type OperatorOptions[T any] struct {
	BaseOperatorOptions []BaseOption
	OperatorOptions     []Option[T]
}
