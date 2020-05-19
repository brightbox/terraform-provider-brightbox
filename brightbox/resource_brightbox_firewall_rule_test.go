package brightbox

import (
	"fmt"
	"testing"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxFirewallRule_Basic(t *testing.T) {
	var firewallRule brightbox.FirewallRule
	var firewallPolicy brightbox.FirewallPolicy

	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updatedName := fmt.Sprintf("bar-%d", rInt)
	resourceName := "brightbox_firewall_rule.rule1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleAndPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists(resourceName, &firewallRule),
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.terraform", &firewallPolicy),
					testAccCheckBrightboxEmptyFirewallRuleAttributes(&firewallRule, name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
					resource.TestCheckResourceAttrPtr(
						resourceName, "firewall_policy", &firewallPolicy.Id),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists(resourceName, &firewallRule),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxFirewallRule_clear_names(t *testing.T) {
	var firewall_rule brightbox.FirewallRule
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_firewall_rule.rule1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleAndPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists(resourceName, &firewall_rule),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists(resourceName, &firewall_rule),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBrightboxFirewallRuleAndPolicyDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxFirewallRuleDestroy(s)
	if err != nil {
		return err
	}
	return testAccCheckBrightboxFirewallPolicyDestroy(s)
}

func testAccCheckBrightboxFirewallRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_firewall_rule" {
			continue
		}

		// Try to find the FirewallRule
		_, err := client.FirewallRule(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for firewall_rule %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxFirewallRuleExists(n string, firewall_policy *brightbox.FirewallRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No FirewallRule ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the FirewallRule
		retrieveFirewallRule, err := client.FirewallRule(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveFirewallRule.Id != rs.Primary.ID {
			return fmt.Errorf("FirewallRule not found")
		}

		*firewall_policy = *retrieveFirewallRule

		return nil
	}
}

func testAccCheckBrightboxEmptyFirewallRuleAttributes(firewall_policy *brightbox.FirewallRule, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if firewall_policy.Description != name {
			return fmt.Errorf("Bad description: %s", firewall_policy.Description)
		}
		return nil
	}
}

func testAccCheckBrightboxFirewallRuleConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = "${brightbox_firewall_policy.terraform.id}"
	description = "foo-%d"
	destination = "any"
}

`, rInt)
}

func testAccCheckBrightboxFirewallRuleConfig_updated(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = "${brightbox_firewall_policy.terraform.id}"
	description = "bar-%d"
	destination = "any"
}

`, rInt)
}

const testAccCheckBrightboxFirewallRuleConfig_empty = `

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = "${brightbox_firewall_policy.terraform.id}"
	description = ""
	destination = "any"
}

`
