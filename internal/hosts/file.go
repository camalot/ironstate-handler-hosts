package hosts

import (
	"os"
	"strings"

	"github.com/TacoContent/ironstate/sdk/handler"
)

func readLines(path string) ([]string, error) {
	contents, err := os.ReadFile(path) //nolint:gosec // path is the user-selected hosts file this handler is designed to manage
	if err != nil {
		return nil, err
	}
	text := strings.ReplaceAll(string(contents), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(text, "\n"), nil
}

func writeLines(path string, lines []string, become handler.Become) error {
	contents := ""
	if len(lines) > 0 {
		contents = strings.Join(lines, "\n") + "\n"
	}
	if !become.Enabled {
		return os.WriteFile(path, []byte(contents), 0o644) //nolint:gosec // hosts files require a system-readable mode
	}
	return writeWithBecome(path, contents, become)
}

func hasMapping(lines []string, ip, hostname string) bool {
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ip && contains(fields[1:], hostname) {
			return true
		}
	}
	return false
}

func removeMapping(lines []string, ip, hostname string, comment string) []string {
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ip && contains(fields[1:], hostname) {
			continue
		}
		// Remove lines that match the comment
		if comment != "" && strings.HasPrefix(strings.TrimSpace(line), "# "+comment) {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}

func removeHostname(lines []string, hostname string) []string {
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && !strings.HasPrefix(fields[0], "#") && contains(fields[1:], hostname) {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
