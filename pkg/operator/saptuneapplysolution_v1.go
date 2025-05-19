package operator

import (
	"context"
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/trento-project/workbench/internal/support"
)

const (
	SaptuneApplySolutionOperatorName = "saptuneapplysolution"
)

type SaptuneApplySolutionOption Option[SaptuneApplySolution]

type saptuneApplySolutionArguments struct {
	solution string
}

// SaptuneApplySolution is an operator responsible for applying a saptune solution.
//
// The operator requires an argument in the form of a map containing a key named "solution".
// This value will be passed to the saptune command-line tool.
//
// All considerations related to applying a solution using the saptune CLI apply here as well.
//
// # Execution Phases
//
// - PLAN:
//   The operator checks for the presence of the saptune binary and verifies its version.
//   The minimum required version is 3.1.0. If saptune is not installed or the version does not meet the minimum requirement,
//   the operation will fail. The current state of the applied saptune solution is collected as the "before" diff.
//
// - COMMIT:
//   The operator checks if the requested solution is already applied. If it is, no action is taken,
//   ensuring idempotency without returning an error. If the solution is not applied, the saptune command
//   to apply the solution will be executed.
//
// - VERIFY:
//   The operator verifies whether the solution has been correctly applied to the system.
//   If not, an error is raised. If successful, the current state of the applied solution is collected as the "after" diff.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the saptune revert command is executed
//   to undo the applied solution.

type SaptuneApplySolution struct {
	baseOperator
	saptune         Saptune
	executor        support.CmdExecutor
	parsedArguments *saptuneApplySolutionArguments
}

func WithCustomSaptuneExecutor(executor support.CmdExecutor) SaptuneApplySolutionOption {
	return func(o *SaptuneApplySolution) {
		o.executor = executor
	}
}

func NewSaptuneApplySolution(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[SaptuneApplySolution],
) *Executor {
	saptuneApply := &SaptuneApplySolution{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		executor:     support.CliExecutor{},
	}

	for _, opt := range options.OperatorOptions {
		opt(saptuneApply)
	}

	saptuneApply.saptune = NewSaptuneClient(
		saptuneApply.executor,
		saptuneApply.logger,
	)

	return &Executor{
		phaser:      saptuneApply,
		operationID: operationID,
	}
}

func (sa *SaptuneApplySolution) plan(ctx context.Context) error {
	opArguments, err := parseSaptuneApplyArguments(sa.arguments)
	if err != nil {
		return err
	}
	sa.parsedArguments = opArguments

	err = sa.saptune.CheckVersionSupport(ctx)

	if err != nil {
		return err
	}

	solutionAppliedOutput, err := sa.executor.Exec(ctx, "saptune", "--format", "json", "solution", "applied")
	if err != nil {
		return errors.New("could not call saptune solution applied")
	}

	sa.resources[beforeDiffField] = string(solutionAppliedOutput)

	return nil
}

func (sa *SaptuneApplySolution) commit(ctx context.Context) error {
	// check if solution is already applied
	solutionAppliedOutput, err := sa.executor.Exec(ctx, "saptune", "--format", "json", "solution", "applied")
	if err != nil {
		return errors.New("could not call saptune solution applied")
	}

	if alreadyApplied := isSaptuneSolutionAlreadyApplied(solutionAppliedOutput, sa.parsedArguments.solution); alreadyApplied {
		sa.logger.Infof("solution %s already applied, skipping saptune apply", sa.parsedArguments.solution)
		return nil
	}

	return sa.saptune.ApplySolution(ctx, sa.parsedArguments.solution)
}

func (sa *SaptuneApplySolution) verify(ctx context.Context) error {
	solutionAppliedOutput, err := sa.executor.Exec(ctx, "saptune", "--format", "json", "solution", "applied")
	if err != nil {
		return errors.New("could not call saptune solution applied")
	}

	if alreadyApplied := isSaptuneSolutionAlreadyApplied(solutionAppliedOutput, sa.parsedArguments.solution); alreadyApplied {
		sa.resources[afterFieldDiff] = string(solutionAppliedOutput)
		return nil
	}

	return fmt.Errorf(
		"verify saptune apply failing, the solution %s was not applied in commit phase",
		sa.parsedArguments.solution,
	)
}

func (sa *SaptuneApplySolution) rollback(ctx context.Context) error {
	revertOutput, err := sa.executor.Exec(ctx, "saptune", "solution", "revert", sa.parsedArguments.solution)
	if err != nil {
		return fmt.Errorf("could not revert saptune solution %s during rollback, error: %s",
			sa.parsedArguments.solution,
			revertOutput,
		)
	}

	return nil
}

func (sa *SaptuneApplySolution) operationDiff(ctx context.Context) map[string]any {
	return sa.standardDiff(ctx)
}
func isSaptuneSolutionAlreadyApplied(saptuneOutput []byte, solution string) bool {
	return gjson.GetBytes(saptuneOutput, fmt.Sprintf(`result.Solution applied.#(Solution ID=="%s")`, solution)).Exists()
}

func parseSaptuneApplyArguments(rawArguments OperatorArguments) (*saptuneApplySolutionArguments, error) {
	argument, found := rawArguments["solution"]
	if !found {
		return nil, errors.New("argument solution not provided, could not use the operator")
	}

	solution, ok := argument.(string)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse solution argument as string, argument provided: %v",
			argument,
		)
	}

	return &saptuneApplySolutionArguments{
		solution: solution,
	}, nil
}
