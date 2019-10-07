package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxDataServerGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxDataServerGroupConfig_default,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataServerGroupDataSourceID("data.brightbox_server_group.default"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_group.default", "name", "default"),
				),
			},
		},
	})
}

func testAccCheckDataServerGroupDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find server group data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Server Group data source ID not set")
		}

		return nil
	}
}

const TestAccBrightboxDataServerGroupConfig_default = `
data "brightbox_server_group" "default" {
	name = "^default$"
}
`
