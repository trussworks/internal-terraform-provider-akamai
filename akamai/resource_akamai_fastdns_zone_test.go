package akamai

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/trussworks/akamai-sdk-go/akamai"
)

func TestAccAkamaiFastDNSZone_basic(t *testing.T) {
	var zone akamai.ZoneMetadata

	rString := acctest.RandString(8)
	resourceName := "akamai_fastdns_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSZoneConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "zone", fmt.Sprintf("%s", zoneName)),
				),
			},
		},
	})
}

func TestAccAkamaiFastDNSZone_disappears(t *testing.T) {
	var zone akamai.ZoneMetadata

	rString := acctest.RandString(8)
	resourceName := "akamai_fastdns_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSZoneConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastDNSZoneExists(resourceName, &zone),
					testAccCheckFastDNSZoneDisappears(&zone),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})

}

func TestAccAkamaiFastDNSZone_comment(t *testing.T) {
	var zone akamai.ZoneMetadata

	rString := acctest.RandString(8)
	resourceName := "akamai_fastdns_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSZoneConfigComment(zoneName, "comment1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFastDNSZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "comment", "comment1"),
				),
			},
			{
				Config: testAccFastDNSZoneConfigComment(zoneName, "comment2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFastDNSZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "comment", "comment2"),
				),
			},
		},
	})

}

func TestAccAkamaiFastDNSZone_updates(t *testing.T) {
	var zone akamai.ZoneMetadata

	resourceName := "akamai_fastdns_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastDNSZoneConfigCommentInitial,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFastDNSZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "comment", "Managed by Terraform"),
				),
			},
			// cause a change, which will trigger an update
			{
				Config: testAccFastDNSZoneConfigCommentUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFastDNSZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "comment", "updated comment"),
				),
			},
		},
	})

}

func testAccCheckFastDNSZoneDisappears(zone *akamai.ZoneMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AkamaiClient).client

		output, err := deleteFastDNSZone(conn, *zone.Zone, false)
		if err != nil {
			return err
		}

		rid := output.(*akamai.ZoneDeleteResponse).RequestID

		_, err = checkDeleteFastDNSZone(conn, *rid)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckFastDNSZoneDestroy(s *terraform.State) error {
	return testAccCheckFastDNSZoneDestroyWithProvider(s, testAccProvider)
}

func testAccCheckFastDNSZoneDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AkamaiClient).client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "akamai_fastdns_zone" {
			continue
		}

		zid := rs.Primary.ID
		_, _, err := conn.FastDNSv2.GetZone(context.Background(), zid)
		if err == nil {
			return fmt.Errorf("Hosted zone still exists")
		}
	}
	return nil
}

func testAccCheckFastDNSZoneExists(n string, zone *akamai.ZoneMetadata) resource.TestCheckFunc {
	return testAccCheckFastDNSZoneExistsWithProvider(n, zone, func() *schema.Provider { return testAccProvider })
}

func testAccCheckFastDNSZoneExistsWithProvider(n string, zone *akamai.ZoneMetadata, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No zone ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AkamaiClient).client

		z, _, err := conn.FastDNSv2.GetZone(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Hosted zone err: %v", err)
		}

		akamai_comment := *z.Comment
		rs_comment := rs.Primary.Attributes["comment"]
		if rs_comment != "" && akamai_comment != rs_comment {
			return fmt.Errorf("Hosted zone with comment '%s' found but does not match '%s'", akamai_comment, rs_comment)
		}

		if rs.Primary.ID != *z.Zone {
			return fmt.Errorf("Got: %v, Expected: %v", rs.Primary.ID, *z.Zone)
		}

		*zone = *z
		return nil
	}
}

func testAccFastDNSZoneConfig(zoneName string) string {
	return fmt.Sprintf(`
resource "akamai_fastdns_zone" "test" {
  zone = "%s"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}
`, zoneName)
}

func testAccFastDNSZoneConfigComment(zoneName, comment string) string {
	return fmt.Sprintf(`
resource "akamai_fastdns_zone" "test" {
  zone = "%s"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
  comment = %q
}
`, zoneName, comment)
}

const testAccFastDNSZoneConfigCommentInitial = `
resource "akamai_fastdns_zone" "test" {
  zone = "zoneconfig.akamaiexample.com"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}`

const testAccFastDNSZoneConfigCommentUpdate = `
resource "akamai_fastdns_zone" "test" {
  zone = "zoneconfig.akamaiexample.com"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
  comment = "updated comment"
}
`
