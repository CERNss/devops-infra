package os

import (
	"devops-infra/internal/executor"
	"fmt"

	"devops-infra/internal/os/debian"
	"devops-infra/internal/os/rhel"
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
