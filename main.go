package main

import (
	"os"

	"github.com/hashicorp/terraform/plugin"
)

func main() {
	if len(os.Args) > 1 {
		// TODO: editor mode
	} else {
		// terraform plugin mode
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: Provider,
		})
	}
}
