package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	"devops-infra/internal/infra/executor"
)

func TestVerifyInstallerSkipInitValidatesApplicableChecksOnly(t *testing.T) {
	exec := &recordingExec{}
	driver := &fakeDriver{exec: exec, family: "debian"}
	installer := NewVerify(driver, VerifyOptions{
		SkipInit:        true,
		SetupKubeconfig: true,
		CNI:             "flannel",
		SkipCNI:         true,
	})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("verify install failed: %v", err)
	}
	if !contains(exec.commands, "test -z \"$(swapon --noheadings 2>/dev/null)\"") {
		t.Fatalf("expected swap check command, got commands=%v", exec.commands)
	}

	for _, cmd := range exec.commands {
		if strings.Contains(cmd, "/etc/kubernetes/admin.conf") {
			t.Fatalf("skip-init should not check admin.conf, got commands=%v", exec.commands)
		}
		if strings.Contains(cmd, "kube-flannel") {
			t.Fatalf("skip-init should not verify CNI, got commands=%v", exec.commands)
		}
	}
}

func TestVerifyInstallerChecksCNIWhenApplicable(t *testing.T) {
	exec := &recordingExec{}
	driver := &fakeDriver{exec: exec, family: "debian"}
	installer := NewVerify(driver, VerifyOptions{
		SkipInit:        false,
		SetupKubeconfig: true,
		CNI:             "calico",
		SkipCNI:         false,
	})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("verify install failed: %v", err)
	}

	if !contains(exec.commands, "KUBECONFIG=/etc/kubernetes/admin.conf kubectl get deployment -n tigera-operator tigera-operator >/dev/null 2>&1") {
		t.Fatalf("expected calico verification command, commands=%v", exec.commands)
	}
}

func TestInitInstallerBuildsDeterministicCommandOrder(t *testing.T) {
	exec := &recordingExec{}
	driver := &fakeDriver{exec: exec, family: "debian"}
	installer := NewInit(driver, InitOptions{
		Version:                   "1.28.15",
		CRISocket:                 "unix:///run/containerd/containerd.sock",
		ControlPlaneEndpoint:      "10.0.0.10:6443",
		APIServerAdvertiseAddress: "10.0.0.11",
		PodNetworkCIDR:            "10.244.0.0/16",
		ServiceCIDR:               "10.96.0.0/12",
		ServiceDNSDomain:          "cluster.local",
		ImageRepository:           "registry.k8s.io",
		Token:                     "abcdef.0123456789abcdef",
		TokenTTL:                  "24h",
		UploadCerts:               true,
		CertificateKey:            "0123456789abcdef0123456789abcdef",
		IgnorePreflightErrors:     "Swap",
		FeatureGates:              "SomeFeature=true",
		PatchesDir:                "/etc/kubeadm/patches",
	})

	if err := installer.Install(context.Background()); err != nil {
		t.Fatalf("init install failed: %v", err)
	}
	if len(exec.commands) != 1 {
		t.Fatalf("expected one command, got %d (%v)", len(exec.commands), exec.commands)
	}

	cmd := exec.commands[0]
	order := []string{
		"kubeadm init",
		"--kubernetes-version=\"1.28.15\"",
		"--cri-socket \"unix:///run/containerd/containerd.sock\"",
		"--control-plane-endpoint \"10.0.0.10:6443\"",
		"--apiserver-advertise-address \"10.0.0.11\"",
		"--pod-network-cidr \"10.244.0.0/16\"",
		"--service-cidr \"10.96.0.0/12\"",
		"--service-dns-domain \"cluster.local\"",
		"--image-repository \"registry.k8s.io\"",
		"--token \"abcdef.0123456789abcdef\"",
		"--token-ttl \"24h\"",
		"--upload-certs",
		"--certificate-key \"0123456789abcdef0123456789abcdef\"",
		"--ignore-preflight-errors \"Swap\"",
		"--feature-gates \"SomeFeature=true\"",
		"--patches \"/etc/kubeadm/patches\"",
	}
	assertSubsequence(t, cmd, order)
}

type recordingExec struct {
	commands    []string
	failCommand string
	err         error
}

func (r *recordingExec) Run(cmd string) error {
	r.commands = append(r.commands, cmd)
	if r.failCommand == cmd {
		if r.err != nil {
			return r.err
		}
		return errors.New("forced failure")
	}
	return nil
}

func (r *recordingExec) RunWithOutput(cmd string) (string, error) {
	r.commands = append(r.commands, cmd)
	if r.failCommand == cmd {
		if r.err != nil {
			return "", r.err
		}
		return "", errors.New("forced failure")
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

func assertSubsequence(t *testing.T, content string, parts []string) {
	t.Helper()
	cursor := 0
	for _, part := range parts {
		idx := strings.Index(content[cursor:], part)
		if idx < 0 {
			t.Fatalf("expected %q in command sequence after offset %d. command=%s", part, cursor, content)
		}
		cursor += idx + len(part)
	}
}
