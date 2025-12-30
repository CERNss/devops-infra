// internal/base/base.go
package base

import (
	"context"
	"fmt"

	logmw "devops-infra/internal/infra/middleware/log"
)

type Component interface {
	Name() string
	IsInstalled(ctx context.Context) bool
	Install(ctx context.Context) error
}

type Installer struct {
	components []Component
	logger     logmw.Logger
}

func New(components ...Component) *Installer {
	return &Installer{components: components}
}

func (i *Installer) WithLogger(logger logmw.Logger) *Installer {
	i.logger = logger
	return i
}

func (i *Installer) Install(ctx context.Context) error {
	logger := i.logger
	if logger == nil {
		logger = logmw.NoopLogger()
	}
	for _, c := range i.components {
		if c.IsInstalled(ctx) {
			logger.Info(fmt.Sprintf("component %s: already installed", c.Name()))
			continue
		}
		logger.Info(fmt.Sprintf("component %s: installing", c.Name()))
		if err := c.Install(ctx); err != nil {
			logger.Error(fmt.Sprintf("component %s: install failed: %v", c.Name(), err))
			return err
		}
		logger.Info(fmt.Sprintf("component %s: installed", c.Name()))
	}
	return nil
}
