package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	osdriver "devops-infra/internal/os"
	"devops-infra/internal/utils/pathutil"
)

type InstallMode string

const (
	InstallModeOfficial InstallMode = "docker"
	InstallModeNerdctl  InstallMode = "nerdctl"
)

const (
	defaultNerdctlVersion = "1.7.7"
	defaultRuncVersion    = "1.1.13"
	defaultCNIVersion     = "1.5.1"
)

type Installer struct {
	os              osdriver.Driver
	mode            InstallMode
	registryMirrors []string
}

func New(os osdriver.Driver, mode InstallMode, registryMirrors []string) *Installer {
	return &Installer{os: os, mode: mode, registryMirrors: registryMirrors}
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
		scriptPath, err := pathutil.ResolvePath("scripts/mirror/docker.sh")
		if err != nil {
			return err
		}
		if err := exec.Run(fmt.Sprintf("bash %q", scriptPath)); err != nil {
			return err
		}

		if err := d.configureRegistryMirrors(); err != nil {
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

func (d *Installer) configureRegistryMirrors() error {
	if len(d.registryMirrors) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(map[string][]string{
		"registry-mirrors": d.registryMirrors,
	}, "", "  ")
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	cmd := fmt.Sprintf(
		"mkdir -p /etc/docker && printf '%%s' '%s' | base64 -d > /etc/docker/daemon.json",
		encoded,
	)

	return d.os.Exec().Run(cmd)
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
`, defaultNerdctlVersion)); err != nil {
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
