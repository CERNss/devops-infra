package orchestration

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	tracemw "devops-infra/internal/middleware/trace"
)

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

func TestNormalizeK8sOptionsSkipInitAndSkipCNIAreIndependent(t *testing.T) {
	normalized := normalizeK8sOptions(InstallK8sOptions{
		SkipInit: true,
		SkipCNI:  false,
		CNI:      "calico",
	})

	if normalized.SkipCNI {
		t.Fatal("expected skip-cni to remain unchanged when skip-init is true")
	}
	if normalized.CNI != "calico" {
		t.Fatalf("expected cni value preserved, got %s", normalized.CNI)
	}
}

type stubFailureAggregator struct {
	closed      bool
	hasFailures bool
	summary     tracemw.FailureSummary
}

func (s *stubFailureAggregator) Close() error {
	s.closed = true
	return nil
}

func (s *stubFailureAggregator) HasFailures() bool {
	return s.hasFailures
}

func (s *stubFailureAggregator) Summary() tracemw.FailureSummary {
	return s.summary
}

func TestFinalizeFailureAggregatorClosesOnEarlyReturn(t *testing.T) {
	agg := &stubFailureAggregator{}

	err := func() error {
		defer finalizeFailureAggregator(agg, io.Discard)
		return errors.New("early return")
	}()
	if err == nil {
		t.Fatal("expected early return error")
	}
	if !agg.closed {
		t.Fatal("expected aggregator close on early return")
	}
}

func TestFinalizeFailureAggregatorPrintsSummaryOnFailures(t *testing.T) {
	agg := &stubFailureAggregator{
		hasFailures: true,
		summary: tracemw.FailureSummary{
			Workflow:       "install-k8s",
			FailedCommands: 1,
			TotalCommands:  3,
		},
	}

	var out bytes.Buffer
	finalizeFailureAggregator(agg, &out)

	if !agg.closed {
		t.Fatal("expected aggregator to be closed")
	}
	if !strings.Contains(out.String(), "[summary] installation failed") {
		t.Fatalf("expected summary output, got %q", out.String())
	}
}
