package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccBrightboxDatabaseServer_importBasic(t *testing.T) {
	resourceName := "brightbox_database_server.default"
	rInt := acctest.RandInt()
	name := fmt.Sprintf("bar-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxDatabaseServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_basic(name),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"admin_password"},
			},
		},
	})
}

func TestAccBrightboxDatabaseServer_importAccess(t *testing.T) {
	resourceName := "brightbox_database_server.default"
	rInt := acctest.RandInt()
	name := fmt.Sprintf("rab-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxDatabaseServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_update_access(name),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"admin_password"},
			},
		},
	})
}
