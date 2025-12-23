package os

import (
	"fmt"

	"devops-infra/internal/os/debian"
	"devops-infra/internal/os/rhel"
)

func NewDriver(info *OSInfo) (Driver, error) {
	switch info.Family {
	case "debian":
		return debian.New(), nil
	case "rhel":
		return rhel.New(), nil
	default:
		return nil, fmt.Errorf("unsupported os family: %s", info.Family)
	}
}
