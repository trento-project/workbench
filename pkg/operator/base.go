package operator

import (
	"context"
	"log/slog"

	"github.com/trento-project/workbench/internal/support"
)

const (
	beforeDiffField = "before"
	afterDiffField  = "after"
)

type BaseOperatorOption Option[baseOperator]

func WithCustomLogger(logger *slog.Logger) BaseOperatorOption {
	return func(b *baseOperator) {
		b.logger = logger
	}
}

type baseOperator struct {
	arguments      OperatorArguments
	resources      map[string]any
	logger         *slog.Logger
}

func newBaseOperator(
	operationID string,
	arguments OperatorArguments,
	options ...BaseOperatorOption,
) baseOperator {
	base := &baseOperator{
		arguments: arguments,
		resources: make(map[string]any),
		logger:    support.NewDefaultLogger(slog.LevelInfo),
	}

	for _, opt := range options {
		opt(base)
	}

	base.logger = base.logger.With("operation_id", operationID)

	return *base
}

func (b *baseOperator) after(_ context.Context) {}
