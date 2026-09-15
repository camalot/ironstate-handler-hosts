package hosts

import (
	"strings"

	"github.com/TacoContent/ironstate/sdk/handler"
)

func (EntryHandler) FactName(item map[string]any) (string, bool) {
	name, _ := item["name"].(string)
	return strings.TrimPrefix(name, "$"), name != ""
}

var _ handler.FactProducer = EntryHandler{}
