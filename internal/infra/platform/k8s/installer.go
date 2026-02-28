package k8s

import (
	"context"
	"fmt"
	"strings"

	"devops-infra/internal/infra/executor"
	osdriver "devops-infra/internal/infra/os"
)

type PreflightOptions struct {
	DisableSELinux  bool
	DisableFirewall bool
}

type PreflightInstaller struct {
	os   osdriver.Driver
	opts PreflightOptions
}

func NewPreflight(os osdriver.Driver, opts PreflightOptions) *PreflightInstaller {
	return &PreflightInstaller{os: os, opts: opts}
}

func (p *PreflightInstaller) Name() string { return "k8s-preflight" }

func (p *PreflightInstaller) IsInstalled(ctx context.Context) bool {
	if !p.opts.DisableSELinux && !p.opts.DisableFirewall {
		return true
	}
	return false
}

func (p *PreflightInstaller) Install(ctx context.Context) error {
	exec := p.os.Exec()
	if p.opts.DisableSELinux && p.os.Family() == "rhel" {
		if err := exec.Run("setenforce 0 || true"); err != nil {
			return err
		}
		if err := exec.Run(`sed -i 's/^SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config`); err != nil {
			return err
		}
	}
	if p.opts.DisableFirewall {
		switch p.os.Family() {
		case "rhel":
			if err := exec.Run("systemctl stop firewalld"); err != nil {
				return err
			}
			if err := exec.Run("systemctl disable firewalld"); err != nil {
				return err
			}
		case "debian":
			if err := exec.Run("systemctl stop ufw"); err != nil {
				return err
			}
			if err := exec.Run("systemctl disable ufw"); err != nil {
				return err
			}
		}
	}
	return nil
}

type RepoOptions struct {
	Version string
}

type RepoInstaller struct {
	os   osdriver.Driver
	opts RepoOptions
}

func NewRepo(os osdriver.Driver, opts RepoOptions) *RepoInstaller {
	return &RepoInstaller{os: os, opts: opts}
}

func (r *RepoInstaller) Name() string { return "k8s-repo" }

func (r *RepoInstaller) IsInstalled(ctx context.Context) bool {
	exec := r.os.Exec()
	switch r.os.Family() {
	case "rhel":
		return executor.ProbeSuccess(exec, "test -f /etc/yum.repos.d/kubernetes.repo")
	case "debian":
		return executor.ProbeSuccess(exec, "test -f /etc/apt/sources.list.d/kubernetes.list")
	default:
		return false
	}
}

