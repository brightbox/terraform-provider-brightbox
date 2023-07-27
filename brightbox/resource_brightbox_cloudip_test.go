package brightbox

import (
	"context"
	"fmt"
	"log"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/cloudipstatus"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBrightboxCloudip_Basic(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	var cloudIPInstance brightbox.CloudIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckNoResourceAttr(
						resourceName, "target"),
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

func TestAccBrightboxCloudip_clear_name(t *testing.T) {
	var cloudIPInstance brightbox.CloudIP
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckNoResourceAttr(
						resourceName, "target"),
				),
			},
			{
				Config: testAccCheckBrightboxCloudipConfig_empty_name,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", ""),
					resource.TestCheckNoResourceAttr(
						resourceName, "target"),
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

func TestAccBrightboxCloudip_Mapped(t *testing.T) {
	var cloudIPInstance brightbox.CloudIP
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("bar-%d", rInt)),
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

func TestAccBrightboxCloudip_PortMapped(t *testing.T) {
	var cloudIPInstance brightbox.CloudIP
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_port_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("bar-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "port_translator.#", "2"),
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

func TestAccBrightboxCloudip_RemappedSingle(t *testing.T) {
	var cloudIPInstance brightbox.CloudIP
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_remapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("baz-%d", rInt)),
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

func TestAccBrightboxCloudip_Remapped(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	var cloudIPInstance brightbox.CloudIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},
			{
				Config: testAccCheckBrightboxCloudipConfig_mapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("bar-%d", rInt)),
				),
			},
			{
				Config: testAccCheckBrightboxCloudipConfig_remapped(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("baz-%d", rInt)),
				),
			},
			{
				Config: testAccCheckBrightboxCloudipConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Cloud IP",
						&cloudIPInstance,
						(*brightbox.Client).CloudIP,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},
		},
	})
}

var testAccCheckBrightboxCloudIPDestroy = testAccCheckBrightboxDestroyBuilder(
	"brightbox_cloud_ip",
	(*brightbox.Client).CloudIP,
)

func TestAccBrightboxCloudip_MappedGroup(t *testing.T) {
	resourceName := "brightbox_cloudip.groupmap"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_mappedGroup(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_MappedLb(t *testing.T) {
	resourceName := "brightbox_cloudip.lbmap"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxCloudIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_mappedLb(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBrightboxCloudipConfig_mappedGroup(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_cloudip" "groupmap" {
	name = "bar-%d"
	target = brightbox_server_group.barfoo.id
}

%s`, rInt, testAccCheckBrightboxServerConfig_serverGroup(rInt))
}

func testAccCheckBrightboxCloudipConfig_mappedLb(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_cloudip" "lbmap" {
	name = "bar-%d"
	target = brightbox_load_balancer.default.id
}

%s`, rInt, testAccCheckBrightboxLoadBalancerConfig_basic)
}

func testAccCheckBrightboxCloudipConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_cloudip" "foobar" {
	name = "foo-%d"
}
`, rInt)
}

const testAccCheckBrightboxCloudipConfig_empty_name = `

resource "brightbox_cloudip" "foobar" {
	name = ""
}
`

func testAccCheckBrightboxCloudipConfig_mapped(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_cloudip" "foobar" {
	name = "bar-%d"
	target = brightbox_server.boofar.interface
}

resource "brightbox_server" "boofar" {
	image = data.brightbox_image.foobar.id
	name = "bar-%d"
	server_groups = [data.brightbox_server_group.default.id]
}
%s%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxCloudipConfig_port_mapped(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_cloudip" "foobar" {
	name = "bar-%d"
	target = brightbox_server.boofar.interface
	port_translator {
		protocol = "tcp"
		incoming = 80
		outgoing = 8080
	}
	port_translator {
		protocol = "udp"
		incoming = 53
		outgoing = 8053
	}
}

resource "brightbox_server" "boofar" {
	image = data.brightbox_image.foobar.id
	name = "bar-%d"
	server_groups = [data.brightbox_server_group.default.id]
}
%s%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

func testAccCheckBrightboxCloudipConfig_remapped(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_cloudip" "foobar" {
	name = "baz-%d"
	target = brightbox_server.fred.interface
}

resource "brightbox_server" "boofar" {
	image = data.brightbox_image.foobar.id
	name = "bar-%d"
	server_groups = [data.brightbox_server_group.default.id]
}

resource "brightbox_server" "fred" {
	image = data.brightbox_image.foobar.id
	name = "baz-%d"
	server_groups = [data.brightbox_server_group.default.id]
}
%s%s`, rInt, rInt, rInt, TestAccBrightboxImageDataSourceConfig_blank_disk,
		TestAccBrightboxDataServerGroupConfig_default)
}

// Sweeper

func init() {
	resource.AddTestSweepers("cloud_ip", &resource.Sweeper{
		Name: "cloud_ip",
		F: func(_ string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			objects, err := client.APIClient.CloudIPs(ctx)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Status != cloudipstatus.Unmapped {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.ID, object.Name)
					if _, err := client.APIClient.DestroyCloudIP(ctx, object.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.ID, err)
					}
				}
			}
			return nil
		},
	})
}
