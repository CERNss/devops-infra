package orchestration

import (
	"fmt"

	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/executor/remote"
)

type ExecutorFactory interface {
	Build(node Node) (executor.Executor, error)
}

type DefaultExecutorFactory struct{}

func (DefaultExecutorFactory) Build(node Node) (executor.Executor, error) {
	opts := executor.Options{}
	if node.ExecOpts != nil {
		opts = *node.ExecOpts
	}

	switch node.ExecutorType {
	case ExecutorLocal, "":
		return executor.NewLocal(opts), nil
	case ExecutorSSH:
		if node.SSH == nil {
			return nil, fmt.Errorf("ssh config is required for node %q", node.Name)
		}
		return remote.NewSSHExecutor(*node.SSH, opts)
	default:
		return nil, fmt.Errorf("unsupported executor type: %s", node.ExecutorType)
	}
}
