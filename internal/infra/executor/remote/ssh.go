package remote

import (
	"bytes"
	executor2 "devops-infra/internal/infra/executor"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	client *ssh.Client
	opts   executor2.Options
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
	final := executor2.Prepare(cmd, s.opts)

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
