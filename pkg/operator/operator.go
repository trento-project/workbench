package operator

import (
	"context"
)

type OperatorPhases string

type OperatorArguments map[string]any
type Option[T any] func(*T)

const (
	PLAN     OperatorPhases = "PLAN"
	COMMIT   OperatorPhases = "COMMIT"
	VERIFY   OperatorPhases = "VERIFY"
	ROLLBACK OperatorPhases = "ROLLBACK"
)

type Operator interface {
	Run(ctx context.Context) *ExecutionReport
}

type OperatorOptions[T any] struct {
	BaseOperatorOptions []BaseOperatorOption
	OperatorOptions     []Option[T]
}
