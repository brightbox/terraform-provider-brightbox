package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxDataServerGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxDataServerGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataServerGroupDataSourceID("data.brightbox_server_group.foobar"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_group.foobar", "name", "default"),
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

const TestAccBrightboxDataServerGroupConfig_basic = `
data "brightbox_server_group" "foobar" {
	name = "^default$"
}
`
