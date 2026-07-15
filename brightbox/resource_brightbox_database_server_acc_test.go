package brightbox

import (
	"fmt"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBrightboxDatabaseServer_Snapshots(t *testing.T) {
	var databaseServer brightbox.DatabaseServer
	rInt := acctest.RandInt()
	name := fmt.Sprintf("bar-%d", rInt)
	resourceName := "brightbox_database_server.default"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxDatabaseServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_noSnapshots(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Database Server",
						&databaseServer,
						(*brightbox.Client).DatabaseServer,
					),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_withSnapshots(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Database Server",
						&databaseServer,
						(*brightbox.Client).DatabaseServer,
					),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_schedule", "0 7 * * *"),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_retention", "5"),
				),
			},
			{
				// Omits both fields to confirm they correctly pull API values
				Config: testAccCheckBrightboxDatabaseServerConfig_noSnapshots(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Database Server",
						&databaseServer,
						(*brightbox.Client).DatabaseServer,
					),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_schedule", "0 7 * * *"),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_retention", "5"),
				),
			},
		},
	})
}

func testAccCheckBrightboxDatabaseServerConfig_noSnapshots(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	database_engine = "mysql"
	database_version = "8.0"
	allow_access = [ data.brightbox_server_group.default.id ]
	timeouts {
	  create = "60m"
	}
}
%s
`, name, TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxDatabaseServerConfig_withSnapshots(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	database_engine = "mysql"
	database_version = "8.0"
	snapshots_schedule = "0 7 * * *"
	snapshots_retention = "5"
	allow_access = [ data.brightbox_server_group.default.id ]
	timeouts {
	  create = "60m"
	}
}
%s
`, name, TestAccBrightboxDataServerGroupConfig_default)
}
