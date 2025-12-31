package execfactory

import (
	"fmt"

	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/executor/remote"
	"devops-infra/internal/infra/orchestration/topology"
)

type ExecutorFactory interface {
	Build(node topology.Node, runtime executor.Runtime) (executor.Executor, error)
}

type DefaultExecutorFactory struct{}

func (DefaultExecutorFactory) Build(node topology.Node, runtime executor.Runtime) (executor.Executor, error) {
	runtime = executor.NormalizeRuntime(runtime)
	opts := executor.Options{}
	if node.ExecOpts != nil {
		opts = *node.ExecOpts
	}

	addr := ""
	if node.SSH != nil && node.SSH.Host != "" {
		addr = node.SSH.Host
	}
	name := node.Name
	if name == "" {
		if addr != "" {
			name = addr
		} else {
			name = "local"
		}
	}
	runtime = executor.WithNode(runtime, name, addr)

	switch node.ExecutorType {
	case topology.ExecutorLocal, "":
		return executor.NewLocalWithRuntime(opts, runtime), nil
	case topology.ExecutorSSH:
		if node.SSH == nil {
			return nil, fmt.Errorf("ssh config is required for node %q", node.Name)
		}
		return remote.NewSSHExecutorWithRuntime(*node.SSH, opts, runtime)
	default:
		return nil, fmt.Errorf("unsupported executor type: %s", node.ExecutorType)
	}
}
