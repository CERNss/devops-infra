package os

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func Detect() (*OSInfo, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], `"'`)
		data[key] = val
	}

	info := &OSInfo{
		ID:      data["ID"],
		IDLike:  data["ID_LIKE"],
		Version: data["VERSION_ID"],
	}

	// Major version
	if info.Version != "" {
		parts := strings.Split(info.Version, ".")
		if v, err := strconv.Atoi(parts[0]); err == nil {
			info.Major = v
		}
	}

	// Normalize family
	switch {
	case strings.Contains(info.IDLike, "debian") ||
		info.ID == "debian" ||
		info.ID == "ubuntu":
		info.Family = "debian"
	case strings.Contains(info.IDLike, "rhel") ||
		strings.Contains(info.IDLike, "fedora") ||
		info.ID == "centos" ||
		info.ID == "rocky" ||
		info.ID == "almalinux" ||
		info.ID == "fedora":
		info.Family = "rhel"
	default:
		info.Family = "unknown"
	}

	return info, nil
}
