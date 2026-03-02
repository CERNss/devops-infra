package k8s

import (
	"context"
	"strings"
	"testing"

	"devops-infra/internal/infra/executor"
)

type packagesRecordingExec struct {
	commands     []string
	probeResults []bool
	dryRun       bool
}

func (r *packagesRecordingExec) Run(cmd string) error {
	r.commands = append(r.commands, cmd)
	return nil
}

func (r *packagesRecordingExec) RunWithOutput(cmd string) (string, error) {
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

func (r *packagesRecordingExec) DryRun() bool { return r.dryRun }

type packagesFakeDriver struct {
	exec           executor.Executor
	family         string
	updateCount    int
	installedPkgs  []string
	enabledService []string
}

func (f *packagesFakeDriver) Name() string            { return "fake" }
func (f *packagesFakeDriver) Family() string          { return f.family }
func (f *packagesFakeDriver) Exec() executor.Executor { return f.exec }
func (f *packagesFakeDriver) Update() error {
	f.updateCount++
	return nil
}
func (f *packagesFakeDriver) InstallPackages(pkgs ...string) error {
	f.installedPkgs = append(f.installedPkgs, pkgs...)
	return nil
}
func (f *packagesFakeDriver) EnableService(name string) error {
	f.enabledService = append(f.enabledService, name)
	return nil
}
func (f *packagesFakeDriver) StartService(string) error         { return nil }
func (f *packagesFakeDriver) RestartService(string) error       { return nil }
func (f *packagesFakeDriver) LoadKernelModules(...string) error { return nil }
func (f *packagesFakeDriver) Sysctl(map[string]string) error    { return nil }
func (f *packagesFakeDriver) SwitchMirror() error               { return nil }

func TestPackagesInstallerIsInstalledRequiresKubeletActive(t *testing.T) {
	exec := &packagesRecordingExec{probeResults: []bool{true, true, true, false}}
	driver := &packagesFakeDriver{exec: exec, family: "rhel"}
	installer := NewPackages(driver, PackagesOptions{Version: "1.28.15"})

	if installer.IsInstalled(context.Background()) {
		t.Fatal("expected inactive kubelet service to be treated as not installed")
	}
}

func TestPackagesInstallerInstallRestartRecovered(t *testing.T) {
	exec := &packagesRecordingExec{probeResults: []bool{false, true}}
	driver := &packagesFakeDriver{exec: exec, family: "debian"}
	installer := NewPackages(driver, PackagesOptions{})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if !containsExact(exec.commands, "systemctl restart kubelet") {
		t.Fatalf("expected kubelet restart, commands=%v", exec.commands)
	}
	if driver.updateCount != 0 || len(driver.installedPkgs) > 0 {
		t.Fatalf("did not expect reinstall path, update=%d pkgs=%v", driver.updateCount, driver.installedPkgs)
	}
	if !containsSubstr(exec.commands, "event=restart_recovered") {
		t.Fatalf("expected restart recovered marker, commands=%v", exec.commands)
	}
}

func TestPackagesInstallerInstallEscalatesToReinstall(t *testing.T) {
	exec := &packagesRecordingExec{probeResults: []bool{false, false, true}}
	driver := &packagesFakeDriver{exec: exec, family: "rhel"}
	installer := NewPackages(driver, PackagesOptions{Version: "1.28.15"})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if driver.updateCount != 1 {
		t.Fatalf("expected update once, got %d", driver.updateCount)
	}
	required := []string{"kubelet-1.28.15", "kubeadm-1.28.15", "kubectl-1.28.15"}
	for _, pkg := range required {
		if !containsExact(driver.installedPkgs, pkg) {
			t.Fatalf("expected package %q, got %v", pkg, driver.installedPkgs)
		}
	}
	if !containsExact(driver.enabledService, "kubelet") {
		t.Fatalf("expected kubelet service enable, got %v", driver.enabledService)
	}
	if !containsSubstr(exec.commands, "event=reinstall_recovered") {
		t.Fatalf("expected reinstall recovered marker, commands=%v", exec.commands)
	}
}

func TestPackagesInstallerInstallDryRunShowsPlan(t *testing.T) {
	exec := &packagesRecordingExec{dryRun: true}
	driver := &packagesFakeDriver{exec: exec, family: "debian"}
	installer := NewPackages(driver, PackagesOptions{})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("dry-run install should not fail: %v", err)
	}
	if driver.updateCount != 0 || len(driver.installedPkgs) > 0 {
		t.Fatalf("did not expect package changes in dry-run: update=%d pkgs=%v", driver.updateCount, driver.installedPkgs)
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
