package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/jessevdk/go-flags"
	"github.com/trento-project/workbench/pkg/operator"
)

type opts struct {
	Arguments string `long:"arguments" short:"a" description:"Json arguments of an operator" required:"true"`
	Verbose   bool   `long:"verbose" short:"v" description:"Log verbosity"`
}

func main() {
	ctx := context.Background()

	var opts opts
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	logger := logrus.StandardLogger()
	if opts.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	operatorName := os.Args[1]

	registry := operator.StandardRegistry(operator.WithCustomLogger(logger))

	builder, err := registry.GetOperatorBuilder(operatorName)
	if err != nil {
		logger.Fatalf("operator: %s not available, exiting", operatorName)
	}

	opArgs := make(operator.OperatorArguments)
	err = json.Unmarshal([]byte(opts.Arguments), &opArgs)
	if err != nil {
		logger.Fatalf("could not unmarhsal %s into arguments", opts.Arguments)
	}

	logger.Infof(
		"starting execution with operator: %s - arguments: %s",
		operatorName,
		opts.Arguments,
	)

	op := builder("test-cli", opArgs)

	report := op.Run(ctx)
	if report.Error != nil {
		logger.Fatalf(
			"operation execution error, phase: %s, reason: %s",
			report.Error.ErrorPhase,
			report.Error.Message,
		)
	}

	logger.Infof(
		"execution succeded in phase: %s, diff: before: %s, after: %s",
		report.Success.LastPhase,
		report.Success.Diff["before"],
		report.Success.Diff["after"],
	)
}
