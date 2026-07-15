package brightbox

import (
	"fmt"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBrightboxServer_Snapshots(t *testing.T) {
	var server brightbox.Server
	rInt := acctest.RandInt()
	name := fmt.Sprintf("foo-%d", rInt)
	resourceName := "brightbox_server.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxServerConfig_noSnapshots(name),
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
				Config: testAccCheckBrightboxServerConfig_withSnapshots(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
					),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_schedule", "0 7 * * *"),
					resource.TestCheckResourceAttr(
						resourceName, "snapshots_retention", "5"),
				),
			},
			{
				// Omits both fields to confirm they correctly pull API values
				Config: testAccCheckBrightboxServerConfig_noSnapshots(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Server",
						&server,
						(*brightbox.Client).Server,
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

func testAccCheckBrightboxServerConfig_noSnapshots(name string) string {
	return fmt.Sprintf(`

resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "%s"
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
	user_data = "foo:-with-character's"
}
%s%s`, name, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxServerConfig_withSnapshots(name string) string {
	return fmt.Sprintf(`

resource "brightbox_server" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "%s"
	type = "1gb.ssd"
	server_groups = [data.brightbox_server_group.default.id]
	user_data = "foo:-with-character's"
	snapshots_schedule = "0 7 * * *"
	snapshots_retention = "5"
}
%s%s`, name, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}
