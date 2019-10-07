package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccBrightboxFirewallPolicy_importBasic(t *testing.T) {
	resourceName := "brightbox_firewall_policy.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxFirewallPolicy_importEmpty(t *testing.T) {
	resourceName := "brightbox_firewall_policy.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_empty,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
