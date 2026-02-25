package validate

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/infra/install/base/docker"
	osdriver "devops-infra/internal/infra/os"
)

type Options struct {
	Mode docker.InstallMode
}

type Installer struct {
	os   osdriver.Driver
	opts Options
}

func New(os osdriver.Driver, opts Options) *Installer {
	return &Installer{os: os, opts: opts}
}

func (v *Installer) Name() string { return "base-postflight" }

func (v *Installer) IsInstalled(context.Context) bool {
	return false
}

func (v *Installer) Install(context.Context) error {
	exec := v.os.Exec()

	checks := []struct {
		desc string
		cmd  string
	}{
		{desc: "kernel sysctl readiness", cmd: "sysctl -n net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward"},
		{desc: "curl binary", cmd: "command -v curl"},
		{desc: "gpg binary", cmd: "command -v gpg"},
		{desc: "tar binary", cmd: "command -v tar"},
		{desc: "ip binary", cmd: "command -v ip"},
		{desc: "containerd active", cmd: "systemctl is-active containerd"},
	}

	for _, check := range checks {
		if _, err := exec.RunWithOutput(check.cmd); err != nil {
			return fmt.Errorf("validation failed: %s: %w", check.desc, err)
		}
	}

	switch strings.TrimSpace(string(v.opts.Mode)) {
	case string(docker.InstallModeNerdctl):
		if _, err := exec.RunWithOutput("command -v nerdctl"); err != nil {
			return fmt.Errorf("validation failed: nerdctl binary: %w", err)
		}
		if _, err := exec.RunWithOutput("test -L /usr/bin/docker"); err != nil {
			return fmt.Errorf("validation failed: docker symlink for nerdctl mode: %w", err)
		}
	case "", string(docker.InstallModeOfficial):
		if _, err := exec.RunWithOutput("command -v docker"); err != nil {
			return fmt.Errorf("validation failed: docker binary: %w", err)
		}
		if _, err := exec.RunWithOutput("systemctl is-active docker"); err != nil {
			return fmt.Errorf("validation failed: docker active: %w", err)
		}
	default:
		return fmt.Errorf("validation failed: unsupported docker mode %q", v.opts.Mode)
	}

	return nil
}
