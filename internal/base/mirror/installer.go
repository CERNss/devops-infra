package mirror

import (
	"context"

	osdriver "devops-infra/internal/os"
)

type Installer struct {
	os     osdriver.Driver
	enable bool
}

func New(os osdriver.Driver, enable bool) *Installer {
	return &Installer{os: os, enable: enable}
}

func (m *Installer) Name() string { return "system-mirror" }

func (m *Installer) IsInstalled(ctx context.Context) bool {
	return !m.enable
}

func (m *Installer) Install(ctx context.Context) error {
	if !m.enable {
		return nil
	}

	return m.os.Exec().Run(
		"bash <(curl -sSL https://linuxmirrors.cn/main.sh)",
	)
}
