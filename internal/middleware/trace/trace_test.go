package trace

import (
	"errors"
	"testing"
	"time"
)

func TestNewTraceEventRequiredFields(t *testing.T) {
	start := time.Now()
	end := start.Add(15 * time.Millisecond)
	event := NewTraceEvent(
		"echo ok",
		"trace-1",
		"node-1",
		"127.0.0.1",
		"kernel",
		"success",
		"",
		"/tmp/stdout.log",
		"/tmp/stderr.log",
		start,
		end,
		"",
		"",
		nil,
		false,
		false,
	)

	if event.SchemaVersion != "v1" {
		t.Fatalf("expected schema version v1, got %q", event.SchemaVersion)
	}
	if event.Result != "success" {
		t.Fatalf("expected result success, got %q", event.Result)
	}
	if event.Component != "kernel" {
		t.Fatalf("expected component kernel, got %q", event.Component)
	}
	if event.DurationMs <= 0 {
		t.Fatalf("expected positive duration, got %d", event.DurationMs)
	}

	failed := NewTraceEvent(
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
		end,
		"",
		"",
		errors.New("exit status 1"),
		false,
		false,
	)
	if failed.Result != "failed" {
		t.Fatalf("expected failed result, got %q", failed.Result)
	}
	if failed.ErrorType != "exec_nonzero" {
		t.Fatalf("expected error type exec_nonzero, got %q", failed.ErrorType)
	}
	if failed.Err == "" {
		t.Fatalf("expected error text to be set")
	}
}
