package rhel

import (
	"fmt"
	"strings"

	"devops-infra/internal/executor"
)

type Driver struct {
	exec executor.Executor
}

func New(exec executor.Executor) *Driver {
	return &Driver{exec: exec}
}

func (r *Driver) Exec() executor.Executor {
	return r.exec
}

func (r *Driver) Name() string   { return "rhel-driver" }
func (r *Driver) Family() string { return "rhel" }

func (r *Driver) Update() error {
	return r.exec.Run("yum makecache")
}

func (r *Driver) InstallPackages(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	cmd := fmt.Sprintf(
		"yum install -y %s",
		strings.Join(pkgs, " "),
	)
	return r.exec.Run(cmd)
}

func (r *Driver) EnableService(name string) error {
	return r.exec.Run("systemctl enable " + name)
}

func (r *Driver) StartService(name string) error {
	return r.exec.Run("systemctl start " + name)
}

func (r *Driver) RestartService(name string) error {
	return r.exec.Run("systemctl restart " + name)
}

func (r *Driver) LoadKernelModules(mods ...string) error {
	for _, m := range mods {
		if err := r.exec.Run("modprobe " + m); err != nil {
			return err
		}
	}
	return nil
}

func (r *Driver) Sysctl(settings map[string]string) error {
	for k, v := range settings {
		if err := r.exec.Run(fmt.Sprintf("sysctl -w %s=%s", k, v)); err != nil {
			return err
		}
	}
	return r.exec.Run("sysctl --system")
}
