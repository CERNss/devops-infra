package k8s

import (
	"context"
	"fmt"
	"strings"

	osdriver "devops-infra/internal/infra/os"
)

type VerifyOptions struct {
	SkipInit        bool
	SetupKubeconfig bool
	CNI             string
	SkipCNI         bool
}

type VerifyInstaller struct {
	os   osdriver.Driver
	opts VerifyOptions
}

func NewVerify(os osdriver.Driver, opts VerifyOptions) *VerifyInstaller {
	return &VerifyInstaller{os: os, opts: opts}
}

func (v *VerifyInstaller) Name() string { return "k8s-verify" }

func (v *VerifyInstaller) IsInstalled(context.Context) bool {
	return false
}

func (v *VerifyInstaller) Install(context.Context) error {
	exec := v.os.Exec()

	checks := []struct {
		desc string
		cmd  string
	}{
		{desc: "kubeadm binary", cmd: "command -v kubeadm"},
		{desc: "kubelet binary", cmd: "command -v kubelet"},
		{desc: "kubectl binary", cmd: "command -v kubectl"},
		{desc: "swap disabled", cmd: "test -z \"$(swapon --noheadings 2>/dev/null)\""},
	}

	if !v.opts.SkipInit {
		checks = append(checks,
			struct {
				desc string
				cmd  string
			}{desc: "kubeadm admin.conf", cmd: "test -f /etc/kubernetes/admin.conf"},
			struct {
				desc string
				cmd  string
			}{desc: "kubelet active", cmd: "systemctl is-active kubelet"},
		)
		if v.opts.SetupKubeconfig {
			checks = append(checks, struct {
				desc string
				cmd  string
			}{desc: "root kubeconfig", cmd: "test -f /root/.kube/config"})
		}
	}

	for _, check := range checks {
		if _, err := exec.RunWithOutput(check.cmd); err != nil {
			return fmt.Errorf("validation failed: %s: %w", check.desc, err)
		}
	}

	if v.opts.SkipInit || v.opts.SkipCNI {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(v.opts.CNI)) {
	case "", "flannel":
		if _, err := exec.RunWithOutput("KUBECONFIG=/etc/kubernetes/admin.conf kubectl get daemonset -n kube-flannel kube-flannel-ds >/dev/null 2>&1"); err != nil {
			return fmt.Errorf("validation failed: flannel deployment: %w", err)
		}
	case "calico":
		if _, err := exec.RunWithOutput("KUBECONFIG=/etc/kubernetes/admin.conf kubectl get deployment -n tigera-operator tigera-operator >/dev/null 2>&1"); err != nil {
			return fmt.Errorf("validation failed: calico deployment: %w", err)
		}
	case "none":
		return nil
	default:
		return fmt.Errorf("validation failed: unsupported cni %q", v.opts.CNI)
	}

	return nil
}
