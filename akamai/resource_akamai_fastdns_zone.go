package akamai

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAkamaiFastDNSZone() *schema.Resource {
	return &schema.Resource{
		Create: resourceAkamaiFastDNSZoneCreate,
		Read:   resourceAkamaiFastDNSZoneRead,
		Update: resourceAkamaiFastDNSZoneUpdate,
		Delete: resourceAkamaiFastDNSZoneDelete,
		Schema: map[string]*schema.Schema{
			"contract_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAkamaiFastDNSZoneCreate(d *schema.ResourceData, m interface{}) error {
	input := d.Get("name").(string)
	log.Printf("[DEBUG] Creating Route53 hosted zone: %s", input)
	return resourceAkamaiFastDNSZoneRead(d, m)
}

func resourceAkamaiFastDNSZoneRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAkamaiFastDNSZoneUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceAkamaiFastDNSZoneRead(d, m)
}

func resourceAkamaiFastDNSZoneDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
