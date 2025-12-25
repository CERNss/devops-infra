package debian

import (
	"fmt"
	"strings"

	"devops-infra/internal/executor"
	"devops-infra/internal/utils/pathutil"
)

type Driver struct {
	exec executor.Executor
}

func New(exec executor.Executor) *Driver {
	return &Driver{exec: exec}
}

func (d *Driver) Exec() executor.Executor {
	return d.exec
}

func (d *Driver) Name() string   { return "debian-driver" }
func (d *Driver) Family() string { return "debian" }

func (d *Driver) Update() error {
	return d.exec.Run("apt-get update -y")
}

func (d *Driver) InstallPackages(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	cmd := fmt.Sprintf(
		"apt-get install -y %s",
		strings.Join(pkgs, " "),
	)
	return d.exec.Run(cmd)
}

func (d *Driver) EnableService(name string) error {
	return d.exec.Run("systemctl enable " + name)
}

func (d *Driver) StartService(name string) error {
	return d.exec.Run("systemctl start " + name)
}

func (d *Driver) RestartService(name string) error {
	return d.exec.Run("systemctl restart " + name)
}

func (d *Driver) LoadKernelModules(mods ...string) error {
	for _, m := range mods {
		if err := d.exec.Run("modprobe " + m); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) Sysctl(settings map[string]string) error {
	for k, v := range settings {
		if err := d.exec.Run(fmt.Sprintf("sysctl -w %s=%s", k, v)); err != nil {
			return err
		}
	}
	return d.exec.Run("sysctl --system")
}

func (d *Driver) SwitchMirror() error {
	scriptPath, err := pathutil.ResolvePath("scripts/mirror/main.sh")
	if err != nil {
		return err
	}
	return d.exec.Run(fmt.Sprintf("bash %q", scriptPath))
}
