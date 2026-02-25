package orchestration

import "testing"

func TestNormalizeK8sOptionsDefaults(t *testing.T) {
	normalized := normalizeK8sOptions(InstallK8sOptions{})

	if normalized.Version != "1.28.15" {
		t.Fatalf("unexpected default version: %s", normalized.Version)
	}
	if normalized.CRISocket != "unix:///run/containerd/containerd.sock" {
		t.Fatalf("unexpected default cri socket: %s", normalized.CRISocket)
	}
	if normalized.PodNetworkCIDR != "10.244.0.0/16" {
		t.Fatalf("unexpected default pod cidr: %s", normalized.PodNetworkCIDR)
	}
	if normalized.ServiceCIDR != "10.96.0.0/12" {
		t.Fatalf("unexpected default service cidr: %s", normalized.ServiceCIDR)
	}
	if normalized.ServiceDNSDomain != "cluster.local" {
		t.Fatalf("unexpected default dns domain: %s", normalized.ServiceDNSDomain)
	}
	if normalized.ImageRepository != "registry.k8s.io" {
		t.Fatalf("unexpected default image repository: %s", normalized.ImageRepository)
	}
	if normalized.CNI != "flannel" {
		t.Fatalf("unexpected default cni: %s", normalized.CNI)
	}
}

func TestNormalizeK8sOptionsSkipInitForcesSkipCNI(t *testing.T) {
	normalized := normalizeK8sOptions(InstallK8sOptions{
		SkipInit: true,
		SkipCNI:  false,
		CNI:      "calico",
	})

	if !normalized.SkipCNI {
		t.Fatal("expected skip-cni to be true when skip-init is true")
	}
	if normalized.CNI != "calico" {
		t.Fatalf("expected cni value preserved, got %s", normalized.CNI)
	}
}
