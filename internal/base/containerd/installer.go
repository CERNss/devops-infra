package containerd

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

func (c *Installer) Name() string { return "containerd" }

func (c *Installer) IsInstalled(ctx context.Context) bool {
	_, err := c.os.Exec().RunWithOutput("containerd --version")
	return err == nil
}

func (c *Installer) Install(ctx context.Context) error {
	exec := c.os.Exec()

	// 官方 release（示例版本，可参数化）
	exec.Run(`
set -e
VERSION=1.7.28
ARCH=amd64
curl -L https://github.com/containerd/containerd/releases/download/v${VERSION}/containerd-${VERSION}-linux-${ARCH}.tar.gz \
 | tar -C /usr/local -xz
`)

	exec.Run(`
mkdir -p /etc/containerd
containerd config default > /etc/containerd/config.toml
`)

	// 强制 systemd cgroup
	exec.Run(`
sed -ri 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
`)

	exec.Run(`
curl -sSL https://raw.githubusercontent.com/containerd/containerd/main/containerd.service \
  -o /etc/systemd/system/containerd.service
`)

	exec.Run("systemctl daemon-reexec")
	exec.Run("systemctl enable containerd")
	exec.Run("systemctl restart containerd")

	return nil
}
