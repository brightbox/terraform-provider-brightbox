package brightbox

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxContainer_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxContainerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxContainerConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "updated"),
				),
			},
		},
	})
}

func TestAccBrightboxContainer_metadata(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxContainerConfig_metadata,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxContainerConfig_metadata_add,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxContainerConfig_metadata,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
		},
	})
}

func testAccCheckBrightboxContainerDestroy(s *terraform.State) error {
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

func testAccCheckBrightboxContainerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		return nil
	}
}

const testAccCheckBrightboxContainerConfig_basic = `

resource "brightbox_orbit_container" "foobar" {
	name = "initial"
}
`

const testAccCheckBrightboxContainerConfig_updated = `

resource "brightbox_orbit_container" "foobar" {
	name = "updated"
}
`

const testAccCheckBrightboxContainerConfig_metadata = `

resource "brightbox_orbit_container" "foobar" {
	name = "initial"
	container_read = [ "acc-testy", "acc-12345"]
	metadata {
		"foo"= "bar"
		"bar"= "baz" 
		"uni" = "€uro"
	}
}
`

const testAccCheckBrightboxContainerConfig_metadata_add = `

resource "brightbox_orbit_container" "foobar" {
	name = "initial"
	metadata = {
		"foo"= "bar"
		"bar"= "foo"
		"uni" = "€uro"
	}
	container_read = [ "acc-testy", "acc-12345", "acc-98765" ]
}
`
