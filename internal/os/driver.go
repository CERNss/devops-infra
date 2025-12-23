package os

type Driver interface {
	Name() string
	Family() string

	// 包管理
	Update() error
	InstallPackages(pkgs ...string) error

	// systemd
	EnableService(name string) error
	StartService(name string) error
	RestartService(name string) error

	// kernel / sysctl
	LoadKernelModules(mods ...string) error
	Sysctl(settings map[string]string) error
}
