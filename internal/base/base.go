// internal/base/base.go
package base

import "context"

type Component interface {
	Name() string
	IsInstalled(ctx context.Context) bool
	Install(ctx context.Context) error
}

type Installer struct {
	components []Component
}

func New(components ...Component) *Installer {
	return &Installer{components: components}
}

func (i *Installer) Install(ctx context.Context) error {
	for _, c := range i.components {
		if c.IsInstalled(ctx) {
			continue
		}
		if err := c.Install(ctx); err != nil {
			return err
		}
	}
	return nil
}
