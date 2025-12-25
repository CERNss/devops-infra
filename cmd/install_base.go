package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"devops-infra/internal/base/docker"
	"devops-infra/internal/orchestration"
	"devops-infra/internal/utils/mirror"
)

var (
	enableMirror          bool
	linuxMirrorSource     string
	dockerInstallMode     string
	dockerMirrorSource    string
	dockerRegistryMirrors []string
	containerdVersion     string
	containerdArch        string
	containerdChecksum    string
	skipKernel            bool
	skipTools             bool
)

var installBaseCmd = &cobra.Command{
	Use:   "base",
	Short: "Install base infrastructure (kernel, tools, docker, containerd)",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch dockerInstallMode {
		case string(docker.InstallModeOfficial), string(docker.InstallModeNerdctl):
		default:
			return fmt.Errorf("invalid docker install mode: %s", dockerInstallMode)
		}

		linuxMirrorSource = strings.TrimSpace(linuxMirrorSource)
		if linuxMirrorSource != "" {
			resolved, ok := mirror.ResolveSystem(linuxMirrorSource)
			if !ok {
				return fmt.Errorf("invalid mirror source: %s", linuxMirrorSource)
			}
			linuxMirrorSource = resolved
			enableMirror = true
		}

		dockerMirrorSource = strings.TrimSpace(dockerMirrorSource)
		if dockerMirrorSource != "" {
			resolved, ok := mirror.ResolveDockerCE(dockerMirrorSource)
			if !ok {
				return fmt.Errorf("invalid docker source: %s", dockerMirrorSource)
			}
			dockerMirrorSource = resolved
		}

		cleanRegistryMirrors := make([]string, 0, len(dockerRegistryMirrors))
		seenRegistryMirrors := make(map[string]struct{})
		for _, mirror := range dockerRegistryMirrors {
			mirror = strings.TrimSpace(mirror)
			if mirror == "" {
				continue
			}
			resolved, ok := mirror.ResolveDockerRegistry(mirror)
			if !ok {
				return fmt.Errorf("invalid docker registry mirror: %s", mirror)
			}
			if _, ok := seenRegistryMirrors[resolved]; ok {
				continue
			}
			seenRegistryMirrors[resolved] = struct{}{}
			cleanRegistryMirrors = append(cleanRegistryMirrors, resolved)
		}

		containerdVersion = strings.TrimSpace(containerdVersion)
		containerdArch = strings.TrimSpace(containerdArch)
		containerdChecksum = strings.TrimSpace(containerdChecksum)

		return orchestration.InstallBase(
			context.Background(),
			orchestration.InstallBaseOptions{
				ExecOpts:              execOpts,
				EnableMirror:          enableMirror,
				LinuxMirrorSource:     linuxMirrorSource,
				DockerInstallMode:     docker.InstallMode(dockerInstallMode),
				DockerMirrorSource:    dockerMirrorSource,
				DockerRegistryMirrors: cleanRegistryMirrors,
				ContainerdVersion:     containerdVersion,
				ContainerdArch:        containerdArch,
				ContainerdChecksum:    containerdChecksum,
				SkipKernel:            skipKernel,
				SkipTools:             skipTools,
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
		"switch system mirror",
	)

	installBaseCmd.Flags().StringVar(
		&linuxMirrorSource,
		"mirror-source",
		"",
		"system mirror source (domain or alias)",
	)

	installBaseCmd.Flags().StringVar(
		&dockerInstallMode,
		"docker-install-mode",
		string(docker.InstallModeOfficial),
		"docker install mode: docker|nerdctl",
	)

	installBaseCmd.Flags().StringVar(
		&dockerMirrorSource,
		"docker-source",
		"",
		"docker CE mirror source (domain or alias)",
	)

	installBaseCmd.Flags().StringSliceVar(
		&dockerRegistryMirrors,
		"docker-registry-mirror",
		nil,
		"docker registry mirror (comma-separated, domain or alias)",
	)

	installBaseCmd.Flags().StringVar(
		&containerdVersion,
		"containerd-version",
		"",
		"containerd version (default: 1.7.28)",
	)

	installBaseCmd.Flags().StringVar(
		&containerdArch,
		"containerd-arch",
		"",
		"containerd arch (default: amd64)",
	)

	installBaseCmd.Flags().StringVar(
		&containerdChecksum,
		"containerd-checksum",
		"",
		"containerd tarball sha256 checksum",
	)

	installBaseCmd.Flags().BoolVar(
		&skipKernel,
		"skip-kernel",
		false,
		"skip kernel/sysctl configuration",
	)

	installBaseCmd.Flags().BoolVar(
		&skipTools,
		"skip-tools",
		false,
		"skip base tools installation",
	)
}
