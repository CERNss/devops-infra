package executor

import (
	"context"
	"testing"

	logmw "devops-infra/internal/middleware/log"
	tracemw "devops-infra/internal/middleware/trace"
)

func TestWithNodeAndComponentPropagateContextFields(t *testing.T) {
	rt := NormalizeRuntime(Runtime{
		Ctx:    context.Background(),
		Trace:  tracemw.NoopTraceSink(),
		Output: logmw.NoopOutputSinkFactory{},
		Logger: logmw.NoopLogger(),
	})
	rt = WithNode(rt, "node-a", "10.0.0.1")
	rt = WithComponent(rt, "containerd")

	fields := logmw.FieldsFromContext(rt.Ctx)
	if fields["node"] != "node-a" {
		t.Fatalf("node mismatch: %v", fields["node"])
	}
	if fields["node_addr"] != "10.0.0.1" {
		t.Fatalf("node_addr mismatch: %v", fields["node_addr"])
	}
	if fields["component"] != "containerd" {
		t.Fatalf("component mismatch: %v", fields["component"])
	}
}

func TestNoopRuntimeOutputHasNoArtifactPaths(t *testing.T) {
	rt := NormalizeRuntime(Runtime{
		Ctx:    context.Background(),
		Trace:  tracemw.NoopTraceSink(),
		Output: logmw.NoopOutputSinkFactory{},
		Logger: logmw.NoopLogger(),
	})

	sink, err := rt.Output.Open(logmw.RuntimeInfo{
		Ctx:    rt.Ctx,
		Logger: rt.Logger,
		LogDir: rt.LogDir,
	}, "echo hello")
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer func() { _ = sink.Close() }()

	if sink.StdoutPath() != "" || sink.StderrPath() != "" {
		t.Fatalf("expected no artifact paths, stdout=%q stderr=%q", sink.StdoutPath(), sink.StderrPath())
	}
}
