package orchestration

import (
	"context"
	"strings"

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/install/base"
	"devops-infra/internal/infra/os"
	platformk8s "devops-infra/internal/infra/platform/k8s"
)

type InstallK8sOptions struct {
	ExecOpts                  executor.Options
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
	DisableSELinux            bool
	DisableFirewall           bool
	SkipInit                  bool
	SetupKubeconfig           bool
	CNI                       string
	SkipCNI                   bool
}

func InstallK8s(ctx context.Context, opts InstallK8sOptions) error {
	osInfo, err := os.Detect()
	if err != nil {
		return err
	}

	exec := executor.NewLocal(opts.ExecOpts)
	driver, err := os.NewDriver(osInfo, exec)
	if err != nil {
		return err
	}

	normalized := normalizeK8sOptions(opts)

	components := []base.Component{}
	if normalized.DisableSELinux || normalized.DisableFirewall {
		components = append(components, platformk8s.NewPreflight(driver, platformk8s.PreflightOptions{
			DisableSELinux:  normalized.DisableSELinux,
			DisableFirewall: normalized.DisableFirewall,
		}))
	}
	components = append(components, platformk8s.NewRepo(driver, platformk8s.RepoOptions{
		Version: normalized.Version,
	}))
	components = append(components, platformk8s.NewPackages(driver, platformk8s.PackagesOptions{
		Version: normalized.Version,
	}))

	if !normalized.SkipInit {
		components = append(components, platformk8s.NewInit(driver, platformk8s.InitOptions{
			Version:                   normalized.Version,
			CRISocket:                 normalized.CRISocket,
			ControlPlaneEndpoint:      normalized.ControlPlaneEndpoint,
			APIServerAdvertiseAddress: normalized.APIServerAdvertiseAddress,
			PodNetworkCIDR:            normalized.PodNetworkCIDR,
			ServiceCIDR:               normalized.ServiceCIDR,
			ServiceDNSDomain:          normalized.ServiceDNSDomain,
			ImageRepository:           normalized.ImageRepository,
			Token:                     normalized.Token,
			TokenTTL:                  normalized.TokenTTL,
			UploadCerts:               normalized.UploadCerts,
			CertificateKey:            normalized.CertificateKey,
			IgnorePreflightErrors:     normalized.IgnorePreflightErrors,
			FeatureGates:              normalized.FeatureGates,
			PatchesDir:                normalized.PatchesDir,
			ConfigPath:                normalized.ConfigPath,
		}))
	}

	if normalized.SetupKubeconfig {
		components = append(components, platformk8s.NewKubeconfig(driver, platformk8s.KubeconfigOptions{
			User: "root",
		}))
	}

	if !normalized.SkipCNI {
		components = append(components, platformk8s.NewCNI(driver, platformk8s.CNIOptions{
			CNI:            normalized.CNI,
			PodNetworkCIDR: normalized.PodNetworkCIDR,
		}))
	}

	installer := base.New(components...)
	return installer.Install(ctx)
}

func normalizeK8sOptions(opts InstallK8sOptions) InstallK8sOptions {
	if strings.TrimSpace(opts.Version) == "" {
		opts.Version = constant.DefaultK8sVersion
	}
	if strings.TrimSpace(opts.CRISocket) == "" {
		opts.CRISocket = "unix:///run/containerd/containerd.sock"
	}
	if strings.TrimSpace(opts.PodNetworkCIDR) == "" {
		opts.PodNetworkCIDR = "10.244.0.0/16"
	}
	if strings.TrimSpace(opts.ServiceCIDR) == "" {
		opts.ServiceCIDR = "10.96.0.0/12"
	}
	if strings.TrimSpace(opts.ServiceDNSDomain) == "" {
		opts.ServiceDNSDomain = "cluster.local"
	}
	if strings.TrimSpace(opts.ImageRepository) == "" {
		opts.ImageRepository = "registry.k8s.io"
	}
	if strings.TrimSpace(opts.CNI) == "" {
		opts.CNI = "flannel"
	}
	return opts
}
