package executor

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

const (
	ErrorTypeExecTimeout   = "exec_timeout"
	ErrorTypeExecNonZero   = "exec_nonzero"
	ErrorTypeNetworkFetch  = "network_fetch"
	ErrorTypeUnsupportedOS = "unsupported_os"
	ErrorTypeValidation    = "validation_failed"
	ErrorTypeUnknown       = "unknown"
)

func ClassifyError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrorTypeExecTimeout
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return ErrorTypeExecNonZero
	}

	message := strings.ToLower(err.Error())

	if strings.Contains(message, "unsupported os") || strings.Contains(message, "unsupported os family") {
		return ErrorTypeUnsupportedOS
	}
	if strings.Contains(message, "dial tcp") ||
		strings.Contains(message, "no such host") ||
		strings.Contains(message, "network is unreachable") ||
		strings.Contains(message, "connection refused") ||
		strings.Contains(message, "i/o timeout") ||
		strings.Contains(message, "tls handshake timeout") ||
		strings.Contains(message, "temporary failure") {
		return ErrorTypeNetworkFetch
	}
	if strings.Contains(message, "checksum") ||
		strings.Contains(message, "validation") ||
		strings.Contains(message, "unsupported cni") {
		return ErrorTypeValidation
	}
	if strings.Contains(message, "exit status") {
		return ErrorTypeExecNonZero
	}

	return ErrorTypeUnknown
}
