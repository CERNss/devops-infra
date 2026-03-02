package containerd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"devops-infra/internal/infra/executor"
)

type containerdRecordingExec struct {
	commands      []string
	probeResults  []bool
	versionOutput string
	runErrs       map[string]error
	dryRun        bool
}

func (r *containerdRecordingExec) Run(cmd string) error {
	r.commands = append(r.commands, cmd)
	if err, ok := r.runErrs[cmd]; ok {
		return err
	}
	return nil
}

func (r *containerdRecordingExec) RunWithOutput(cmd string) (string, error) {
	r.commands = append(r.commands, cmd)
	if err, ok := r.runErrs[cmd]; ok {
		return "", err
	}
	if strings.Contains(cmd, "__DEVOPS_INFRA_PROBE_TRUE__") && strings.Contains(cmd, "__DEVOPS_INFRA_PROBE_FALSE__") {
		if len(r.probeResults) == 0 {
			return "__DEVOPS_INFRA_PROBE_FALSE__", nil
		}
		next := r.probeResults[0]
		r.probeResults = r.probeResults[1:]
		if next {
			return "__DEVOPS_INFRA_PROBE_TRUE__", nil
		}
		return "__DEVOPS_INFRA_PROBE_FALSE__", nil
	}
	if cmd == "containerd --version" {
		return r.versionOutput, nil
	}
	return "", nil
}

func (r *containerdRecordingExec) DryRun() bool { return r.dryRun }

type containerdFakeDriver struct {
	exec executor.Executor
}

func (f *containerdFakeDriver) Name() string                      { return "fake" }
func (f *containerdFakeDriver) Family() string                    { return "rhel" }
func (f *containerdFakeDriver) Exec() executor.Executor           { return f.exec }
func (f *containerdFakeDriver) Update() error                     { return nil }
func (f *containerdFakeDriver) InstallPackages(...string) error   { return nil }
func (f *containerdFakeDriver) EnableService(string) error        { return nil }
func (f *containerdFakeDriver) StartService(string) error         { return nil }
func (f *containerdFakeDriver) RestartService(string) error       { return nil }
func (f *containerdFakeDriver) LoadKernelModules(...string) error { return nil }
func (f *containerdFakeDriver) Sysctl(map[string]string) error    { return nil }
func (f *containerdFakeDriver) SwitchMirror() error               { return nil }

func TestIsInstalledRequiresActiveService(t *testing.T) {
	exec := &containerdRecordingExec{
		probeResults:  []bool{true, false},
		versionOutput: "containerd github.com/containerd/containerd 2.1.0",
	}
	driver := &containerdFakeDriver{exec: exec}
	installer := New(driver, Options{Version: "2.1.0"})

	if installer.IsInstalled(context.Background()) {
		t.Fatal("expected inactive containerd service to be treated as not installed")
	}
}

func TestInstallRestartRecoveredWithoutReinstall(t *testing.T) {
	exec := &containerdRecordingExec{probeResults: []bool{false, true}}
	driver := &containerdFakeDriver{exec: exec}
	installer := New(driver, Options{Version: "2.1.0"})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if !containsExact(exec.commands, "systemctl restart containerd") {
		t.Fatalf("expected restart command, got %v", exec.commands)
	}
	if containsSubstr(exec.commands, "/tmp/containerd.tar.gz") {
		t.Fatalf("did not expect reinstall path, commands=%v", exec.commands)
	}
	if !containsSubstr(exec.commands, "event=restart_recovered") {
		t.Fatalf("expected restart recovered marker, commands=%v", exec.commands)
	}
}

func TestInstallEscalatesToReinstallWhenRestartNotEnough(t *testing.T) {
	exec := &containerdRecordingExec{probeResults: []bool{false, false, true}}
	driver := &containerdFakeDriver{exec: exec}
	installer := New(driver, Options{Version: "2.1.0"})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if !containsSubstr(exec.commands, "curl -L -o /tmp/containerd.tar.gz") {
		t.Fatalf("expected reinstall command, got %v", exec.commands)
	}
	if !containsSubstr(exec.commands, "event=reinstall_recovered") {
		t.Fatalf("expected reinstall recovered marker, commands=%v", exec.commands)
	}
}

func TestInstallDryRunReturnsPlannedOutcome(t *testing.T) {
	exec := &containerdRecordingExec{
		dryRun: true,
		runErrs: map[string]error{
			"systemctl restart containerd": errors.New("should not run in dry-run"),
		},
	}
	driver := &containerdFakeDriver{exec: exec}
	installer := New(driver, Options{Version: "2.1.0"})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install dry-run should not fail: %v", err)
	}
	if !containsSubstr(exec.commands, "dry_run_plan_probe_restart_reinstall") {
		t.Fatalf("expected dry-run plan marker, commands=%v", exec.commands)
	}
}

func containsExact(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsSubstr(values []string, target string) bool {
	for _, value := range values {
		if strings.Contains(value, target) {
			return true
		}
	}
	return false
}
