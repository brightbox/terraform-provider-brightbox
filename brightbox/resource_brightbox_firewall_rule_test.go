package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxFirewallRule_Basic(t *testing.T) {
	var firewall_rule brightbox.FirewallRule
	var firewall_policy brightbox.FirewallPolicy

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleAndPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallRuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists("brightbox_firewall_rule.rule1", &firewall_rule),
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.terraform", &firewall_policy),
					testAccCheckBrightboxEmptyFirewallRuleAttributes(&firewall_rule),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_rule.rule1", "description", "outbound_rule"),
					resource.TestCheckResourceAttrPtr(
						"brightbox_firewall_rule.rule1", "firewall_policy", &firewall_policy.Id),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallRuleConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists("brightbox_firewall_rule.rule1", &firewall_rule),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_rule.rule1", "description", "updated"),
				),
			},
		},
	})
}

func TestAccBrightboxFirewallRule_clear_names(t *testing.T) {
	var firewall_rule brightbox.FirewallRule

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleAndPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallRuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists("brightbox_firewall_rule.rule1", &firewall_rule),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_rule.rule1", "description", "outbound_rule"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallRuleConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallRuleExists("brightbox_firewall_rule.rule1", &firewall_rule),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_rule.rule1", "description", ""),
				),
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
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

func testAccCheckBrightboxEmptyFirewallRuleAttributes(firewall_policy *brightbox.FirewallRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if firewall_policy.Description != "outbound_rule" {
			return fmt.Errorf("Bad description: %s", firewall_policy.Description)
		}
		return nil
	}
}

const testAccCheckBrightboxFirewallRuleConfig_basic = `

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = "${brightbox_firewall_policy.terraform.id}"
	description = "outbound_rule"
	destination = "any"
}

`

const testAccCheckBrightboxFirewallRuleConfig_updated = `

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = "${brightbox_firewall_policy.terraform.id}"
	description = "updated"
	destination = "any"
}

`

const testAccCheckBrightboxFirewallRuleConfig_empty = `

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = "${brightbox_firewall_policy.terraform.id}"
	description = ""
	destination = "any"
}

`
