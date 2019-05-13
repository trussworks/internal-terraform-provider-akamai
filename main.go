package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/trussworks/terraform-provider-akamai/akamai"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: akamai.Provider})
}
