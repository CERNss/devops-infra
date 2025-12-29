package cmd

import (
	"context"
	"time"

	"devops-infra/internal/infra/executor"

	"github.com/spf13/cobra"
)

var execOpts executor.Options
var rootTimeout time.Duration

type ctxKey string

const timeoutCancelKey ctxKey = "timeout-cancel"

var rootCmd = &cobra.Command{
	Use:   "devops-infra",
	Short: "Infrastructure base",
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
	rootCmd.PersistentFlags().DurationVar(
		&rootTimeout,
		"timeout",
		5*time.Minute,
		"global timeout for commands (0 to disable)",
	)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if rootTimeout <= 0 {
			return
		}
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, rootTimeout)
		cmd.SetContext(context.WithValue(ctx, timeoutCancelKey, cancel))
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		cancel, ok := cmd.Context().Value(timeoutCancelKey).(context.CancelFunc)
		if ok {
			cancel()
		}
	}
}
