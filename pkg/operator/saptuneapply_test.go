package operator_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/trento-project/workbench/pkg/operator"
)

func TestSaptuneApplyRunFailure(t *testing.T) {
	arguments := operator.OperatorArguments{
		"test": 1,
	}
	opID := "223"
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	sa := operator.NewSaptuneApply(arguments, opID, operator.OperatorOptions[operator.SaptuneApply]{
		BaseOperatorOptions: []operator.BaseOption{operator.WithLogger(logrus.StandardLogger())},
	})
	res := sa.Run(context.TODO())

	assert.NotNil(t, res.Error)
}
