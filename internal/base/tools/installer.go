package tools

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

func (t *Installer) Name() string { return "common-tools" }

func (t *Installer) IsInstalled(ctx context.Context) bool {
	return false
}

func (t *Installer) Install(ctx context.Context) error {
	err := t.os.Update()
	if err != nil {
		return err
	}
	return t.os.InstallPackages(
		"curl",
		"ca-certificates",
		"gnupg",
		"tar",
		"iproute",
	)
}
