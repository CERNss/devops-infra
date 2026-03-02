package k8s

import (
	"context"
	"errors"
	"testing"

	"devops-infra/internal/infra/executor"
)

type preflightRecordingExec struct {
	commands []string
}

func (r *preflightRecordingExec) Run(cmd string) error {
	r.commands = append(r.commands, cmd)
	return nil
}

func (r *preflightRecordingExec) RunWithOutput(cmd string) (string, error) {
	r.commands = append(r.commands, cmd)
	return "", errors.New("not used")
}

type preflightFakeDriver struct {
	exec   executor.Executor
	family string
}

func (f *preflightFakeDriver) Name() string                      { return "fake" }
func (f *preflightFakeDriver) Family() string                    { return f.family }
func (f *preflightFakeDriver) Exec() executor.Executor           { return f.exec }
func (f *preflightFakeDriver) Update() error                     { return nil }
func (f *preflightFakeDriver) InstallPackages(...string) error   { return nil }
func (f *preflightFakeDriver) EnableService(string) error        { return nil }
func (f *preflightFakeDriver) StartService(string) error         { return nil }
func (f *preflightFakeDriver) RestartService(string) error       { return nil }
func (f *preflightFakeDriver) LoadKernelModules(...string) error { return nil }
func (f *preflightFakeDriver) Sysctl(map[string]string) error    { return nil }
func (f *preflightFakeDriver) SwitchMirror() error               { return nil }

func TestPreflightInstallerIsInstalledDefaultsToFalseWhenSwapHandlingEnabled(t *testing.T) {
	exec := &preflightRecordingExec{}
	driver := &preflightFakeDriver{exec: exec, family: "rhel"}
	installer := NewPreflight(driver, PreflightOptions{
		EnsureSwapOff: true,
	})
	if installer.IsInstalled(context.Background()) {
		t.Fatal("expected preflight to run when swap handling is enabled")
	}
}

func TestPreflightInstallerInstallRHELIncludesSwapAndZRAMHandling(t *testing.T) {
	exec := &preflightRecordingExec{}
	driver := &preflightFakeDriver{exec: exec, family: "rhel"}
	installer := NewPreflight(driver, PreflightOptions{
		EnsureSwapOff:   true,
		DisableSELinux:  true,
		DisableFirewall: true,
	})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("preflight install failed: %v", err)
	}

	required := []string{
		"swapoff -a || true",
		`sed -ri '/^[^#].*[[:space:]]swap[[:space:]].*$/ s/^/#/' /etc/fstab 2>/dev/null || true`,
		"systemctl disable --now systemd-zram-setup@zram0.service 2>/dev/null || true",
		"systemctl mask systemd-zram-generator.service 2>/dev/null || true",
		"setenforce 0 || true",
		"systemctl stop firewalld 2>/dev/null || true",
	}
	for _, cmd := range required {
		if !contains(exec.commands, cmd) {
			t.Fatalf("expected command %q, got %v", cmd, exec.commands)
		}
	}
}

func TestPreflightInstallerInstallDebianIncludesSwapAndUFWHandling(t *testing.T) {
	exec := &preflightRecordingExec{}
	driver := &preflightFakeDriver{exec: exec, family: "debian"}
	installer := NewPreflight(driver, PreflightOptions{
		EnsureSwapOff:   true,
		DisableFirewall: true,
	})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("preflight install failed: %v", err)
	}

	required := []string{
		"swapoff -a || true",
		`sed -ri '/^[^#].*[[:space:]]swap[[:space:]].*$/ s/^/#/' /etc/fstab 2>/dev/null || true`,
		"if command -v apt-get >/dev/null 2>&1; then apt-get -y purge zram-config >/dev/null 2>&1 || true; fi",
		"systemctl stop ufw 2>/dev/null || true",
		"systemctl disable ufw 2>/dev/null || true",
	}
	for _, cmd := range required {
		if !contains(exec.commands, cmd) {
			t.Fatalf("expected command %q, got %v", cmd, exec.commands)
		}
	}
}
