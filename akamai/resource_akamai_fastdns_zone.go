package akamai

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceZone() *schema.Resource {
	return &schema.Resource{
		Create: resourceZoneCreate,
		Read:   resourceZoneRead,
		Update: resourceZoneUpdate,
		Delete: resourceZoneDelete,

		Schema: map[string]*schema.Schema{
			"placeholder": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceZoneCreate(d *schema.ResourceData, m interface{}) error {
	return resourceZoneRead(d, m)
}

func resourceZoneRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceZoneUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceZoneRead(d, m)
}

func resourceZoneDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
