package interceptor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	logmw "devops-infra/internal/middleware/log"
)

func TestZapLoggerIncludesContextFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.log")
	logger, err := NewJSONLogger(path)
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}

	ctx := logmw.WithTraceID(context.Background(), "trace-123")
	ctx = logmw.WithFields(ctx, map[string]any{
		"command":    "echo ok",
		"component":  "kernel",
		"node":       "local",
		"result":     "success",
		"duration_ms": int64(12),
	})
	logger.Info(ctx, "exec done")

	ctxFailed := logmw.WithTraceID(context.Background(), "trace-456")
	ctxFailed = logmw.WithFields(ctxFailed, map[string]any{
		"command":    "kubeadm init",
		"component":  "k8s-init",
		"node":       "local",
		"result":     "failed",
		"error_type": "exec_nonzero",
	})
	logger.Error(ctxFailed, "exec failed")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d: %s", len(lines), string(data))
	}

	first := map[string]any{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("decode first log line: %v", err)
	}
	assertField(t, first, "trace_id", "trace-123")
	assertField(t, first, "component", "kernel")
	assertField(t, first, "result", "success")
	assertField(t, first, "command", "echo ok")

	second := map[string]any{}
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatalf("decode second log line: %v", err)
	}
	assertField(t, second, "trace_id", "trace-456")
	assertField(t, second, "component", "k8s-init")
	assertField(t, second, "result", "failed")
	assertField(t, second, "error_type", "exec_nonzero")
}

func assertField(t *testing.T, payload map[string]any, key string, want string) {
	t.Helper()
	got, ok := payload[key]
	if !ok {
		t.Fatalf("missing field %q in payload: %+v", key, payload)
	}
	if got != want {
		t.Fatalf("field %q = %v, want %s", key, got, want)
	}
}
