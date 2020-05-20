package brightbox

import (
	"fmt"
	"log"
	"testing"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxFirewallPolicy_Basic(t *testing.T) {
	var firewallPolicy brightbox.FirewallPolicy
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_firewall_policy.foobar"
	updatedName := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					testAccCheckBrightboxFirewallPolicyAttributes(&firewallPolicy, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
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

func TestAccBrightboxFirewallPolicy_clear_names(t *testing.T) {
	var firewallPolicy brightbox.FirewallPolicy
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_firewall_policy.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					testAccCheckBrightboxFirewallPolicyAttributes(&firewallPolicy, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
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

func TestAccBrightboxFirewallPolicy_Mapped(t *testing.T) {
	resourceName := "brightbox_firewall_policy.foobar"
	rInt := acctest.RandInt()
	var firewallPolicy brightbox.FirewallPolicy
	var serverGroup brightbox.ServerGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group1", &serverGroup),
					resource.TestCheckResourceAttrPtr(
						resourceName, "server_group", &serverGroup.Id),
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

func TestAccBrightboxFirewallPolicy_mappings(t *testing.T) {
	var firewallPolicy brightbox.FirewallPolicy
	var serverGroup brightbox.ServerGroup
	rInt := acctest.RandInt()
	resourceName := "brightbox_firewall_policy.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxFirewallPolicyAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group1", &serverGroup),
					resource.TestCheckResourceAttrPtr(
						resourceName, "server_group", &serverGroup.Id),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_remap(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.group2", &serverGroup),
					resource.TestCheckResourceAttrPtr(
						resourceName, "server_group", &serverGroup.Id),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_unmap(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(
						resourceName, "server_group", ""),
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
	client := testAccProvider.Meta().(*CompositeClient).APIClient

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
					"Error waiting for firewallPolicy %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxFirewallPolicyExists(n string, firewallPolicy *brightbox.FirewallPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No FirewallPolicy ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the FirewallPolicy
		retrieveFirewallPolicy, err := client.FirewallPolicy(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveFirewallPolicy.Id != rs.Primary.ID {
			return fmt.Errorf("FirewallPolicy not found")
		}

		*firewallPolicy = *retrieveFirewallPolicy

		return nil
	}
}

func testAccCheckBrightboxFirewallPolicyAttributes(firewallPolicy *brightbox.FirewallPolicy, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if firewallPolicy.Name != name {
			return fmt.Errorf("Bad name: %s", firewallPolicy.Name)
		}
		if firewallPolicy.Description != name {
			return fmt.Errorf("Bad description: %s", firewallPolicy.Description)
		}
		if firewallPolicy.Default != false {
			return fmt.Errorf("Bad default: %v", firewallPolicy.Default)
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

resource "brightbox_server_group" "group1" {
	name = "foo-%d"
}

resource "brightbox_server_group" "group2" {
	name = "bar-%d"
}

resource "brightbox_firewall_policy" "foobar" {
	name = "foo-%d"
	description = "foo-%d"
	server_group = "${brightbox_server_group.group1.id}"
}

`, rInt, rInt, rInt, rInt)
}

func testAccCheckBrightboxFirewallPolicyConfig_remap(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_server_group" "group1" {
	name = "foo-%d"
}

resource "brightbox_server_group" "group2" {
	name = "bar-%d"
}

resource "brightbox_firewall_policy" "foobar" {
	name = "bar-%d"
	description = "bar-%d"
	server_group = "${brightbox_server_group.group2.id}"
}

`, rInt, rInt, rInt, rInt)
}

func testAccCheckBrightboxFirewallPolicyConfig_unmap(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_firewall_policy" "foobar" {
	name = "baz-%d"
	description = "baz-%d"
}
`, rInt, rInt)
}

// Sweeper

func init() {
	resource.AddTestSweepers("firewall_policy", &resource.Sweeper{
		Name: "firewall_policy",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			objects, err := client.APIClient.FirewallPolicies()
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Default {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.Id, object.Name)
					if err := client.APIClient.DestroyFirewallPolicy(object.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.Id, err)
					}
				}
			}
			return nil
		},
	})
}
