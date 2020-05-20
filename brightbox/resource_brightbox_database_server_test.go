package brightbox

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/brightbox/gobrightbox/status"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBrightboxDatabaseServer_BasicUpdates(t *testing.T) {
	var databaseServer brightbox.DatabaseServer
	rInt := acctest.RandInt()
	name := fmt.Sprintf("bar-%d", rInt)
	updatedName := fmt.Sprintf("baz-%d", rInt)
	resourceName := "brightbox_database_server.default"
	var cloudip brightbox.CloudIP

	resource.Test(t, resource.TestCase{
		DisableBinaryDriver: true,
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckBrightboxDatabaseServerAndOthersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_locked(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "name", name),
					resource.TestCheckResourceAttr(
						resourceName, "description", name),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_weekday", "6"),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_hour", "6"),
					resource.TestCheckResourceAttr(
						resourceName, "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						resourceName, "database_version", "8.0"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					testAccCheckBrightboxEmptyDatabaseServerAttributes(&databaseServer, name),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_locked(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_clear_names,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_weekday", "6"),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_hour", "6"),
					resource.TestCheckResourceAttr(
						resourceName, "database_engine", "mysql"),
					resource.TestMatchResourceAttr(
						resourceName, "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						resourceName, "database_version", "8.0"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_update_maintenance(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_weekday", "5"),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_hour", "4"),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_schedule", "4 5 * * *"),
					resource.TestMatchResourceAttr(
						resourceName, "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						resourceName, "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						resourceName, "database_version", "8.0"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_update_access(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_weekday", "5"),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_hour", "4"),
					resource.TestMatchResourceAttr(
						resourceName, "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						resourceName, "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						resourceName, "database_version", "8.0"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.#", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.2131663435", "158.152.1.65/32"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_map_cloudip(updatedName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.barfar", &cloudip),
					testAccCheckBrightboxDatabaseServerExists(resourceName, &databaseServer),
					resource.TestCheckResourceAttr(
						resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "description", updatedName),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_weekday", "5"),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_hour", "4"),
					resource.TestMatchResourceAttr(
						resourceName, "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						resourceName, "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						resourceName, "database_version", "8.0"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.#", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "allow_access.2131663435", "158.152.1.65/32"),
				),
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

func testAccCheckBrightboxDatabaseServerAndOthersDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxDatabaseServerDestroy(s)
	if err != nil {
		return err
	}
	err = testAccCheckBrightboxServerGroupDestroy(s)
	if err != nil {
		return err
	}
	err = testAccCheckBrightboxCloudipDestroy(s)
	if err != nil {
		return err
	}
	return testAccCheckBrightboxServerDestroy(s)
}

func testAccCheckBrightboxDatabaseServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_database_server" {
			continue
		}

		// Try to find the DatabaseServer
		_, err := client.DatabaseServer(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for database_server %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxDatabaseServerExists(n string, databaseServer *brightbox.DatabaseServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DatabaseServer ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the DatabaseServer
		retrieveDatabaseServer, err := client.DatabaseServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveDatabaseServer.Id != rs.Primary.ID {
			return fmt.Errorf("DatabaseServer not found")
		}

		*databaseServer = *retrieveDatabaseServer

		return nil
	}
}

