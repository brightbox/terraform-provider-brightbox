package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxFirewallPolicy_Basic(t *testing.T) {
	var firewall_policy brightbox.FirewallPolicy
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updated_name := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxFirewallPolicyAttributes(&firewall_policy, name),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", updated_name),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", updated_name),
				),
			},
		},
	})
}

func TestAccBrightboxFirewallPolicy_clear_names(t *testing.T) {
	var firewall_policy brightbox.FirewallPolicy
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxFirewallPolicyAttributes(&firewall_policy, name),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_firewall_policy.foobar", "description", name),
				),
			},
			{
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
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group1", &server_group),
					resource.TestCheckResourceAttrPtr(
						"brightbox_firewall_policy.foobar", "server_group", &server_group.Id),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_remap(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists("brightbox_firewall_policy.foobar", &firewall_policy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group2", &server_group),
					resource.TestCheckResourceAttrPtr(
						"brightbox_firewall_policy.foobar", "server_group", &server_group.Id),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_unmap(rInt),
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
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

func testAccCheckBrightboxFirewallPolicyAttributes(firewall_policy *brightbox.FirewallPolicy, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if firewall_policy.Name != name {
			return fmt.Errorf("Bad name: %s", firewall_policy.Name)
		}
		if firewall_policy.Description != name {
			return fmt.Errorf("Bad description: %s", firewall_policy.Description)
		}
		if firewall_policy.Default != false {
			return fmt.Errorf("Bad default: %v", firewall_policy.Default)
		}
		return nil
	}
}

func testAccCheckBrightboxFirewallPolicyConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "foobar" {
	name = "foo-%d"
	description = "foo-%d"
}
`, rInt, rInt)
}

func testAccCheckBrightboxFirewallPolicyConfig_updated(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "foobar" {
	name = "bar-%d"
	description = "bar-%d"
}
`, rInt, rInt)
}

const testAccCheckBrightboxFirewallPolicyConfig_empty = `

resource "brightbox_firewall_policy" "foobar" {
	name = ""
	description = ""
}
`

func testAccCheckBrightboxFirewallPolicyConfig_mapped(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "foobar" {
	name = "foo-%d"
	description = "foo-%d"
	server_group = "${brightbox_server_group.group1.id}"
}

resource "brightbox_server_group" "group1" {
	name = "foo-%d"
}
`, rInt, rInt, rInt)
}

func testAccCheckBrightboxFirewallPolicyConfig_remap(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "foobar" {
	name = "bar-%d"
	description = "bar-%d"
	server_group = "${brightbox_server_group.group2.id}"
}

resource "brightbox_server_group" "group1" {
	name = "foo-%d"
}

resource "brightbox_server_group" "group2" {
	name = "bar-%d"
}
`, rInt, rInt, rInt, rInt)
}

func testAccCheckBrightboxFirewallPolicyConfig_unmap(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "foobar" {
	name = "baz-%d"
	description = "baz-%d"
	server_group = ""
}
`, rInt, rInt)
}
