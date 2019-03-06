package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxApiClient_Basic(t *testing.T) {
	var api_client brightbox.ApiClient
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updated_name := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxApiClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxApiClientConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxApiClientExists("brightbox_api_client.foobar", &api_client),
					testAccCheckBrightboxApiClientAttributes(&api_client, name),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxApiClientConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxApiClientExists("brightbox_api_client.foobar", &api_client),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "name", updated_name),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "description", updated_name),
				),
			},
		},
	})
}

func TestAccBrightboxApiClient_clear_names(t *testing.T) {
	var api_client brightbox.ApiClient
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxApiClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxApiClientConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxApiClientExists("brightbox_api_client.foobar", &api_client),
					testAccCheckBrightboxApiClientAttributes(&api_client, name),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxApiClientConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxApiClientExists("brightbox_api_client.foobar", &api_client),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "name", ""),
					resource.TestCheckResourceAttr(
						"brightbox_api_client.foobar", "description", ""),
				),
			},
		},
	})
}

func testAccCheckBrightboxApiClientDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_api_client" {
			continue
		}

		// Try to find the ApiClient
		_, err := client.ApiClient(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for api_client %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxApiClientExists(n string, api_client *brightbox.ApiClient) resource.TestCheckFunc {
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

func testAccCheckBrightboxApiClientAttributes(api_client *brightbox.ApiClient, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if api_client.Name != name {
			return fmt.Errorf("Bad name: %s", api_client.Name)
		}
		if api_client.Description != name {
			return fmt.Errorf("Bad description: %s", api_client.Description)
		}
		return nil
	}
}

func testAccCheckBrightboxApiClientConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_api_client" "foobar" {
	name = "foo-%d"
	description = "foo-%d"
	permissions_group = "storage"
}
`, rInt, rInt)
}

func testAccCheckBrightboxApiClientConfig_updated(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_api_client" "foobar" {
	name = "bar-%d"
	description = "bar-%d"
	permissions_group = "full"
}
`, rInt, rInt)
}

const testAccCheckBrightboxApiClientConfig_empty = `

resource "brightbox_api_client" "foobar" {
	name = ""
	description = ""
}
`
