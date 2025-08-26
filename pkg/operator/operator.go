package operator

import (
	"context"
)

type PhaseName string

type OperatorArguments map[string]any
type Option[T any] func(*T)

const (
	PLAN     PhaseName = "PLAN"
	COMMIT   PhaseName = "COMMIT"
	VERIFY   PhaseName = "VERIFY"
	ROLLBACK PhaseName = "ROLLBACK"
)

type Operator interface {
	Run(ctx context.Context) *ExecutionReport
}

type OperatorOptions[T any] struct {
	BaseOperatorOptions []BaseOperatorOption
	OperatorOptions     []Option[T]
}
