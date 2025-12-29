package executor

type Executor interface {
	Run(cmd string) error
	RunWithOutput(cmd string) (string, error)
}

type DryRunner interface {
	DryRun() bool
}

func IsDryRun(exec Executor) bool {
	if exec == nil {
		return false
	}
	if dr, ok := exec.(DryRunner); ok {
		return dr.DryRun()
	}
	return false
}

type Options struct {
	Sudo    bool
	DryRun  bool
	Verbose bool
}
