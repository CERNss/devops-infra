package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"devops-infra/internal/base/docker"
	"devops-infra/internal/orchestration"
)

var (
	enableMirror      bool
	dockerInstallMode string
)

var installBaseCmd = &cobra.Command{
	Use:   "base",
	Short: "Install base infrastructure (kernel, tools, docker, containerd)",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch dockerInstallMode {
		case string(docker.InstallModeMirrorScript), string(docker.InstallModeNerdctlSymlink):
		default:
			return fmt.Errorf("invalid docker install mode: %s", dockerInstallMode)
		}

		return orchestration.InstallBase(
			context.Background(),
			orchestration.InstallBaseOptions{
				ExecOpts:          execOpts,
				EnableMirror:      enableMirror,
				DockerInstallMode: docker.InstallMode(dockerInstallMode),
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
		"switch system and docker mirror",
	)

	installBaseCmd.Flags().StringVar(
		&dockerInstallMode,
		"docker-install-mode",
		string(docker.InstallModeMirrorScript),
		"docker install mode: mirror|nerdctl",
	)
}
