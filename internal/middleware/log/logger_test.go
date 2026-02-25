package log

import (
	"context"
	"testing"
)

func TestWithFieldsMergesAndOverrides(t *testing.T) {
	ctx := WithFields(context.Background(), map[string]any{
		"component": "kernel",
		"result":    "running",
	})
	ctx = WithFields(ctx, map[string]any{
		"result":     "success",
		"durationMs": int64(12),
	})

	fields := FieldsFromContext(ctx)
	if len(fields) != 3 {
		t.Fatalf("unexpected fields length: %d (%v)", len(fields), fields)
	}
	if fields["component"] != "kernel" {
		t.Fatalf("component mismatch: %v", fields["component"])
	}
	if fields["result"] != "success" {
		t.Fatalf("result mismatch: %v", fields["result"])
	}
	if fields["durationMs"] != int64(12) {
		t.Fatalf("duration mismatch: %v", fields["durationMs"])
	}
}

func TestFieldsFromContextReturnsCopy(t *testing.T) {
	ctx := WithFields(context.Background(), map[string]any{"component": "docker"})
	first := FieldsFromContext(ctx)
	first["component"] = "mutated"

	second := FieldsFromContext(ctx)
	if second["component"] != "docker" {
		t.Fatalf("expected immutable context fields, got %v", second["component"])
	}
}

func TestWithTraceIDNilContext(t *testing.T) {
	ctx := WithTraceID(nil, "trace-1")
	if got := TraceIDFromContext(ctx); got != "trace-1" {
		t.Fatalf("trace_id mismatch: %q", got)
	}
}
