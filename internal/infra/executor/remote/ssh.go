package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"devops-infra/internal/infra/executor"
	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	client *ssh.Client
	opts   executor.Options
	runtime executor.Runtime
}

func NewSSHExecutor(cfg SSHConfig, opts executor.Options) (*SSHExecutor, error) {
	return NewSSHExecutorWithRuntime(cfg, opts, executor.DefaultRuntime())
}

func NewSSHExecutorWithRuntime(cfg SSHConfig, opts executor.Options, runtime executor.Runtime) (*SSHExecutor, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("ssh host is required")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("ssh user is required")
	}

	auths, err := sshAuthMethods(cfg)
	if err != nil {
		return nil, err
	}
	if len(auths) == 0 {
		return nil, fmt.Errorf("no ssh auth method provided")
	}

	port := cfg.Port
	if port == 0 {
		port = 22
	}

	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}

	return &SSHExecutor{
		client:  client,
		opts:    opts,
		runtime: executor.NewRuntime(runtime.Ctx, runtime.Trace),
	}, nil
}

func (s *SSHExecutor) DryRun() bool {
	return s.opts.DryRun
}

// Run / RunWithOutput 是对外接口
func (s *SSHExecutor) Run(cmd string) error {
	_, err := s.run(cmd, false)
	return err
}

func (s *SSHExecutor) RunWithOutput(cmd string) (string, error) {
	return s.run(cmd, true)
}

// run 是 SSHExecutor 的私有方法
func (s *SSHExecutor) run(cmd string, capture bool) (string, error) {
	final := executor.Prepare(cmd, s.opts)

	if s.opts.Verbose || s.opts.DryRun {
		fmt.Printf("[ssh %s] %s\n", s.client.RemoteAddr(), final)
	}

	if s.opts.DryRun {
		s.traceCommand(final, time.Now(), "", "", nil, true)
		return "", nil
	}

	start := time.Now()
	session, err := s.client.NewSession()
	if err != nil {
		s.traceCommand(final, start, "", "", err, false)
		return "", err
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	combinedBuf := &bytes.Buffer{}
	if capture {
		session.Stdout = io.MultiWriter(combinedBuf, stdoutBuf)
		session.Stderr = io.MultiWriter(combinedBuf, stderrBuf)
		err := s.runWithContext(session, final)
		s.traceCommand(final, start, stdoutBuf.String(), stderrBuf.String(), err, false)
		return combinedBuf.String(), err
	}

	session.Stdout = io.MultiWriter(os.Stdout, combinedBuf, stdoutBuf)
	session.Stderr = io.MultiWriter(os.Stderr, combinedBuf, stderrBuf)
	err = s.runWithContext(session, final)
	s.traceCommand(final, start, stdoutBuf.String(), stderrBuf.String(), err, false)
	return "", err
}

func sshAuthMethods(cfg SSHConfig) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if cfg.KeyPath != "" {
		key, err := os.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}

	return methods, nil
}

func (s *SSHExecutor) runWithContext(session *ssh.Session, cmd string) error {
	ctx := s.runtime.Ctx
	if ctx == nil {
		return session.Run(cmd)
	}

	if err := session.Start(cmd); err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Wait()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		_ = session.Close()
		return ctx.Err()
	}
}

func (s *SSHExecutor) traceCommand(
	command string,
	start time.Time,
	stdout string,
	stderr string,
	err error,
	dryRun bool,
) {
	trace := s.runtime.Trace
	if trace == nil {
		return
	}

	end := time.Now()
	timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
	event := executor.NewTraceEvent(command, start, end, stdout, stderr, err, dryRun, timedOut)
	trace.OnCommand(event)
}
