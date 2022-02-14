package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBrightboxServerType_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxServerTypeConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerTypeDataSourceID("data.brightbox_server_type.foobar"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_type.foobar", "handle", "4gb.nbs"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_type.foobar", "ram", "4096"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_type.foobar", "disk_size", "0"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_type.foobar", "storage_type", "network"),
				),
			},
		},
	})
}

func testAccCheckServerTypeDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find server type data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Server type data source ID not set")
		}

		return nil
	}
}

const TestAccBrightboxServerTypeConfig_basic = `
data "brightbox_server_type" "foobar" {
	handle = "^4gb.nbs$"
}
`
