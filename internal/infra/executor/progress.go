package executor

import (
	"fmt"
	"strings"
	"time"
)

func PrintCommandStart(verbose bool, dryRun bool, cmd string) {
	if verbose {
		return
	}
	summary := commandSummary(cmd)
	if isProbeCommand(summary) {
		if dryRun {
			fmt.Printf("[cmd] dry-run check: %s\n", summary)
			return
		}
		fmt.Printf("[cmd] check: %s\n", summary)
		return
	}
	if dryRun {
		fmt.Printf("[cmd] dry-run: %s\n", summary)
		return
	}
	fmt.Printf("[cmd] start: %s\n", summary)
}

func PrintCommandDone(verbose bool, start time.Time, cmd string, err error) {
	if verbose {
		return
	}
	summary := commandSummary(cmd)
	duration := time.Since(start).Round(time.Millisecond)
	if isProbeCommand(summary) {
		if err != nil {
			fmt.Printf("[cmd] check miss (%s): %s\n", duration, summary)
			return
		}
		fmt.Printf("[cmd] check ok (%s): %s\n", duration, summary)
		return
	}
	if err != nil {
		fmt.Printf("[cmd] failed (%s): %s\n", duration, summary)
		return
	}
	fmt.Printf("[cmd] done (%s): %s\n", duration, summary)
}

func commandSummary(cmd string) string {
	trimmed := strings.TrimSpace(cmd)
	trimmed = strings.ReplaceAll(trimmed, "\n", " ")
	trimmed = strings.Join(strings.Fields(trimmed), " ")
	const maxLen = 160
	if len(trimmed) > maxLen {
		return trimmed[:maxLen-3] + "..."
	}
	return trimmed
}

func isProbeCommand(summary string) bool {
	switch {
	case strings.HasPrefix(summary, "test "):
		return true
	case strings.HasPrefix(summary, "command -v "):
		return true
	case strings.HasPrefix(summary, "which "):
		return true
	case strings.HasPrefix(summary, "sysctl -n "):
		return true
	case strings.HasPrefix(summary, "ls /etc/cni/net.d/"):
		return true
	case strings.HasSuffix(summary, " --version"):
		return true
	}
	return false
}