func testAccCheckBrightboxEmptyDatabaseServerAttributes(databaseServer *brightbox.DatabaseServer, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		databaseTypeRs, ok := s.RootModule().Resources["data.brightbox_database_type.foobar"]
		if !ok {
			return fmt.Errorf("can't find database type foobar in state")
		}
		if databaseServer.Name != name {
			return fmt.Errorf("Bad name: %s", databaseServer.Name)
		}
		if databaseServer.Description != name {
			return fmt.Errorf("Bad name: %s", databaseServer.Description)
		}
		if databaseServer.Locked != false {
			return fmt.Errorf("Bad locked: %v", databaseServer.Locked)
		}
		if databaseServer.Status != "active" {
			return fmt.Errorf("Bad status: %s", databaseServer.Status)
		}
		if databaseServer.DatabaseEngine != "mysql" {
			return fmt.Errorf("Bad database engine: %s", databaseServer.DatabaseEngine)
		}
		if databaseServer.DatabaseVersion != "8.0" {
			return fmt.Errorf("Bad database version: %s", databaseServer.DatabaseVersion)
		}
		if databaseServer.DatabaseServerType.Id != databaseTypeRs.Primary.Attributes["id"] {
			return fmt.Errorf("Bad database server type: %v", databaseServer.DatabaseServerType)
		}
		if databaseServer.MaintenanceWeekday != 6 {
			return fmt.Errorf("Bad MaintenanceWeekday: %d", databaseServer.MaintenanceWeekday)
		}
		if databaseServer.MaintenanceHour != 6 {
			return fmt.Errorf("Bad MaintenanceHour: %d", databaseServer.MaintenanceHour)
		}
		if databaseServer.Zone.Handle == "" {
			return fmt.Errorf("Bad Zone: %s", databaseServer.Zone.Handle)
		}
		if databaseServer.AdminUsername == "" {
			return fmt.Errorf("Bad AdminUsername: %s", databaseServer.AdminUsername)
		}
		if databaseServer.AdminPassword != "" {
			return fmt.Errorf("Exposed API AdminPassword: %s", databaseServer.AdminPassword)
		}
		if len(databaseServer.AllowAccess) != 1 {
			return fmt.Errorf("Bad AllowAccess list: %#v", databaseServer.AllowAccess)
		}
		return nil
	}
}

func testAccCheckBrightboxDatabaseServerConfig_basic(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	description = "%s"
	database_engine = "mysql"
	database_version = "8.0"
	database_type = "${data.brightbox_database_type.foobar.id}"
	maintenance_weekday = 6
	maintenance_hour = 6
	allow_access = [ "${data.brightbox_server_group.default.id}" ]
}

data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
%s
`, name, name, TestAccBrightboxDataServerGroupConfig_default)
}

var testAccCheckBrightboxDatabaseServerConfig_clear_names = testAccCheckBrightboxDatabaseServerConfig_basic("")

func testAccCheckBrightboxDatabaseServerConfig_locked(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	description = "%s"
	database_engine = "mysql"
	database_version = "8.0"
	database_type = "${data.brightbox_database_type.foobar.id}"
	maintenance_weekday = 6
	maintenance_hour = 6
	allow_access = [ "${data.brightbox_server_group.default.id}" ]
	locked = true
}

data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
%s
`, name, name, TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxDatabaseServerConfig_update_maintenance(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	description = "%s"
	database_engine = "mysql"
	database_version = "8.0"
	database_type = "${data.brightbox_database_type.foobar.id}"
	maintenance_weekday = 5
	maintenance_hour = 4
	snapshots_schedule = "4 5 * * *"
	allow_access = [ "${data.brightbox_server_group.default.id}" ]
}

data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
%s
`, name, name, TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxDatabaseServerConfig_update_access(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	description = "%s"
	database_engine = "mysql"
	database_version = "8.0"
	maintenance_weekday = 5
	maintenance_hour = 4
	allow_access = [
		"${brightbox_server_group.barfoo.id}", "${brightbox_server.foobar.id}", "158.152.1.65/32"
	]
}

resource "brightbox_server" "foobar" {
	name = "bar-20200513"
	image = "${data.brightbox_image.foobar.id}"
	server_groups = [ "${data.brightbox_server_group.default.id}" ]
}

resource "brightbox_server_group" "barfoo" {
	name = "bar-20200513"
}

%s%s`, name, name, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxDatabaseServerConfig_map_cloudip(name string, rInt int) string {
	return fmt.Sprintf(`
%s

	resource "brightbox_cloudip" "barfar" {
		name = "baz-%d"
		target = "${brightbox_database_server.default.id}"
	}
`, testAccCheckBrightboxDatabaseServerConfig_update_access(name), rInt)
}

// Sweeper

func init() {
	resource.AddTestSweepers("database_server", &resource.Sweeper{
		Name: "database_server",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			objects, err := client.APIClient.DatabaseServers()
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Status != status.Active {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.Id, object.Name)
					if err := setLockState(client.APIClient, false, brightbox.DatabaseServer{Id: object.Id}); err != nil {
						log.Printf("error unlocking %s during sweep: %s", object.Id, err)
					}
					if err := client.APIClient.DestroyDatabaseServer(object.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.Id, err)
					}
				}
			}
			return nil
		},
	})
}
