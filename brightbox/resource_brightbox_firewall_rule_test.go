package brightbox

import (
	"errors"
	"fmt"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Rule",
						&firewallRule,
						(*brightbox.Client).FirewallRule,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_firewall_policy.terraform",
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					testAccCheckBrightboxEmptyFirewallRuleAttributes(&firewallRule, name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
					resource.TestCheckResourceAttrPtr(
						resourceName, "firewall_policy", &firewallPolicy.ID),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Rule",
						&firewallRule,
						(*brightbox.Client).FirewallRule,
					),
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
	var firewallRule brightbox.FirewallRule
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Rule",
						&firewallRule,
						(*brightbox.Client).FirewallRule,
					),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Rule",
						&firewallRule,
						(*brightbox.Client).FirewallRule,
					),
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

func TestAccBrightboxFirewallRule_mappings(t *testing.T) {
	var firewallRule brightbox.FirewallRule
	var firewallPolicy brightbox.FirewallPolicy
	rInt := acctest.RandInt()
	resourceName := "brightbox_firewall_rule.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallRuleAndPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Rule",
						&firewallRule,
						(*brightbox.Client).FirewallRule,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_firewall_policy.policy1",
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "firewall_policy", &firewallPolicy.ID),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallRuleConfig_remap(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Rule",
						&firewallRule,
						(*brightbox.Client).FirewallRule,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_firewall_policy.policy2",
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "firewall_policy", &firewallPolicy.ID),
				),
			},
		},
	})
}

func testAccCheckBrightboxFirewallRuleAndPolicyDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxDestroyBuilder(
		"Firewall Policy",
		(*brightbox.Client).DestroyFirewallPolicy,
	)(s)
	if err != nil {
		var apierror *brightbox.APIError
		if errors.As(err, &apierror) {
			if apierror.StatusCode != 404 {
				return err
			}
		}
	}
	return nil
}

func testAccCheckBrightboxEmptyFirewallRuleAttributes(firewallPolicy *brightbox.FirewallRule, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if firewallPolicy.Description != name {
			return fmt.Errorf("Bad description: %s", firewallPolicy.Description)
		}
		return nil
	}
}

func testAccCheckBrightboxFirewallRuleConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = brightbox_firewall_policy.terraform.id
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
	firewall_policy = brightbox_firewall_policy.terraform.id
	description = "bar-%d"
	destination = "any"
}

`, rInt)
}

const testAccCheckBrightboxFirewallRuleConfig_empty = `

resource "brightbox_firewall_policy" "terraform" {
}

resource "brightbox_firewall_rule" "rule1" {
	firewall_policy = brightbox_firewall_policy.terraform.id
	description = ""
	destination = "any"
}

`

func testAccCheckBrightboxFirewallRuleConfig_mapped(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "policy1" {
	name = "foo-%d"
}

resource "brightbox_firewall_policy" "policy2" {
	name = "bar-%d"
}

resource "brightbox_firewall_rule" "foobar" {
	description = "foo-%d"
	firewall_policy = brightbox_firewall_policy.policy1.id
	destination = "any"
}

`, rInt, rInt, rInt)
}

func testAccCheckBrightboxFirewallRuleConfig_remap(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "policy1" {
	name = "foo-%d"
}

resource "brightbox_firewall_policy" "policy2" {
	name = "bar-%d"
}

resource "brightbox_firewall_rule" "foobar" {
	description = "bar-%d"
	firewall_policy = brightbox_firewall_policy.policy2.id
	destination = "any"
}

`, rInt, rInt, rInt)
}
