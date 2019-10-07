package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxServerGroup_Basic(t *testing.T) {
	var server_group brightbox.ServerGroup
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updated_name := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					testAccCheckBrightboxServerGroupAttributes(&server_group, name),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxServerGroupConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", updated_name),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", updated_name),
				),
			},
		},
	})
}

func TestAccBrightboxServerGroup_clear_names(t *testing.T) {
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
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					testAccCheckBrightboxServerGroupAttributes(&server_group, name),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxServerGroupConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", ""),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", ""),
				),
			},
		},
	})
}

func testAccCheckBrightboxServerGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

func testAccCheckBrightboxServerGroupExists(n string, server_group *brightbox.ServerGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ServerGroup ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

		// Try to find the ServerGroup
		retrieveServerGroup, err := client.ServerGroup(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveServerGroup.Id != rs.Primary.ID {
			return fmt.Errorf("ServerGroup not found")
		}

		*server_group = *retrieveServerGroup

		return nil
	}
}

func testAccCheckBrightboxServerGroupAttributes(server_group *brightbox.ServerGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if server_group.Name != name {
			return fmt.Errorf("Bad name: %s", server_group.Name)
		}
		if server_group.Description != name {
			return fmt.Errorf("Bad description: %s", server_group.Description)
		}
		if server_group.Default != false {
			return fmt.Errorf("Bad default: %v", server_group.Default)
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
