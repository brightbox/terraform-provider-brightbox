package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccBrightboxContainer_importBasic(t *testing.T) {
	resourceName := "brightbox_orbit_container.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxContainerConfig_basic,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxContainer_importMetadata(t *testing.T) {
	resourceName := "brightbox_orbit_container.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxContainerConfig_metadata,
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
