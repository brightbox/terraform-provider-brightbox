package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBrightboxServerType_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxServerTypeConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDataSourceID("Server Type", "data.brightbox_server_type.foobar"),
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

const TestAccBrightboxServerTypeConfig_basic = `
data "brightbox_server_type" "foobar" {
	handle = "^4gb.nbs$"
}
`
