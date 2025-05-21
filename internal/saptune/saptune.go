package saptune

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"golang.org/x/mod/semver"

	"github.com/trento-project/workbench/internal/support"
)

const minimalSaptuneVersion = "v3.1.0"

type Saptune interface {
	CheckVersionSupport(ctx context.Context) error
	ApplySolution(ctx context.Context, solution string) error
	GetAppliedSolution(ctx context.Context) (string, error)
}

type saptuneClient struct {
	executor support.CmdExecutor
	logger   *logrus.Entry
}

func NewSaptuneClient(
	executor support.CmdExecutor,
	logger *logrus.Entry,
) Saptune {
	return &saptuneClient{
		executor: executor,
		logger:   logger,
	}
}

func (saptune *saptuneClient) CheckVersionSupport(ctx context.Context) error {
	versionOutput, err := saptune.executor.Exec(ctx, "rpm", "-q", "--qf", "%{VERSION}", "saptune")
	if err != nil {
		return fmt.Errorf(
			"could not get the installed saptune version: %w",
			err,
		)
	}

	detectedVersion := string(versionOutput)

	if supported := isSaptuneVersionSupported(detectedVersion); !supported {
		return fmt.Errorf(
			"saptune version not supported, installed: %s, minimum supported: %s",
			detectedVersion,
			minimalSaptuneVersion,
		)
	}

	saptune.logger.Debugf("installed saptune version: %s", detectedVersion)

	return nil
}

func (saptune *saptuneClient) GetAppliedSolution(ctx context.Context) (string, error) {
	solutionAppliedOutput, err := saptune.executor.Exec(ctx, "saptune", "--format", "json", "solution", "applied")
	if err != nil {
		return "", fmt.Errorf("could not call saptune solution applied: %w", err)
	}
	if err != nil {
		return "", err
	}
	return gjson.GetBytes(solutionAppliedOutput, "result.Solution applied.0.Solution ID").String(), nil
}

func (saptune *saptuneClient) ApplySolution(ctx context.Context, solution string) error {
	applyOutput, err := saptune.executor.Exec(ctx, "saptune", "solution", "apply", solution)
	if err != nil {
		saptune.logger.Errorf(
			"could not perform saptune solution apply %s, error output: %s",
			solution,
			applyOutput,
		)

		return fmt.Errorf("could not perform saptune apply solution %s, error: %s",
			solution,
			err,
		)
	}

	return nil
}

func isSaptuneVersionSupported(version string) bool {
	compareOutput := semver.Compare(minimalSaptuneVersion, "v"+strings.TrimSpace(version))

	return compareOutput != 1
}
