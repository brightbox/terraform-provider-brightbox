package brightbox

import (
	"context"
	"fmt"
	"log"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBrightboxFirewallPolicy_Basic(t *testing.T) {
	var firewallPolicy brightbox.FirewallPolicy
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_firewall_policy.foobar"
	updatedName := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "server_group", &serverGroup.ID),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxFirewallPolicyAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "server_group", &serverGroup.ID),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_remap(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "server_group", &serverGroup.ID),
				),
			},
			{
				Config: testAccCheckBrightboxFirewallPolicyConfig_unmap(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Firewall Policy",
						&firewallPolicy,
						(*brightbox.Client).FirewallPolicy,
					),
					resource.TestCheckResourceAttr(
						resourceName, "server_group", ""),
				),
			},
		},
	})
}

var testAccCheckBrightboxFirewallPolicyDestroy = testAccCheckBrightboxDestroyBuilder(
	"brightbox_firewall_policy",
	(*brightbox.Client).FirewallPolicy,
)

func testAccCheckBrightboxFirewallPolicyAndGroupDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxFirewallPolicyDestroy(s)
	if err != nil {
		return err
	}
	return testAccCheckBrightboxServerGroupDestroy(s)
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			objects, err := client.APIClient.FirewallPolicies(ctx)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Default {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.ID, object.Name)
					if _, err := client.APIClient.DestroyFirewallPolicy(ctx, object.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.ID, err)
					}
				}
			}
			return nil
		},
	})
}
