package akamai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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

	// send the delete zone request. Akamai throws 500s sometimes, until
	// they fix that bug we must retry until HTTP 201 (or timeout)
	log.Printf("[DEBUG] Deleting Akamai FastDNS Hosted Zone: %s", d.Id())
	output, err := deleteFastDNSZone(conn, d.Id())
	if err != nil {
		return err
	}

	// make sure the zone really was deleted
	rid := output.(*akamai.ZoneDeleteResponse).RequestID
	_, err = checkDeleteFastDNSZone(conn, *rid)
	if err != nil {
		return err
	}

	return nil
}

func deleteFastDNSZone(conn *akamai.Client, zone string) (interface{}, error) {
	wait := resource.StateChangeConf{
		Pending:    []string{"rejected"},
		Target:     []string{"accepted"},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			z := []string{zone}
			input := &akamai.ZoneDeleteRequest{
				Zones: z,
			}

			output, resp, err := conn.FastDNSv2.DeleteZone(context.Background(), input)
			// This is bad Go, as we'd really want to check the err first. Akamai throws
			// intermittent HTTP 500 and 503 errors though, and often retrying gives us
			// our expected HTTP 201. Until Akamai provides a stable endpoint we need this.
			if resp.StatusCode == 500 || resp.StatusCode == 503 {
				return 42, "rejected", nil
			}

			if err != nil {
				e := fmt.Errorf("error deleting Akamai FastDNS Zone (%s) error: ", z, err)
				return 42, "failure", e
			}

			return output, "accepted", nil
		},
	}
	return wait.WaitForState()

}

func checkDeleteFastDNSZone(conn *akamai.Client, rid string) (interface{}, error) {
	wait := resource.StateChangeConf{
		Pending:    []string{"rejected"},
		Target:     []string{"accepted"},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			zs, _, err := conn.FastDNSv2.DeleteZoneStatus(context.Background(), rid)
			if err != nil {
				e := fmt.Errorf("error checking Akamai FastDNS delete status: %s", err)
				return 42, "failure", e
			}

			if !*zs.IsComplete {
				// if the delete has not completed, retry
				return 42, "rejected", nil
			}

			return 42, "accepted", nil
		},
	}

	return wait.WaitForState()
}
