package cni

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/executor"
	osdriver "devops-infra/internal/infra/os"
)

type Options struct {
	Version string
	Arch    string
}

type Installer struct {
	os   osdriver.Driver
	opts Options
}

func New(os osdriver.Driver, opts Options) *Installer {
	return &Installer{os: os, opts: opts}
}

func (c *Installer) Name() string { return "cni-plugins" }

func (c *Installer) IsInstalled(ctx context.Context) bool {
	exec := c.os.Exec()
	if executor.IsDryRun(exec) {
		return false
	}
	if !executor.ProbeSuccess(exec, "test -x /opt/cni/bin/bridge") {
		return false
	}
	if !executor.ProbeSuccess(exec, "test -x /opt/cni/bin/portmap") {
		return false
	}
	return true
}

func (c *Installer) Install(ctx context.Context) error {
	exec := c.os.Exec()
	version, arch := c.resolveOptions()
	return exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=%s
mkdir -p /opt/cni/bin
curl -L https://github.com/containernetworking/plugins/releases/download/v${VERSION}/cni-plugins-linux-${ARCH}-v${VERSION}.tgz \
 | tar -C /opt/cni/bin -xz
`, version, arch))
}

func (c *Installer) resolveOptions() (string, string) {
	version := strings.TrimSpace(c.opts.Version)
	if version == "" {
		version = constant.DefaultCNIVersion
	}
	arch := strings.TrimSpace(c.opts.Arch)
	if arch == "" {
		arch = constant.DefaultContainerdArch
	}
	return version, arch
}
