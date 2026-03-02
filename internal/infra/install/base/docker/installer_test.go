package docker

import (
	"context"
	"strings"
	"testing"

	"devops-infra/internal/infra/executor"
)

type dockerRecordingExec struct {
	commands     []string
	probeResults []bool
	dryRun       bool
}

func (r *dockerRecordingExec) Run(cmd string) error {
	r.commands = append(r.commands, cmd)
	return nil
}

func (r *dockerRecordingExec) RunWithOutput(cmd string) (string, error) {
	r.commands = append(r.commands, cmd)
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
	return "", nil
}

func (r *dockerRecordingExec) DryRun() bool { return r.dryRun }

type dockerFakeDriver struct {
	exec executor.Executor
}

func (f *dockerFakeDriver) Name() string                      { return "fake" }
func (f *dockerFakeDriver) Family() string                    { return "debian" }
func (f *dockerFakeDriver) Exec() executor.Executor           { return f.exec }
func (f *dockerFakeDriver) Update() error                     { return nil }
func (f *dockerFakeDriver) InstallPackages(...string) error   { return nil }
func (f *dockerFakeDriver) EnableService(string) error        { return nil }
func (f *dockerFakeDriver) StartService(string) error         { return nil }
func (f *dockerFakeDriver) RestartService(string) error       { return nil }
func (f *dockerFakeDriver) LoadKernelModules(...string) error { return nil }
func (f *dockerFakeDriver) Sysctl(map[string]string) error    { return nil }
func (f *dockerFakeDriver) SwitchMirror() error               { return nil }

func TestIsInstalledOfficialRequiresActiveService(t *testing.T) {
	exec := &dockerRecordingExec{probeResults: []bool{true, false}}
	driver := &dockerFakeDriver{exec: exec}
	installer := New(driver, Options{Mode: InstallModeOfficial})

	if installer.IsInstalled(context.Background()) {
		t.Fatal("expected inactive docker service to be treated as not installed")
	}
}

func TestInstallOfficialSkipsWhenHealthy(t *testing.T) {
	exec := &dockerRecordingExec{probeResults: []bool{true}}
	driver := &dockerFakeDriver{exec: exec}
	installer := New(driver, Options{Mode: InstallModeOfficial})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if containsExact(exec.commands, "systemctl restart docker") {
		t.Fatalf("did not expect restart command, commands=%v", exec.commands)
	}
	if !containsSubstr(exec.commands, "event=skip_healthy") {
		t.Fatalf("expected skip_healthy marker, commands=%v", exec.commands)
	}
}

func TestInstallOfficialRestartsWhenUnhealthy(t *testing.T) {
	exec := &dockerRecordingExec{probeResults: []bool{false, true}}
	driver := &dockerFakeDriver{exec: exec}
	installer := New(driver, Options{Mode: InstallModeOfficial})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if !containsExact(exec.commands, "systemctl restart docker") {
		t.Fatalf("expected restart command, commands=%v", exec.commands)
	}
	if !containsSubstr(exec.commands, "event=restart_recovered") {
		t.Fatalf("expected restart recovered marker, commands=%v", exec.commands)
	}
}

func TestInstallOfficialDryRunShowsPlan(t *testing.T) {
	exec := &dockerRecordingExec{dryRun: true}
	driver := &dockerFakeDriver{exec: exec}
	installer := New(driver, Options{Mode: InstallModeOfficial})

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
