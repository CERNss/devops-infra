package reconcile

import (
	"errors"
	"strings"
	"testing"
)

type fakeExec struct {
	commands     []string
	probeResults []bool
	runErrs      map[string]error
	dryRun       bool
}

func (f *fakeExec) Run(cmd string) error {
	f.commands = append(f.commands, cmd)
	if err, ok := f.runErrs[cmd]; ok {
		return err
	}
	return nil
}

func (f *fakeExec) RunWithOutput(cmd string) (string, error) {
	f.commands = append(f.commands, cmd)
	if strings.Contains(cmd, "__DEVOPS_INFRA_PROBE_TRUE__") && strings.Contains(cmd, "__DEVOPS_INFRA_PROBE_FALSE__") {
		if len(f.probeResults) == 0 {
			return "__DEVOPS_INFRA_PROBE_FALSE__", nil
		}
		next := f.probeResults[0]
		f.probeResults = f.probeResults[1:]
		if next {
			return "__DEVOPS_INFRA_PROBE_TRUE__", nil
		}
		return "__DEVOPS_INFRA_PROBE_FALSE__", nil
	}
	return "", nil
}

func (f *fakeExec) DryRun() bool { return f.dryRun }

func TestEnsureServiceHealthySkipHealthy(t *testing.T) {
	exec := &fakeExec{probeResults: []bool{true}}
	outcome, err := EnsureServiceHealthy(ServiceOptions{
		Exec:           exec,
		Label:          "svc",
		HealthCheckCmd: "systemctl is-active svc",
		RestartCmd:     "systemctl restart svc",
		Reinstall: func() error {
			t.Fatal("reinstall should not be called")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeSkipHealthy {
		t.Fatalf("unexpected outcome: %s", outcome)
	}
	if !containsAny(exec.commands, "event=skip_healthy") {
		t.Fatalf("expected skip_healthy event, commands=%v", exec.commands)
	}
}

func TestEnsureServiceHealthyRestartRecovered(t *testing.T) {
	exec := &fakeExec{probeResults: []bool{false, true}}
	reinstallCalled := false
	outcome, err := EnsureServiceHealthy(ServiceOptions{
		Exec:           exec,
		Label:          "svc",
		HealthCheckCmd: "systemctl is-active svc",
		RestartCmd:     "systemctl restart svc",
		Reinstall: func() error {
			reinstallCalled = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reinstallCalled {
		t.Fatal("reinstall should not be called")
	}
	if outcome != OutcomeRestartRecovered {
		t.Fatalf("unexpected outcome: %s", outcome)
	}
	if !containsAny(exec.commands, "event=restart_recovered") {
		t.Fatalf("expected restart_recovered event, commands=%v", exec.commands)
	}
}

func TestEnsureServiceHealthyReinstallRecovered(t *testing.T) {
	exec := &fakeExec{probeResults: []bool{false, false, true}}
	reinstallCalled := 0
	outcome, err := EnsureServiceHealthy(ServiceOptions{
		Exec:           exec,
		Label:          "svc",
		HealthCheckCmd: "systemctl is-active svc",
		RestartCmd:     "systemctl restart svc",
		Reinstall: func() error {
			reinstallCalled++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reinstallCalled != 1 {
		t.Fatalf("expected reinstall called once, got %d", reinstallCalled)
	}
	if outcome != OutcomeReinstall {
		t.Fatalf("unexpected outcome: %s", outcome)
	}
	if !containsAny(exec.commands, "event=reinstall_recovered") {
		t.Fatalf("expected reinstall_recovered event, commands=%v", exec.commands)
	}
}

func TestEnsureServiceHealthyUnrecoverable(t *testing.T) {
	exec := &fakeExec{probeResults: []bool{false, false, false}}
	_, err := EnsureServiceHealthy(ServiceOptions{
		Exec:           exec,
		Label:          "svc",
		HealthCheckCmd: "systemctl is-active svc",
		RestartCmd:     "systemctl restart svc",
		Reinstall: func() error {
			return nil
		},
	})
	if err == nil {
		t.Fatal("expected unrecoverable error")
	}
	if !containsAny(exec.commands, "event=failed stage=unhealthy_after_reinstall") {
		t.Fatalf("expected failed stage marker, commands=%v", exec.commands)
	}
}

func TestEnsureServiceHealthyReinstallError(t *testing.T) {
	exec := &fakeExec{probeResults: []bool{false, false}}
	_, err := EnsureServiceHealthy(ServiceOptions{
		Exec:           exec,
		Label:          "svc",
		HealthCheckCmd: "systemctl is-active svc",
		RestartCmd:     "systemctl restart svc",
		Reinstall: func() error {
			return errors.New("boom")
		},
	})
	if err == nil || !strings.Contains(err.Error(), "reinstall failed") {
		t.Fatalf("expected reinstall failed error, got %v", err)
	}
	if !containsAny(exec.commands, "event=failed stage=reinstall_failed") {
		t.Fatalf("expected reinstall_failed stage marker, commands=%v", exec.commands)
	}
}

func TestEnsureServiceHealthyDryRunPlanned(t *testing.T) {
	exec := &fakeExec{dryRun: true}
	reinstallCalled := false
	outcome, err := EnsureServiceHealthy(ServiceOptions{
		Exec:           exec,
		Label:          "svc",
		HealthCheckCmd: "systemctl is-active svc",
		RestartCmd:     "systemctl restart svc",
		Reinstall: func() error {
			reinstallCalled = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeDryRunPlanned {
		t.Fatalf("unexpected outcome: %s", outcome)
	}
	if reinstallCalled {
		t.Fatal("reinstall should not be called in dry-run mode")
	}
	if !containsAny(exec.commands, "dry_run_plan_probe_restart_reinstall") {
		t.Fatalf("expected dry-run plan marker, commands=%v", exec.commands)
	}
}

func containsAny(values []string, target string) bool {
	for _, value := range values {
		if strings.Contains(value, target) {
			return true
		}
	}
	return false
}
