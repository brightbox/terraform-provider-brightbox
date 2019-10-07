package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxDatabaseType_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxDatabaseTypeConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseTypeDataSourceID("data.brightbox_database_type.foobar"),
					resource.TestCheckResourceAttr(
						"data.brightbox_database_type.foobar", "name", "SSD 4GB"),
					resource.TestCheckResourceAttr(
						"data.brightbox_database_type.foobar", "ram", "4096"),
					resource.TestCheckResourceAttr(
						"data.brightbox_database_type.foobar", "disk_size", "61440"),
				),
			},
		},
	})
}

func testAccCheckDatabaseTypeDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find database type data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Database Type data source ID not set")
		}

		return nil
	}
}

const TestAccBrightboxDatabaseTypeConfig_basic = `
data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
`
