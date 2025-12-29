package orchestration

import "devops-infra/internal/infra/executor"

type Topology struct {
	Nodes []Node
}

func NewSingleNodeTopology(execOpts executor.Options) Topology {
	return Topology{
		Nodes: []Node{
			NewLocalNode("local", execOpts),
		},
	}
}
