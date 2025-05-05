package operator

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/trento-project/workbench/internal/support"
)

const (
	beforeDiffField = "before"
	afterFieldDiff  = "after"
)

type BaseOperatorOption Option[baseOperator]

func WithCustomLogger(logger *logrus.Logger) BaseOperatorOption {
	return func(b *baseOperator) {
		b.loggerInstance = logger
	}
}

func WithCustomExecutor(executor support.CmdExecutor) BaseOperatorOption {
	return func(b *baseOperator) {
		b.executor = executor
	}
}

type baseOperator struct {
	arguments      OperatorArguments
	resources      map[string]any
	executor       support.CmdExecutor
	loggerInstance *logrus.Logger
	logger         *logrus.Entry
}

func newBaseOperator(
	operationID string,
	arguments OperatorArguments,
	options ...BaseOperatorOption,
) baseOperator {
	base := &baseOperator{
		arguments:      arguments,
		resources:      make(map[string]any),
		executor:       support.CliExecutor{},
		loggerInstance: logrus.StandardLogger(),
	}

	for _, opt := range options {
		opt(base)
	}

	base.logger = base.loggerInstance.WithField("operation_id", operationID)

	return *base
}

func (b *baseOperator) standardDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)
	diff["before"] = b.resources[beforeDiffField]
	diff["after"] = b.resources[afterFieldDiff]

	return diff
}
