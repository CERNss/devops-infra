package docker

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/executor"
	osdriver "devops-infra/internal/infra/os"
	"devops-infra/internal/infra/reconcile"
	"devops-infra/internal/utils/mirror"
)

type InstallMode string

const (
	InstallModeOfficial InstallMode = "docker"
	InstallModeNerdctl  InstallMode = "nerdctl"
)

const (
	defaultRuncVersion = "1.1.13"
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
	exec := d.os.Exec()
	if executor.IsDryRun(exec) {
		return false
	}
	switch d.mode {
	case InstallModeNerdctl:
		return executor.ProbeSuccess(exec, "test -L /usr/bin/docker")
	case InstallModeOfficial, "":
		return executor.ProbeSuccess(exec, "docker --version") &&
			executor.ProbeSuccess(exec, "systemctl is-active docker >/dev/null 2>&1")
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
		if executor.IsDryRun(exec) {
			return exec.Run("ln -sf $(command -v nerdctl) /usr/bin/docker")
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
		engineVersion := strings.TrimSpace(d.engineVersion)
		if engineVersion == "" {
			engineVersion = constant.DefaultDockerEngineVersion
		}
		_, err := reconcile.EnsureServiceHealthy(reconcile.ServiceOptions{
			Exec:           exec,
			Label:          "docker",
			HealthCheckCmd: "command -v docker >/dev/null 2>&1 && systemctl is-active docker >/dev/null 2>&1",
			RestartCmd:     "systemctl restart docker",
			Reinstall: func() error {
				return d.installOfficialFresh(engineVersion)
			},
		})
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported docker install mode: %s", d.mode)
	}
}

func (d *Installer) installOfficialFresh(engineVersion string) error {
	exec := d.os.Exec()
	scriptPath, err := mirror.EnsureMirrorDockerScript()
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("bash %q", scriptPath)
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
	if err := exec.Run("systemctl enable docker"); err != nil {
		return err
	}
	return exec.Run("systemctl restart docker")
}

func (d *Installer) ensureNerdctl() error {
	exec := d.os.Exec()

	if executor.IsDryRun(exec) {
		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
curl -L https://github.com/containerd/nerdctl/releases/download/v${VERSION}/nerdctl-${VERSION}-linux-${ARCH}.tar.gz \
 | tar -C /usr/local/bin -xz
`, constant.DefaultNerdctlVersion)); err != nil {
			return err
		}

		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
curl -L -o /usr/local/sbin/runc https://github.com/opencontainers/runc/releases/download/v${VERSION}/runc.${ARCH}
chmod +x /usr/local/sbin/runc
`, defaultRuncVersion)); err != nil {
			return err
		}

		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
mkdir -p /opt/cni/bin
curl -L https://github.com/containernetworking/plugins/releases/download/v${VERSION}/cni-plugins-linux-${ARCH}-v${VERSION}.tgz \
 | tar -C /opt/cni/bin -xz
`, constant.DefaultCNIVersion)); err != nil {
			return err
		}

		return nil
	}

	if !executor.ProbeSuccess(exec, "command -v nerdctl") {
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

	if !executor.ProbeSuccess(exec, "command -v runc") {
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

	if !executor.ProbeSuccess(exec, "test -x /opt/cni/bin/bridge") {
		if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=amd64
mkdir -p /opt/cni/bin
curl -L https://github.com/containernetworking/plugins/releases/download/v${VERSION}/cni-plugins-linux-${ARCH}-v${VERSION}.tgz \
 | tar -C /opt/cni/bin -xz
`, constant.DefaultCNIVersion)); err != nil {
			return err
		}
	}

	return nil
}
