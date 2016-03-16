package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxServerGroup_Basic(t *testing.T) {
	var server_group brightbox.ServerGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxServerGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					testAccCheckBrightboxServerGroupAttributes(&server_group),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", "empty"),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", "empty"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxServerGroupConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", "updated"),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", "updated"),
				),
			},
		},
	})
}

func TestAccBrightboxServerGroup_clear_names(t *testing.T) {
	var server_group brightbox.ServerGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxServerGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.foobar", &server_group),
					testAccCheckBrightboxServerGroupAttributes(&server_group),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "name", "empty"),
					resource.TestCheckResourceAttr(
						"brightbox_server_group.foobar", "description", "empty"),
				),
			},
			resource.TestStep{
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

func testAccCheckBrightboxServerGroupAttributes(server_group *brightbox.ServerGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if server_group.Name != "empty" {
			return fmt.Errorf("Bad name: %s", server_group.Name)
		}
		if server_group.Description != "empty" {
			return fmt.Errorf("Bad description: %s", server_group.Description)
		}
		if server_group.Default != false {
			return fmt.Errorf("Bad default: %v", server_group.Default)
		}
		return nil
	}
}

const testAccCheckBrightboxServerGroupConfig_basic = `

resource "brightbox_server_group" "foobar" {
	name = "empty"
	description = "empty"
}
`

const testAccCheckBrightboxServerGroupConfig_updated = `

resource "brightbox_server_group" "foobar" {
	name = "updated"
	description = "updated"
}
`

const testAccCheckBrightboxServerGroupConfig_empty = `

resource "brightbox_server_group" "foobar" {
	name = ""
	description = ""
}
`
