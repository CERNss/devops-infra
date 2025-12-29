package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type LocalExecutor struct {
	opts    Options
	runtime Runtime
}

func NewLocal(opts Options) *LocalExecutor {
	return NewLocalWithRuntime(opts, DefaultRuntime())
}

func NewLocalWithRuntime(opts Options, runtime Runtime) *LocalExecutor {
	runtime = normalizeRuntime(runtime)
	return &LocalExecutor{opts: opts, runtime: runtime}
}

func (e *LocalExecutor) DryRun() bool {
	return e.opts.DryRun
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
		e.traceCommand(finalCmd, time.Now(), "", "", nil, true)
		return "", nil
	}

	start := time.Now()
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	combinedBuf := &bytes.Buffer{}

	var stdoutWriter io.Writer
	var stderrWriter io.Writer
	if capture {
		stdoutWriter = io.MultiWriter(combinedBuf, stdoutBuf)
		stderrWriter = io.MultiWriter(combinedBuf, stderrBuf)
	} else {
		stdoutWriter = io.MultiWriter(os.Stdout, combinedBuf, stdoutBuf)
		stderrWriter = io.MultiWriter(os.Stderr, combinedBuf, stderrBuf)
	}

	c := exec.CommandContext(e.runtime.Ctx, "bash", "-c", finalCmd)
	c.Stdout = stdoutWriter
	c.Stderr = stderrWriter

	if capture {
		err := c.Run()
		e.traceCommand(finalCmd, start, stdoutBuf.String(), stderrBuf.String(), err, false)
		return combinedBuf.String(), err
	}

	err := c.Run()
	e.traceCommand(finalCmd, start, stdoutBuf.String(), stderrBuf.String(), err, false)
	return "", err
}

func (e *LocalExecutor) prepare(cmd string) string {
	if e.opts.Sudo && !isRoot() {
		return "sudo -E bash -c " + shellQuote(cmd)
	}
	return cmd
}

func (e *LocalExecutor) traceCommand(
	command string,
	start time.Time,
	stdout string,
	stderr string,
	err error,
	dryRun bool,
) {
	trace := e.runtime.Trace
	if trace == nil {
		return
	}

	end := time.Now()
	timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
	event := NewTraceEvent(command, start, end, stdout, stderr, err, dryRun, timedOut)
	trace.OnCommand(event)
}
