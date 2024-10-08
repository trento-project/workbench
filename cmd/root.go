package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "workbench",
	Short: "Workbench allow run trento operator directly",
}

func Execute() {
	executeCmd.Flags().StringVarP(&arguments, "arguments", "a", "", "arguments as json object (required)")
	executeCmd.MarkFlagRequired("arguments")

	rootCmd.AddCommand(executeCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
