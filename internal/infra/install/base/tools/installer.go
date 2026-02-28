package tools

import (
	"context"

	"devops-infra/internal/infra/executor"
	osdriver "devops-infra/internal/infra/os"
)

type Installer struct {
	os osdriver.Driver
}

func New(os osdriver.Driver) *Installer {
	return &Installer{os: os}
}

func (t *Installer) Name() string { return "common-tools" }

func (t *Installer) IsInstalled(ctx context.Context) bool {
	exec := t.os.Exec()
	if executor.IsDryRun(exec) {
		return false
	}
	checks := []string{"curl", "gpg", "tar", "ip"}
	for _, binary := range checks {
		if !executor.ProbeSuccess(exec, "command -v "+binary) {
			return false
		}
	}
	return true
}

func (t *Installer) Install(ctx context.Context) error {
	if t.IsInstalled(ctx) {
		return nil
	}
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
