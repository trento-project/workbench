package operator

import (
	"context"
	"errors"
	"fmt"

	"github.com/trento-project/workbench/internal/saptune"
	"github.com/trento-project/workbench/internal/support"
)

type saptuneSolutionArguments struct {
	solution string
}

func parseSaptuneSolutionArguments(rawArguments OperatorArguments) (*saptuneSolutionArguments, error) {
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

	return &saptuneSolutionArguments{
		solution: solution,
	}, nil
}

const SaptuneApplySolutionOperatorName = "saptuneapplysolution"

type SaptuneApplySolutionOption Option[SaptuneApplySolution]

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
//   the operation will fail.
//   The initially applied solution, if any, is collected as the "before" diff.
//
// - COMMIT:
//   The operator checks if the requested solution is already applied. If it is, no action is taken,
//   ensuring idempotency without returning an error.
//   If there is any other solution already applied, an error is raised, because only one solution can be applied at a time.
// 	 If otherwise there is no solution applied the saptune command to apply the solution will be executed.
//
// - VERIFY:
//   The operator verifies whether the solution has been correctly applied to the system.
//   If not, an error is raised. If successful, the current state of the applied solution is collected as the "after" diff.
//
// - ROLLBACK:
//   There is nothing to rollback in this operator.
//   If the COMMIT phase fails before attempting to apply the solution, it would not make sense to revert a not applied solution.
//   If the COMMIT phase fails when applying the solution, again it would not make sense to revert a solution that was not applied.
//   If the VERIFY phase fails when retrieving the applied solution, it might be risky to revert the solution that possibly should be applied.
//   If the VERIFY phase fails when because the solution requested to be applied is not the resulting applied one, it might still be risky to revert the solution that possibly should be applied.

type SaptuneApplySolution struct {
	baseOperator
	saptune         saptune.Saptune
	parsedArguments *saptuneSolutionArguments
}

func WithSaptuneClient(saptuneClient saptune.Saptune) SaptuneApplySolutionOption {
	return func(o *SaptuneApplySolution) {
		o.saptune = saptuneClient
	}
}

func NewSaptuneApplySolution(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[SaptuneApplySolution],
) *Executor {
	saptuneApply := &SaptuneApplySolution{
		baseOperator: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
	}

	saptuneApply.saptune = saptune.NewSaptuneClient(
		support.CliExecutor{},
		saptuneApply.logger,
	)

	for _, opt := range options.OperatorOptions {
		opt(saptuneApply)
	}

	return &Executor{
		phaser:      saptuneApply,
		operationID: operationID,
	}
}

func (sa *SaptuneApplySolution) plan(ctx context.Context) error {
	opArguments, err := parseSaptuneSolutionArguments(sa.arguments)
	if err != nil {
		return err
	}
	sa.parsedArguments = opArguments

	if err = sa.saptune.CheckVersionSupport(ctx); err != nil {
		return err
	}

	initiallyAppliedSolution, err := sa.saptune.GetAppliedSolution(ctx)
	if err != nil {
		return err
	}

	sa.resources[beforeDiffField] = initiallyAppliedSolution

	return nil
}

func (sa *SaptuneApplySolution) commit(ctx context.Context) error {
	initiallyAppliedSolution, ok := sa.resources[beforeDiffField].(string)
	if !ok {
		return fmt.Errorf("expected a string for initiallyAppliedSolution, but got %T", sa.resources[beforeDiffField])
	}

	if sa.parsedArguments.solution == initiallyAppliedSolution {
		sa.logger.Infof("solution %s is already applied, skipping commit phase", sa.parsedArguments.solution)
		return nil
	}

	if initiallyAppliedSolution != "" {
		return fmt.Errorf(
			"cannot apply solution %s because another solution %s is already applied",
			sa.parsedArguments.solution,
			initiallyAppliedSolution,
		)
	}

	return sa.saptune.ApplySolution(ctx, sa.parsedArguments.solution)
}

func (sa *SaptuneApplySolution) verify(ctx context.Context) error {
	appliedSolution, err := sa.saptune.GetAppliedSolution(ctx)
	if err != nil {
		return err
	}

	if appliedSolution != sa.parsedArguments.solution {
		return fmt.Errorf(
			"verify saptune apply failing, the solution %s was not applied in commit phase",
			sa.parsedArguments.solution,
		)
	}
	sa.resources[afterDiffField] = appliedSolution
	return nil
}

func (b *baseOperator) rollback(ctx context.Context) error {
	return nil
}
