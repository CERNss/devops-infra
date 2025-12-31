package cmd

import (
	"devops-infra/internal/infra/orchestration"
	"strings"

	"github.com/spf13/cobra"
)

var (
	k8sVersion                   string
	k8sCRISocket                 string
	k8sControlPlaneEndpoint      string
	k8sAPIServerAdvertiseAddress string
	k8sPodNetworkCIDR            string
	k8sServiceCIDR               string
	k8sServiceDNSDomain          string
	k8sImageRepository           string
	k8sToken                     string
	k8sTokenTTL                  string
	k8sUploadCerts               bool
	k8sCertificateKey            string
	k8sIgnorePreflightErrors     string
	k8sFeatureGates              string
	k8sPatchesDir                string
	k8sConfigPath                string
	k8sDisableSELinux            bool
	k8sDisableFirewall           bool
	k8sSkipInit                  bool
	k8sSetupKubeconfig           bool
	k8sCNI                       string
	k8sSkipCNI                   bool
)

var installK8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Install Kubernetes (kubeadm-based)",
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sVersion = strings.TrimSpace(k8sVersion)
		k8sCRISocket = strings.TrimSpace(k8sCRISocket)
		k8sControlPlaneEndpoint = strings.TrimSpace(k8sControlPlaneEndpoint)
		k8sAPIServerAdvertiseAddress = strings.TrimSpace(k8sAPIServerAdvertiseAddress)
		k8sPodNetworkCIDR = strings.TrimSpace(k8sPodNetworkCIDR)
		k8sServiceCIDR = strings.TrimSpace(k8sServiceCIDR)
		k8sServiceDNSDomain = strings.TrimSpace(k8sServiceDNSDomain)
		k8sImageRepository = strings.TrimSpace(k8sImageRepository)
		k8sToken = strings.TrimSpace(k8sToken)
		k8sTokenTTL = strings.TrimSpace(k8sTokenTTL)
		k8sCertificateKey = strings.TrimSpace(k8sCertificateKey)
		k8sIgnorePreflightErrors = strings.TrimSpace(k8sIgnorePreflightErrors)
		k8sFeatureGates = strings.TrimSpace(k8sFeatureGates)
		k8sPatchesDir = strings.TrimSpace(k8sPatchesDir)
		k8sConfigPath = strings.TrimSpace(k8sConfigPath)
		k8sCNI = strings.TrimSpace(k8sCNI)

		return orchestration.InstallK8s(cmd.Context(), orchestration.InstallK8sOptions{
			ExecOpts:                  execOpts,
			Version:                   k8sVersion,
			CRISocket:                 k8sCRISocket,
			ControlPlaneEndpoint:      k8sControlPlaneEndpoint,
			APIServerAdvertiseAddress: k8sAPIServerAdvertiseAddress,
			PodNetworkCIDR:            k8sPodNetworkCIDR,
			ServiceCIDR:               k8sServiceCIDR,
			ServiceDNSDomain:          k8sServiceDNSDomain,
			ImageRepository:           k8sImageRepository,
			Token:                     k8sToken,
			TokenTTL:                  k8sTokenTTL,
			UploadCerts:               k8sUploadCerts,
			CertificateKey:            k8sCertificateKey,
			IgnorePreflightErrors:     k8sIgnorePreflightErrors,
			FeatureGates:              k8sFeatureGates,
			PatchesDir:                k8sPatchesDir,
			ConfigPath:                k8sConfigPath,
			DisableSELinux:            k8sDisableSELinux,
			DisableFirewall:           k8sDisableFirewall,
			SkipInit:                  k8sSkipInit,
			SetupKubeconfig:           k8sSetupKubeconfig,
			CNI:                       k8sCNI,
			SkipCNI:                   k8sSkipCNI,
		})
	},
}

func init() {
	installCmd.AddCommand(installK8sCmd)

	installK8sCmd.Flags().StringVar(
		&k8sVersion,
		"kubernetes-version",
		"",
		"kubernetes version for kubeadm init",
	)
	installK8sCmd.Flags().StringVar(
		&k8sCRISocket,
		"cri-socket",
		"",
		"CRI socket path",
	)
	installK8sCmd.Flags().StringVar(
		&k8sControlPlaneEndpoint,
		"control-plane-endpoint",
		"",
		"control plane endpoint (vip/lb)",
	)
	installK8sCmd.Flags().StringVar(
		&k8sAPIServerAdvertiseAddress,
		"apiserver-advertise-address",
		"",
		"API server advertise address",
	)
	installK8sCmd.Flags().StringVar(
		&k8sPodNetworkCIDR,
		"pod-network-cidr",
		"",
		"pod network CIDR (default: flannel 10.244.0.0/16)",
	)
	installK8sCmd.Flags().StringVar(
		&k8sServiceCIDR,
		"service-cidr",
		"",
		"service CIDR (default: 10.96.0.0/12)",
	)
	installK8sCmd.Flags().StringVar(
		&k8sServiceDNSDomain,
		"service-dns-domain",
		"",
		"service DNS domain (default: cluster.local)",
	)
	installK8sCmd.Flags().StringVar(
		&k8sImageRepository,
		"image-repository",
		"",
		"kubernetes image repository",
	)
	installK8sCmd.Flags().StringVar(
		&k8sToken,
		"token",
		"",
		"kubeadm bootstrap token",
	)
	installK8sCmd.Flags().StringVar(
		&k8sTokenTTL,
		"token-ttl",
		"",
		"kubeadm token TTL",
	)
	installK8sCmd.Flags().BoolVar(
		&k8sUploadCerts,
		"upload-certs",
		false,
		"upload control plane certificates",
	)
	installK8sCmd.Flags().StringVar(
		&k8sCertificateKey,
		"certificate-key",
		"",
		"certificate key for upload-certs",
	)
	installK8sCmd.Flags().StringVar(
		&k8sIgnorePreflightErrors,
		"ignore-preflight-errors",
		"",
		"ignore specific preflight errors",
	)
	installK8sCmd.Flags().StringVar(
		&k8sFeatureGates,
		"feature-gates",
		"",
		"feature gates",
	)
	installK8sCmd.Flags().StringVar(
		&k8sPatchesDir,
		"patches",
		"",
		"kubeadm patches directory",
	)
	installK8sCmd.Flags().StringVar(
		&k8sConfigPath,
		"config",
		"",
		"kubeadm config file path",
	)
	installK8sCmd.Flags().BoolVar(
		&k8sDisableSELinux,
		"disable-selinux",
		false,
		"disable SELinux on RHEL",
	)
	installK8sCmd.Flags().BoolVar(
		&k8sDisableFirewall,
		"disable-firewall",
		false,
		"disable firewalld/ufw",
	)
	installK8sCmd.Flags().BoolVar(
		&k8sSkipInit,
		"skip-init",
		false,
		"skip kubeadm init",
	)
	installK8sCmd.Flags().BoolVar(
		&k8sSetupKubeconfig,
		"setup-kubeconfig",
		true,
		"setup kubeconfig for root user",
	)
	installK8sCmd.Flags().StringVar(
		&k8sCNI,
		"cni",
		"flannel",
		"cni plugin: flannel|calico|none",
	)
	installK8sCmd.Flags().BoolVar(
		&k8sSkipCNI,
		"skip-cni",
		false,
		"skip CNI installation",
	)
}
