package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBrightboxDatabaseType_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxDatabaseTypeConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDataSourceID("Database Type", "data.brightbox_database_type.foobar"),
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

const TestAccBrightboxDatabaseTypeConfig_basic = `
data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
`
