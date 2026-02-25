package executor

import (
	"testing"

	logmw "devops-infra/internal/middleware/log"
)

func TestWithNodeAndComponentPropagateContextFields(t *testing.T) {
	rt := DefaultRuntime()
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
