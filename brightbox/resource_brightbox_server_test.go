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

var imageRe = regexp.MustCompile("^img-.....$")
var zoneRe = regexp.MustCompile("^gb1s?-[ab]$")

func TestAccBrightboxServer_Basic(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists(resourceName, &server),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRe),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "type", "1gb.ssd"),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRe),
					resource.TestCheckResourceAttr(
						resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckBrightboxServerConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists(resourceName, &server),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRe),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "type", "1gb.ssd"),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRe),
					resource.TestCheckResourceAttr(
						resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists(resourceName, &server),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
		},
	})
}

func TestAccBrightboxServer_Blank(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_blank(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxServer_userDataBase64(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_base64_userdata(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists(
						resourceName, &server),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data_base64",
						"aGVsbG8gd29ybGQ="),
				),
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

func TestAccBrightboxServer_serverGroup(t *testing.T) {
	serverResourceName := "brightbox_server.foobar"
	resourceName := "brightbox_server_group.barfoo"
	var serverGroup brightbox.ServerGroup
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_serverGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(resourceName, &serverGroup),
					testAccCheckBrightboxServerExists(serverResourceName, &server),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxServer_multiServerGroupUp(t *testing.T) {
	serverResourceName := "brightbox_server.foobar"
	serverGroupResourceName := "brightbox_server_group.barfoo"
	otherServerGroupResourceName := "brightbox_server_group.barfoo2"
	var serverGroup, serverGroup2 brightbox.ServerGroup
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_serverGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists(serverResourceName, &server),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_multiServerGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(serverGroupResourceName, &serverGroup),
					testAccCheckBrightboxServerGroupExists(otherServerGroupResourceName, &serverGroup2),
					testAccCheckBrightboxServerExists(serverResourceName, &server),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "2"),
				),
			},
			{
				ResourceName:      otherServerGroupResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxServer_multiServerGroupDown(t *testing.T) {
	resourceName := "brightbox_server_group.barfoo"
	otherResourceName := "brightbox_server_group.barfoo2"
	serverResourceName := "brightbox_server.foobar"
	var serverGroup, serverGroup2 brightbox.ServerGroup
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_multiServerGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists(resourceName, &serverGroup),
					testAccCheckBrightboxServerGroupExists(otherResourceName, &serverGroup2),
					testAccCheckBrightboxServerExists(serverResourceName, &server),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckBrightboxServerConfig_serverGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists(serverResourceName, &server),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "1"),
				),
			},
		},
	})
}

func testAccCheckBrightboxServerAndGroupDestroy(s *terraform.State) error {
	err := testAccCheckBrightboxServerDestroy(s)
	if err != nil {
		return err
	}
	return testAccCheckBrightboxServerGroupDestroy(s)
}

func testAccCheckBrightboxServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_server" {
			continue
		}

		// Try to find the Server
		server, err := client.Server(rs.Primary.ID)

		// Wait

		if err == nil && server.Status != "deleted" {
			return fmt.Errorf(
				"Server %s not in deleted state. Status is %s", rs.Primary.ID, server.Status)
		} else if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for server (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxServerExists(n string, server *brightbox.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).APIClient

		// Try to find the Server
		retrieveServer, err := client.Server(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveServer.Id != rs.Primary.ID {
			return fmt.Errorf("Server not found")
		}

		*server = *retrieveServer

		return nil
	}
}

func testAccCheckBrightboxServerAttributes(server *brightbox.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if !imageRe.MatchString(server.Image.Id) {
			return fmt.Errorf("Bad image id: %s", server.Image.Id)
		}

		if server.ServerType.Handle != "1gb.ssd" {
			return fmt.Errorf("Bad server type: %s", server.ServerType.Handle)
		}

		if !zoneRe.MatchString(server.Zone.Handle) {
			return fmt.Errorf("Bad zone: %s", server.Zone.Handle)
		}

		return nil
	}
}

func TestAccBrightboxServer_Update(t *testing.T) {
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},

			{
				Config: testAccCheckBrightboxServerConfig_rename(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", fmt.Sprintf("baz-%d", rInt)),
				),
			},
		},
	})
}

func TestAccBrightboxServer_UpdateUserData(t *testing.T) {
	var afterCreate, afterUpdate brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &afterCreate),
					testAccCheckBrightboxServerAttributes(&afterCreate),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},

			{
				Config: testAccCheckBrightboxServerConfig_userdata_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &afterUpdate),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar",
						"user_data",
						"a26afb778136f34ce2ef86e72be0802498b0291b"),
					testAccCheckBrightboxServerRecreated(
						t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func testAccCheckBrightboxServerRecreated(t *testing.T,
	before, after *brightbox.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.Id != after.Id {
			t.Fatalf("ID changed to %v, but expected in place update", after.Id)
		}
		return nil
	}
}

func testAccCheckBrightboxServerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
	user_data = "foo:-with-character's"
}

data "brightbox_server_group" "barfoo" {
	name = "^default$"
}

%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_locked(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
	user_data = "foo:-with-character's"
	locked = true
}

data "brightbox_server_group" "barfoo" {
	name = "^default$"
}

%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_base64_userdata(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
	user_data_base64 = "${base64encode("hello world")}"
}

data "brightbox_server_group" "barfoo" {
	name = "^default$"
}

%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_userdata_update(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
	user_data = "foo:-with-different-character's"
}

data "brightbox_server_group" "barfoo" {
	name = "^default$"
}

%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_rename(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	name = "baz-%d"
	type = "1gb.ssd"
	image = "${data.brightbox_image.foobar.id}"
	server_groups = ["${data.brightbox_server_group.default.id}"]
	user_data = "foo:-with-character's"
}

data "brightbox_server_group" "barfoo" {
	name = "^default$"
}

%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_blank(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = ""
	type = "1gb.ssd"
	server_groups = ["${data.brightbox_server_group.default.id}"]
	user_data = "foo:-with-character's"
}

data "brightbox_server_group" "barfoo" {
	name = "^default$"
}

%s%s`, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_serverGroup(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	name = "foo-%d"
	image = "${data.brightbox_image.foobar.id}"
	server_groups = ["${brightbox_server_group.barfoo.id}"]
	type = "512mb.ssd"
}

resource "brightbox_server_group" "barfoo" {
	name = "bar-%d"
}

%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk)
}

func testAccCheckBrightboxServerConfig_multiServerGroup(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	name = "foo-%d"
	image = "${data.brightbox_image.foobar.id}"
	server_groups = ["${brightbox_server_group.barfoo.id}",
	"${brightbox_server_group.barfoo2.id}"]
	type = "512mb.ssd"
}

resource "brightbox_server_group" "barfoo" {
	name = "bar-%d"
}

resource "brightbox_server_group" "barfoo2" {
	name = "baz-%d"
}

%s`, rInt, rInt, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk)
}

// Sweeper

func init() {
	resource.AddTestSweepers("server", &resource.Sweeper{
		Name: "server",
		F: func(_ string) error {
			client, err := obtainCloudClient()
			if err != nil {
				return err
			}
			objects, err := client.APIClient.Servers()
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Status != status.Active {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.Id, object.Name)
					if err := setLockState(client.APIClient, false, brightbox.Server{Id: object.Id}); err != nil {
						log.Printf("error unlocking %s during sweep: %s", object.Id, err)
					}
					if err := client.APIClient.DestroyServer(object.Id); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.Id, err)
					}
				}
			}
			return nil
		},
	})
}
