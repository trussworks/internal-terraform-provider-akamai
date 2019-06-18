package akamai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/trussworks/akamai-sdk-go/akamai"
)

var akamaiNoRecordFound = errors.New("No matching record found.")

func resourceAkamaiFastDNSRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceAkamaiFastDNSRecordCreate,
		Read:   resourceAkamaiFastDNSRecordRead,
		Update: resourceAkamaiFastDNSRecordUpdate,
		Delete: resourceAkamaiFastDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
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

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
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
			// This is for intermittent Akamai 5xx errors that often occur
			// The 409 check is for when multiple records are created at once, sometimes Akamai
			// throws when the Zone is modified too quickly.
			if resp.StatusCode == 500 || resp.StatusCode == 503 || resp.StatusCode == 409 {
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
	// If we don't have a zone ID we're doing an import. Parse it from the ID.
	if _, ok := d.GetOk("zone"); !ok {
		parts := parseRecordId(d.Id())
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return fmt.Errorf("Error importing akamai_fastdns_record. Please make sure the record ID is in the form ZONEID_RECORDNAME_TYPE.")
		}

		d.Set("zone", parts[0])
		d.Set("name", parts[1])
		d.Set("type", parts[2])
	}

	record, err := findRecord(d, m)
	if err != nil {
		switch err {
		case akamaiNoRecordFound:
			log.Printf("[DEBUG] %s for: %s, removing from state file", err, d.Id())
			d.SetId("")
			return nil
		default:
			return err

		}
	}

	//rdata := cleanResourceRecords(record.Rdata, *record.Type)
	//d.Set("rdata", rdata)
	d.Set("ttl", record.TTL)

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
	_, err = deleteFastDNSRecord(conn, input)
	if err != nil {
		return fmt.Errorf("[ERR]: Error deleting record set: %s", err)
	}

	return nil
}

func deleteFastDNSRecord(conn *akamai.Client, rs *akamai.RecordSetOptions) (interface{}, error) {
	wait := resource.StateChangeConf{
		Pending:    []string{"rejected"},
		Target:     []string{"accepted"},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.FastDNSv2.DeleteRecordSet(context.Background(), rs)
			if resp.StatusCode == 409 {
				// when deleting multiple records we can sometimes get a Concurrent Zone Modification Error
				return 42, "rejected", nil

			}

			if err != nil {
				e := fmt.Errorf("error deleting Akamai FastDNS record (%s) error: %s", rs.Name, err)

				return 42, "failure", e
			}
			return resp, "accepted", nil
		},
	}
	return wait.WaitForState()
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
		if typeStr == "TXT" || typeStr == "SPF" {
			s = flattenTxtEntry(s)
		}

		if typeStr == "CNAME" {
			s = strings.TrimSuffix(s, ".")
		}

		records = append(records, s)
	}
	return records
}

// parseRecordId takes the ID which we use to store a record in Terraform and
// returns back the zone, name, and record type
func parseRecordId(id string) [3]string {
	var recZone, recName, recType string
	parts := strings.Split(id, "_")
	recZone, recName, recType = parts[0], parts[1], parts[2]

	recName = strings.TrimSuffix(recName, ".")

	return [3]string{recZone, recName, recType}
}

// findRecord takes a ResourceData struct for akamai_fastdns_record.
// It then queries Akamai for the information on its records.
func findRecord(d *schema.ResourceData, meta interface{}) (*akamai.RecordSet, error) {
	conn := meta.(*AkamaiClient).client

	zone := d.Get("zone").(string)
	en := expandRecordName(d.Get("name").(string), zone)

	log.Printf("[DEBUG] Expanded record name: %s", en)
	d.Set("fqdn", en)

	recordType := d.Get("type").(string)

	rso := &akamai.RecordSetOptions{
		Zone: zone,
		Name: en,
		Type: recordType,
	}

	rs, resp, err := conn.FastDNSv2.GetRecordSet(context.Background(), rso)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, akamaiNoRecordFound
	}

	return rs, err
}

func cleanResourceRecords(recs []*string, typeStr string) []string {
	strs := make([]string, 0, len(recs))

	for _, r := range recs {
		if r != nil {
			s := *r
			if typeStr == "TXT" || typeStr == "SPF" {
				s = expandTxtEntry(s)
			}
			strs = append(strs, s)
		}
	}
	return strs
}

func flattenTxtEntry(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}

func expandTxtEntry(s string) string {
	last := len(s) - 1
	if last != 0 && s[0] == '"' && s[last] == '"' {
		s = s[1:last]
	}
	return s
}
