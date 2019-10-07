package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccBrightboxOrbitContainer_importBasic(t *testing.T) {
	resourceName := "brightbox_orbit_container.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxOrbitContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_basic,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at"},
			},
		},
	})
}

func TestAccBrightboxOrbitContainer_importMetadata(t *testing.T) {
	resourceName := "brightbox_orbit_container.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxOrbitContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxOrbitContainerConfig_metadata,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at"},
			},
		},
	})
}
