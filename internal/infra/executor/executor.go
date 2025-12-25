package executor

type Executor interface {
	Run(cmd string) error
	RunWithOutput(cmd string) (string, error)
}

type Options struct {
	Sudo    bool
	DryRun  bool
	Verbose bool
}
