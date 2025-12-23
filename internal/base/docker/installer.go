package docker

import (
	"context"

	osdriver "devops-infra/internal/os"
)

type Installer struct {
	os osdriver.Driver
}

func New(os osdriver.Driver) *Installer {
	return &Installer{os: os}
}

func (d *Installer) Name() string { return "docker" }

func (d *Installer) IsInstalled(ctx context.Context) bool {
	_, err := d.os.Exec().RunWithOutput("docker version")
	return err == nil
}

func (d *Installer) Install(ctx context.Context) error {
	exec := d.os.Exec()

	if err := exec.Run(
		"bash <(curl -sSL https://linuxmirrors.cn/docker.sh)",
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
}
