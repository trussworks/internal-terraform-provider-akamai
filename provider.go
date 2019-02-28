package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider is the entry point to the provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["client_secret"],
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["host"],
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["access_token"],
			},
			"client_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["client_token"],
			},
			"edgerc_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".edgerc",
				Description: descriptions["edgerc_file"],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"akamai_zone": resourceZone(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"access_token": "The access token for API operations. This can be found in the\n" +
			"Identity Management section of Akamai Luna Control Center.",
		"client_token": "The client token for API operations. This can be found in the\n" +
			"Identity Management section of Akamai Luna Control Center.",
		"client_secret": "The client secret for API operations. This can be found in the\n" +
			"Identity Management section of Akamai Luna Control Center.",
		"host": "The base API hostname without the protocol scheme. This can be found in the\n" +
			"Identity Management section of Akamai Luna Control Center.",
		"edgerc_file": "The path to the edgerc credentials file. If not set\n" +
			"this defaults to ~/.edgerc.",
	}
}

type Config struct {
	AccessToken  string
	ClientSecret string
	ClientToken  string
	Host         string
	EdgercFile   string
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		AccessToken:  d.Get("access_token").(string),
		ClientSecret: d.Get("client_secret").(string),
		ClientToken:  d.Get("client_token").(string),
		Host:         d.Get("host").(string),
		EdgercFile:   d.Get("edgerc_file").(string),
	}
	return config, nil
}
