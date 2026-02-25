package rhel

import "testing"

func TestSysctlRunsInDeterministicOrder(t *testing.T) {
	exec := &recordingExecutor{}
	driver := New(exec)

	settings := map[string]string{
		"net.ipv4.ip_forward":                 "1",
		"net.bridge.bridge-nf-call-iptables":  "1",
		"net.bridge.bridge-nf-call-ip6tables": "1",
	}
	if err := driver.Sysctl(settings); err != nil {
		t.Fatalf("sysctl failed: %v", err)
	}

	want := []string{
		"sysctl -w net.bridge.bridge-nf-call-ip6tables=1",
		"sysctl -w net.bridge.bridge-nf-call-iptables=1",
		"sysctl -w net.ipv4.ip_forward=1",
		"sysctl --system",
	}
	assertCommands(t, exec.commands, want)
}

type recordingExecutor struct {
	commands []string
}

func (r *recordingExecutor) Run(cmd string) error {
	r.commands = append(r.commands, cmd)
	return nil
}

func (r *recordingExecutor) RunWithOutput(cmd string) (string, error) {
	r.commands = append(r.commands, cmd)
	return "", nil
}

func assertCommands(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected command count: got=%d want=%d, commands=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("command[%d]=%q, want %q", i, got[i], want[i])
		}
	}
}