func (r *RepoInstaller) Install(ctx context.Context) error {
	exec := r.os.Exec()
	repoVersion := minorVersion(r.opts.Version)

	switch r.os.Family() {
	case "rhel":
		if err := r.os.InstallPackages("ca-certificates", "curl", "gpg"); err != nil {
			return err
		}
		if err := exec.Run(fmt.Sprintf(`
cat <<'EOF' > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes v%s
baseurl=https://pkgs.k8s.io/core:/stable:/v%s/rpm/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://pkgs.k8s.io/core:/stable:/v%s/rpm/repodata/repomd.xml.key
EOF
`, repoVersion, repoVersion, repoVersion)); err != nil {
			return err
		}
	case "debian":
		if err := r.os.InstallPackages("ca-certificates", "curl", "gpg"); err != nil {
			return err
		}
		if err := exec.Run("mkdir -p /etc/apt/keyrings"); err != nil {
			return err
		}
		if err := exec.Run(fmt.Sprintf(
			"curl -fsSL https://pkgs.k8s.io/core:/stable:/v%s/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-yum-keyring.gpg",
			repoVersion,
		)); err != nil {
			return err
		}
		if err := exec.Run(fmt.Sprintf(
			"echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-yum-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v%s/deb/ /' > /etc/apt/sources.list.d/kubernetes.list",
			repoVersion,
		)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported OS family for k8s repo: %s", r.os.Family())
	}

	return nil
}

type PackagesOptions struct {
	Version string
}

type PackagesInstaller struct {
	os   osdriver.Driver
	opts PackagesOptions
}

func NewPackages(os osdriver.Driver, opts PackagesOptions) *PackagesInstaller {
	return &PackagesInstaller{os: os, opts: opts}
}

func (p *PackagesInstaller) Name() string { return "k8s-packages" }

func (p *PackagesInstaller) IsInstalled(ctx context.Context) bool {
	exec := p.os.Exec()
	if !executor.ProbeSuccess(exec, "command -v kubeadm") {
		return false
	}
	if !executor.ProbeSuccess(exec, "command -v kubelet") {
		return false
	}
	if !executor.ProbeSuccess(exec, "command -v kubectl") {
		return false
	}
	return true
}

func (p *PackagesInstaller) Install(ctx context.Context) error {
	if err := p.os.Update(); err != nil {
		return err
	}

	switch p.os.Family() {
	case "rhel":
		version := strings.TrimSpace(p.opts.Version)
		if version != "" {
			if err := p.os.InstallPackages(
				"kubelet-"+version,
				"kubeadm-"+version,
				"kubectl-"+version,
			); err != nil {
				return err
			}
			break
		}
		if err := p.os.InstallPackages("kubelet", "kubeadm", "kubectl"); err != nil {
			return err
		}
	case "debian":
		if err := p.os.InstallPackages("kubelet", "kubeadm", "kubectl"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported OS family for k8s packages: %s", p.os.Family())
	}

	return p.os.EnableService("kubelet")
}

type InitOptions struct {
	Version                   string
	CRISocket                 string
	ControlPlaneEndpoint      string
	APIServerAdvertiseAddress string
	PodNetworkCIDR            string
	ServiceCIDR               string
	ServiceDNSDomain          string
	ImageRepository           string
	Token                     string
	TokenTTL                  string
	UploadCerts               bool
	CertificateKey            string
	IgnorePreflightErrors     string
	FeatureGates              string
	PatchesDir                string
	ConfigPath                string
}

type InitInstaller struct {
	os   osdriver.Driver
	opts InitOptions
}

func NewInit(os osdriver.Driver, opts InitOptions) *InitInstaller {
	return &InitInstaller{os: os, opts: opts}
}

func (i *InitInstaller) Name() string { return "k8s-init" }

func (i *InitInstaller) IsInstalled(ctx context.Context) bool {
	return executor.ProbeSuccess(i.os.Exec(), "test -f /etc/kubernetes/admin.conf")
}

func (i *InitInstaller) Install(ctx context.Context) error {
	exec := i.os.Exec()
	if i.opts.ConfigPath != "" {
		return exec.Run(fmt.Sprintf("kubeadm init --config %q", i.opts.ConfigPath))
	}

	cmd := "kubeadm init"
	if strings.TrimSpace(i.opts.Version) != "" {
		cmd += fmt.Sprintf(" --kubernetes-version=%q", strings.TrimSpace(i.opts.Version))
	}
	if strings.TrimSpace(i.opts.CRISocket) != "" {
		cmd += fmt.Sprintf(" --cri-socket %q", strings.TrimSpace(i.opts.CRISocket))
	}
	if strings.TrimSpace(i.opts.ControlPlaneEndpoint) != "" {
		cmd += fmt.Sprintf(" --control-plane-endpoint %q", strings.TrimSpace(i.opts.ControlPlaneEndpoint))
	}
	if strings.TrimSpace(i.opts.APIServerAdvertiseAddress) != "" {
		cmd += fmt.Sprintf(" --apiserver-advertise-address %q", strings.TrimSpace(i.opts.APIServerAdvertiseAddress))
	}
	if strings.TrimSpace(i.opts.PodNetworkCIDR) != "" {
		cmd += fmt.Sprintf(" --pod-network-cidr %q", strings.TrimSpace(i.opts.PodNetworkCIDR))
	}
	if strings.TrimSpace(i.opts.ServiceCIDR) != "" {
		cmd += fmt.Sprintf(" --service-cidr %q", strings.TrimSpace(i.opts.ServiceCIDR))
	}
	if strings.TrimSpace(i.opts.ServiceDNSDomain) != "" {
		cmd += fmt.Sprintf(" --service-dns-domain %q", strings.TrimSpace(i.opts.ServiceDNSDomain))
	}
	if strings.TrimSpace(i.opts.ImageRepository) != "" {
		cmd += fmt.Sprintf(" --image-repository %q", strings.TrimSpace(i.opts.ImageRepository))
	}
	if strings.TrimSpace(i.opts.Token) != "" {
		cmd += fmt.Sprintf(" --token %q", strings.TrimSpace(i.opts.Token))
	}
	if strings.TrimSpace(i.opts.TokenTTL) != "" {
		cmd += fmt.Sprintf(" --token-ttl %q", strings.TrimSpace(i.opts.TokenTTL))
	}
	if i.opts.UploadCerts {
		cmd += " --upload-certs"
	}
	if strings.TrimSpace(i.opts.CertificateKey) != "" {
		cmd += fmt.Sprintf(" --certificate-key %q", strings.TrimSpace(i.opts.CertificateKey))
	}
	if strings.TrimSpace(i.opts.IgnorePreflightErrors) != "" {
		cmd += fmt.Sprintf(" --ignore-preflight-errors %q", strings.TrimSpace(i.opts.IgnorePreflightErrors))
	}
	if strings.TrimSpace(i.opts.FeatureGates) != "" {
		cmd += fmt.Sprintf(" --feature-gates %q", strings.TrimSpace(i.opts.FeatureGates))
	}
	if strings.TrimSpace(i.opts.PatchesDir) != "" {
		cmd += fmt.Sprintf(" --patches %q", strings.TrimSpace(i.opts.PatchesDir))
	}

	return exec.Run(cmd)
}

type KubeconfigOptions struct {
	User string
}

type KubeconfigInstaller struct {
	os   osdriver.Driver
	opts KubeconfigOptions
}

func NewKubeconfig(os osdriver.Driver, opts KubeconfigOptions) *KubeconfigInstaller {
	return &KubeconfigInstaller{os: os, opts: opts}
}

func (k *KubeconfigInstaller) Name() string { return "k8s-kubeconfig" }

func (k *KubeconfigInstaller) IsInstalled(ctx context.Context) bool {
	if strings.TrimSpace(k.opts.User) == "" {
		return true
	}
	return executor.ProbeSuccess(k.os.Exec(), "test -f /root/.kube/config")
}

func (k *KubeconfigInstaller) Install(ctx context.Context) error {
	if strings.TrimSpace(k.opts.User) == "" {
		return nil
	}
	exec := k.os.Exec()
	if err := exec.Run("mkdir -p /root/.kube"); err != nil {
		return err
	}
	if err := exec.Run("cp -f /etc/kubernetes/admin.conf /root/.kube/config"); err != nil {
		return err
	}
	return exec.Run("chown 0:0 /root/.kube/config")
}

type CNIOptions struct {
	CNI            string
	PodNetworkCIDR string
}

type CNIInstaller struct {
	os   osdriver.Driver
	opts CNIOptions
}

func NewCNI(os osdriver.Driver, opts CNIOptions) *CNIInstaller {
	return &CNIInstaller{os: os, opts: opts}
}

func (c *CNIInstaller) Name() string { return "k8s-cni" }

func (c *CNIInstaller) IsInstalled(ctx context.Context) bool {
	if strings.EqualFold(strings.TrimSpace(c.opts.CNI), "none") {
		return true
	}
	return false
}

func (c *CNIInstaller) Install(ctx context.Context) error {
	exec := c.os.Exec()
	cni := strings.ToLower(strings.TrimSpace(c.opts.CNI))
	cidr := strings.TrimSpace(c.opts.PodNetworkCIDR)
	if cidr == "" {
		cidr = "10.244.0.0/16"
	}

	switch cni {
	case "flannel", "":
		if err := exec.Run("curl -fsSL -o /tmp/kube-flannel.yml https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml"); err != nil {
			return err
		}
		if err := exec.Run(fmt.Sprintf(
			`sed -ri 's#"Network": "[^"]+"#"Network": "%s"#' /tmp/kube-flannel.yml`,
			cidr,
		)); err != nil {
			return err
		}
		return exec.Run("KUBECONFIG=/etc/kubernetes/admin.conf kubectl apply -f /tmp/kube-flannel.yml")
	case "calico":
		if err := exec.Run("KUBECONFIG=/etc/kubernetes/admin.conf kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.3/manifests/tigera-operator.yaml"); err != nil {
			return err
		}
		if err := exec.Run("curl -fsSL -o /tmp/calico-custom-resources.yaml https://raw.githubusercontent.com/projectcalico/calico/v3.28.3/manifests/custom-resources.yaml"); err != nil {
			return err
		}
		if err := exec.Run(fmt.Sprintf(
			`sed -ri 's#^([[:space:]]*cidr:).*#\1 %s#' /tmp/calico-custom-resources.yaml`,
			cidr,
		)); err != nil {
			return err
		}
		return exec.Run("KUBECONFIG=/etc/kubernetes/admin.conf kubectl apply -f /tmp/calico-custom-resources.yaml")
	case "none":
		return nil
	default:
		return fmt.Errorf("unsupported CNI: %s", cni)
	}
}

func minorVersion(version string) string {
	clean := strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.Split(clean, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	if clean == "" {
		return "1.28"
	}
	return clean
}
