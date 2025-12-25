package docker

import (
	"context"
	"fmt"

	osdriver "devops-infra/internal/os"
)

type InstallMode string

const (
	InstallModeMirrorScript InstallMode = "mirror"
	InstallModeNerdctlSymlink InstallMode = "nerdctl"
)

type Installer struct {
	os   osdriver.Driver
	mode InstallMode
}

func New(os osdriver.Driver, mode InstallMode) *Installer {
	return &Installer{os: os, mode: mode}
}

func (d *Installer) Name() string { return "docker" }

func (d *Installer) IsInstalled(ctx context.Context) bool {
	switch d.mode {
	case InstallModeNerdctlSymlink:
		_, err := d.os.Exec().RunWithOutput("test -L /usr/bin/docker")
		return err == nil
	case InstallModeMirrorScript, "":
		_, err := d.os.Exec().RunWithOutput("docker version")
		return err == nil
	default:
		return false
	}
}

func (d *Installer) Install(ctx context.Context) error {
	exec := d.os.Exec()

	switch d.mode {
	case InstallModeNerdctlSymlink:
		if err := exec.Run("command -v nerdctl >/dev/null 2>&1"); err != nil {
			return err
		}
		return exec.Run("ln -sf /usr/bin/nerdctl /usr/bin/docker")
	case InstallModeMirrorScript, "":
		if err := exec.Run(
			"bash scripts/mirror/docker.sh",
		); err != nil {
			return err
		}

		err := exec.Run("systemctl enable docker")
		if err != nil {
			return err
		}
		err = exec.Run("systemctl restart docker")
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported docker install mode: %s", d.mode)
	}
}
