package operator

import (
	"context"

	"github.com/sirupsen/logrus"
)

const (
	beforeDiffField = "before"
	afterFieldDiff  = "after"
)

type BaseOperationOption Option[baseOperation]

func WithCustomLogger(logger *logrus.Logger) BaseOperationOption {
	return func(b *baseOperation) {
		b.loggerInstance = logger
	}
}

type baseOperation struct {
	arguments      OperatorArguments
	resources      map[string]any
	loggerInstance *logrus.Logger
	logger         *logrus.Entry
}

func newBaseOperator(
	operationID string,
	arguments OperatorArguments,
	options ...BaseOperationOption,
) baseOperation {
	base := &baseOperation{
		arguments:      arguments,
		resources:      make(map[string]any),
		loggerInstance: logrus.StandardLogger(),
	}

	for _, opt := range options {
		opt(base)
	}

	base.logger = base.loggerInstance.WithField("operation_id", operationID)

	return *base
}

func (b *baseOperation) standardDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)
	diff["before"] = b.resources[beforeDiffField]
	diff["after"] = b.resources[afterFieldDiff]

	return diff
}
