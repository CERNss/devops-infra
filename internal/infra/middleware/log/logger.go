package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
)

type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type writerLogger struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *writerLogger) log(level string, msg string) {
	if l == nil || l.w == nil {
		return
	}
	ts := time.Now().Format(time.RFC3339Nano)
	l.mu.Lock()
	_, _ = fmt.Fprintf(l.w, "%s %-5s %s\n", ts, level, msg)
	l.mu.Unlock()
}

func (l *writerLogger) Info(msg string)  { l.log("INFO", msg) }
func (l *writerLogger) Warn(msg string)  { l.log("WARN", msg) }
func (l *writerLogger) Error(msg string) { l.log("ERROR", msg) }

type fileLogger struct {
	*writerLogger
	file *os.File
}

func (l *fileLogger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

type multiLogger struct {
	loggers []Logger
}

func (m multiLogger) Info(msg string) {
	for _, logger := range m.loggers {
		logger.Info(msg)
	}
}

func (m multiLogger) Warn(msg string) {
	for _, logger := range m.loggers {
		logger.Warn(msg)
	}
}

func (m multiLogger) Error(msg string) {
	for _, logger := range m.loggers {
		logger.Error(msg)
	}
}

type noopLogger struct{}

func (noopLogger) Info(string)  {}
func (noopLogger) Warn(string)  {}
func (noopLogger) Error(string) {}

func NoopLogger() Logger {
	return noopLogger{}
}

func NewStderrLogger() Logger {
	return &writerLogger{w: os.Stderr}
}

func NewFileLogger(path string) (Logger, error) {
	resolved, err := pathutil.ResolveUserPath(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &fileLogger{
		writerLogger: &writerLogger{w: f},
		file:         f,
	}, nil
}

func DefaultLogger(logDir string) Logger {
	path := constant.DefaultLogFile
	if logDir != "" {
		path = filepath.Join(logDir, "run.log")
	}
	fileLogger, err := NewFileLogger(path)
	if err != nil {
		return NewStderrLogger()
	}
	return multiLogger{loggers: []Logger{
		NewStderrLogger(),
		fileLogger,
	}}
}
