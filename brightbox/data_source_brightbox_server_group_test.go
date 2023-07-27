package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBrightboxDataServerGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: TestAccBrightboxDataServerGroupConfig_default,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDataSourceID("ServerGroup", "data.brightbox_server_group.default"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_group.default", "name", "default"),
					resource.TestCheckResourceAttr(
						"data.brightbox_server_group.default", "default", "true"),
					resource.TestMatchResourceAttr(
						"data.brightbox_server_group.default", "firewall_policy", firewallPolicyRegexp),
				),
			},
		},
	})
}

const TestAccBrightboxDataServerGroupConfig_default = `
data "brightbox_server_group" "default" {
	name = "^default$"
}
`
