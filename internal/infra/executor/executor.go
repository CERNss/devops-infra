package executor

import (
	"fmt"
	"strings"
)

type Executor interface {
	Run(cmd string) error
	RunWithOutput(cmd string) (string, error)
}

type DryRunner interface {
	DryRun() bool
}

func IsDryRun(exec Executor) bool {
	if exec == nil {
		return false
	}
	if dr, ok := exec.(DryRunner); ok {
		return dr.DryRun()
	}
	return false
}

// ProbeSuccess runs a command as a probe and returns whether it succeeded.
// The wrapper always exits 0, so expected probe misses do not surface as failed commands.
func ProbeSuccess(exec Executor, cmd string) bool {
	if exec == nil {
		return false
	}
	const probeTrue = "__DEVOPS_INFRA_PROBE_TRUE__"
	const probeFalse = "__DEVOPS_INFRA_PROBE_FALSE__"
	wrapped := fmt.Sprintf(
		"if { %s; } >/dev/null 2>&1; then echo %s; else echo %s; fi",
		strings.TrimSpace(cmd),
		probeTrue,
		probeFalse,
	)
	out, err := exec.RunWithOutput(wrapped)
	if err != nil {
		return false
	}
	return strings.Contains(out, probeTrue)
}

type Options struct {
	Sudo    bool
	DryRun  bool
	Verbose bool
}
