package mirror

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/infra/assets"
	osdriver "devops-infra/internal/infra/os"
)

type Options struct {
	Enable bool
	Source string
}

type Installer struct {
	os     osdriver.Driver
	enable bool
	source string
}

func New(os osdriver.Driver, opts Options) *Installer {
	return &Installer{
		os:     os,
		enable: opts.Enable,
		source: strings.TrimSpace(opts.Source),
	}
}

func (m *Installer) Name() string { return "system-mirror" }

func (m *Installer) IsInstalled(ctx context.Context) bool {
	return !m.enable
}

func (m *Installer) Install(ctx context.Context) error {
	if !m.enable {
		return nil
	}

	scriptPath, err := assets.EnsureMirrorMainScript()
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("bash %q", scriptPath)
	if m.source != "" {
		cmd += fmt.Sprintf(" --source %q", m.source)
	}

	return m.os.Exec().Run(cmd)
}
