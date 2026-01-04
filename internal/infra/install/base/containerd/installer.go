package containerd

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/executor"
	osdriver "devops-infra/internal/infra/os"
)

type Options struct {
	Version  string
	Arch     string
	Checksum string
	// EnsureCNIConfig creates a minimal CNI config when /etc/cni/net.d is empty.
	EnsureCNIConfig bool
	CNISubnet       string
	CNIRouteDst     string
}

type Installer struct {
	os   osdriver.Driver
	opts Options
}

func New(os osdriver.Driver, opts Options) *Installer {
	return &Installer{os: os, opts: opts}
}

func (c *Installer) Name() string { return "containerd" }

func (c *Installer) IsInstalled(ctx context.Context) bool {
	exec := c.os.Exec()
	if executor.IsDryRun(exec) {
		return false
	}
	version, _ := c.resolveOptions()
	output, err := exec.RunWithOutput("containerd --version")
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

	if checksum != "" && !executor.IsDryRun(exec) {
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
set -e
mkdir -p /etc/containerd
if [ -f /etc/containerd/config.toml ]; then
  ts=$(date +%Y%m%d%H%M%S)
  mv /etc/containerd/config.toml /etc/containerd/config.toml.bak.${ts}
fi
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

	if c.opts.EnsureCNIConfig {
		if err := c.ensureCNIConfig(); err != nil {
			return err
		}
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
		version = constant.DefaultContainerdVersion
	}

	arch := strings.TrimSpace(c.opts.Arch)
	if arch == "" {
		arch = constant.DefaultContainerdArch
	}

	return version, arch
}

func (c *Installer) ensureCNIConfig() error {
	exec := c.os.Exec()
	checkCmd := "ls /etc/cni/net.d/*.conf /etc/cni/net.d/*.conflist 2>/dev/null | head -n 1"
	output, err := exec.RunWithOutput(checkCmd)
	if err == nil && strings.TrimSpace(output) != "" {
		return nil
	}

	if err := exec.Run("mkdir -p /etc/cni/net.d"); err != nil {
		return err
	}

	subnet := strings.TrimSpace(c.opts.CNISubnet)
	if subnet == "" {
		subnet = constant.DefaultNerdctlCNISubnet
	}
	routeDst := strings.TrimSpace(c.opts.CNIRouteDst)
	if routeDst == "" {
		routeDst = constant.DefaultNerdctlCNIRouteDst
	}

	return exec.Run(fmt.Sprintf(`
cat <<'EOF' > /etc/cni/net.d/99-nerdctl-bridge.conflist
{
  "cniVersion": "0.4.0",
  "name": "nerdctl",
  "plugins": [
    {
      "type": "bridge",
      "bridge": "cni0",
      "isGateway": true,
      "ipMasq": true,
      "ipam": {
        "type": "host-local",
        "ranges": [[{"subnet": "%s"}]],
        "routes": [{"dst": "%s"}]
      }
    },
    {
      "type": "portmap",
      "capabilities": {"portMappings": true}
    }
  ]
}
EOF
`, subnet, routeDst))
}
