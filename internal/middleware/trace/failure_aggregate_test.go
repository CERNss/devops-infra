package trace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFailureAggregatorAndSummary(t *testing.T) {
	tmp := t.TempDir()
	agg, err := NewFailureAggregator(tmp, "install-base")
	if err != nil {
		t.Fatalf("create aggregator: %v", err)
	}

	start := time.Now()
	agg.OnCommand(NewTraceEvent(
		"echo ok",
		"trace-1",
		"node-1",
		"127.0.0.1",
		"kernel",
		"success",
		"",
		"",
		"",
		start,
		start.Add(10*time.Millisecond),
		"",
		"",
		nil,
		false,
		false,
	))
	agg.OnCommand(NewTraceEvent(
		"kubeadm init",
		"trace-2",
		"node-1",
		"127.0.0.1",
		"k8s-init",
		"failed",
		"exec_nonzero",
		"",
		"",
		start,
		start.Add(20*time.Millisecond),
		"",
		"",
		errDummy{},
		false,
		false,
	))

	if !agg.HasFailures() {
		t.Fatal("expected failures")
	}

	summary := agg.Summary()
	if summary.FailedCommands != 1 || summary.TotalCommands != 2 {
		t.Fatalf("unexpected counts: %+v", summary)
	}
	if len(summary.FailedComponents) != 1 || summary.FailedComponents[0] != "k8s-init" {
		t.Fatalf("unexpected failed components: %+v", summary.FailedComponents)
	}

	if got := FormatFailureSummary(summary); !strings.Contains(got, "installation failed") {
		t.Fatalf("expected formatted summary, got %q", got)
	}

	if err := agg.Close(); err != nil {
		t.Fatalf("close aggregator: %v", err)
	}

	if _, err := os.Stat(summary.ErrorsPath); err != nil {
		t.Fatalf("errors artifact missing: %v", err)
	}
	if _, err := os.Stat(summary.SummaryPath); err != nil {
		t.Fatalf("summary artifact missing: %v", err)
	}

	data, err := os.ReadFile(filepath.Clean(summary.SummaryPath))
	if err != nil {
		t.Fatalf("read summary file: %v", err)
	}
	if !strings.Contains(string(data), "\"workflow\": \"install-base\"") {
		t.Fatalf("summary file missing workflow: %s", string(data))
	}
}

func TestFormatFailureSummaryNoFailures(t *testing.T) {
	if got := FormatFailureSummary(FailureSummary{FailedCommands: 0}); got != "" {
		t.Fatalf("expected empty summary for successful run, got %q", got)
	}
}

type errDummy struct{}

func (errDummy) Error() string { return "exit status 1" }
