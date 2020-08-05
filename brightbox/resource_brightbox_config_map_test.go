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

func TestAccBrightboxConfigMap_Basic(t *testing.T) {
	resourceName := "brightbox_config_map.foobar"
	var configMap brightbox.ConfigMap
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	updatedName := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxConfigMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxConfigMapConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxConfigMapExists(resourceName, &configMap),
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
					testAccCheckBrightboxConfigMapExists(resourceName, &configMap),
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
	var config_map brightbox.ConfigMap
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)

	resource.Test(t, resource.TestCase{
		DisableBinaryDriver: true,
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckBrightboxConfigMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxConfigMapConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxConfigMapExists(resourceName, &config_map),
					testAccCheckBrightboxConfigMapAttributes(&config_map, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_empty_name(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxConfigMapExists(resourceName, &config_map),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxConfigMapExists(resourceName, &config_map),
					testAccCheckBrightboxConfigMapAttributes(&config_map, name),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_empty_data(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxConfigMapExists(resourceName, &config_map),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "data.%", "0"),
				),
			},
			{
				Config: testAccCheckBrightboxConfigMapConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxConfigMapExists(resourceName, &config_map),
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
		DisableBinaryDriver: true,
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckBrightboxConfigMapDestroy,
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

func testAccCheckBrightboxConfigMapDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_config_map" {
			continue
		}

		// Try to find the ConfigMap
		_, err := client.ConfigMap(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for config_map %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxConfigMapExists(n string, configMap *brightbox.ConfigMap) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ConfigMap ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the ConfigMap
		retrieveConfigMap, err := client.ConfigMap(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveConfigMap.Id != rs.Primary.ID {
			return fmt.Errorf("ConfigMap not found")
		}

		*configMap = *retrieveConfigMap

		return nil
	}
}

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
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			objects, err := client.APIClient.ConfigMaps()
			if err != nil {
				return err
			}
			for _, object := range objects {
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.Id, object.Name)
					if err := client.APIClient.DestroyConfigMap(object.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.Id, err)
					}
				}
			}
			return nil
		},
	})
}
