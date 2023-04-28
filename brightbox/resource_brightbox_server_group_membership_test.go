package brightbox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/exp/slices"
)

func TestAccBrightboxServerGroupMembership_basic(t *testing.T) {
	resourceName := "brightbox_server_group_membership.foobar"
	resource2Name := "brightbox_server_group_membership.barfoo"
	var group1, group2 brightbox.ServerGroup
	var server1, server2, server3 brightbox.Server
	groupName1 := acctest.RandomWithPrefix("foo")
	groupName2 := acctest.RandomWithPrefix("foo")
	serverName1 := acctest.RandomWithPrefix("foo")
	serverName2 := acctest.RandomWithPrefix("foo")
	serverName3 := acctest.RandomWithPrefix("foo")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxServerGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerGroupMembershipConfig_init(
					groupName1,
					groupName2,
					serverName1,
					serverName2,
					serverName3,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&group1,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&group2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server1",
						"Server",
						&server1,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server2",
						"Server",
						&server2,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server3",
						"Server",
						&server3,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "group", &group1.ID),
					testAccCheckServerListforGroup(
						&group1,
						[]*brightbox.Server{&server1},
						[]*brightbox.Server{&server2, &server3},
					),
					testAccCheckServerListforGroup(
						&group2,
						[]*brightbox.Server{},
						[]*brightbox.Server{&server1, &server2, &server3},
					),
				),
			},
			// test resource import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccBrightboxServerGroupMembershipImportStateIDFunc(resourceName),
				// We do not have a way to align IDs since the Create function uses id.UniqueId()
				// Failed state verification, resource with ID USER/GROUP not found
				// ImportStateVerify: true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}

					return nil
				},
			},
			// test adding an additional group to an existing resource
			{
				Config: testAccBrightboxServerGroupMembershipConfig_addOne(
					groupName1,
					groupName2,
					serverName1,
					serverName2,
					serverName3,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&group1,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&group2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server1",
						"Server",
						&server1,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server2",
						"Server",
						&server2,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server3",
						"Server",
						&server3,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "group", &group1.ID),
					testAccCheckServerListforGroup(
						&group1,
						[]*brightbox.Server{&server1, &server2},
						[]*brightbox.Server{&server3},
					),
					testAccCheckServerListforGroup(
						&group2,
						[]*brightbox.Server{},
						[]*brightbox.Server{&server1, &server2, &server3},
					),
				),
			},
			// test adding multiple servers for the same group, and the same server for another group
			{
				Config: testAccBrightboxServerGroupMembershipConfig_addAll(
					groupName1,
					groupName2,
					serverName1,
					serverName2,
					serverName3,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&group1,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&group2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server1",
						"Server",
						&server1,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server2",
						"Server",
						&server2,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server3",
						"Server",
						&server3,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "group", &group1.ID),
					resource.TestCheckResourceAttrPtr(
						resource2Name, "group", &group2.ID),
					testAccCheckServerListforGroup(
						&group1,
						[]*brightbox.Server{&server1, &server2, &server3},
						[]*brightbox.Server{},
					),
					testAccCheckServerListforGroup(
						&group2,
						[]*brightbox.Server{&server1, &server2, &server3},
						[]*brightbox.Server{},
					),
				),
			},
			// test that nothing happens when we apply the same config again
			{
				Config: testAccBrightboxServerGroupMembershipConfig_addAll(
					groupName1,
					groupName2,
					serverName1,
					serverName2,
					serverName3,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&group1,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&group2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server1",
						"Server",
						&server1,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server2",
						"Server",
						&server2,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server3",
						"Server",
						&server3,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "group", &group1.ID),
					resource.TestCheckResourceAttrPtr(
						resource2Name, "group", &group2.ID),
					testAccCheckServerListforGroup(
						&group1,
						[]*brightbox.Server{&server1, &server2, &server3},
						[]*brightbox.Server{},
					),
					testAccCheckServerListforGroup(
						&group2,
						[]*brightbox.Server{&server1, &server2, &server3},
						[]*brightbox.Server{},
					),
				),
			},
			// test removing a server
			{
				Config: testAccBrightboxServerGroupMembershipConfig_remove(
					groupName1,
					groupName2,
					serverName1,
					serverName2,
					serverName3,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&group1,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&group2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server1",
						"Server",
						&server1,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server2",
						"Server",
						&server2,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server3",
						"Server",
						&server3,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "group", &group1.ID),
					resource.TestCheckResourceAttrPtr(
						resource2Name, "group", &group2.ID),
					testAccCheckServerListforGroup(
						&group1,
						[]*brightbox.Server{&server1, &server3},
						[]*brightbox.Server{&server2},
					),
					testAccCheckServerListforGroup(
						&group2,
						[]*brightbox.Server{&server1, &server3},
						[]*brightbox.Server{&server2},
					),
				),
			},
			// test removing a resource
			{
				Config: testAccBrightboxServerGroupMembershipConfig_deleteResource(
					groupName1,
					groupName2,
					serverName1,
					serverName2,
					serverName3,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group1",
						"Server Group",
						&group1,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server_group.group2",
						"Server Group",
						&group2,
						(*brightbox.Client).ServerGroup,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server1",
						"Server",
						&server1,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server2",
						"Server",
						&server2,
						(*brightbox.Client).Server,
					),
					testAccCheckBrightboxObjectExists(
						"brightbox_server.server3",
						"Server",
						&server3,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttrPtr(
						resourceName, "group", &group1.ID),
					resource.TestCheckResourceAttrPtr(
						resource2Name, "group", &group2.ID),
					testAccCheckServerListforGroup(
						&group1,
						[]*brightbox.Server{&server1, &server3},
						[]*brightbox.Server{&server2},
					),
					testAccCheckServerListforGroup(
						&group2,
						[]*brightbox.Server{&server1},
						[]*brightbox.Server{&server2, &server3},
					),
				),
			},
		},
	})
}

func testAccBrightboxServerGroupMembershipImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		serverCount, _ := strconv.Atoi(rs.Primary.Attributes["servers.#"])
		stateID := rs.Primary.Attributes["group"]
		for i := 0; i < serverCount; i++ {
			serverID := rs.Primary.Attributes[fmt.Sprintf("server.%d", i)]
			stateID = fmt.Sprintf("%s/%s", stateID, serverID)
		}
		return stateID, nil
	}
}

func testAccCheckBrightboxServerGroupMembershipDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_server_group_membership" {
			continue
		}

		group := rs.Primary.Attributes["group"]
		object, err := client.ServerGroup(context.Background(), group)
		if err != nil {
			var apierror *brightbox.APIError
			if errors.As(err, &apierror) {
				if apierror.StatusCode == 404 {
					continue
				}
				return err
			}
		}
		foundServers := len(object.Servers)
		if foundServers > 0 {
			return fmt.Errorf("Expected all group membership for user to be removed, found: %d", foundServers)
		}
	}
	return nil
}

func testAccCheckServerListforGroup(
	group *brightbox.ServerGroup,
	serversPlus []*brightbox.Server,
	serversNeg []*brightbox.Server,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groupName := group.ID
		var serverIDs []string
		for _, server := range group.Servers {
			serverIDs = append(serverIDs, server.ID)
		}
		var serverIDsPlus []string
		for _, server := range serversPlus {
			serverIDsPlus = append(serverIDsPlus, server.ID)
		}
		var serverIDsNeg []string
		for _, server := range serversNeg {
			serverIDsNeg = append(serverIDsNeg, server.ID)
		}
		for _, excluded := range serverIDsNeg {
			if slices.Contains(serverIDs, excluded) {
				return fmt.Errorf("Unexpected server ID found in %s: %s", groupName, excluded)
			}
		}
		for _, included := range serverIDsPlus {
			if !slices.Contains(serverIDs, included) {
				return fmt.Errorf("Required server ID not found in %s: %s", groupName, included)

			}
		}
		return nil
	}
}

