package base

import (
	"context"
	"errors"
	"testing"
)

func TestInstallerSkipsInstalledComponents(t *testing.T) {
	installed := &fakeComponent{name: "already", installed: true}
	pending := &fakeComponent{name: "pending", installed: false}

	installer := New(installed, pending)
	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	if installed.installCalls != 0 {
		t.Fatalf("installed component should be skipped, got install calls %d", installed.installCalls)
	}
	if pending.installCalls != 1 {
		t.Fatalf("pending component should be installed once, got %d", pending.installCalls)
	}
}

func TestInstallerStopsOnFailure(t *testing.T) {
	first := &fakeComponent{name: "first", installed: false, installErr: errors.New("boom")}
	second := &fakeComponent{name: "second", installed: false}

	installer := New(first, second)
	err := installer.Install(context.Background())
	if err == nil {
		t.Fatal("expected installer to return error")
	}

	if first.installCalls != 1 {
		t.Fatalf("first component install calls = %d, want 1", first.installCalls)
	}
	if second.installCalls != 0 {
		t.Fatalf("second component should not run after failure, got %d", second.installCalls)
	}
}

func TestInstallerRerunSkipsSatisfiedComponents(t *testing.T) {
	component := &statefulComponent{name: "stateful"}
	installer := New(component)

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	if component.installCalls != 1 {
		t.Fatalf("expected exactly one install call across reruns, got %d", component.installCalls)
	}
}

type fakeComponent struct {
	name         string
	installed    bool
	installErr   error
	installCalls int
}

func (f *fakeComponent) Name() string { return f.name }

func (f *fakeComponent) IsInstalled(context.Context) bool { return f.installed }

func (f *fakeComponent) Install(context.Context) error {
	f.installCalls++
	return f.installErr
}

type statefulComponent struct {
	name         string
	installCalls int
	installed    bool
}

func (s *statefulComponent) Name() string { return s.name }

func (s *statefulComponent) IsInstalled(context.Context) bool { return s.installed }

func (s *statefulComponent) Install(context.Context) error {
	s.installCalls++
	s.installed = true
	return nil
}
