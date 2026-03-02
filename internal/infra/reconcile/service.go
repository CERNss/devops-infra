package reconcile

import (
	"fmt"
	"strings"

	"devops-infra/internal/infra/executor"
)

type Outcome string

const (
	OutcomeDryRunPlanned    Outcome = "dry_run_planned"
	OutcomeSkipHealthy      Outcome = "skip_healthy"
	OutcomeRestartRecovered Outcome = "restart_recovered"
	OutcomeReinstall        Outcome = "reinstall_recovered"
)

type ServiceOptions struct {
	Exec           executor.Executor
	Label          string
	HealthCheckCmd string
	RestartCmd     string
	Reinstall      func() error
}

// EnsureServiceHealthy applies bounded reconcile strategy:
// healthy -> skip, unhealthy -> restart, still unhealthy -> reinstall, still unhealthy -> fail.
func EnsureServiceHealthy(opts ServiceOptions) (Outcome, error) {
	label := normalizeLabel(opts.Label)
	if opts.Exec == nil {
		return "", fmt.Errorf("service reconcile(%s): nil executor", label)
	}
	healthCmd := strings.TrimSpace(opts.HealthCheckCmd)
	if healthCmd == "" {
		return "", fmt.Errorf("service reconcile(%s): empty health check command", label)
	}

	if executor.IsDryRun(opts.Exec) {
		emitInfo(opts.Exec, label, "dry_run_plan_probe_restart_reinstall")
		return OutcomeDryRunPlanned, nil
	}

	if executor.ProbeSuccess(opts.Exec, healthCmd) {
		emitInfo(opts.Exec, label, "skip_healthy")
		return OutcomeSkipHealthy, nil
	}

	emitInfo(opts.Exec, label, "restart_attempt")
	restartCmd := strings.TrimSpace(opts.RestartCmd)
	if restartCmd != "" {
		if err := opts.Exec.Run(restartCmd); err != nil {
			emitInfo(opts.Exec, label, "restart_error")
		}
	}
	if executor.ProbeSuccess(opts.Exec, healthCmd) {
		emitInfo(opts.Exec, label, "restart_recovered")
		return OutcomeRestartRecovered, nil
	}

	emitInfo(opts.Exec, label, "reinstall_attempt")
	if opts.Reinstall == nil {
		emitFailure(opts.Exec, label, "reinstall_unavailable")
		return "", fmt.Errorf("service reconcile(%s): unhealthy after restart and reinstall is unavailable", label)
	}
	if err := opts.Reinstall(); err != nil {
		emitFailure(opts.Exec, label, "reinstall_failed")
		return "", fmt.Errorf("service reconcile(%s): reinstall failed: %w", label, err)
	}
	if !executor.ProbeSuccess(opts.Exec, healthCmd) {
		emitFailure(opts.Exec, label, "unhealthy_after_reinstall")
		return "", fmt.Errorf("service reconcile(%s): unhealthy after reinstall", label)
	}
	emitInfo(opts.Exec, label, "reinstall_recovered")
	return OutcomeReinstall, nil
}

func normalizeLabel(label string) string {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}

func sanitize(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return strings.Join(strings.Fields(value), "_")
}

func emitInfo(exec executor.Executor, label string, event string) {
	if exec == nil {
		return
	}
	message := fmt.Sprintf("DEVOPS_INFRA_RECONCILE service=%s event=%s", sanitize(label), sanitize(event))
	_ = exec.Run(fmt.Sprintf("printf '%%s\\n' %q", message))
}

func emitFailure(exec executor.Executor, label string, stage string) {
	if exec == nil {
		return
	}
	message := fmt.Sprintf("DEVOPS_INFRA_RECONCILE service=%s event=failed stage=%s", sanitize(label), sanitize(stage))
	_ = exec.Run(fmt.Sprintf("printf '%%s\\n' %q >&2; exit 1", message))
}
