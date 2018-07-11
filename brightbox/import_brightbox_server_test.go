package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccBrightboxServer_importBasic(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxServer_importBlank(t *testing.T) {
	resourceName := "brightbox_server.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_blank,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxServer_importUserDataBase64(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_base64_userdata(rInt),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_base64"},
			},
		},
	})
}

func TestAccBrightboxServer_importMultiServerGroup(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_multi_server_group(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxServer_importServerGroup(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_server_group(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
