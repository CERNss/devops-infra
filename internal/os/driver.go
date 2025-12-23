package os

import "devops-infra/internal/executor"

type Driver interface {
	Name() string
	Family() string

	Exec() executor.Executor

	Update() error
	InstallPackages(pkgs ...string) error

	EnableService(name string) error
	StartService(name string) error
	RestartService(name string) error

	LoadKernelModules(mods ...string) error
	Sysctl(settings map[string]string) error

	SwitchMirror() error
}
