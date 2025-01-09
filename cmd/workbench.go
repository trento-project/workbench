package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/jessevdk/go-flags"
	"github.com/trento-project/workbench/pkg/operator"
)

var Version string

type cliOptions struct {
	Arguments string `long:"arguments" short:"a" description:"Json arguments of an operator" required:"true"`
	Verbose   bool   `long:"verbose" short:"v" description:"Log verbosity"`
}

var options cliOptions
var flagParser = flags.NewParser(&options, flags.Default) //nolint

func main() {
	ctx := context.Background()

	args, err := flagParser.Parse()
	if err != nil {
		os.Exit(1)
	}

	logger := logrus.StandardLogger()
	if options.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.Infof(
		"starting workbench CLI, version: %s",
		Version,
	)

	operatorName := args[0]
	registry := operator.StandardRegistry(operator.WithCustomLogger(logger))

	builder, err := registry.GetOperatorBuilder(operatorName)
	if err != nil {
		logger.Fatalf("operator: %s not available, exiting", operatorName)
	}

	opArgs := make(operator.OperatorArguments)
	err = json.Unmarshal([]byte(options.Arguments), &opArgs)
	if err != nil {
		logger.Fatalf("could not unmarhsal %s into arguments", options.Arguments)
	}

	logger.Infof(
		"starting execution with operator: %s - arguments: %s",
		operatorName,
		options.Arguments,
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
