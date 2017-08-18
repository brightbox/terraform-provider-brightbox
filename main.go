package main

import (
	"github.com/brightbox/terraform-provider-brightbox/brightbox"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: brightbox.Provider,
	})
}
