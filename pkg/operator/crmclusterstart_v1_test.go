package operator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CrmClusterStartOperatorTestSuite struct {
	suite.Suite
}

func TestCrmClusterStartOperator(t *testing.T) {
	suite.Run(t, new(CrmClusterStartOperatorTestSuite))
}

func (suite *CrmClusterStartOperatorTestSuite) SetupTest() {
	// Setup code for the test suite can be added here
}
