package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"devops-infra/internal/orchestration"
)

var (
	enableMirror bool
)

var installBaseCmd = &cobra.Command{
	Use:   "base",
	Short: "Install base infrastructure (kernel, tools, docker, containerd)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return orchestration.InstallBase(
			context.Background(),
			orchestration.InstallBaseOptions{
				ExecOpts:     execOpts,
				EnableMirror: enableMirror,
			},
		)
	},
}

func init() {
	installCmd.AddCommand(installBaseCmd)

	installBaseCmd.Flags().BoolVar(
		&enableMirror,
		"mirror",
		false,
		"switch system and docker mirror (linuxmirrors.cn)",
	)
}
