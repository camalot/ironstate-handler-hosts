package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type hostSpec struct {
	Path     string
	IP       string
	Hostname string
	Entry    string
	Comment  string
}

func parseSpec(item map[string]any) (hostSpec, error) {
	path, _ := item["path"].(string)
	if path == "" {
		path = defaultHostsPath()
	}
	ip, _ := item["ip"].(string)
	hostname, _ := item["hostname"].(string)
	if strings.TrimSpace(ip) == "" || strings.TrimSpace(hostname) == "" {
		return hostSpec{}, fmt.Errorf("hosts handler requires non-empty 'ip' and 'hostname'")
	}
	comment, _ := item["comment"].(string)
	// comment is optional and will be added as a line starting with "# " if provided
	return hostSpec{Path: path, IP: strings.TrimSpace(ip), Hostname: strings.TrimSpace(hostname), Entry: strings.TrimSpace(ip) + " " + strings.TrimSpace(hostname), Comment: strings.TrimSpace(comment)}, nil
}

func defaultHostsPath() string {
	if runtime.GOOS == "windows" {
		root := os.Getenv("SystemRoot")
		if root == "" {
			root = `C:\Windows`
		}
		return filepath.Join(root, "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}
