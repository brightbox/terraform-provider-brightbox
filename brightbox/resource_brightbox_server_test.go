package brightbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/serverstatus"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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
						resourceName, "disk_encrypted", "true"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "type", "1gb.ssd"),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRegexp),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "disk_encrypted", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
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
						resourceName, "disk_encrypted", "true"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "type", "1gb.ssd"),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRegexp),
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
				),
			},
		},
	})
}

func TestAccBrightboxServer_Blank(t *testing.T) {
	// resourceName := "brightbox_server.foobar"
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
				Config: testAccCheckBrightboxServerConfig_blank(rInt),
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
		} else if server.Status != serverstatus.Deleted {
			return fmt.Errorf(
				"Server %s not in deleted state. Status is %s", rs.Primary.ID, server.Status)
		}
	}

	return nil
}

func testAccCheckBrightboxServerAttributes(server *brightbox.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if !imageRegexp.MatchString(server.Image.ID) {
			return fmt.Errorf("Bad image id: %s", server.Image.ID)
		}

		if server.ServerType.ID != "typ-8985i" {
			return fmt.Errorf("Bad server type: %s", server.ServerType.ID)
		}

		if !zoneRegexp.MatchString(server.Zone.Handle) {
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
						resourceName, "disk_encrypted", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestMatchResourceAttr(
						resourceName, "type", serverTypeRegexp),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRegexp),
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
						resourceName, "disk_encrypted", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestMatchResourceAttr(
						resourceName, "type", serverTypeRegexp),
					resource.TestMatchResourceAttr(
						resourceName, "zone", zoneRegexp),
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
					testAccCheckBrightboxServerType(&serverResource.ServerType, 2, 8192),
					resource.TestCheckResourceAttrSet(
						resourceName, "type"),
				),
			},
		},
	})
}

// func TestAccBrightboxServer_bootVolume(t *testing.T) {
// 	resourceName := "brightbox_server.foobar"
// 	var serverResource brightbox.Server
// 	rInt := acctest.RandInt()

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckBrightboxServerDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCheckBrightboxServerConfig_bootVolume(rInt),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckBrightboxObjectExists(
// 						resourceName,
// 						"Server",
// 						&serverResource,
// 						(*brightbox.Client).Server,
// 					),
// 					resource.TestMatchResourceAttr(
// 						resourceName, "volume", volumeRegexp),
// 					resource.TestMatchResourceAttr(
// 						resourceName, "image", imageRegexp),
// 					resource.TestCheckResourceAttr(
// 						resourceName, "disk_size", "40960"),
// 					resource.TestCheckResourceAttr(
// 						resourceName, "encrypted", "false"),
// 					resource.TestCheckResourceAttr(
// 						resourceName, "encrypted", "false"),
// 					testAccCheckBrightboxServerType(&serverResource.ServerType, 2, 4096),
// 					resource.TestCheckResourceAttrSet(
// 						resourceName, "type"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccBrightboxServer_DataDisk(t *testing.T) {
	resourceName := "brightbox_server.foobar"
	var server brightbox.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_networkDataDisk(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "username", "ubuntu"),
					testAccCheckBrightboxVolumes(&server.Volumes, 1),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_twoDataDisk(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxVolumes(&server.Volumes, 2),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_specialDataDisk(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxVolumes(&server.Volumes, 1),
				),
			},
			{
				Config: testAccCheckBrightboxServerConfig_typechange40G(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxServerType(&server.ServerType, 2, 8192),
					resource.TestCheckResourceAttrSet(
						resourceName, "type"),
					testAccCheckBrightboxVolumes(&server.Volumes, 0),
				),
			},
		},
	})
}

func testAccCheckBrightboxVolumes(volumeRef *[]brightbox.Volume, numVolumes int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bootVolumes := filter(*volumeRef, func(v brightbox.Volume) bool { return v.Boot })
		dataVolumes := filter(*volumeRef, func(v brightbox.Volume) bool { return !v.Boot })

		if len(bootVolumes) != 1 {
			return fmt.Errorf("Expected a boot volume, got %d", len(bootVolumes))
		}

		if len(dataVolumes) != numVolumes {
			return fmt.Errorf("Expected %d data volume(s), got %d", numVolumes, len(dataVolumes))
		}

		return nil
	}
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

const TestAccBrightboxDataServerTypeConfig_8gbserver = `
data "brightbox_server_type" "barfoo" {
	handle = "^8gb.nbs$"
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
		TestAccBrightboxDataServerTypeConfig_8gbserver,
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

// func testAccCheckBrightboxServerConfig_bootVolume(rInt int) string {
// 	return fmt.Sprintf(`
// resource "brightbox_server" "foobar" {
// 	name = "foo-%d"
// 	type = data.brightbox_server_type.foobar.id
// 	server_groups = [data.brightbox_server_group.default.id]
// 	user_data = "foo:-with-character's"
// 	volume = brightbox_volume.foobar.id
// }

// %s%s%s`, rInt, testAccCheckBrightboxVolumeConfig_basic(rInt),
// 		TestAccBrightboxDataServerGroupConfig_default,
// 		TestAccBrightboxDataServerTypeConfig_network_disk)
// }

func testAccCheckBrightboxServerConfig_networkDataDisk(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = data.brightbox_server_type.foobar.id
	server_groups = [data.brightbox_server_group.default.id]
	disk_size = 40960
	data_volumes = [brightbox_volume.foobar.id]
}

%s%s%s%s`, rInt, TestAccBrightboxImageDataSourceConfig_ubuntu_latest_official,
		TestAccBrightboxDataServerGroupConfig_default,
		TestAccBrightboxDataServerTypeConfig_network_disk,
		testAccCheckBrightboxVolumeConfig_rawMinimal(rInt),
	)
}

func testAccCheckBrightboxServerConfig_twoDataDisk(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = data.brightbox_server_type.foobar.id
	server_groups = [data.brightbox_server_group.default.id]
	disk_size = 40960
	data_volumes = [brightbox_volume.barfoo.id, brightbox_volume.foobar.id]
}

resource "brightbox_volume" "barfoo" {
	name = "foo-%d"
	size = 61440
	filesystem_type = "ext4"
}

%s%s%s%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_ubuntu_latest_official,
		TestAccBrightboxDataServerGroupConfig_default,
		TestAccBrightboxDataServerTypeConfig_network_disk,
		testAccCheckBrightboxVolumeConfig_rawMinimal(rInt),
	)
}

func testAccCheckBrightboxServerConfig_specialDataDisk(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	type = data.brightbox_server_type.foobar.id
	server_groups = [data.brightbox_server_group.default.id]
	disk_size = 40960
	data_volumes = [brightbox_volume.barfoo.id]
}

resource "brightbox_volume" "barfoo" {
	name = "foo-%d"
	size = 61440
	filesystem_type = "ext4"
}

%s%s%s%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_ubuntu_latest_official,
		TestAccBrightboxDataServerGroupConfig_default,
		TestAccBrightboxDataServerTypeConfig_network_disk,
		testAccCheckBrightboxVolumeConfig_rawMinimal(rInt),
	)
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
				var apierror *brightbox.APIError
				if errors.As(err, &apierror) {
					if apierror.StatusCode >= 200 && apierror.StatusCode <= 299 {
						log.Printf("[DEBUG] Parse Error %+v", apierror.Unwrap())
						fred := apierror.Unwrap()
						return fred.(*brightbox.APIError)
					}
					return apierror
				}
			}
			for _, object := range objects {
				if object.Status != serverstatus.Active {
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
