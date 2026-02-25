package flow

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	tracemw "devops-infra/internal/middleware/trace"
)

type stubFailureAggregator struct {
	closed      bool
	hasFailures bool
	summary     tracemw.FailureSummary
}

func (s *stubFailureAggregator) Close() error {
	s.closed = true
	return nil
}

func (s *stubFailureAggregator) HasFailures() bool {
	return s.hasFailures
}

func (s *stubFailureAggregator) Summary() tracemw.FailureSummary {
	return s.summary
}

func TestFinalizeFailureAggregatorClosesOnEarlyReturn(t *testing.T) {
	agg := &stubFailureAggregator{}

	err := func() error {
		defer finalizeFailureAggregator(agg, io.Discard)
		return errors.New("early return")
	}()
	if err == nil {
		t.Fatal("expected early return error")
	}
	if !agg.closed {
		t.Fatal("expected aggregator close on early return")
	}
}

func TestFinalizeFailureAggregatorPrintsSummaryOnFailures(t *testing.T) {
	agg := &stubFailureAggregator{
		hasFailures: true,
		summary: tracemw.FailureSummary{
			Workflow:       "install-base",
			FailedCommands: 1,
			TotalCommands:  2,
		},
	}

	var out bytes.Buffer
	finalizeFailureAggregator(agg, &out)

	if !agg.closed {
		t.Fatal("expected aggregator to be closed")
	}
	if !strings.Contains(out.String(), "[summary] installation failed") {
		t.Fatalf("expected summary output, got %q", out.String())
	}
}
