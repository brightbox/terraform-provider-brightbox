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

func TestAccBrightboxAPIClient_Basic(t *testing.T) {
	var apiClient brightbox.ApiClient
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updatedName := fmt.Sprintf("bar-%d", rInt)
	resourceName := "brightbox_api_client.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxAPIClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxAPIClientConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxAPIClientExists(resourceName, &apiClient),
					testAccCheckBrightboxAPIClientAttributes(&apiClient, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				Config: testAccCheckBrightboxAPIClientConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxAPIClientExists(resourceName, &apiClient),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
				),
			},
		},
	})
}

func TestAccBrightboxAPIClient_clear_names(t *testing.T) {
	var apiClient brightbox.ApiClient
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_api_client.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxAPIClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxAPIClientConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxAPIClientExists(resourceName, &apiClient),
					testAccCheckBrightboxAPIClientAttributes(&apiClient, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxAPIClientConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxAPIClientExists(resourceName, &apiClient),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
				),
			},
		},
	})
}

func testAccCheckBrightboxAPIClientDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

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
					"Error waiting for apiClient %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxAPIClientExists(n string, apiClient *brightbox.ApiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ApiClient ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the ApiClient
		retrieveAPIClient, err := client.ApiClient(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveAPIClient.Id != rs.Primary.ID {
			return fmt.Errorf("ApiClient not found")
		}

		*apiClient = *retrieveAPIClient

		return nil
	}
}

func testAccCheckBrightboxAPIClientAttributes(apiClient *brightbox.ApiClient, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if apiClient.Name != name {
			return fmt.Errorf("Bad name: %s", apiClient.Name)
		}
		if apiClient.Description != name {
			return fmt.Errorf("Bad description: %s", apiClient.Description)
		}
		return nil
	}
}

func testAccCheckBrightboxAPIClientConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_api_client" "foobar" {
	name = "foo-%d"
	description = "foo-%d"
	permissions_group = "storage"
}
`, rInt, rInt)
}

func testAccCheckBrightboxAPIClientConfig_updated(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_api_client" "foobar" {
	name = "bar-%d"
	description = "bar-%d"
	permissions_group = "full"
}
`, rInt, rInt)
}

const testAccCheckBrightboxAPIClientConfig_empty = `

resource "brightbox_api_client" "foobar" {
	name = ""
	description = ""
}
`

// Sweeper

func init() {
	resource.AddTestSweepers("api_client", &resource.Sweeper{
		Name: "api_client",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			apiClients, err := client.APIClient.ApiClients()
			if err != nil {
				return err
			}
			for _, apiClient := range apiClients {
				if apiClient.RevokedAt != nil {
					continue
				}
				if isTestName(apiClient.Name) {
					log.Printf("[INFO] removing %s named %s", apiClient.Id, apiClient.Name)
					if err := client.APIClient.DestroyApiClient(apiClient.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", apiClient.Id, err)
					}
				}
			}
			return nil
		},
	})
}
