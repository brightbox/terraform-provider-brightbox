package brightbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	serverConst "github.com/brightbox/gobrightbox/v2/status/server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var imageRe = regexp.MustCompile("^img-.....$")
var zoneRe = regexp.MustCompile("^gb1s?-[ab]$")
var typeRe = regexp.MustCompile("^typ-.....$")

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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
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
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type"},
			},
			{
				Config: testAccCheckBrightboxServerConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data_base64",
						"aGVsbG8gd29ybGQ="),
				),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						serverResourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
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
					testAccCheckBrightboxObjectExists(
						serverResourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "1"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_multiServerGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						serverGroupResourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						otherServerGroupResourceName,
						"Server Group",
						&serverGroup2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						serverResourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "2"),
				),
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server Group",
						&serverGroup,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						otherResourceName,
						"Server Group",
						&serverGroup2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						serverResourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						serverResourceName, "server_groups.#", "2"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_serverGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						serverResourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
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
		server, err := client.Server(context.Background(), rs.Primary.ID)

		// Wait

		if err != nil {
			var apierror *brightbox.APIError
			if errors.As(err, &apierror) {
				if apierror.StatusCode != 404 {
					return fmt.Errorf(
						"Error waiting for server (%s) to be destroyed: %s",
						rs.Primary.ID, err)
				}
			}
		} else if server.Status != serverConst.Deleted {
			return fmt.Errorf(
				"Server %s not in deleted state. Status is %s", rs.Primary.ID, server.Status)
		}
	}

	return nil
}

func testAccCheckBrightboxServerAttributes(server *brightbox.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if !imageRe.MatchString(server.Image.ID) {
			return fmt.Errorf("Bad image id: %s", server.Image.ID)
		}

		if server.ServerType.ID != "typ-8985i" {
			return fmt.Errorf("Bad server type: %s", server.ServerType.ID)
		}

		if !zoneRe.MatchString(server.Zone.Handle) {
			return fmt.Errorf("Bad zone: %s", server.Zone.Handle)
		}

		return nil
	}
}

func TestAccBrightboxServer_Update(t *testing.T) {
	resourceName := "brightbox_server.foobar"
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},

			{
				Config: testAccCheckBrightboxServerConfig_rename(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxServerAttributes(&server),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("baz-%d", rInt)),
				),
			},
		},
	})
}

func TestAccBrightboxServer_UpdateUserData(t *testing.T) {
	resourceName := "brightbox_server.foobar"
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
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&afterCreate,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxServerAttributes(&afterCreate),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},

			{
				Config: testAccCheckBrightboxServerConfig_userdata_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&afterUpdate,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data",
						"a26afb778136f34ce2ef86e72be0802498b0291b"),
					testAccCheckBrightboxServerRecreated(
						t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccBrightboxServer_DiskResize(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_networkdisk40G(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRe),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestMatchResourceAttr(
						resourceName, "type", typeRe),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRe),
					resource.TestCheckResourceAttr(
						resourceName, "disk_size", "40960"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_networkdisk60G(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRe),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestMatchResourceAttr(
						resourceName, "type", typeRe),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRe),
					resource.TestCheckResourceAttr(
						resourceName, "disk_size", "61440"),
				),
			},
		},
	})
}

func TestAccBrightboxServer_ServerResize(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	var serverResource brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_networkdisk40G(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&serverResource,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "disk_size", "40960"),
					testAccCheckBrightboxServerType(&serverResource.ServerType, 2, 4096),
					resource.TestCheckResourceAttrSet(
						resourceName, "type"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_typechange40G(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&serverResource,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "disk_size", "40960"),
					testAccCheckBrightboxServerType(&serverResource.ServerType, 6, 16384),
					resource.TestCheckResourceAttrSet(
						resourceName, "type"),
				),
			},
		},
	})
}

func testAccCheckBrightboxServerType(serverTypeRef **brightbox.ServerType, cores uint, ram uint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serverType := *serverTypeRef
		if serverType == nil {
			return fmt.Errorf("Expected a server Type, got nil")
		}
		if serverType.Cores != cores {
			return fmt.Errorf("Expected %v cores, got %v", cores, serverType.Cores)
		}

		if serverType.RAM != ram {
			return fmt.Errorf("Expected ram size of %v , got %v", ram, serverType.RAM)
		}
		return nil
	}
}

func testAccCheckBrightboxServerRecreated(t *testing.T,
	before, after *brightbox.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID != after.ID {
			t.Fatalf("ID changed to %v, but expected in place update", after.ID)
		}
		return nil
	}
}

const TestAccBrightboxDataServerTypeConfig_network_disk = `
data "brightbox_server_type" "foobar" {
	handle = "^4gb.nbs$"
}
`

const TestAccBrightboxDataServerTypeConfig_16gbserver = `
data "brightbox_server_type" "barfoo" {
	handle = "^16gb.nbs$"
}
`

func testAccCheckBrightboxServerConfig_networkdisk40G(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = data.brightbox_server_type.foobar.id
	server_groups = [data.brightbox_server_group.default.id]
	disk_size = 40960
}

%s%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default,
		TestAccBrightboxDataServerTypeConfig_network_disk,
	)
}

func testAccCheckBrightboxServerConfig_typechange40G(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = data.brightbox_server_type.barfoo.id
	server_groups = [data.brightbox_server_group.default.id]
	disk_size = 40960
}

%s%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default,
		TestAccBrightboxDataServerTypeConfig_16gbserver,
	)
}

func testAccCheckBrightboxServerConfig_networkdisk60G(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = data.brightbox_server_type.foobar.id
	server_groups = [data.brightbox_server_group.default.id]
	disk_size = 61440
}

%s%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default,
		TestAccBrightboxDataServerTypeConfig_network_disk,
	)
}

func testAccCheckBrightboxServerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
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
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
	user_data = "foo:-with-character's"
	disk_encrypted = true
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
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
	user_data_base64 = base64encode("hello world")
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
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
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
	image = data.brightbox_image.foobar.id
	server_groups = [data.brightbox_server_group.default.id]
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
	image = data.brightbox_image.foobar.id
	name = ""
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
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
	image = data.brightbox_image.foobar.id
	server_groups = [brightbox_server_group.barfoo.id]
	type = "typ-n0977"
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
	image = data.brightbox_image.foobar.id
	server_groups = [
	  brightbox_server_group.barfoo.id,
	  brightbox_server_group.barfoo2.id
	  ]
	type = "typ-n0977"
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			objects, err := client.APIClient.Servers(ctx)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Status != serverConst.Active {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.ID, object.Name)
					if _, err := client.APIClient.UnlockServer(ctx, object.ID); err != nil {
						log.Printf("error unlocking %s during sweep: %s", object.ID, err)
					}
					if _, err := client.APIClient.DestroyServer(ctx, object.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.ID, err)
					}
				}
			}
			return nil
		},
	})
}
