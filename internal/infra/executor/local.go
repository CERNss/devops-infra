package executor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"

	logmw "devops-infra/internal/middleware/log"
	tracemw "devops-infra/internal/middleware/trace"
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
	traceID := logmw.NewTraceID()
	logCtx := logmw.WithTraceID(e.runtime.Ctx, traceID)
	logCtx = logmw.WithFields(logCtx, map[string]any{
		"command":   finalCmd,
		"node":      e.runtime.NodeName,
		"node_addr": e.runtime.NodeAddr,
		"component": e.runtime.Component,
	})
	logger := e.runtime.Logger
	if logger == nil {
		logger = logmw.NoopLogger()
	}

	if e.opts.DryRun {
		PrintCommandStart(e.opts.Verbose, true, finalCmd)
		logger.Info(logmw.WithFields(logCtx, map[string]any{
			"event":       "command_dry_run",
			"result":      "dry_run",
			"duration_ms": int64(0),
		}), "exec dry-run")
		e.traceCommand(finalCmd, traceID, time.Now(), "", "", nil, true)
		return "", nil
	}

	start := time.Now()
	PrintCommandStart(e.opts.Verbose, false, finalCmd)
	logger.Info(logmw.WithFields(logCtx, map[string]any{
		"event":  "command_start",
		"result": "running",
	}), "exec start")
	sink, err := e.runtime.Output.Open(logmw.RuntimeInfo{
		Ctx:     e.runtime.Ctx,
		Logger:  logger,
		LogDir:  e.runtime.LogDir,
		TraceID: traceID,
	}, finalCmd)
	if err != nil {
		sink = logmw.NoopOutputSink()
	}
	defer sink.Close()

	combinedBuf := &bytes.Buffer{}
	stdoutWriter := sink.Stdout()
	stderrWriter := sink.Stderr()
	if stdoutWriter == nil {
		stdoutWriter = io.Discard
	}
	if stderrWriter == nil {
		stderrWriter = io.Discard
	}
	if capture {
		stdoutWriter = io.MultiWriter(stdoutWriter, combinedBuf)
		stderrWriter = io.MultiWriter(stderrWriter, combinedBuf)
	} else if e.opts.Verbose {
		stdoutWriter = io.MultiWriter(stdoutWriter, os.Stdout)
		stderrWriter = io.MultiWriter(stderrWriter, os.Stderr)
	}

	c := exec.CommandContext(e.runtime.Ctx, "bash", "-c", finalCmd)
	c.Stdout = stdoutWriter
	c.Stderr = stderrWriter

	if capture {
		err = c.Run()
		e.traceCommand(finalCmd, traceID, start, sink.StdoutPath(), sink.StderrPath(), err, false)
		if err != nil {
			errorType := ClassifyError(err)
			logger.Error(logmw.WithFields(logCtx, map[string]any{
				"event":       "command_done",
				"result":      "failed",
				"duration_ms": time.Since(start).Milliseconds(),
				"error_type":  errorType,
				"error":       err.Error(),
			}), "exec failed")
			PrintCommandDone(e.opts.Verbose, start, finalCmd, err)
		} else {
			logger.Info(logmw.WithFields(logCtx, map[string]any{
				"event":       "command_done",
				"result":      "success",
				"duration_ms": time.Since(start).Milliseconds(),
			}), "exec done")
			PrintCommandDone(e.opts.Verbose, start, finalCmd, nil)
		}
		return combinedBuf.String(), err
	}

	err = c.Run()
	e.traceCommand(finalCmd, traceID, start, sink.StdoutPath(), sink.StderrPath(), err, false)
	if err != nil {
		errorType := ClassifyError(err)
		logger.Error(logmw.WithFields(logCtx, map[string]any{
			"event":       "command_done",
			"result":      "failed",
			"duration_ms": time.Since(start).Milliseconds(),
			"error_type":  errorType,
			"error":       err.Error(),
		}), "exec failed")
		PrintCommandDone(e.opts.Verbose, start, finalCmd, err)
	} else {
		logger.Info(logmw.WithFields(logCtx, map[string]any{
			"event":       "command_done",
			"result":      "success",
			"duration_ms": time.Since(start).Milliseconds(),
		}), "exec done")
		PrintCommandDone(e.opts.Verbose, start, finalCmd, nil)
	}
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
	traceID string,
	start time.Time,
	stdoutPath string,
	stderrPath string,
	err error,
	dryRun bool,
) {
	trace := e.runtime.Trace
	if trace == nil {
		return
	}

	end := time.Now()
	timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
	result := "success"
	if dryRun {
		result = "dry_run"
	} else if err != nil {
		result = "failed"
	}
	errorType := ClassifyError(err)
	event := tracemw.NewTraceEvent(
		command,
		traceID,
		e.runtime.NodeName,
		e.runtime.NodeAddr,
		e.runtime.Component,
		result,
		errorType,
		stdoutPath,
		stderrPath,
		start,
		end,
		"",
		"",
		err,
		dryRun,
		timedOut,
	)
	trace.OnCommand(event)
}
