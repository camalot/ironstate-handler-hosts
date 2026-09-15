package hosts

import (
	"os"
	"strings"

	"github.com/TacoContent/ironstate/sdk/handler"
)

func (EntryHandler) ScanRole() string { return "roles/system/hosts" }

func (EntryHandler) Scan(ctx handler.Context) ([]handler.ScanItem, error) {
	path := defaultHostsPath()
	if value, ok := ctx.Flat["hosts_path"].(string); ok && value != "" {
		path = value
	}
	lines, err := readLines(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	items := make([]handler.ScanItem, 0)
	for _, line := range lines {
		line = strings.TrimPrefix(strings.TrimSpace(line), "\uFEFF")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		for _, hostname := range fields[1:] {
			items = append(items, handler.ScanItem{
				Module: "entry",
				Name:   hostname,
				Config: map[string]any{"path": path, "ip": fields[0], "hostname": hostname},
			})
		}
	}
	return items, nil
}

var _ handler.ScanCapable = EntryHandler{}
