package trace

import (
	"errors"
	"strings"
	"testing"
)

func TestNewMultiTraceSinkDispatchesToAllSinks(t *testing.T) {
	sinkA := &capturingSink{}
	sinkB := &capturingSink{}
	multi := NewMultiTraceSink(nil, sinkA, sinkB)

	event := TraceEvent{TraceID: "trace-1", Result: "failed"}
	multi.OnCommand(event)

	if len(sinkA.events) != 1 || len(sinkB.events) != 1 {
		t.Fatalf("expected event broadcast to all sinks, got A=%d B=%d", len(sinkA.events), len(sinkB.events))
	}
	if sinkA.events[0].TraceID != "trace-1" || sinkB.events[0].TraceID != "trace-1" {
		t.Fatalf("unexpected trace ids: A=%v B=%v", sinkA.events[0].TraceID, sinkB.events[0].TraceID)
	}
}

func TestMultiTraceSinkCloseAggregatesErrors(t *testing.T) {
	sinkA := &closableSink{closeErr: errors.New("close-a")}
	sinkB := &closableSink{closeErr: errors.New("close-b")}

	multi := NewMultiTraceSink(sinkA, sinkB)
	closable, ok := multi.(CloseableTraceSink)
	if !ok {
		t.Fatal("expected multi sink to implement CloseableTraceSink")
	}

	err := closable.Close()
	if err == nil {
		t.Fatal("expected joined close error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "close-a") || !strings.Contains(msg, "close-b") {
		t.Fatalf("expected joined errors, got %q", msg)
	}
}

type capturingSink struct {
	events []TraceEvent
}

func (c *capturingSink) OnCommand(event TraceEvent) {
	c.events = append(c.events, event)
}

type closableSink struct {
	capturingSink
	closeErr error
}

func (c *closableSink) Close() error {
	return c.closeErr
}
