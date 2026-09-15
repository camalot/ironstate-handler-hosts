package hosts

import (
	"os"
	"strings"

	"github.com/TacoContent/ironstate/sdk/handler"
)

func (EntryHandler) FactName(item map[string]any) (string, bool) {
	name, _ := item["name"].(string)
	return strings.TrimPrefix(name, "$"), name != ""
}

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
	seen := make(map[string]bool)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		for _, hostname := range fields[1:] {
			if strings.HasPrefix(hostname, "#") {
				break
			}
			key := fields[0] + "\x00" + hostname
			if seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, handler.ScanItem{
				Module: "entry",
				Name:   hostname,
				Config: map[string]any{"path": path, "ip": fields[0], "hostname": hostname},
			})
		}
	}
	return items, nil
}

var _ handler.FactProducer = EntryHandler{}
var _ handler.ScanCapable = EntryHandler{}
