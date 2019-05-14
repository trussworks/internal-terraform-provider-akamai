package akamai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/trussworks/akamai-sdk-go/akamai"
)

func resourceAkamaiFastDNSRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceAkamaiFastDNSRecordCreate,
		Read:   resourceAkamaiFastDNSRecordRead,
		Update: resourceAkamaiFastDNSRecordUpdate,
		Delete: resourceAkamaiFastDNSRecordDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"rdata": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"zone": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAkamaiFastDNSRecordCreate(d *schema.ResourceData, m interface{}) error {
	conn := m.(*AkamaiClient).client
	zone := d.Get("zone").(string)

	zoneRecord, _, err := conn.FastDNSv2.GetZone(context.Background(), zone)
	if err != nil {
		return err
	}

	if zoneRecord.Zone == nil {
		return fmt.Errorf("No Akamai Zone found for id (%s)", zone)
	}

	// build the record
	en := expandRecordName(d.Get("name").(string), *zoneRecord.Zone)

	rec := &akamai.RecordSetCreateRequest{
		Zone: d.Get("zone").(string),
		Name: en,
		Type: d.Get("type").(string),
		TTL:  d.Get("ttl").(int),
	}

	// add the resource records
	if v, ok := d.GetOk("rdata"); ok {
		recs := v.([]interface{})
		rec.Rdata = expandResourceRecords(recs, d.Get("type").(string))
	}

	_, err = createFastDNSRecord(conn, rec)
	if err != nil {
		return err
	}

	// generate an ID to use
	vars := []string{
		zone,
		strings.ToLower(d.Get("name").(string)),
		d.Get("type").(string),
	}
	d.SetId(strings.Join(vars, "_"))
	return nil
}

func createFastDNSRecord(conn *akamai.Client, rec *akamai.RecordSetCreateRequest) (interface{}, error) {
	wait := resource.StateChangeConf{
		Pending:    []string{"rejected"},
		Target:     []string{"accepted"},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {

			output, resp, err := conn.FastDNSv2.CreateRecordSet(context.Background(), rec)
			if resp.StatusCode == 500 || resp.StatusCode == 503 {
				return 42, "rejected", nil
			}

			if err != nil {
				e := fmt.Errorf("[ERR]: Error creating record set: %s", err)
				return 42, "failure", e

			}

			return output, "accepted", nil
		},
	}
	return wait.WaitForState()
}
func resourceAkamaiFastDNSRecordRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAkamaiFastDNSRecordUpdate(d *schema.ResourceData, m interface{}) error {
	// If the type or name of the record has changed
	// we want to create a new record.
	if d.HasChange("type") || d.HasChange("name") {
		return resourceAkamaiFastDNSRecordCreate(d, m)
	}

	// Otherwise, continue to PUT a new record.
	conn := m.(*AkamaiClient).client
	zone := d.Get("zone").(string)

	zoneRecord, _, err := conn.FastDNSv2.GetZone(context.Background(), zone)
	if err != nil {
		return err
	}

	if zoneRecord.Zone == nil {
		return fmt.Errorf("No Akamai Zone found for id (%s)", zone)
	}

	// build the record
	en := expandRecordName(d.Get("name").(string), *zoneRecord.Zone)

	rec := &akamai.RecordSetCreateRequest{
		Zone: d.Get("zone").(string),
		Name: en,
		Type: d.Get("type").(string),
		TTL:  d.Get("ttl").(int),
	}

	// add the resource records
	if v, ok := d.GetOk("rdata"); ok {
		recs := v.([]interface{})
		rec.Rdata = expandResourceRecords(recs, d.Get("type").(string))
	}

	log.Printf("[DEBUG] Updating resource records for zone: %s, name : %s", zone, rec.Name)

	// Update the record
	_, resp, err := conn.FastDNSv2.UpdateRecordSet(context.Background(), rec)
	if err != nil {
		return fmt.Errorf("[ERR]: Error updating record set: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("[ERR]: Could not update record set: HTTP %s", resp.Status)
	}

	// generate the ID
	vars := []string{
		zone,
		strings.ToLower(d.Get("name").(string)),
		d.Get("type").(string),
	}
	d.SetId(strings.Join(vars, "_"))

	return nil
}

func resourceAkamaiFastDNSRecordDelete(d *schema.ResourceData, m interface{}) error {
	conn := m.(*AkamaiClient).client
	zone := d.Get("zone").(string)

	zoneRecord, _, err := conn.FastDNSv2.GetZone(context.Background(), zone)
	if err != nil {
		return err
	}

	if zoneRecord.Zone == nil {
		return fmt.Errorf("No Akamai Zone found for id (%s)", zone)
	}

	// build the record
	en := expandRecordName(d.Get("name").(string), *zoneRecord.Zone)

	input := &akamai.RecordSetOptions{
		Zone: zone,
		Name: en,
		Type: d.Get("type").(string),
	}

	// delete the record
	resp, err := conn.FastDNSv2.DeleteRecordSet(context.Background(), input)
	if err != nil {
		return fmt.Errorf("[ERR]: Error deleting record set: %s", err)
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("[ERR]: Could no delete record set: HTTP %s", resp.Status)
	}

	return nil
}

// Check if the current record name contains the zone suffix.
// If it does not, add the zone name to form a fully qualified name.
func expandRecordName(name, zone string) string {
	rn := strings.ToLower(strings.TrimSuffix(name, "."))
	zone = strings.TrimSuffix(zone, ".")
	if !strings.HasSuffix(rn, zone) {
		if len(name) == 0 {
			rn = zone
		} else {
			rn = strings.Join([]string{rn, zone}, ".")
		}
	}
	return rn
}

// expandResourceRecords will take the records from the schema and
// return a valid []string of records.
func expandResourceRecords(recs []interface{}, typeStr string) []string {
	records := make([]string, 0, len(recs))
	for _, r := range recs {
		s := r.(string)
		// here we can clean for TXT and SPF records
		records = append(records, s)
	}
	return records
}
