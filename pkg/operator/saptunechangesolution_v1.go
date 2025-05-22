package operator

import "github.com/trento-project/workbench/internal/saptune"

type SaptuneChangeSolution struct {
	baseOperator
	saptune         saptune.Saptune
	parsedArguments *saptuneSolutionArguments
}
