package orchestration

import (
	"context"
	"fmt"
	"os"
	"strings"

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/install/base"
	osinfra "devops-infra/internal/infra/os"
	platformk8s "devops-infra/internal/infra/platform/k8s"
	"devops-infra/internal/interceptor"
	tracemw "devops-infra/internal/middleware/trace"
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
	osInfo, err := osinfra.Detect()
	if err != nil {
		return err
	}

	trace := tracemw.DefaultTraceSink()
	logger := interceptor.DefaultLogger(constant.DefaultLogDir)
	aggregator, aggErr := tracemw.NewFailureAggregator(constant.DefaultLogDir, "install-k8s")
	if aggErr != nil {
		logger.Warn(ctx, fmt.Sprintf("failed to initialize failure aggregator: %v", aggErr))
	} else {
		trace = tracemw.NewMultiTraceSink(trace, aggregator)
	}
	runtime := executor.NewRuntime(ctx, trace)
	runtime = executor.WithLogDir(runtime, constant.DefaultLogDir)
	runtime = executor.WithLogger(runtime, logger)

	normalized := normalizeK8sOptions(opts)

	components := []base.Component{}
	newDriver := func(component string) (osinfra.Driver, error) {
		componentRuntime := executor.WithComponent(runtime, component)
		exec := executor.NewLocalWithRuntime(opts.ExecOpts, componentRuntime)
		return osinfra.NewDriver(osInfo, exec)
	}

	if normalized.DisableSELinux || normalized.DisableFirewall {
		driver, driverErr := newDriver("k8s-preflight")
		if driverErr != nil {
			return driverErr
		}
		components = append(components, platformk8s.NewPreflight(driver, platformk8s.PreflightOptions{
			DisableSELinux:  normalized.DisableSELinux,
			DisableFirewall: normalized.DisableFirewall,
		}))
	}

	driverRepo, err := newDriver("k8s-repo")
	if err != nil {
		return err
	}
	components = append(components, platformk8s.NewRepo(driverRepo, platformk8s.RepoOptions{
		Version: normalized.Version,
	}))

	driverPackages, err := newDriver("k8s-packages")
	if err != nil {
		return err
	}
	components = append(components, platformk8s.NewPackages(driverPackages, platformk8s.PackagesOptions{
		Version: normalized.Version,
	}))

	if !normalized.SkipInit {
		driverInit, driverErr := newDriver("k8s-init")
		if driverErr != nil {
			return driverErr
		}
		components = append(components, platformk8s.NewInit(driverInit, platformk8s.InitOptions{
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
		driverKubeconfig, driverErr := newDriver("k8s-kubeconfig")
		if driverErr != nil {
			return driverErr
		}
		components = append(components, platformk8s.NewKubeconfig(driverKubeconfig, platformk8s.KubeconfigOptions{
			User: "root",
		}))
	}

	if !normalized.SkipCNI {
		driverCNI, driverErr := newDriver("k8s-cni")
		if driverErr != nil {
			return driverErr
		}
		components = append(components, platformk8s.NewCNI(driverCNI, platformk8s.CNIOptions{
			CNI:            normalized.CNI,
			PodNetworkCIDR: normalized.PodNetworkCIDR,
		}))
	}

	driverVerify, err := newDriver("k8s-verify")
	if err != nil {
		return err
	}
	components = append(components, platformk8s.NewVerify(driverVerify, platformk8s.VerifyOptions{
		SkipInit:        normalized.SkipInit,
		SetupKubeconfig: normalized.SetupKubeconfig,
		CNI:             normalized.CNI,
		SkipCNI:         normalized.SkipCNI,
	}))

	installer := base.New(components...)
	installer = installer.WithLogger(logger)
	if aggregator != nil {
		defer func() {
			_ = aggregator.Close()
			if !aggregator.HasFailures() {
				return
			}
			fmt.Fprint(os.Stdout, tracemw.FormatFailureSummary(aggregator.Summary()))
		}()
	}
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
	if opts.SkipInit {
		opts.SkipCNI = true
	}
	return opts
}
