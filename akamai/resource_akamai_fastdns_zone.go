package akamai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/trussworks/akamai-sdk-go/akamai"
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

			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"sign_and_serve": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
		},
	}
}

func resourceAkamaiFastDNSZoneCreate(d *schema.ResourceData, m interface{}) error {
	conn := m.(*AkamaiClient).client

	input := &akamai.ZoneCreateRequest{
		Zone:         d.Get("zone").(string),
		Type:         d.Get("type").(string),
		Comment:      d.Get("comment").(string),
		SignAndServe: d.Get("sign_and_serve").(bool),
	}

	cid := d.Get("contract_id").(string)
	log.Printf("[DEBUG] Creating Akamai FastDNS Hosted Zone: %s", input.Zone)

	output, _, err := conn.FastDNSv2.CreateZone(context.Background(), cid, input)
	if err != nil {
		return fmt.Errorf("error creating Akamai FastDNS Hosted Zone: %s", err)
	}

	log.Printf("[DEBUG] Akamai FastDNS Hosted Zone Created: %v", *output.Zone)

	d.SetId(*output.Zone)

	return resourceAkamaiFastDNSZoneRead(d, m)
}

func resourceAkamaiFastDNSZoneRead(d *schema.ResourceData, m interface{}) error {
	conn := m.(*AkamaiClient).client

	input := d.Get("zone").(string)
	log.Printf("[DEBUG] Getting Akamai FastDNS Hosted Zone: %s", input)

	output, _, err := conn.FastDNSv2.GetZone(context.Background(), input)
	if err != nil {
		return fmt.Errorf("error getting Akamai FastDNS Zone (%s): %s", d.Id(), err)
	}

	if output == nil || output.Zone == nil {
		log.Printf("[WARN] Akamai FastDNS Hosted Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] Listing zone returned from Akamai: %v", *output.Zone)

	d.Set("comment", *output.Comment)
	d.Set("zone", *output.Zone)
	d.Set("type", *output.Type)
	d.Set("contract_id", *output.ContractID)

	return nil
}

func resourceAkamaiFastDNSZoneUpdate(d *schema.ResourceData, m interface{}) error {
	conn := m.(*AkamaiClient).client
	if d.HasChange("comment") {
		input := &akamai.ZoneCreateRequest{
			Zone:    d.Id(),
			Comment: d.Get("comment").(string),
		}

		_, _, err := conn.FastDNSv2.UpdateZone(context.Background(), input)
		if err != nil {
			return fmt.Errorf("error updating Akamai FastDNS Zone (&s) error: %s", d.Id(), err)
		}

		d.SetPartial("comment")

	}

	d.Partial(false)

	return resourceAkamaiFastDNSZoneRead(d, m)
}

func resourceAkamaiFastDNSZoneDelete(d *schema.ResourceData, m interface{}) error {
	conn := m.(*AkamaiClient).client

	input := &akamai.ZoneDeleteRequest{
		Zones: []string{d.Id()},
	}
	log.Printf("[DEBUG] Deleting Akamai FastDNS Hosted Zone: %s", d.Id())

	output, resp, err := conn.FastDNSv2.DeleteZone(context.Background(), input)
	for resp.StatusCode != 201 {
		log.Printf("[WARN] Akamai FastDNS API error (%s). Retrying...", resp.StatusCode)
		time.Sleep(5 * time.Second)
		output, resp, err = conn.FastDNSv2.DeleteZone(context.Background(), input)
	}
	if err != nil {
		return fmt.Errorf("error deleting Akamai FastDNS Zone (&s) error: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Checking delete status of Akamai FastDNS Hosted Zone: %s", d.Id())

	for {
		zs, _, err := conn.FastDNSv2.DeleteZoneStatus(context.Background(), *output.RequestID)
		if err != nil {
			return fmt.Errorf("error deleting Akamai FastDNS Zone (&s) error: %s", d.Id(), err)
		}

		if *zs.IsComplete {
			break
		}
	}
	return nil
}
