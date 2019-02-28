package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/mojotalantikite/terraform-provider-akamai/akamai"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: akamai.Provider})
}
