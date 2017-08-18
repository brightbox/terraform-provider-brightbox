package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxContainer_Basic(t *testing.T) {
	var api_client brightbox.ApiClient

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxContainerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxContainerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_container.foobar", &api_client),
					testAccCheckBrightboxApiClientAttributes(&api_client),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "name", "initial"),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "description", "initial"),
					resource.TestCheckResourceAttrPtr(
						"brightbox_container.foobar", "auth_user", &api_client.Id),
					resource.TestCheckResourceAttrPtr(
						"brightbox_container.foobar", "account_id", &api_client.Account.Id),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxContainerConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_container.foobar", &api_client),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "name", "updated"),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "description", "updated"),
				),
			},
		},
	})
}

func TestAccBrightboxContainer_clear_names(t *testing.T) {
	var container brightbox.ApiClient

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxContainerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxContainerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_container.foobar", &container),
					testAccCheckBrightboxApiClientAttributes(&container),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "name", "initial"),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "description", "initial"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxContainerConfig_blank_desc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_container.foobar", &container),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "name", "initial"),
					resource.TestCheckResourceAttr(
						"brightbox_container.foobar", "description", ""),
				),
			},
		},
	})
}

func testAccCheckBrightboxContainerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_container" {
			continue
		}

		// Try to find the ApiClient
		_, err := client.ApiClient(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for container %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxContainerExists(n string, api_client *brightbox.ApiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ApiClient ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

		// Try to find the ApiClient
		retrieveApiClient, err := client.ApiClient(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveApiClient.Id != rs.Primary.ID {
			return fmt.Errorf("ApiClient not found")
		}

		*api_client = *retrieveApiClient

		return nil
	}
}

func testAccCheckBrightboxApiClientAttributes(api_client *brightbox.ApiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if api_client.Name != "initial" {
			return fmt.Errorf("Bad name: %s", api_client.Name)
		}
		if api_client.Description != "initial" {
			return fmt.Errorf("Bad description: %s", api_client.Description)
		}
		return nil
	}
}

const testAccCheckBrightboxContainerConfig_basic = `

resource "brightbox_container" "foobar" {
	name = "initial"
	description = "initial"
}
`

const testAccCheckBrightboxContainerConfig_updated = `

resource "brightbox_container" "foobar" {
	name = "updated"
	description = "updated"
}
`

const testAccCheckBrightboxContainerConfig_blank_desc = `

resource "brightbox_container" "foobar" {
	name = "initial"
	description = ""
}
`
