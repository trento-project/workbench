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

type baseOperation struct {
	arguments OperatorArguments
	resources map[string]any
	logger    *logrus.Entry
}

func newBaseOperator(
	operationID string,
	arguments OperatorArguments,
	options ...BaseOperationOption,
) baseOperation {
	base := &baseOperation{
		arguments: arguments,
		resources: make(map[string]any),
		logger:    logrus.StandardLogger().WithField("operation_id", operationID),
	}

	for _, opt := range options {
		opt(base)
	}

	return *base
}

func (b *baseOperation) standardDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)
	diff["before"] = b.resources[beforeDiffField]
	diff["after"] = b.resources[afterFieldDiff]

	return nil
}
