package akamai

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAkamaiFastDNSZone(t *testing.T) {
	rInt := acctest.RandInt()
	publicResourceName := "akamai_fastdns_zone.test"
	publicDomain := fmt.Sprintf("akamaiterraformtestacc-%d.com", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAkamaiFastDNSZoneConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAkamaiFastDNSZoneCheck(
						publicResourceName, "data.akamai_fastdns_zone.by_zone", publicDomain),
				),
			},
		},
	})
}

// rsName for the name of the created resource
// dsName for the name of the created data source
// zName for the name of the zone
func testAccDataSourceAkamaiFastDNSZoneCheck(rsName, dsName, zName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rsName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", rsName)
		}

		hostedZone, ok := s.RootModule().Resources[dsName]
		if !ok {
			return fmt.Errorf("can't find zone %q in state", dsName)
		}

		attr := rs.Primary.Attributes
		if attr["id"] != hostedZone.Primary.Attributes["zone"] {
			return fmt.Errorf("Akamai FastDNS Zone id is %s; want %s", attr["id"], hostedZone.Primary.Attributes["zone"])
		}

		if attr["zone"] != zName {
			return fmt.Errorf("Akamai FastDNS Zone name is %q; want %q", attr["zone"], zName)
		}

		return nil
	}
}

func testAccDataSourceAkamaiFastDNSZoneConfig(rInt int) string {
	return fmt.Sprintf(`
resource "akamai_fastdns_zone" "test" {
  zone = "akamaiterraformtestacc-%d.com"
  contract_id = "G-2LP9RJ3"
  type = "PRIMARY"
}

data "akamai_fastdns_zone" "by_zone" {
  zone = "${akamai_fastdns_zone.test.zone}"
}
`, rInt)
}