func testAccServerGroupMembershipConfig_base(groupName1, groupName2, serverName1, serverName2, serverName3 string) string {
	return configCompose(
		fmt.Sprintf(`
resource "brightbox_server_group" "group1" {
	name = %[1]q
}

resource "brightbox_server_group" "group2" {
	name = %[2]q
}

resource "brightbox_server" "server1" {
    image = data.brightbox_image.foobar.id
	name = %[3]q
	type = "1gb.ssd"
}

resource "brightbox_server" "server2" {
    image = data.brightbox_image.foobar.id
	name = %[4]q
	type = "1gb.ssd"
}

resource "brightbox_server" "server3" {
    image = data.brightbox_image.foobar.id
	name = %[5]q
	type = "1gb.ssd"
}

`, groupName1, groupName2, serverName1, serverName2, serverName3),
		TestAccBrightboxImageDataSourceConfig_ubuntu_latest_official,
		// TestAccBrightboxDataServerGroupConfig_default
	)
}

func configCompose(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}

func testAccCheckBrightboxServerGroupMembershipConfig_init(
	groupName1,
	groupName2,
	serverName1,
	serverName2,
	serverName3 string,
) string {
	return configCompose(
		`
resource "brightbox_server_group_membership" "foobar" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server1.id
	]
}

`,
		testAccServerGroupMembershipConfig_base(
			groupName1,
			groupName2,
			serverName1,
			serverName2,
			serverName3,
		),
	)
}

func testAccBrightboxServerGroupMembershipConfig_addOne(
	groupName1,
	groupName2,
	serverName1,
	serverName2,
	serverName3 string,
) string {
	return configCompose(
		`
resource "brightbox_server_group_membership" "foobar" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server1.id,
		brightbox_server.server2.id,
	]
}

`,
		testAccServerGroupMembershipConfig_base(
			groupName1,
			groupName2,
			serverName1,
			serverName2,
			serverName3,
		),
	)
}

func testAccBrightboxServerGroupMembershipConfig_addAll(
	groupName1,
	groupName2,
	serverName1,
	serverName2,
	serverName3 string,
) string {
	return configCompose(
		`
resource "brightbox_server_group_membership" "foobar" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server1.id,
		brightbox_server.server2.id,
	]
}

resource "brightbox_server_group_membership" "foobar2" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server3.id,
	]
}

resource "brightbox_server_group_membership" "barfoo" {
	group = brightbox_server_group.group2.id
	servers = [
		brightbox_server.server1.id,
		brightbox_server.server2.id,
	]
}

resource "brightbox_server_group_membership" "barfoo2" {
	group = brightbox_server_group.group2.id
	servers = [
		brightbox_server.server3.id,
	]
}

`,
		testAccServerGroupMembershipConfig_base(
			groupName1,
			groupName2,
			serverName1,
			serverName2,
			serverName3,
		),
	)
}

func testAccBrightboxServerGroupMembershipConfig_remove(
	groupName1,
	groupName2,
	serverName1,
	serverName2,
	serverName3 string,
) string {
	return configCompose(
		`
resource "brightbox_server_group_membership" "foobar" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server1.id,
	]
}

resource "brightbox_server_group_membership" "foobar2" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server3.id,
	]
}

resource "brightbox_server_group_membership" "barfoo" {
	group = brightbox_server_group.group2.id
	servers = [
		brightbox_server.server1.id,
	]
}

resource "brightbox_server_group_membership" "barfoo2" {
	group = brightbox_server_group.group2.id
	servers = [
		brightbox_server.server3.id,
	]
}

`,
		testAccServerGroupMembershipConfig_base(
			groupName1,
			groupName2,
			serverName1,
			serverName2,
			serverName3,
		),
	)
}

func testAccBrightboxServerGroupMembershipConfig_deleteResource(
	groupName1,
	groupName2,
	serverName1,
	serverName2,
	serverName3 string,
) string {
	return configCompose(
		`
resource "brightbox_server_group_membership" "foobar" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server1.id,
	]
}

resource "brightbox_server_group_membership" "foobar2" {
	group = brightbox_server_group.group1.id
	servers = [
		brightbox_server.server3.id,
	]
}

resource "brightbox_server_group_membership" "barfoo" {
	group = brightbox_server_group.group2.id
	servers = [
		brightbox_server.server1.id,
	]
}

`,
		testAccServerGroupMembershipConfig_base(
			groupName1,
			groupName2,
			serverName1,
			serverName2,
			serverName3,
		),
	)
}
