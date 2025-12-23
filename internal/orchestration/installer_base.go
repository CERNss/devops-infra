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
	ExecOpts     executor.Options
	EnableMirror bool
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

	// 4. Build base installer
	installer := base.New(
		kernel.New(driver),
		mirror.New(driver, opts.EnableMirror),
		tools.New(driver),
		docker.New(driver),
		containerd.New(driver),
	)

	// 5. Run
	return installer.Install(ctx)
}
