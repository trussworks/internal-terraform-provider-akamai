package akamai

import (
	"context"
	"fmt"
	"log"

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

			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"activation_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAkamaiFastDNSZoneRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AkamaiClient).client

	zone, zoneExists := d.GetOk("zone")
	if !zoneExists {
		return fmt.Errorf("zone must be set")

	}
	input := zone.(string)

	log.Printf("[DEBUG] Getting Akamai FastDNS Hosted Zone: %s", input)

	output, resp, err := conn.FastDNSv2.GetZone(context.Background(), input)
	if err != nil || resp.StatusCode == 404 {
		return fmt.Errorf("Error finding FastDNS Zone: %v", err)
	}

	d.SetId("zone")
	d.Set("zone", output.Zone)
	d.Set("comment", output.Comment)
	d.Set("type", output.Type)
	d.Set("contract_id", output.ContractID)
	d.Set("activation_state", output.ActivationState)

	return nil
}
