package operator

import (
	"context"
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/trento-project/workbench/internal/support"
	"golang.org/x/mod/semver"
)

const (
	minimalSaptuneVersion            = "v3.1.0"
	SaptuneApplySolutionOperatorName = "saptuneapplysolution"
)

type SaptuneApplySolutionOption Option[SaptuneApplySolution]

type saptuneApplySolutionArguments struct {
	solution string
}

type SaptuneApplySolution struct {
	baseOperation
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
		baseOperation: newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		executor:      support.CliExecutor{},
	}

	for _, opt := range options.OperatorOptions {
		opt(saptuneApply)
	}

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

	// check saptune version
	versionOutput, err := sa.executor.Exec(ctx, "rpm", "-q", "--qf", "%{VERSION}", "saptune")
	if err != nil {
		return fmt.Errorf(
			"could not get the installed saptune version: %w",
			err,
		)
	}
	sa.logger.Debugf("installed saptune version: %s", string(versionOutput))

	if supported := isSaptuneVersionSupported(string(versionOutput)); !supported {
		return fmt.Errorf(
			"saptune version not supported, installed: %s, minimum supported: %s",
			versionOutput,
			minimalSaptuneVersion,
		)
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

	applyOutput, err := sa.executor.Exec(ctx, "saptune", "solution", "apply", sa.parsedArguments.solution)
	if err != nil {
		sa.logger.Errorf(
			"could not perform saptune solution apply %s, error output: %s",
			sa.parsedArguments.solution,
			applyOutput,
		)

		return fmt.Errorf("could not perform the saptune apply solution %s, error: %s",
			sa.parsedArguments.solution,
			err,
		)
	}

	return nil
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

func isSaptuneVersionSupported(version string) bool {
	compareOutput := semver.Compare(minimalSaptuneVersion, "v"+version)

	return compareOutput != 1
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
