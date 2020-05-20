package brightbox

import (
	"fmt"
	"log"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const containerName = "test-acc-initial"

func TestAccBrightboxOrbitContainer_Basic(t *testing.T) {
	resourceName := "brightbox_orbit_container.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxOrbitContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "name", containerName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "name", "test-acc-updated"),
				),
			},
		},
	})
}

func TestAccBrightboxOrbitContainer_metadata(t *testing.T) {
	resourceName := "brightbox_orbit_container.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxOrbitContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "name", containerName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata_add,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", containerName),
				),
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", containerName),
				),
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata_delete,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", containerName),
				),
			},
		},
	})
}

func testAccCheckBrightboxOrbitContainerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).OrbitClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_orbit_container" {
			continue
		}

		// Try to find the container
		getresult := containers.Get(client, rs.Primary.ID, nil)

		// Wait

		err := getresult.Err
		if err != nil && err.Error() != "Resource not found" {
			return fmt.Errorf(
				"Error waiting for container %s to be destroyed: %s",
				rs.Primary.ID, getresult.Err)
		}
	}

	return nil
}

func testAccCheckBrightboxOrbitContainerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		return nil
	}
}

const testAccCheckBrightboxOrbitContainerConfig_basic = `

resource "brightbox_orbit_container" "foobar" {
	name = "test-acc-initial"
}
`

const testAccCheckBrightboxOrbitContainerConfig_updated = `

resource "brightbox_orbit_container" "foobar" {
	name = "test-acc-updated"
}
`

const testAccCheckBrightboxOrbitContainerConfig_metadata = `

resource "brightbox_orbit_container" "foobar" {
	name = "test-acc-initial"
	container_read = [ "acc-testy", "acc-12345"]
	metadata = {
		"foo"= "bar"
		"bar"= "baz" 
		"uni" = "€uro"
	}
}
`

const testAccCheckBrightboxOrbitContainerConfig_metadata_add = `

resource "brightbox_orbit_container" "foobar" {
	name = "test-acc-initial"
	metadata = {
		"foo"= "bar"
		"bar"= "foo"
		"uni" = "€uro"
	}
	container_read = [ "acc-testy", "acc-12345", "acc-98765" ]
}
`

const testAccCheckBrightboxOrbitContainerConfig_metadata_delete = `

resource "brightbox_orbit_container" "foobar" {
	name = "test-acc-initial"
	metadata = {
		"bar"= "foo"
		"uni" = "€uro"
	}
	container_read = [ "acc-testy", "acc-12345", "acc-98765" ]
}
`

func init() {
	resource.AddTestSweepers("orbit_containers", &resource.Sweeper{
		Name: "orbit_containers",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			result, err := containers.Delete(client.OrbitClient, containerName).Extract()
			if err != nil {
				if _, ok := err.(gophercloud.ErrDefault404); !ok {
					return err
				}
				return nil
			}
			log.Printf("[INFO] Container deleted with TransID %s", result.TransID)
			return nil
		},
	})
}
