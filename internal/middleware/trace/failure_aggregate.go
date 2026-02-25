package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
)

type FailureSummary struct {
	RunID            string   `json:"run_id"`
	Workflow         string   `json:"workflow"`
	TotalCommands    int      `json:"total_commands"`
	FailedCommands   int      `json:"failed_commands"`
	FailedComponents []string `json:"failed_components,omitempty"`
	FailedCommandsAt []string `json:"failed_command_summaries,omitempty"`
	ErrorsPath       string   `json:"errors_path"`
	SummaryPath      string   `json:"summary_path"`
	GeneratedAt      string   `json:"generated_at"`
}

type FailureAggregator struct {
	mu          sync.Mutex
	workflow    string
	runID       string
	total       int
	failed      int
	components  map[string]struct{}
	commands    []string
	errors      *os.File
	errorsPath  string
	summaryPath string
}

func NewFailureAggregator(logDir string, workflow string) (*FailureAggregator, error) {
	baseDir := strings.TrimSpace(logDir)
	if baseDir == "" {
		baseDir = constant.DefaultLogDir
	}
	resolved, err := pathutil.ResolveUserPath(filepath.Join(baseDir, "errors"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(resolved, 0o755); err != nil {
		return nil, err
	}

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	errorsPath := filepath.Join(resolved, "run-"+runID+".errors.jsonl")
	summaryPath := filepath.Join(resolved, "run-"+runID+".summary.json")

	errorsFile, err := os.OpenFile(errorsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	return &FailureAggregator{
		workflow:    workflow,
		runID:       runID,
		components:  map[string]struct{}{},
		errors:      errorsFile,
		errorsPath:  errorsPath,
		summaryPath: summaryPath,
	}, nil
}

func (a *FailureAggregator) OnCommand(event TraceEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.total++
	if event.Result != "failed" {
		return
	}

	a.failed++
	if component := strings.TrimSpace(event.Component); component != "" {
		a.components[component] = struct{}{}
	}
	a.commands = append(a.commands, summarizeCommand(event.Command))

	if a.errors == nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = a.errors.Write(append(payload, '\n'))
}

func (a *FailureAggregator) HasFailures() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.failed > 0
}

func (a *FailureAggregator) Summary() FailureSummary {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.summaryLocked()
}

func (a *FailureAggregator) summaryLocked() FailureSummary {
	components := make([]string, 0, len(a.components))
	for component := range a.components {
		components = append(components, component)
	}
	sort.Strings(components)

	commands := make([]string, len(a.commands))
	copy(commands, a.commands)

	return FailureSummary{
		RunID:            a.runID,
		Workflow:         a.workflow,
		TotalCommands:    a.total,
		FailedCommands:   a.failed,
		FailedComponents: components,
		FailedCommandsAt: commands,
		ErrorsPath:       a.errorsPath,
		SummaryPath:      a.summaryPath,
		GeneratedAt:      time.Now().Format(time.RFC3339Nano),
	}
}

func (a *FailureAggregator) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	summary := a.summaryLocked()
	if payload, err := json.MarshalIndent(summary, "", "  "); err == nil {
		_ = os.WriteFile(a.summaryPath, append(payload, '\n'), 0o644)
	}

	if a.errors == nil {
		return nil
	}
	err := a.errors.Close()
	a.errors = nil
	return err
}

func FormatFailureSummary(summary FailureSummary) string {
	if summary.FailedCommands == 0 {
		return ""
	}
	lines := []string{
		"[summary] installation failed",
		fmt.Sprintf("[summary] workflow: %s", summary.Workflow),
		fmt.Sprintf("[summary] failed commands: %d/%d", summary.FailedCommands, summary.TotalCommands),
	}
	if len(summary.FailedComponents) > 0 {
		lines = append(lines, "[summary] failed components: "+strings.Join(summary.FailedComponents, ", "))
	}
	if len(summary.FailedCommandsAt) > 0 {
		max := len(summary.FailedCommandsAt)
		if max > 5 {
			max = 5
		}
		for i := 0; i < max; i++ {
			lines = append(lines, fmt.Sprintf("[summary] command[%d]: %s", i+1, summary.FailedCommandsAt[i]))
		}
	}
	lines = append(lines,
		"[summary] error events: "+summary.ErrorsPath,
		"[summary] run summary: "+summary.SummaryPath,
	)
	return strings.Join(lines, "\n") + "\n"
}

func summarizeCommand(command string) string {
	flat := strings.Join(strings.Fields(strings.ReplaceAll(command, "\n", " ")), " ")
	if len(flat) <= 160 {
		return flat
	}
	return flat[:157] + "..."
}
