package cmd

import (
	"devops-infra/internal/executor"

	"github.com/spf13/cobra"
)

var execOpts executor.Options

var rootCmd = &cobra.Command{
	Use:   "devops-infra",
	Short: "Infrastructure installer",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&execOpts.Sudo,
		"sudo",
		true,
		"run commands with sudo",
	)
	rootCmd.PersistentFlags().BoolVar(
		&execOpts.DryRun,
		"dry-run",
		false,
		"print commands without executing",
	)
	rootCmd.PersistentFlags().BoolVar(
		&execOpts.Verbose,
		"verbose",
		false,
		"verbose output",
	)
}
