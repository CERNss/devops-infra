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
	ExecOpts          executor.Options
	EnableMirror      bool
	DockerInstallMode docker.InstallMode
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
		mode = docker.InstallModeMirrorScript
	}

	// 4. Build base installer
	components := []base.Component{
		kernel.New(driver),
		mirror.New(driver, opts.EnableMirror),
		tools.New(driver),
	}

	if mode == docker.InstallModeNerdctlSymlink {
		components = append(components, containerd.New(driver), docker.New(driver, mode))
	} else {
		components = append(components, docker.New(driver, mode), containerd.New(driver))
	}

	installer := base.New(components...)

	// 5. Run
	return installer.Install(ctx)
}
