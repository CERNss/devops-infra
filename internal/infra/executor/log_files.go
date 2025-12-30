package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
)

func OpenCommandLogs(rt Runtime) (string, string, *os.File, *os.File) {
	logDir := rt.LogDir
	if logDir == "" {
		logDir = constant.DefaultTraceDir
	}

	resolved, err := pathutil.ResolveUserPath(logDir)
	if err != nil {
		return "", "", nil, nil
	}
	if err := os.MkdirAll(resolved, 0o755); err != nil {
		return "", "", nil, nil
	}

	identity := fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	stdoutPath := filepath.Join(resolved, "cmd-"+identity+".stdout.log")
	stderrPath := filepath.Join(resolved, "cmd-"+identity+".stderr.log")

	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", "", nil, nil
	}

	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		_ = stdoutFile.Close()
		return "", "", nil, nil
	}

	return stdoutPath, stderrPath, stdoutFile, stderrFile
}
