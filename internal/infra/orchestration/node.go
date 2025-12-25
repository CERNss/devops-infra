package orchestration

import (
	"devops-infra/internal/infra/executor"
)

type Node struct {
	Name string
	Exec executor.Executor
}
