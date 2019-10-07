package brightbox

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxOrbitContainer_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxOrbitContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "updated"),
				),
			},
		},
	})
}

func TestAccBrightboxOrbitContainer_metadata(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxOrbitContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata_add,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
				),
			},
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata_delete,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxOrbitContainerExists("brightbox_orbit_container.foobar"),
					resource.TestCheckResourceAttr(
						"brightbox_orbit_container.foobar", "name", "initial"),
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
	name = "initial"
}
`

const testAccCheckBrightboxOrbitContainerConfig_updated = `

resource "brightbox_orbit_container" "foobar" {
	name = "updated"
}
`

const testAccCheckBrightboxOrbitContainerConfig_metadata = `

resource "brightbox_orbit_container" "foobar" {
	name = "initial"
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
	name = "initial"
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
	name = "initial"
	metadata = {
		"bar"= "foo"
		"uni" = "€uro"
	}
	container_read = [ "acc-testy", "acc-12345", "acc-98765" ]
}
`
