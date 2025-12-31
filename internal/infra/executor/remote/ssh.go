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

	"devops-infra/internal/constant"
	"devops-infra/internal/infra/executor"
	logmw "devops-infra/internal/middleware/log"
	tracemw "devops-infra/internal/middleware/trace"
	pathutil "devops-infra/internal/utils/path"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHExecutor struct {
	client  *ssh.Client
	opts    executor.Options
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
		port = constant.DefaultSSHPort
	}

	hostKeyCallback, err := hostKeyCallback(cfg)
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
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
		runtime: executor.NormalizeRuntime(runtime),
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
	traceID := logmw.NewTraceID()
	logCtx := logmw.WithTraceID(s.runtime.Ctx, traceID)
	logger := s.runtime.Logger
	if logger == nil {
		logger = logmw.NoopLogger()
	}

	if s.opts.DryRun {
		executor.PrintCommandStart(s.opts.Verbose, true, final)
		logger.Info(logCtx, fmt.Sprintf("exec dry-run: %s", final))
		s.traceCommand(final, traceID, time.Now(), "", "", nil, true)
		return "", nil
	}

	start := time.Now()
	executor.PrintCommandStart(s.opts.Verbose, false, final)
	logger.Info(logCtx, fmt.Sprintf("exec start: %s", final))
	sink, err := s.runtime.Output.Open(logmw.RuntimeInfo{
		Ctx:     s.runtime.Ctx,
		Logger:  logger,
		LogDir:  s.runtime.LogDir,
		TraceID: traceID,
	}, final)
	if err != nil {
		sink = logmw.NoopOutputSink()
	}
	defer sink.Close()

	session, err := s.client.NewSession()
	if err != nil {
		s.traceCommand(final, traceID, start, sink.StdoutPath(), sink.StderrPath(), err, false)
		logger.Error(logCtx, fmt.Sprintf("exec failed: %s: %v", final, err))
		executor.PrintCommandDone(s.opts.Verbose, start, final, err)
		return "", err
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	combinedBuf := &bytes.Buffer{}
	stdoutWriter := sink.Stdout()
	stderrWriter := sink.Stderr()
	if stdoutWriter == nil {
		stdoutWriter = io.Discard
	}
	if stderrWriter == nil {
		stderrWriter = io.Discard
	}

	if capture {
		session.Stdout = io.MultiWriter(combinedBuf, stdoutWriter)
		session.Stderr = io.MultiWriter(combinedBuf, stderrWriter)
		err = s.runWithContext(session, final)
		s.traceCommand(final, traceID, start, sink.StdoutPath(), sink.StderrPath(), err, false)
		if err != nil {
			logger.Error(logCtx, fmt.Sprintf("exec failed: %s: %v", final, err))
			executor.PrintCommandDone(s.opts.Verbose, start, final, err)
		} else {
			logger.Info(logCtx, fmt.Sprintf("exec done: %s", final))
			executor.PrintCommandDone(s.opts.Verbose, start, final, nil)
		}
		return combinedBuf.String(), err
	}

	if s.opts.Verbose {
		session.Stdout = io.MultiWriter(os.Stdout, stdoutWriter)
		session.Stderr = io.MultiWriter(os.Stderr, stderrWriter)
	} else {
		session.Stdout = stdoutWriter
		session.Stderr = stderrWriter
	}
	err = s.runWithContext(session, final)
	s.traceCommand(final, traceID, start, sink.StdoutPath(), sink.StderrPath(), err, false)
	if err != nil {
		logger.Error(logCtx, fmt.Sprintf("exec failed: %s: %v", final, err))
		executor.PrintCommandDone(s.opts.Verbose, start, final, err)
	} else {
		logger.Info(logCtx, fmt.Sprintf("exec done: %s", final))
		executor.PrintCommandDone(s.opts.Verbose, start, final, nil)
	}
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

func hostKeyCallback(cfg SSHConfig) (ssh.HostKeyCallback, error) {
	if cfg.InsecureIgnoreHostKey {
		return ssh.InsecureIgnoreHostKey(), nil
	}

	knownHostsPath := cfg.KnownHostsPath
	if knownHostsPath == "" {
		knownHostsPath = constant.DefaultKnownHostsPath
	}
	resolved, err := pathutil.ResolveUserPath(knownHostsPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(resolved); err != nil {
		return nil, fmt.Errorf("known_hosts not found: %w", err)
	}
	return knownhosts.New(resolved)
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
	traceID string,
	start time.Time,
	stdoutPath string,
	stderrPath string,
	err error,
	dryRun bool,
) {
	trace := s.runtime.Trace
	if trace == nil {
		return
	}

	end := time.Now()
	timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
	event := tracemw.NewTraceEvent(
		command,
		traceID,
		s.runtime.NodeName,
		s.runtime.NodeAddr,
		stdoutPath,
		stderrPath,
		start,
		end,
		"",
		"",
		err,
		dryRun,
		timedOut,
	)
	trace.OnCommand(event)
}
