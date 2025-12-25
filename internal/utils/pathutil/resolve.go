package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResolvePath(rel string) (string, error) {
	if filepath.IsAbs(rel) {
		if fileExists(rel) {
			return rel, nil
		}
		return "", fmt.Errorf("path not found: %s", rel)
	}

	rel = filepath.Clean(rel)
	roots := []string{}

	if execPath, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(execPath))
	}
	if wd, err := os.Getwd(); err == nil {
		if len(roots) == 0 || roots[0] != wd {
			roots = append(roots, wd)
		}
	}

	for _, start := range roots {
		dir := start
		for {
			candidate := filepath.Join(dir, rel)
			if fileExists(candidate) {
				return candidate, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "", fmt.Errorf("path not found: %s", rel)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
