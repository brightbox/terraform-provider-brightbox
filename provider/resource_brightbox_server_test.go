package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxServer_Basic(t *testing.T) {
	var server brightbox.Server

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "image", "img-zhoh0"),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "name", "create_server_test"),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "type", "1gb.ssd"),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "zone", "gb1-a"),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
		},
	})
}

func TestAccBrightboxServer_server_group(t *testing.T) {
	var server_group brightbox.ServerGroup
	var server brightbox.Server

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxServerConfig_server_group,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxServerConfig_server_group,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "1"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxServerConfig_multi_server_group,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerAndGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxServerConfig_multi_server_group,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo", &server_group),
					testAccCheckBrightboxServerGroupExists("brightbox_server_group.barfoo2", &server_group2),
					testAccCheckBrightboxServerExists("brightbox_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"brightbox_server.foobar", "server_groups.#", "2"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxServerConfig_server_group,
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
	client := testAccProvider.Meta().(*brightbox.Client)

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

		client := testAccProvider.Meta().(*brightbox.Client)

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

		if server.Image.Id != "img-zhoh0" {
			return fmt.Errorf("Bad image id: %s", server.Image.Id)
		}

		if server.ServerType.Handle != "1gb.ssd" {
			return fmt.Errorf("Bad server type: %s", server.ServerType.Handle)
		}

		if server.Zone.Handle != "gb1-a" {
			return fmt.Errorf("Bad zone: %s", server.Zone.Handle)
		}

		if server.Name != "create_server_test" {
			return fmt.Errorf("Bad name: %s", server.Name)
		}
		return nil
	}
}

const testAccCheckBrightboxServerConfig_basic = `

resource "brightbox_server" "foobar" {
	image = "img-zhoh0"
	name = "create_server_test"
	type = "1gb.ssd"
	zone = "gb1-a"
	user_data = "foo:-with-character's"
}
`

const testAccCheckBrightboxServerConfig_server_group = `

resource "brightbox_server" "foobar" {
	name = "server_group_test"
	image = "img-zhoh0"
	server_groups = ["${brightbox_server_group.barfoo.id}"]
}

resource "brightbox_server_group" "barfoo" {
	name = "server_group_test"
}
`

const testAccCheckBrightboxServerConfig_multi_server_group = `

resource "brightbox_server" "foobar" {
	name = "server_group_test"
	image = "img-zhoh0"
	server_groups = ["${brightbox_server_group.barfoo.id}",
	"${brightbox_server_group.barfoo2.id}"]
}

resource "brightbox_server_group" "barfoo" {
	name = "server_group_test"
}

resource "brightbox_server_group" "barfoo2" {
	name = "server_group_test2"
}
`
