package operator

import (
	"context"

	"github.com/sirupsen/logrus"
)

const (
	beforeDiffField = "before"
	afterDiffField  = "after"
)

type BaseOperatorOption Option[baseOperator]

func WithCustomLogger(logger *logrus.Logger) BaseOperatorOption {
	return func(b *baseOperator) {
		b.loggerInstance = logger
	}
}

type baseOperator struct {
	arguments      OperatorArguments
	resources      map[string]any
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
		loggerInstance: logrus.StandardLogger(),
	}

	for _, opt := range options {
		opt(base)
	}

	base.logger = base.loggerInstance.WithField("operation_id", operationID)

	return *base
}

func (b *baseOperator) after(_ context.Context) {}
