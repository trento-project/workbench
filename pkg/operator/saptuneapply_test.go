package operator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trento-project/workbench/pkg/operator"
)

func TestSaptuneApplyRunFailure(t *testing.T) {
	arguments := operator.OperatorArguments{
		"test": 1,
	}
	opID := "223"
	sa := operator.NewSaptuneApply(arguments, opID)
	res := sa.Run(context.TODO())

	assert.NotNil(t, res.Error)
}
