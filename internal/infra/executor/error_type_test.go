package executor

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

func TestClassifyError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := ClassifyError(nil); got != "" {
			t.Fatalf("expected empty classification, got %q", got)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		if got := ClassifyError(context.DeadlineExceeded); got != ErrorTypeExecTimeout {
			t.Fatalf("expected %q, got %q", ErrorTypeExecTimeout, got)
		}
	})

	t.Run("exec non-zero", func(t *testing.T) {
		err := exec.Command("bash", "-c", "exit 7").Run()
		if got := ClassifyError(err); got != ErrorTypeExecNonZero {
			t.Fatalf("expected %q, got %q (%v)", ErrorTypeExecNonZero, got, err)
		}
	})

	t.Run("network", func(t *testing.T) {
		err := errors.New("dial tcp: lookup registry.example.invalid: no such host")
		if got := ClassifyError(err); got != ErrorTypeNetworkFetch {
			t.Fatalf("expected %q, got %q", ErrorTypeNetworkFetch, got)
		}
	})

	t.Run("unsupported os", func(t *testing.T) {
		err := errors.New("unsupported os family: unknown")
		if got := ClassifyError(err); got != ErrorTypeUnsupportedOS {
			t.Fatalf("expected %q, got %q", ErrorTypeUnsupportedOS, got)
		}
	})

	t.Run("validation", func(t *testing.T) {
		err := errors.New("containerd checksum mismatch: expected deadbeef")
		if got := ClassifyError(err); got != ErrorTypeValidation {
			t.Fatalf("expected %q, got %q", ErrorTypeValidation, got)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		err := errors.New("some random failure")
		if got := ClassifyError(err); got != ErrorTypeUnknown {
			t.Fatalf("expected %q, got %q", ErrorTypeUnknown, got)
		}
	})
}
