package containerd

import (
	"context"
	"fmt"
	"strings"

	osdriver "devops-infra/internal/os"
)

type Options struct {
	Version  string
	Arch     string
	Checksum string
}

const (
	defaultVersion = "1.7.28"
	defaultArch    = "amd64"
)

type Installer struct {
	os osdriver.Driver
	opts Options
}

func New(os osdriver.Driver, opts Options) *Installer {
	return &Installer{os: os, opts: opts}
}

func (c *Installer) Name() string { return "containerd" }

func (c *Installer) IsInstalled(ctx context.Context) bool {
	version, _ := c.resolveOptions()
	output, err := c.os.Exec().RunWithOutput("containerd --version")
	if err != nil {
		return false
	}
	if version == "" {
		return true
	}

	return strings.Contains(output, version)
}

func (c *Installer) Install(ctx context.Context) error {
	exec := c.os.Exec()
	version, arch := c.resolveOptions()
	checksum := strings.TrimSpace(c.opts.Checksum)

	// 官方 release（示例版本，可参数化）
	if err := exec.Run(fmt.Sprintf(`
set -e
VERSION=%s
ARCH=%s
curl -L -o /tmp/containerd.tar.gz https://github.com/containerd/containerd/releases/download/v${VERSION}/containerd-${VERSION}-linux-${ARCH}.tar.gz
`, version, arch)); err != nil {
		return err
	}

	if checksum != "" {
		sumOut, err := exec.RunWithOutput("sha256sum /tmp/containerd.tar.gz")
		if err != nil {
			return err
		}
		actual := strings.Fields(sumOut)
		if len(actual) == 0 || !strings.EqualFold(actual[0], checksum) {
			return fmt.Errorf("containerd checksum mismatch: expected %s", checksum)
		}
	}

	if err := exec.Run("tar -C /usr/local -xzf /tmp/containerd.tar.gz"); err != nil {
		return err
	}
	if err := exec.Run("rm -f /tmp/containerd.tar.gz"); err != nil {
		return err
	}

	if err := exec.Run(`
mkdir -p /etc/containerd
containerd config default > /etc/containerd/config.toml
`); err != nil {
		return err
	}

	// 强制 systemd cgroup
	if err := exec.Run(`
sed -ri 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
`); err != nil {
		return err
	}

	if err := exec.Run(`
curl -sSL https://raw.githubusercontent.com/containerd/containerd/main/containerd.service \
  -o /etc/systemd/system/containerd.service
`); err != nil {
		return err
	}

	if err := exec.Run("systemctl daemon-reexec"); err != nil {
		return err
	}
	if err := exec.Run("systemctl enable containerd"); err != nil {
		return err
	}
	if err := exec.Run("systemctl restart containerd"); err != nil {
		return err
	}

	return nil
}

func (c *Installer) resolveOptions() (string, string) {
	version := strings.TrimSpace(c.opts.Version)
	if version == "" {
		version = defaultVersion
	}

	arch := strings.TrimSpace(c.opts.Arch)
	if arch == "" {
		arch = defaultArch
	}

	return version, arch
}
