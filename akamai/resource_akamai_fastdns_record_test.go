package akamai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/trussworks/akamai-sdk-go/akamai"
)

func TestExpandRecordName(t *testing.T) {
	cases := []struct {
		Input, Output string
	}{
		{"www", "www.porchetta.io"},
		{"www.", "www.porchetta.io"},
		{"dev.www", "dev.www.porchetta.io"},
		{"*", "*.porchetta.io"},
		{"porchetta.io", "porchetta.io"},
		{"test.porchetta.io", "test.porchetta.io"},
		{"test.porchetta.io.", "test.porchetta.io"},
	}
	zoneName := "porchetta.io"
	for _, tc := range cases {
		actual := expandRecordName(tc.Input, zoneName)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}

	}

}

func TestAccAkamaiFastDNSRecord_basic(t *testing.T) {
	var record akamai.RecordSet

	resourceName := "akamai_fastdns_record.default"
	zoneName := fmt.Sprintf("testzone-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckFastDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSRecordConfig_basic(zoneName, "127.0.0.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSRecordExists(resourceName, &record),
				),
			},
		},
	})
}

func TestAccFastDNSRecord_multiple(t *testing.T) {
	var record1, record2, record3 akamai.RecordSet
	var zone1 akamai.ZoneMetadata

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSRecordMultipleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSZoneExists("akamai_fastdns_zone.multiple", &zone1),
					testAccCheckFastDNSRecordExists("akamai_fastdns_record.multiple.0", &record1),
					testAccCheckFastDNSRecordExists("akamai_fastdns_record.multiple.1", &record2),
					testAccCheckFastDNSRecordExists("akamai_fastdns_record.multiple.2", &record3),
				),
			},
		},
	})

}

func TestAccFastDNSRecord_modify(t *testing.T) {
	var record akamai.RecordSet

	resourceName := "akamai_fastdns_record.default"
	zoneName := fmt.Sprintf("testzone-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSRecordConfig_basic(zoneName, "127.0.0.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSRecordExists(resourceName, &record),
				),
			},
			{
				Config: testAccFastDNSRecordConfig_basic(zoneName, "127.0.0.11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSRecordExists(resourceName, &record),
				),
			},
		},
	})
}

func TestAccFastDNSRecord_cname(t *testing.T) {
	var record akamai.RecordSet

	resourceName := "akamai_fastdns_record.default"
	zoneName := fmt.Sprintf("testzone-cname-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSRecordConfig_cname(zoneName, "testservice-dev"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSRecordExists(resourceName, &record),
				),
			},
		},
	})
}

func TestAccFastDNSRecord_txt(t *testing.T) {
	var record akamai.RecordSet

	resourceName := "akamai_fastdns_record.default"
	zoneName := fmt.Sprintf("testzone-txt-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSRecordConfig_txt(zoneName, "txtrecordtoadd"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSRecordExists(resourceName, &record),
				),
			},
		},
	})
}

func testAccCheckFastDNSRecordDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AkamaiClient).client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "akamai_fastdns_record" {
			continue
		}

		parts := parseRecordId(rs.Primary.ID)
		zone, name, rType := parts[0], parts[1], parts[2]

		en := expandRecordName(name, "akamaiexample.com")

		ars := &akamai.RecordSetOptions{
			Zone: zone,
			Name: en,
			Type: rType,
		}

		r, _, err := conn.FastDNSv2.GetRecordSet(context.Background(), ars)
		if err != nil {
			if akamaiErr, ok := err.(*akamai.AkamaiError); ok {
				if akamaiErr.Status == 404 {
					// if it can't be found, it doesn't exist
					return nil
				}

			}
			return fmt.Errorf("could not find record: %v", err)
		}

		if *r.Type == rType && strings.ToLower(*r.Name) == strings.ToLower(en) {
			return fmt.Errorf("Record still exists: %v", r)
		}
	}

	return nil

}

func testAccCheckFastDNSRecordExists(n string, record *akamai.RecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AkamaiClient).client
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		parts := parseRecordId(rs.Primary.ID)
		zone, name, rType := parts[0], parts[1], parts[2]
		en := expandRecordName(name, zone)

		ars := &akamai.RecordSetOptions{
			Zone: zone,
			Name: en,
			Type: rType,
		}

		r, _, err := conn.FastDNSv2.GetRecordSet(context.Background(), ars)
		if err != nil {
			return fmt.Errorf("Record err: %v", err)
		}

		if *r.Type == rType && strings.ToLower(*r.Name) == strings.ToLower(en) {
			*record = *r
			return nil
		}

		return fmt.Errorf("Record does not exist: %v", rs.Primary.ID)
	}
}

func testAccFastDNSRecordConfig_basic(zone, ip string) string {
	return fmt.Sprintf(`
resource "akamai_fastdns_zone" "main" {
  zone = "%s"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}

resource "akamai_fastdns_record" "default" {
  zone = "${akamai_fastdns_zone.main.zone}"
  name = "www"
  type = "A"
  ttl = "30"
  rdata = ["%s"]
}
`, zone, ip)
}

const testAccFastDNSRecordMultipleConfig = `
resource "akamai_fastdns_zone" "multiple" {
  zone = "multiple.akamaiexample.com"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}

resource "akamai_fastdns_record" "multiple" {
  count = 3

  rdata = ["127.0.0.${count.index}"]
  name = "record${count.index}"
  type = "A"
  zone = "${akamai_fastdns_zone.multiple.zone}"
  ttl = "30"
}
`

func testAccFastDNSRecordConfig_cname(zone, record string) string {
	return fmt.Sprintf(`
resource "akamai_fastdns_zone" "main" {
  zone = "%s"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}

resource "akamai_fastdns_record" "default" {
  zone = "${akamai_fastdns_zone.main.zone}"
  type = "CNAME"
  name = "%s"
  ttl = "30"
  rdata = ["%s.%s"]
}
`, zone, record, record, zone)
}

func testAccFastDNSRecordConfig_txt(zone, record string) string {
	return fmt.Sprintf(`
resource "akamai_fastdns_zone" "main" {
  zone = "%s"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}

resource "akamai_fastdns_record" "default" {
  zone = "${akamai_fastdns_zone.main.zone}"
  type = "TXT"
  name = "%s"
  ttl = "30"
  rdata = ["%s"]
}
`, zone, record, record)
}
