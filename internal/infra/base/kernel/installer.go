package kernel

import (
	"context"
	osdriver "devops-infra/internal/infra/os"
	"strings"
)

type Installer struct {
	os osdriver.Driver
}

func New(os osdriver.Driver) *Installer {
	return &Installer{os: os}
}

func (k *Installer) Name() string { return "kernel" }

func (k *Installer) IsInstalled(ctx context.Context) bool {
	exec := k.os.Exec()
	output, err := exec.RunWithOutput("sysctl -n net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward")
	if err != nil {
		return false
	}
	values := strings.Fields(output)
	if len(values) < 3 {
		return false
	}
	return values[0] == "1" && values[1] == "1" && values[2] == "1"
}

func (k *Installer) Install(ctx context.Context) error {
	exec := k.os.Exec()
	if k.IsInstalled(ctx) {
		return nil
	}

	if err := exec.Run("swapoff -a"); err != nil {
		return err
	}
	if err := exec.Run(`sed -ri 's/^\s*([^#]+\s+swap\s+)/#\1/' /etc/fstab`); err != nil {
		return err
	}

	if err := k.os.LoadKernelModules("overlay", "br_netfilter"); err != nil {
		return err
	}

	return k.os.Sysctl(map[string]string{
		"net.bridge.bridge-nf-call-iptables":  "1",
		"net.bridge.bridge-nf-call-ip6tables": "1",
		"net.ipv4.ip_forward":                 "1",
	})
}
