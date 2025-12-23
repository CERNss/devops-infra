package kernel

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

func (k *Installer) Name() string { return "kernel" }

func (k *Installer) IsInstalled(ctx context.Context) bool {
	return false // 始终执行（幂等）
}

func (k *Installer) Install(ctx context.Context) error {
	exec := k.os.Exec()

	exec.Run("swapoff -a")
	exec.Run(`sed -ri 's/^\s*([^#]+\s+swap\s+)/#\1/' /etc/fstab`)

	k.os.LoadKernelModules("overlay", "br_netfilter")

	return k.os.Sysctl(map[string]string{
		"net.bridge.bridge-nf-call-iptables":  "1",
		"net.bridge.bridge-nf-call-ip6tables": "1",
		"net.ipv4.ip_forward":                 "1",
	})
}
