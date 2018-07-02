package brightbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxDatabaseServer_BasicUpdates(t *testing.T) {
	var database_server brightbox.DatabaseServer
	rInt := acctest.RandInt()
	name := fmt.Sprintf("bar-%d", rInt)
	updatedName := fmt.Sprintf("baz-%d", rInt)
	var cloudip brightbox.CloudIP

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxDatabaseServerAndOthersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists("brightbox_database_server.default", &database_server),
					testAccCheckBrightboxEmptyDatabaseServerAttributes(&database_server, name),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "name", name),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "description", name),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_weekday", "6"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_hour", "6"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_version", "5.6"),
					resource.TestCheckNoResourceAttr(
						"brightbox_database_server.default", "allow_access"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_clear_names,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists("brightbox_database_server.default", &database_server),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "name", ""),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "description", ""),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_weekday", "6"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_hour", "6"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_engine", "mysql"),
					resource.TestMatchResourceAttr(
						"brightbox_database_server.default", "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_version", "5.6"),
					resource.TestCheckNoResourceAttr(
						"brightbox_database_server.default", "allow_access"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_update_maintenance(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists("brightbox_database_server.default", &database_server),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "name", updatedName),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "description", updatedName),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_weekday", "5"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_hour", "4"),
					resource.TestMatchResourceAttr(
						"brightbox_database_server.default", "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_version", "5.6"),
					resource.TestCheckNoResourceAttr(
						"brightbox_database_server.default", "allow_access"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_update_access(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxDatabaseServerExists("brightbox_database_server.default", &database_server),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "name", updatedName),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "description", updatedName),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_weekday", "5"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_hour", "4"),
					resource.TestMatchResourceAttr(
						"brightbox_database_server.default", "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_version", "5.6"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "allow_access.#", "3"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "allow_access.2131663435", "158.152.1.65/32"),
				),
			},
			{
				Config: testAccCheckBrightboxDatabaseServerConfig_map_cloudip(updatedName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.barfar", &cloudip),
					testAccCheckBrightboxDatabaseServerExists("brightbox_database_server.default", &database_server),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "name", updatedName),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "description", updatedName),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_weekday", "5"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "maintenance_hour", "4"),
					resource.TestMatchResourceAttr(
						"brightbox_database_server.default", "database_type", regexp.MustCompile("^dbt-.....$")),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_engine", "mysql"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "database_version", "5.6"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "allow_access.#", "3"),
					resource.TestCheckResourceAttr(
						"brightbox_database_server.default", "allow_access.2131663435", "158.152.1.65/32"),
				),
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
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

func testAccCheckBrightboxDatabaseServerExists(n string, database_server *brightbox.DatabaseServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DatabaseServer ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

		// Try to find the DatabaseServer
		retrieveDatabaseServer, err := client.DatabaseServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveDatabaseServer.Id != rs.Primary.ID {
			return fmt.Errorf("DatabaseServer not found")
		}

		*database_server = *retrieveDatabaseServer

		return nil
	}
}

func testAccCheckBrightboxEmptyDatabaseServerAttributes(database_server *brightbox.DatabaseServer, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		databaseTypeRs, ok := s.RootModule().Resources["data.brightbox_database_type.foobar"]
		if !ok {
			return fmt.Errorf("can't find database type foobar in state")
		}
		if database_server.Name != name {
			return fmt.Errorf("Bad name: %s", database_server.Name)
		}
		if database_server.Description != name {
			return fmt.Errorf("Bad name: %s", database_server.Description)
		}
		if database_server.Locked != false {
			return fmt.Errorf("Bad locked: %v", database_server.Locked)
		}
		if database_server.Status != "active" {
			return fmt.Errorf("Bad status: %s", database_server.Status)
		}
		if database_server.DatabaseEngine != "mysql" {
			return fmt.Errorf("Bad database engine: %s", database_server.DatabaseEngine)
		}
		if database_server.DatabaseVersion != "5.6" {
			return fmt.Errorf("Bad database version: %s", database_server.DatabaseVersion)
		}
		if database_server.DatabaseServerType.Id != databaseTypeRs.Primary.Attributes["id"] {
			return fmt.Errorf("Bad database server type: %v", database_server.DatabaseServerType)
		}
		if database_server.MaintenanceWeekday != 6 {
			return fmt.Errorf("Bad MaintenanceWeekday: %d", database_server.MaintenanceWeekday)
		}
		if database_server.MaintenanceHour != 6 {
			return fmt.Errorf("Bad MaintenanceHour: %d", database_server.MaintenanceHour)
		}
		if database_server.Zone.Handle == "" {
			return fmt.Errorf("Bad Zone: %s", database_server.Zone.Handle)
		}
		if database_server.AdminUsername == "" {
			return fmt.Errorf("Bad AdminUsername: %s", database_server.AdminUsername)
		}
		if database_server.AdminPassword != "" {
			return fmt.Errorf("Exposed API AdminPassword: %s", database_server.AdminPassword)
		}
		if len(database_server.AllowAccess) != 0 {
			return fmt.Errorf("Bad AllowAccess list: %#v", database_server.AllowAccess)
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
	database_version = "5.6"
	database_type = "${data.brightbox_database_type.foobar.id}"
	maintenance_weekday = 6
	maintenance_hour = 6
}

data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
`, name, name)
}

var testAccCheckBrightboxDatabaseServerConfig_clear_names = testAccCheckBrightboxDatabaseServerConfig_basic("")

func testAccCheckBrightboxDatabaseServerConfig_update_maintenance(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	description = "%s"
	database_engine = "mysql"
	database_version = "5.6"
	database_type = "${data.brightbox_database_type.foobar.id}"
	maintenance_weekday = 5
	maintenance_hour = 4
}

data "brightbox_database_type" "foobar" {
	name = "^SSD 4GB$"
}
`, name, name)
}

func testAccCheckBrightboxDatabaseServerConfig_update_access(name string) string {
	return fmt.Sprintf(`

resource "brightbox_database_server" "default" {
	name = "%s"
	description = "%s"
	database_engine = "mysql"
	database_version = "5.6"
	maintenance_weekday = 5
	maintenance_hour = 4
	allow_access = [
		"${brightbox_server_group.barfoo.id}", "${brightbox_server.foobar.id}", "158.152.1.65/32"
	]
}

resource "brightbox_server" "foobar" {
	name = "database test"
	image = "${data.brightbox_image.foobar.id}"
}

resource "brightbox_server_group" "barfoo" {
	name = "database test"
}

%s`, name, name, TestAccBrightboxImageDataSourceConfig_blank_disk)
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
