package operator_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/tidwall/gjson"
// 	"github.com/trento-project/workbench/internal/support/mocks"
// 	"github.com/trento-project/workbench/pkg/operator"
// 	"github.com/trento-project/workbench/test/helpers"
// )

// // const pippo = `{"$schema":"file:///usr/share/saptune/schemas/1.0/saptune_solution_applied.schema.json","publish time":"2025-05-19 07:44:29.127","argv":"saptune --format json solution applied","pid":18937,"command":"solution applied","exit code":0,"result":{"Solution applied":[{"Solution ID":"HANA","applied partially":false}]},"messages":[]}`

// func TestGetCurrentlyAppliedSaptuneSolution(t *testing.T) {
// 	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
// 	logger := logrus.NewEntry(logrus.StandardLogger())
// 	ctx := context.Background()

// 	hanaSolutionApplied := helpers.ReadFixture("saptune/applied_hana_solution.json")
// 	mockCmdExecutor.On(
// 		"Exec",
// 		ctx,
// 		"saptune",
// 		"--format",
// 		"json",
// 		"solution",
// 		"applied",
// 	).Return(hanaSolutionApplied, nil)

// 	solutionID, err := operator.GetCurrentlyAppliedSaptuneSolution(mockCmdExecutor, logger, ctx)

// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, solutionID)
// }

// func Test2(t *testing.T) {
// 	mockCmdExecutor := mocks.NewMockCmdExecutor(t)
// 	logger := logrus.NewEntry(logrus.StandardLogger())
// 	ctx := context.Background()

// 	hanaSolutionApplied := helpers.ReadFixture("saptune/no_solution_applied.json")
// 	manySolutionApplied := helpers.ReadFixture("saptune/many_solutions_applied.json")
// 	mockCmdExecutor.On(
// 		"Exec",
// 		ctx,
// 		"saptune",
// 		"--format",
// 		"json",
// 		"solution",
// 		"applied",
// 	).Return(hanaSolutionApplied, nil)

// 	ppp := gjson.GetBytes(hanaSolutionApplied, `result.Solution applied.#(Solution ID=="HANA")`).Exists()
// 	ppp2 := gjson.GetBytes(manySolutionApplied, `result.Solution applied.#(Solution ID=="HANA")`).Exists()

// 	assert.True(t, ppp)
// 	assert.True(t, ppp2)

// 	solutionID, err := operator.GetCurrentlyAppliedSaptuneSolution(mockCmdExecutor, logger, ctx)

// 	assert.NoError(t, err)
// 	assert.Empty(t, solutionID)
// }
