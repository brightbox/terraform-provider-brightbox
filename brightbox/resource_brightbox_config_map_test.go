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

func TestAccBrightboxConfigMap_Basic(t *testing.T) {
	resourceName := "brightbox_config_map.foobar"
	var configMap brightbox.ConfigMap
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updatedName := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxConfigMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxConfigMapConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					testAccCheckBrightboxConfigMapAttributes(&configMap, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_updated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "2"),
				),
			},
		},
	})
}

func TestAccBrightboxConfigMap_clear_entries(t *testing.T) {
	resourceName := "brightbox_config_map.foobar"
	var configMap brightbox.ConfigMap
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxConfigMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxConfigMapConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					testAccCheckBrightboxConfigMapAttributes(&configMap, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_empty_name(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					testAccCheckBrightboxConfigMapAttributes(&configMap, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_empty_data(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "0"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Config Map",
						&configMap,
						(*brightbox.Client).ConfigMap,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "0"),
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

func TestAccBrightboxConfigMap_blank(t *testing.T) {
	resourceName := "brightbox_config_map.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxConfigMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxConfigMapConfig_blank,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testAccCheckBrightboxConfigMapDestroy = testAccCheckBrightboxDestroyBuilder(
	"brightbox_config_map",
	(*brightbox.Client).ConfigMap,
)

func testAccCheckBrightboxConfigMapAttributes(configMap *brightbox.ConfigMap, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if configMap.Name != name {
			return fmt.Errorf("Bad name: %s", configMap.Name)
		}
		return nil
	}
}

func testAccCheckBrightboxConfigMapConfig_basic(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_config_map" "foobar" {
	name = "foo-%d"
	data = {"test": "thing-%d"}
}
`, rInt, rInt)
}

func testAccCheckBrightboxConfigMapConfig_empty_name(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_config_map" "foobar" {
	name = ""
	data = {"test": "thing-%d"}
}
`, rInt)
}

func testAccCheckBrightboxConfigMapConfig_empty_data(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_config_map" "foobar" {
	name = "foo-%d"
	data = { }
}
`, rInt)
}

func testAccCheckBrightboxConfigMapConfig_updated(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_config_map" "foobar" {
	name = "bar-%d"
	data = { "test": "bar-%d", "test2": "foo-%d" }
}
`, rInt, rInt, rInt)
}

const testAccCheckBrightboxConfigMapConfig_empty = `

resource "brightbox_config_map" "foobar" {
	name = ""
	data = {}
}
`
const testAccCheckBrightboxConfigMapConfig_blank = `

resource "brightbox_config_map" "foobar" {
	data = { "thing": "{ \"name\" : \"Admin\" }" }
	}
`

// Sweeper

func init() {
	resource.AddTestSweepers("config_map", &resource.Sweeper{
		Name: "config_map",
		F: func(_ string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			objects, err := client.APIClient.ConfigMaps(ctx)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.ID, object.Name)
					if _, err := client.APIClient.DestroyConfigMap(ctx, object.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.ID, err)
					}
				}
			}
			return nil
		},
	})
}
