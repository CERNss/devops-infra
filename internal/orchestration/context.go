package orchestration

import (
	"devops-infra/internal/executor"
	osdriver "devops-infra/internal/os"
)

type ExecContext struct {
	OSInfo *osdriver.OSInfo
	Driver osdriver.Driver
	Exec   executor.Executor
}
