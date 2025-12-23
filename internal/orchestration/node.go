package orchestration

import "devops-infra/internal/executor"

type Node struct {
	Name string
	Exec executor.Executor
}
