package validate

import (
	"context"
	"errors"
	"testing"

	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/install/base/docker"
)

func TestInstallerChecksOfficialDockerReadiness(t *testing.T) {
	exec := &recordingExec{}
	driver := &fakeDriver{exec: exec, family: "debian"}
	installer := New(driver, Options{Mode: docker.InstallModeOfficial})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("validate install failed: %v", err)
	}
	if !contains(exec.commands, "command -v docker") {
		t.Fatalf("expected docker binary check, commands=%v", exec.commands)
	}
	if !contains(exec.commands, "systemctl is-active docker") {
		t.Fatalf("expected docker service check, commands=%v", exec.commands)
	}
}

func TestInstallerChecksNerdctlReadiness(t *testing.T) {
	exec := &recordingExec{}
	driver := &fakeDriver{exec: exec, family: "debian"}
	installer := New(driver, Options{Mode: docker.InstallModeNerdctl})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("validate install failed: %v", err)
	}
	if !contains(exec.commands, "command -v nerdctl") {
		t.Fatalf("expected nerdctl binary check, commands=%v", exec.commands)
	}
	if !contains(exec.commands, "test -L /usr/bin/docker") {
		t.Fatalf("expected docker symlink check, commands=%v", exec.commands)
	}
}

func TestInstallerReturnsValidationError(t *testing.T) {
	exec := &recordingExec{failCommand: "command -v curl", err: errors.New("not found")}
	driver := &fakeDriver{exec: exec, family: "debian"}
	installer := New(driver, Options{Mode: docker.InstallModeOfficial})

	err := installer.Install(context.Background())
	if err == nil {
		t.Fatal("expected validation error")
	}
}

type recordingExec struct {
	commands    []string
	failCommand string
	err         error
}

func (r *recordingExec) Run(string) error { return nil }

func (r *recordingExec) RunWithOutput(cmd string) (string, error) {
	r.commands = append(r.commands, cmd)
	if r.failCommand == cmd {
		return "", r.err
	}
	return "", nil
}

type fakeDriver struct {
	exec   executor.Executor
	family string
}

func (f *fakeDriver) Name() string                      { return "fake" }
func (f *fakeDriver) Family() string                    { return f.family }
func (f *fakeDriver) Exec() executor.Executor           { return f.exec }
func (f *fakeDriver) Update() error                     { return nil }
func (f *fakeDriver) InstallPackages(...string) error   { return nil }
func (f *fakeDriver) EnableService(string) error        { return nil }
func (f *fakeDriver) StartService(string) error         { return nil }
func (f *fakeDriver) RestartService(string) error       { return nil }
func (f *fakeDriver) LoadKernelModules(...string) error { return nil }
func (f *fakeDriver) Sysctl(map[string]string) error    { return nil }
func (f *fakeDriver) SwitchMirror() error               { return nil }

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
