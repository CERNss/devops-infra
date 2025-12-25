package orchestration

import (
	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/os"
)

type ExecContext struct {
	OSInfo *os.OSInfo
	Driver os.Driver
	Exec   executor.Executor
}
