package mirror

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"devops-infra/internal/utils/path"
)

const (
	mirrorMainURL   = "https://linuxmirrors.cn/main.sh"
	mirrorDockerURL = "https://linuxmirrors.cn/docker.sh"
)

func EnsureMirrorMainScript() (string, error) {
	return ensureScript("main.sh", mirrorMainURL)
}

func EnsureMirrorDockerScript() (string, error) {
	return ensureScript("docker.sh", mirrorDockerURL)
}

func ensureScript(filename, url string) (string, error) {
	dir, err := resolveMirrorDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	target := filepath.Join(dir, filename)
	ok, err := isValidFile(target)
	if err != nil {
		return "", err
	}
	if ok {
		return target, nil
	}

	if err := downloadFile(target, url); err != nil {
		return "", err
	}

	return target, nil
}

func resolveMirrorDir() (string, error) {
	root, err := path.ResolvePath("go.mod")
	if err == nil {
		return filepath.Join(filepath.Dir(root), "scripts", "mirror"), nil
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return "", err
	}
	return filepath.Join(wd, "scripts", "mirror"), nil
}

func isValidFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("expected regular file: %s", path)
	}
	if info.Size() == 0 {
		return false, nil
	}
	return true, nil
}

func downloadFile(dest, url string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s failed: %s", url, resp.Status)
	}

	tmp := dest + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Chmod(dest, 0o755)
}
