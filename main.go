package main

import (
	"os"

	"github.com/TacoContent/ironstate/sdk/handler"
	"github.com/TacoContent/ironstate/sdk/plugin"

	"github.com/camalot/ironstate-handler-hosts/internal/hosts"
)

func main() {
	_ = version
	_ = commit
	_ = date
	if len(os.Args) > 1 && os.Args[1] == hosts.BecomeCommand {
		os.Exit(hosts.RunWriteHelper(os.Args[1:]))
	}
	plugin.Serve(map[string]handler.Handler{
		"entry": hosts.EntryHandler{},
	})
}
