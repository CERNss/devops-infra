package executor

import (
	"errors"
	"strings"
	"testing"
)

type probeExec struct {
	lastCmd string
	output  string
	err     error
}

func (p *probeExec) Run(string) error { return nil }

func (p *probeExec) RunWithOutput(cmd string) (string, error) {
	p.lastCmd = cmd
	return p.output, p.err
}

func TestProbeSuccessTrue(t *testing.T) {
	exec := &probeExec{output: "__DEVOPS_INFRA_PROBE_TRUE__\n"}
	ok := ProbeSuccess(exec, "test -f /etc/hosts")
	if !ok {
		t.Fatal("expected probe to succeed")
	}
	if !strings.Contains(exec.lastCmd, "if { test -f /etc/hosts; } >/dev/null 2>&1") {
		t.Fatalf("unexpected wrapped probe command: %q", exec.lastCmd)
	}
}

func TestProbeSuccessFalseOnProbeMiss(t *testing.T) {
	exec := &probeExec{output: "__DEVOPS_INFRA_PROBE_FALSE__\n"}
	if ProbeSuccess(exec, "command -v kubeadm") {
		t.Fatal("expected probe miss")
	}
}

func TestProbeSuccessFalseOnError(t *testing.T) {
	exec := &probeExec{err: errors.New("exec error")}
	if ProbeSuccess(exec, "command -v kubeadm") {
		t.Fatal("expected probe failure on execution error")
	}
}
