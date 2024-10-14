package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/trento-project/workbench/pkg/operator"
)

var arguments string

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "execute an operator providing operator name and arguments",
	Long: `
		workbench execute <operator name> --arguments <json object>

		Example

		worbench execute saptunesolutionapply --arguments "{"solution": "HANA"}"
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logrus.StandardLogger()
		if verbose {
			logger.SetLevel(logrus.DebugLevel)
		}

		operatorName := args[0]

		opArgs := make(operator.OperatorArguments)
		err := json.Unmarshal([]byte(arguments), &opArgs)
		if err != nil {
			return fmt.Errorf("could not unmarhsal %s into arguments", arguments)
		}

		registry := operator.StandardRegistry(operator.WithCustomLogger(logger))

		builder, err := registry.GetOperatorBuilder(operatorName)
		if err != nil {
			return err
		}

		op := builder("test-cli", opArgs)

		report := op.Run(context.Background())
		if report.Error != nil {
			return fmt.Errorf("operation execution error, phase: %s, reason: %s",
				report.Error.ErrorPhase,
				report.Error.Message,
			)
		}

		logger.Infof("exeuction succeded in phase: %s, diff: before: %s, after: %s",
			report.Success.LastPhase,
			report.Success.Diff["before"],
			report.Success.Diff["after"],
		)
		return nil
	},
}
