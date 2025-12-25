package docker

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/constant"
	osdriver "devops-infra/internal/infra/os"
	"devops-infra/internal/utils/mirror"
)

type InstallMode string

const (
	InstallModeOfficial InstallMode = "docker"
	InstallModeNerdctl  InstallMode = "nerdctl"
)

const (
	defaultRuncVersion = "1.1.13"
	defaultCNIVersion  = "1.5.1"
)

type Installer struct {
	os              osdriver.Driver
	mode            InstallMode
	source          string
	registryMirrors []string
	engineVersion   string
}

type Options struct {
	Mode            InstallMode
	Source          string
	RegistryMirrors []string
	EngineVersion   string
}

func New(os osdriver.Driver, opts Options) *Installer {
	return &Installer{
		os:              os,
		mode:            opts.Mode,
		source:          strings.TrimSpace(opts.Source),
		registryMirrors: opts.RegistryMirrors,
		engineVersion:   strings.TrimSpace(opts.EngineVersion),
	}
}

func (d *Installer) Name() string { return "docker" }

func (d *Installer) IsInstalled(ctx context.Context) bool {
	switch d.mode {
	case InstallModeNerdctl:
		_, err := d.os.Exec().RunWithOutput("test -L /usr/bin/docker")
		return err == nil
	case InstallModeOfficial, "":
		_, err := d.os.Exec().RunWithOutput("docker --version")
		return err == nil
	default:
		return false
	}
}

func (d *Installer) Install(ctx context.Context) error {
	exec := d.os.Exec()

	switch d.mode {
	case InstallModeNerdctl:
		if err := d.ensureNerdctl(); err != nil {
			return err
		}
		nerdctlPath, err := exec.RunWithOutput("command -v nerdctl")
		if err != nil {
			return err
		}
		nerdctlPath = strings.TrimSpace(nerdctlPath)
		if nerdctlPath == "" {
			return fmt.Errorf("nerdctl not found after installation")
		}
		return exec.Run("ln -sf " + nerdctlPath + " /usr/bin/docker")
	case InstallModeOfficial, "":
		scriptPath, err := mirror.EnsureMirrorDockerScript()
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf("bash %q", scriptPath)
		engineVersion := strings.TrimSpace(d.engineVersion)
		if engineVersion == "" {
			engineVersion = constant.DefaultDockerEngineVersion
		}
		if d.source != "" {
			cmd += fmt.Sprintf(" --source %q", d.source)
		}
		if len(d.registryMirrors) > 0 {
			cmd += fmt.Sprintf(" --source-registry %q", strings.Join(d.registryMirrors, ","))
		}
		if engineVersion != "" {
			cmd += fmt.Sprintf(" --designated-version %q", engineVersion)
		}
		if err := exec.Run(cmd); err != nil {
			return err
		}

		err = exec.Run("systemctl enable docker")
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

func (d *Installer) ensureNerdctl() error {
	exec := d.os.Exec()

	if _, err := exec.RunWithOutput("command -v nerdctl"); err != nil {
		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
curl -L https://github.com/containerd/nerdctl/releases/download/v${VERSION}/nerdctl-${VERSION}-linux-${ARCH}.tar.gz \
 | tar -C /usr/local/bin -xz
`, constant.DefaultNerdctlVersion)); err != nil {
			return err
		}
	}

	if _, err := exec.RunWithOutput("command -v runc"); err != nil {
		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
curl -L -o /usr/local/sbin/runc https://github.com/opencontainers/runc/releases/download/v${VERSION}/runc.${ARCH}
chmod +x /usr/local/sbin/runc
`, defaultRuncVersion)); err != nil {
			return err
		}
	}

	if _, err := exec.RunWithOutput("test -x /opt/cni/bin/bridge"); err != nil {
		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
mkdir -p /opt/cni/bin
curl -L https://github.com/containernetworking/plugins/releases/download/v${VERSION}/cni-plugins-linux-${ARCH}-v${VERSION}.tgz \
 | tar -C /opt/cni/bin -xz
`, defaultCNIVersion)); err != nil {
			return err
		}
	}

	return nil
}
