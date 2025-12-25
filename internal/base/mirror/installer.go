package mirror

import (
	"context"
	"fmt"

	osdriver "devops-infra/internal/os"
	"devops-infra/internal/utils/pathutil"
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

	scriptPath, err := pathutil.ResolvePath("scripts/mirror/main.sh")
	if err != nil {
		return err
	}

	return m.os.Exec().Run(fmt.Sprintf(
		"bash %q",
		scriptPath,
	))
}
