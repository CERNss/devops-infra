package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type LocalExecutor struct {
	opts Options
}

func NewLocal(opts Options) *LocalExecutor {
	return &LocalExecutor{opts: opts}
}

func (e *LocalExecutor) Run(cmd string) error {
	_, err := e.run(cmd, false)
	return err
}

func (e *LocalExecutor) RunWithOutput(cmd string) (string, error) {
	return e.run(cmd, true)
}

func (e *LocalExecutor) run(cmd string, capture bool) (string, error) {
	finalCmd := e.prepare(cmd)

	if e.opts.Verbose || e.opts.DryRun {
		fmt.Printf("[exec] %s\n", finalCmd)
	}

	if e.opts.DryRun {
		return "", nil
	}

	c := exec.Command("bash", "-c", finalCmd)

	if capture {
		var buf bytes.Buffer
		c.Stdout = &buf
		c.Stderr = &buf
		err := c.Run()
		return buf.String(), err
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return "", c.Run()
}

func (e *LocalExecutor) prepare(cmd string) string {
	if e.opts.Sudo && !isRoot() {
		return "sudo -E bash -c " + shellQuote(cmd)
	}
	return cmd
}
