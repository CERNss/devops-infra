package topology

import (
	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/executor/remote"
)

type ExecutorType string

const (
	ExecutorLocal ExecutorType = "local"
	ExecutorSSH   ExecutorType = "ssh"
)

type Node struct {
	Name         string
	ExecutorType ExecutorType
	ExecOpts     *executor.Options
	SSH          *remote.SSHConfig
}

func NewLocalNode(name string, opts executor.Options) Node {
	return Node{
		Name:         name,
		ExecutorType: ExecutorLocal,
		ExecOpts:     &opts,
	}
}
