package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccBrightboxFirewallRule_importBasic(t *testing.T) {
	resourceName := "brightbox_firewall_rule.rule1"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxFirewallRule_importEmpty(t *testing.T) {
	resourceName := "brightbox_firewall_rule.rule1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_empty,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
