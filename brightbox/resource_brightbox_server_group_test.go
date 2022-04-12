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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxServerGroupAttributes(&serverGroup, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
					resource.TestCheckResourceAttr(
						resourceName, "default", "false"),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "default", "false"),
				),
			},
		},
	})
}

func TestAccBrightboxServerGroup_clear_names(t *testing.T) {
	resourceName := "brightbox_server_group.foobar"
	var serverGroup brightbox.ServerGroup
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxServerGroupAttributes(&serverGroup, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
				),
			},
			{
				Config: testAccCheckBrightboxServerGroupConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
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

var testAccCheckBrightboxServerGroupDestroy = testAccCheckBrightboxDestroyBuilder(
	"brightbox_server_group",
	(*brightbox.Client).ServerGroup,
)

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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			objects, err := client.APIClient.ServerGroups(ctx)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Default {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.ID, object.Name)
					if _, err := client.APIClient.DestroyServerGroup(ctx, object.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.ID, err)
					}
				}
			}
			return nil
		},
	})
}
