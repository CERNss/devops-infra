package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
)

type OutputSink interface {
	Stdout() io.Writer
	Stderr() io.Writer
	StdoutPath() string
	StderrPath() string
	Close() error
}

type OutputSinkFactory interface {
	Open(info RuntimeInfo, command string) (OutputSink, error)
}

type RuntimeInfo struct {
	LogDir string
}

type FileOutputSinkFactory struct {
	Dir string
}

func (f FileOutputSinkFactory) Open(info RuntimeInfo, command string) (OutputSink, error) {
	logDir := f.Dir
	if logDir == "" {
		logDir = info.LogDir
	}
	if logDir == "" {
		logDir = constant.DefaultTraceDir
	}

	resolved, err := pathutil.ResolveUserPath(logDir)
	if err != nil {
		return noopOutputSink{}, err
	}
	if err := os.MkdirAll(resolved, 0o755); err != nil {
		return noopOutputSink{}, err
	}

	identity := fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	stdoutPath := filepath.Join(resolved, "cmd-"+identity+".stdout.log")
	stderrPath := filepath.Join(resolved, "cmd-"+identity+".stderr.log")

	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return noopOutputSink{}, err
	}

	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		_ = stdoutFile.Close()
		return noopOutputSink{}, err
	}

	return &FileOutputSink{
		stdoutPath: stdoutPath,
		stderrPath: stderrPath,
		stdoutFile: stdoutFile,
		stderrFile: stderrFile,
	}, nil
}

type FileOutputSink struct {
	stdoutPath string
	stderrPath string
	stdoutFile *os.File
	stderrFile *os.File
}

func (s *FileOutputSink) Stdout() io.Writer {
	return s.stdoutFile
}

func (s *FileOutputSink) Stderr() io.Writer {
	return s.stderrFile
}

func (s *FileOutputSink) StdoutPath() string {
	return s.stdoutPath
}

func (s *FileOutputSink) StderrPath() string {
	return s.stderrPath
}

func (s *FileOutputSink) Close() error {
	if s == nil {
		return nil
	}
	if s.stdoutFile != nil {
		_ = s.stdoutFile.Close()
	}
	if s.stderrFile != nil {
		_ = s.stderrFile.Close()
	}
	return nil
}

type noopOutputSink struct{}

func (noopOutputSink) Stdout() io.Writer    { return io.Discard }
func (noopOutputSink) Stderr() io.Writer    { return io.Discard }
func (noopOutputSink) StdoutPath() string   { return "" }
func (noopOutputSink) StderrPath() string   { return "" }
func (noopOutputSink) Close() error         { return nil }

func NoopOutputSink() OutputSink {
	return noopOutputSink{}
}
