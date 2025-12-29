package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type TraceEvent struct {
	Command    string `json:"command"`
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

func NewTraceEvent(
	command string,
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
