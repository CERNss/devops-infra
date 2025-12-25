package orchestration

import (
	"context"
	"devops-infra/internal/infra/base"
	"devops-infra/internal/infra/base/containerd"
	"devops-infra/internal/infra/base/docker"
	"devops-infra/internal/infra/base/kernel"
	"devops-infra/internal/infra/base/mirror"
	"devops-infra/internal/infra/base/tools"
	executor2 "devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/os"
)

type InstallBaseOptions struct {
	ExecOpts              executor2.Options
	EnableMirror          bool
	LinuxMirrorSource     string
	DockerInstallMode     docker.InstallMode
	DockerMirrorSource    string
	DockerRegistryMirrors []string
	DockerEngineVersion   string
	ContainerdVersion     string
	ContainerdArch        string
	ContainerdChecksum    string
	SkipKernel            bool
	SkipTools             bool
}

func InstallBase(ctx context.Context, opts InstallBaseOptions) error {
	// 1. Detect OS
	osInfo, err := os.Detect()
	if err != nil {
		return err
	}

	// 2. Create executor (local)
	exec := executor2.NewLocal(opts.ExecOpts)

	// 3. Create OS driver
	driver, err := os.NewDriver(osInfo, exec)
	if err != nil {
		return err
	}

	mode := opts.DockerInstallMode
	if mode == "" {
		mode = docker.InstallModeOfficial
	}

	// 4. Build base installer
	components := []base.Component{}
	if !opts.SkipKernel {
		components = append(components, kernel.New(driver))
	}
	components = append(components, mirror.New(driver, mirror.Options{
		Enable: opts.EnableMirror,
		Source: opts.LinuxMirrorSource,
	}))
	if !opts.SkipTools {
		components = append(components, tools.New(driver))
	}

	containerdInstaller := containerd.New(driver, containerd.Options{
		Version:  opts.ContainerdVersion,
		Arch:     opts.ContainerdArch,
		Checksum: opts.ContainerdChecksum,
	})

	if mode == docker.InstallModeNerdctl {
		components = append(
			components,
			docker.New(driver, docker.Options{
				Mode:            mode,
				Source:          opts.DockerMirrorSource,
				RegistryMirrors: opts.DockerRegistryMirrors,
				EngineVersion:   opts.DockerEngineVersion,
			}),
			containerdInstaller,
		)
	} else {
		components = append(
			components,
			docker.New(driver, docker.Options{
				Mode:            mode,
				Source:          opts.DockerMirrorSource,
				RegistryMirrors: opts.DockerRegistryMirrors,
				EngineVersion:   opts.DockerEngineVersion,
			}),
			containerdInstaller,
		)
	}

	installer := base.New(components...)

	// 5. Run
	return installer.Install(ctx)
}
