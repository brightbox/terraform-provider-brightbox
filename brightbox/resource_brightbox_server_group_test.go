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

func TestAccBrightboxServerGroup_Basic(t *testing.T) {
	resourceName := "brightbox_server_group.foobar"
	var serverGroup brightbox.ServerGroup
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updatedName := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(resourceName, &serverGroup),
					testAccCheckBrightboxServerGroupAttributes(&serverGroup, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckBrightboxServerGroupConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(resourceName, &serverGroup),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
				),
			},
		},
	})
}

func TestAccBrightboxServerGroup_clear_names(t *testing.T) {
	resourceName := "brightbox_server_group.foobar"
	var server_group brightbox.ServerGroup
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(resourceName, &server_group),
					testAccCheckBrightboxServerGroupAttributes(&server_group, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxServerGroupConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(resourceName, &server_group),
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

func testAccCheckBrightboxServerGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_server_group" {
			continue
		}

		// Try to find the ServerGroup
		_, err := client.ServerGroup(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for server_group %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxServerGroupExists(n string, serverGroup *brightbox.ServerGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ServerGroup ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the ServerGroup
		retrieveServerGroup, err := client.ServerGroup(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveServerGroup.Id != rs.Primary.ID {
			return fmt.Errorf("ServerGroup not found")
		}

		*serverGroup = *retrieveServerGroup

		return nil
	}
}

func testAccCheckBrightboxServerGroupAttributes(serverGroup *brightbox.ServerGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if serverGroup.Name != name {
			return fmt.Errorf("Bad name: %s", serverGroup.Name)
		}
		if serverGroup.Description != name {
			return fmt.Errorf("Bad description: %s", serverGroup.Description)
		}
		if serverGroup.Default != false {
			return fmt.Errorf("Bad default: %v", serverGroup.Default)
		}
		return nil
	}
}

func testAccCheckBrightboxServerGroupConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_server_group" "foobar" {
	name = "foo-%d"
	description = "foo-%d"
}
`, rInt, rInt)
}

func testAccCheckBrightboxServerGroupConfig_updated(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_server_group" "foobar" {
	name = "bar-%d"
	description = "bar-%d"
}
`, rInt, rInt)
}

const testAccCheckBrightboxServerGroupConfig_empty = `

resource "brightbox_server_group" "foobar" {
	name = ""
	description = ""
}
`

// Sweeper

func init() {
	resource.AddTestSweepers("server_group", &resource.Sweeper{
		Name: "server_group",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			objects, err := client.APIClient.ServerGroups()
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Default {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.Id, object.Name)
					if err := client.APIClient.DestroyServerGroup(object.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.Id, err)
					}
				}
			}
			return nil
		},
	})
}
