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

var imageRe = regexp.MustCompile("^img-.....$")
var zoneRe = regexp.MustCompile("^gb1s?-[ab]$")

func TestAccBrightboxServer_Basic(t *testing.T) {
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
					resource.TestMatchResourceAttr(
						"brightbox_server.foobar", "image", imageRe),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "type", "1gb.ssd"),
					resource.TestMatchResourceAttr(
						"brightbox_server.foobar", "zone", zoneRe),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
		},
	})
}

func TestAccBrightboxServer_Blank(t *testing.T) {
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
						"brightbox_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_blank,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", ""),
				),
			},
		},
	})
}

func TestAccBrightboxServer_userDataBase64(t *testing.T) {
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
						"brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar",
						"user_data_base64",
						"aGVsbG8gd29ybGQ="),
				),
			},
		},
	})
}

func TestAccBrightboxServer_server_group(t *testing.T) {
	var server_group brightbox.ServerGroup
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_server_group(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo", &server_group),
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "1"),
				),
			},
		},
	})
}

func TestAccBrightboxServer_multi_server_group_up(t *testing.T) {
	var server_group, server_group2 brightbox.ServerGroup
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_server_group(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_multi_server_group(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo", &server_group),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo2", &server_group2),
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "2"),
				),
			},
		},
	})
}

func TestAccBrightboxServer_multi_server_group_down(t *testing.T) {
	var server_group, server_group2 brightbox.ServerGroup
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_multi_server_group(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo", &server_group),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo2", &server_group2),
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "2"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_server_group(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "1"),
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
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

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

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

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
	user_data = "foo:-with-character's"
}

%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk)
}

func testAccCheckBrightboxServerConfig_base64_userdata(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-%d"
	type = "1gb.ssd"
	user_data_base64 = "${base64encode("hello world")}"
}

%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk)
}

func testAccCheckBrightboxServerConfig_userdata_update(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "foo-%d"
	type = "1gb.ssd"
	user_data = "foo:-with-different-character's"
}

%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk)
}

func testAccCheckBrightboxServerConfig_rename(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	name = "baz-%d"
	type = "1gb.ssd"
	image = "${data.brightbox_image.foobar.id}"
	user_data = "foo:-with-character's"
}

%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk)
}

var testAccCheckBrightboxServerConfig_blank = fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = "${data.brightbox_image.foobar.id}"
	name = ""
	type = "1gb.ssd"
	user_data = "foo:-with-character's"
}

%s`, TestAccBrightboxImageDataSourceConfig_blank_disk)

func testAccCheckBrightboxServerConfig_server_group(rInt int) string {
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

func testAccCheckBrightboxServerConfig_multi_server_group(rInt int) string {
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
