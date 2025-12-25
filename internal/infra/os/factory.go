package os

import (
	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/os/debian"
	"devops-infra/internal/infra/os/rhel"
	"fmt"
)

func NewDriver(info *OSInfo, exec executor.Executor) (Driver, error) {
	switch info.Family {
	case "debian":
		return debian.New(exec), nil
	case "rhel":
		return rhel.New(exec), nil
	default:
		return nil, fmt.Errorf("unsupported os family: %s", info.Family)
	}
}
