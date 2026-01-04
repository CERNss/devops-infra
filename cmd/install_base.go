package cmd

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/install/base/docker"
	"devops-infra/internal/infra/orchestration/flow"
	"devops-infra/internal/middleware/log"
	"devops-infra/internal/middleware/trace"
	"devops-infra/internal/utils/mirror"
)

var (
	enableMirror          bool
	linuxMirrorSource     string
	dockerInstallMode     string
	dockerMirrorSource    string
	dockerRegistryMirrors []string
	dockerEngineVersion   string
	containerdVersion     string
	containerdArch        string
	containerdChecksum    string
	cniSubnet             string
	cniRouteDst           string
	logDir                string
	enableLog             bool
	traceDir              string
	enableTrace           bool
	skipKernel            bool
	skipTools             bool
)

var installBaseCmd = &cobra.Command{
	Use:   "base",
	Short: "Install base infrastructure (kernel, tools, docker, containerd)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dockerInstallMode = strings.ToLower(strings.TrimSpace(dockerInstallMode))
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
		if dockerInstallMode == string(docker.InstallModeNerdctl) && dockerMirrorSource != "" {
			return fmt.Errorf("docker-source is not supported when docker-install-mode=nerdctl")
		}
		if dockerMirrorSource != "" {
			resolved, ok := mirror.ResolveDockerCE(dockerMirrorSource)
			if !ok {
				return fmt.Errorf("invalid docker source: %s", dockerMirrorSource)
			}
			dockerMirrorSource = resolved
		}

		cleanRegistryMirrors := make([]string, 0, len(dockerRegistryMirrors))
		seenRegistryMirrors := make(map[string]struct{})
		for _, registryMirror := range dockerRegistryMirrors {
			registryMirror = strings.TrimSpace(registryMirror)
			if registryMirror == "" {
				continue
			}
			if dockerInstallMode == string(docker.InstallModeNerdctl) {
				return fmt.Errorf("docker-registry-mirror is not supported when docker-install-mode=nerdctl")
			}
			resolved, ok := mirror.ResolveDockerRegistry(registryMirror)
			if !ok {
				return fmt.Errorf("invalid docker registry mirror: %s", registryMirror)
			}
			if _, ok := seenRegistryMirrors[resolved]; ok {
				continue
			}
			seenRegistryMirrors[resolved] = struct{}{}
			cleanRegistryMirrors = append(cleanRegistryMirrors, resolved)
		}

		dockerEngineVersion = strings.TrimSpace(dockerEngineVersion)
		if dockerInstallMode == string(docker.InstallModeNerdctl) && dockerEngineVersion != "" {
			return fmt.Errorf("docker-version is not supported when docker-install-mode=nerdctl")
		}

		containerdVersion = strings.TrimSpace(containerdVersion)
		containerdArch = strings.TrimSpace(containerdArch)
		containerdChecksum = strings.TrimSpace(containerdChecksum)
		cniSubnet = strings.TrimSpace(cniSubnet)
		cniRouteDst = strings.TrimSpace(cniRouteDst)
		logDir = strings.TrimSpace(logDir)
		traceDir = strings.TrimSpace(traceDir)
		if containerdChecksum != "" {
			if len(containerdChecksum) != 64 {
				return fmt.Errorf("invalid containerd checksum length: %d", len(containerdChecksum))
			}
			if _, err := hex.DecodeString(containerdChecksum); err != nil {
				return fmt.Errorf("invalid containerd checksum: %w", err)
			}
		}
		if cniSubnet == "" {
			cniSubnet = constant.DefaultNerdctlCNISubnet
		}
		if cniRouteDst == "" {
			cniRouteDst = constant.DefaultNerdctlCNIRouteDst
		}

		var traceSink trace.TraceSink
		if !enableTrace {
			traceSink = trace.NoopTraceSink()
		} else if traceDir != "" {
			tracePath := filepath.Join(traceDir, "trace.jsonl")
			sink, err := trace.NewFileTraceSink(tracePath)
			if err != nil {
				return err
			}
			traceSink = sink
		}

		var outputFactory log.OutputSinkFactory
		var logger log.Logger
		if !enableLog {
			outputFactory = log.NoopOutputSinkFactory{}
			logger = log.NoopLogger()
		}

		return flow.InstallBase(
			cmd.Context(),
			flow.InstallBaseOptions{
				ExecOpts:              execOpts,
				TraceSink:             traceSink,
				OutputFactory:         outputFactory,
				Logger:                logger,
				LogDir:                logDir,
				EnableMirror:          enableMirror,
				LinuxMirrorSource:     linuxMirrorSource,
				DockerInstallMode:     docker.InstallMode(dockerInstallMode),
				DockerMirrorSource:    dockerMirrorSource,
				DockerRegistryMirrors: cleanRegistryMirrors,
				DockerEngineVersion:   dockerEngineVersion,
				ContainerdVersion:     containerdVersion,
				ContainerdArch:        containerdArch,
				ContainerdChecksum:    containerdChecksum,
				CNISubnet:             cniSubnet,
				CNIRouteDst:           cniRouteDst,
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

	installBaseCmd.Flags().StringVar(
		&dockerEngineVersion,
		"docker-version",
		"",
		fmt.Sprintf("docker engine version (default: %s)", constant.DefaultDockerEngineVersion),
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
		fmt.Sprintf("containerd version (default: %s)", constant.DefaultContainerdVersion),
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

	installBaseCmd.Flags().StringVar(
		&cniSubnet,
		"cni-subnet",
		constant.DefaultNerdctlCNISubnet,
		"cni subnet for nerdctl (default: 10.88.0.0/16)",
	)

	installBaseCmd.Flags().StringVar(
		&cniRouteDst,
		"cni-route-dst",
		constant.DefaultNerdctlCNIRouteDst,
		"cni route destination for nerdctl (default: 0.0.0.0/0)",
	)

	installBaseCmd.Flags().StringVar(
		&logDir,
		"log-dir",
		constant.DefaultLogDir,
		"log directory (default: logs)",
	)

	installBaseCmd.Flags().BoolVar(
		&enableLog,
		"enable-log",
		true,
		"enable run log and command output logs",
	)

	installBaseCmd.Flags().StringVar(
		&traceDir,
		"trace-dir",
		"",
		"trace directory (default: trace/trace.jsonl)",
	)

	installBaseCmd.Flags().BoolVar(
		&enableTrace,
		"enable-trace",
		true,
		"enable trace events",
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
