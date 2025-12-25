package orchestration

import (
	"context"

	"devops-infra/internal/base"
	"devops-infra/internal/base/containerd"
	"devops-infra/internal/base/docker"
	"devops-infra/internal/base/kernel"
	"devops-infra/internal/base/mirror"
	"devops-infra/internal/base/tools"
	"devops-infra/internal/executor"
	osdriver "devops-infra/internal/os"
)

type InstallBaseOptions struct {
	ExecOpts              executor.Options
	EnableMirror          bool
	LinuxMirrorSource     string
	DockerInstallMode     docker.InstallMode
	DockerMirrorSource    string
	DockerRegistryMirrors []string
	ContainerdVersion     string
	ContainerdArch        string
	ContainerdChecksum    string
	SkipKernel            bool
	SkipTools             bool
}

func InstallBase(ctx context.Context, opts InstallBaseOptions) error {
	// 1. Detect OS
	osInfo, err := osdriver.Detect()
	if err != nil {
		return err
	}

	// 2. Create executor (local)
	exec := executor.NewLocal(opts.ExecOpts)

	// 3. Create OS driver
	driver, err := osdriver.NewDriver(osInfo, exec)
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
			}),
			containerdInstaller,
		)
	}

	installer := base.New(components...)

	// 5. Run
	return installer.Install(ctx)
}
