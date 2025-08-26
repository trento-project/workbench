package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/trento-project/workbench/internal/support"
	"github.com/trento-project/workbench/pkg/operator"
)

type cliOptions struct {
	Arguments string `long:"arguments" short:"a" description:"Json arguments of an operator" required:"true"`
	Verbose   bool   `long:"verbose" short:"v" description:"Log verbosity"`
}

func main() {
	var version string
	var options cliOptions
	var flagParser = flags.NewParser(&options, flags.Default)

	ctx := context.Background()

	args, err := flagParser.Parse()
	if err != nil {
		os.Exit(1)
	}

	var logLevel slog.Level
	if options.Verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := support.NewDefaultLogger(logLevel)

	logger.Info("starting workbench CLI", "version", version)

	operatorName := args[0]
	registry := operator.StandardRegistry(operator.WithCustomLogger(logger))

	builder, err := registry.GetOperatorBuilder(operatorName)
	if err != nil {
		logger.Error("operator not available, exiting", "operator", operatorName)
		os.Exit(1)
	}

	opArgs := make(operator.Arguments)
	err = json.Unmarshal([]byte(options.Arguments), &opArgs)
	if err != nil {
		logger.Error("could not unmarshal options arguments", "arguments", options.Arguments)
		os.Exit(1)
	}

	logger.Info(
		"starting execution with operator",
		"operator", operatorName,
		"arguments", options.Arguments,
	)

	op := builder("test-cli", opArgs)

	report := op.Run(ctx)
	if report.Error != nil {
		logger.Error(
			"operation execution error",
			"phase", report.Error.ErrorPhase,
			"reason", report.Error.Message,
		)
		os.Exit(1)
	}

	logger.Info(
		"execution succeeded",
		"phase", report.Success.LastPhase,
		"diff_before", report.Success.Diff["before"],
		"diff_after", report.Success.Diff["after"],
	)
}
