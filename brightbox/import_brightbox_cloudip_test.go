package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccBrightboxCloudip_importBasic(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importEmptyName(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_empty_name,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importMapped(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_mapped(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importRemapped(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_remapped(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
