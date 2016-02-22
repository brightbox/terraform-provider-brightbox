package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxFirewallPolicy_Basic(t *testing.T) {
	var firewall_policy brightbox.FirewallPolicy

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxFirewallPolicyAttributes(&firewall_policy),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", "empty"),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", "empty"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", "updated"),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", "updated"),
				),
			},
		},
	})
}

func TestAccBrightboxFirewallPolicy_clear_names(t *testing.T) {
	var firewall_policy brightbox.FirewallPolicy

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxFirewallPolicyAttributes(&firewall_policy),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", "empty"),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", "empty"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", ""),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", ""),
				),
			},
		},
	})
}

func TestAccBrightboxFirewallPolicy_mappings(t *testing.T) {
	var firewall_policy brightbox.FirewallPolicy
	var server_group brightbox.ServerGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyAndGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_mapped,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group1", &server_group),
					resource.TestCheckResourceAttrPtr(
						"brightbox_firewall_policy.foobar", "server_group", &server_group.Id),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_remap,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group2", &server_group),
					resource.TestCheckResourceAttrPtr(
						"brightbox_firewall_policy.foobar", "server_group", &server_group.Id),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxFirewallPolicyConfig_unmap,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "server_group", ""),
				),
			},
		},
	})
}

func testAccCheckBrightboxFirewallPolicyAndGroupDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxFirewallPolicyDestroy(s)
	if err != nil {
		return err
	}
	return testAccCheckBrightboxServerGroupDestroy(s)
}

func testAccCheckBrightboxFirewallPolicyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*brightbox.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_firewall_policy" {
			continue
		}

		// Try to find the FirewallPolicy
		_, err := client.FirewallPolicy(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for firewall_policy %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxFirewallPolicyExists(n string, firewall_policy *brightbox.FirewallPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No FirewallPolicy ID is set")
		}

		client := testAccProvider.Meta().(*brightbox.Client)

		// Try to find the FirewallPolicy
		retrieveFirewallPolicy, err := client.FirewallPolicy(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveFirewallPolicy.Id != rs.Primary.ID {
			return fmt.Errorf("FirewallPolicy not found")
		}

		*firewall_policy = *retrieveFirewallPolicy

		return nil
	}
}

func testAccCheckBrightboxFirewallPolicyAttributes(firewall_policy *brightbox.FirewallPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if firewall_policy.Name != "empty" {
			return fmt.Errorf("Bad name: %s", firewall_policy.Name)
		}
		if firewall_policy.Description != "empty" {
			return fmt.Errorf("Bad description: %s", firewall_policy.Description)
		}
		if firewall_policy.Default != false {
			return fmt.Errorf("Bad default: %v", firewall_policy.Default)
		}
		return nil
	}
}

const testAccCheckBrightboxFirewallPolicyConfig_basic = `

resource "brightbox_firewall_policy" "foobar" {
	name = "empty"
	description = "empty"
}
`

const testAccCheckBrightboxFirewallPolicyConfig_updated = `

resource "brightbox_firewall_policy" "foobar" {
	name = "updated"
	description = "updated"
}
`

const testAccCheckBrightboxFirewallPolicyConfig_empty = `

resource "brightbox_firewall_policy" "foobar" {
	name = ""
	description = ""
}
`

const testAccCheckBrightboxFirewallPolicyConfig_mapped = `

resource "brightbox_firewall_policy" "foobar" {
	name = "mapped"
	description = "mapped"
	server_group = "${brightbox_server_group.group1.id}"
}

resource "brightbox_server_group" "group1" {
	name = "Terraform"
}
`

const testAccCheckBrightboxFirewallPolicyConfig_remap = `

resource "brightbox_firewall_policy" "foobar" {
	name = "remapped"
	description = "remapped"
	server_group = "${brightbox_server_group.group2.id}"
}

resource "brightbox_server_group" "group1" {
	name = "Terraform"
}

resource "brightbox_server_group" "group2" {
	name = "Terraform"
}
`

const testAccCheckBrightboxFirewallPolicyConfig_unmap = `

resource "brightbox_firewall_policy" "foobar" {
	name = "unmapped"
	description = "unmapped"
	server_group = ""
}
`
