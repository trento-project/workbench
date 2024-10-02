package operator

import (
	"context"

	"github.com/trento-project/workbench/internal/support"
)

type SaptuneApplyOption Option[SaptuneApply]

type SaptuneApply struct {
	base
	executor support.CmdExecutor
}

func WithCustomSaptuneExecutor(executor support.CmdExecutor) Option[SaptuneApply] {
	return func(o *SaptuneApply) {
		o.executor = executor
	}
}

func NewSaptuneApply(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[SaptuneApply],
) *SaptuneApply {
	saptuneApply := &SaptuneApply{
		base:     newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		executor: support.Executor{},
	}

	for _, opt := range options.OperatorOptions {
		opt(saptuneApply)
	}

	return saptuneApply
}

func (sa *SaptuneApply) Run(ctx context.Context) *ExecutionReport {
	err := sa.plan(ctx)
	if err != nil {
		return sa.reportError(err)
	}

	err = sa.commit(ctx)
	if err != nil {
		rollbackError := sa.rollback(ctx)
		if rollbackError != nil {
			return sa.reportError(sa.wrapRollbackError(err, rollbackError))
		}
		return sa.reportError(err)
	}

	err = sa.verify(ctx)
	if err != nil {
		rollbackError := sa.rollback(ctx)
		if rollbackError != nil {
			return sa.reportError(sa.wrapRollbackError(err, rollbackError))
		}
		return sa.reportError(err)
	}

	// compute diff
	//
	diff := make(map[string]string)

	return sa.reportSuccess(diff)
}

func (sa *SaptuneApply) plan(_ context.Context) error {
	sa.logger.Debug(
		"printing infos",
	)
	return nil
}
