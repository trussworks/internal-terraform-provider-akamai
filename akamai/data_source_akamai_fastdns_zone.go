package akamai

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAkamaiFastDNSZone() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAkamaiFastDNSZoneRead,
		Schema: map[string]*schema.Schema{
			"contract_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAkamaiFastDNSZoneRead(d *schema.ResourceData, meta interface{}) error {
	id, idExists := d.GetOk("contract_id")

	if !idExists {
		return fmt.Errorf("contract_id must be set")
	}

	d.Set("contract_id", id)

	return nil
}
