package brightbox

import (
	"context"
	"fmt"
	"log"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccBrightboxAPIClient_Basic(t *testing.T) {
	var apiClient brightbox.APIClient
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updatedName := fmt.Sprintf("bar-%d", rInt)
	resourceName := "brightbox_api_client.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxAPIClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxAPIClientConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"API Client",
						&apiClient,
						(*brightbox.Client).APIClient,
					),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"API Client",
						&apiClient,
						(*brightbox.Client).APIClient,
					),
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
	var apiClient brightbox.APIClient
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_api_client.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxAPIClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxAPIClientConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"API Client",
						&apiClient,
						(*brightbox.Client).APIClient,
					),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"API Client",
						&apiClient,
						(*brightbox.Client).APIClient,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
				),
			},
		},
	})
}

var testAccCheckBrightboxAPIClientDestroy = testAccCheckBrightboxDestroyBuilder(
	"brightbox_api_client",
	(*brightbox.Client).APIClient,
)

func testAccCheckBrightboxAPIClientAttributes(apiClient *brightbox.APIClient, name string) resource.TestCheckFunc {
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			apiClients, err := client.APIClient.APIClients(ctx)
			if err != nil {
				return err
			}
			for _, apiClient := range apiClients {
				if apiClient.RevokedAt != nil {
					continue
				}
				if isTestName(apiClient.Name) {
					log.Printf("[INFO] removing %s named %s", apiClient.ID, apiClient.Name)
					if _, err := client.APIClient.DestroyAPIClient(ctx, apiClient.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", apiClient.ID, err)
					}
				}
			}
			return nil
		},
	})
}
