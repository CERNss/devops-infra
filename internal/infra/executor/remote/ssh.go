package remote

import (
	"bytes"
	"devops-infra/internal/infra/executor"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	client *ssh.Client
	opts   executor.Options
}

func NewSSHExecutor(cfg SSHConfig, opts executor.Options) (*SSHExecutor, error) {
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

	return &SSHExecutor{client: client, opts: opts}, nil
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
		return "", nil
	}

	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {
			return
		}
	}(session)

	var buf bytes.Buffer
	if capture {
		session.Stdout = &buf
		session.Stderr = &buf
		err := session.Run(final)
		return buf.String(), err
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	return "", session.Run(final)
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
