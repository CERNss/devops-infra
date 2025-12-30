package trace

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
)

type TraceEvent struct {
	Command    string `json:"command"`
	Node       string `json:"node,omitempty"`
	NodeAddr   string `json:"node_addr,omitempty"`
	StdoutPath string `json:"stdout_path,omitempty"`
	StderrPath string `json:"stderr_path,omitempty"`
	Start      string `json:"start"`
	End        string `json:"end"`
	DurationMs int64  `json:"duration_ms"`
	Err        string `json:"error,omitempty"`
	TimedOut   bool   `json:"timed_out,omitempty"`
	DryRun     bool   `json:"dry_run,omitempty"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
}

type TraceSink interface {
	OnCommand(event TraceEvent)
}

type stderrTraceSink struct {
	mu sync.Mutex
	w  io.Writer
}

func NewStderrTraceSink() TraceSink {
	return &stderrTraceSink{w: os.Stderr}
}

func DefaultTraceSink() TraceSink {
	sink, err := NewFileTraceSink(constant.DefaultTraceFile)
	if err != nil {
		return NewStderrTraceSink()
	}
	return sink
}

func (s *stderrTraceSink) OnCommand(event TraceEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		fmt.Fprintf(s.w, "trace marshal error: %v\n", err)
		return
	}

	s.mu.Lock()
	_, _ = s.w.Write(append(payload, '\n'))
	s.mu.Unlock()
}

type noopTraceSink struct{}

func (noopTraceSink) OnCommand(TraceEvent) {}

func NoopTraceSink() TraceSink {
	return noopTraceSink{}
}

type fileTraceSink struct {
	mu sync.Mutex
	f  *os.File
}

func NewFileTraceSink(path string) (TraceSink, error) {
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
	return &fileTraceSink{f: f}, nil
}

func (s *fileTraceSink) OnCommand(event TraceEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		fmt.Fprintf(s.f, "trace marshal error: %v\n", err)
		return
	}

	s.mu.Lock()
	_, _ = s.f.Write(append(payload, '\n'))
	s.mu.Unlock()
}

func NewTraceEvent(
	command string,
	nodeName string,
	nodeAddr string,
	stdoutPath string,
	stderrPath string,
	start time.Time,
	end time.Time,
	stdout string,
	stderr string,
	err error,
	dryRun bool,
	timedOut bool,
) TraceEvent {
	event := TraceEvent{
		Command:    command,
		Node:       nodeName,
		NodeAddr:   nodeAddr,
		StdoutPath: stdoutPath,
		StderrPath: stderrPath,
		Start:      start.Format(time.RFC3339Nano),
		End:        end.Format(time.RFC3339Nano),
		DurationMs: end.Sub(start).Milliseconds(),
		Stdout:     stdout,
		Stderr:     stderr,
		DryRun:     dryRun,
		TimedOut:   timedOut,
	}
	if err != nil {
		event.Err = err.Error()
	}
	return event
}
