package brightbox

import (
	"fmt"
	"testing"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBrightboxCloudip_Basic(t *testing.T) {
	var cloudip brightbox.CloudIP

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					testAccCheckBrightboxCloudipAttributes(&cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "unmapped"),
					resource.TestCheckNoResourceAttr(
						"brightbox_cloudip.foobar", "target"),
				),
			},
		},
	})
}

func TestAccBrightboxCloudip_clear_name(t *testing.T) {
	var cloudip brightbox.CloudIP

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					testAccCheckBrightboxCloudipAttributes(&cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "unmapped"),
					resource.TestCheckNoResourceAttr(
						"brightbox_cloudip.foobar", "target"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_empty_name,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", ""),
					resource.TestCheckNoResourceAttr(
						"brightbox_cloudip.foobar", "target"),
				),
			},
		},
	})
}

func TestAccBrightboxCloudip_Mapped(t *testing.T) {
	var cloudip brightbox.CloudIP

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_mapped,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "mapped"),
				),
			},
		},
	})
}

func TestAccBrightboxCloudip_Remapped(t *testing.T) {
	var cloudip brightbox.CloudIP

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "unmapped"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_mapped,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "mapped"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_remapped,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "remapped"),
				),
			},
			resource.TestStep{
				Config: testAccCheckBrightboxCloudipConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxCloudipExists("brightbox_cloudip.foobar", &cloudip),
					resource.TestCheckResourceAttr(
						"brightbox_cloudip.foobar", "name", "unmapped"),
				),
			},
		},
	})
}

func testAccCheckBrightboxCloudipDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).ApiClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_cloudip" {
			continue
		}

		// Try to find the CloudIP
		_, err := client.CloudIP(rs.Primary.ID)

		// Wait

		if err != nil {
			apierror := err.(brightbox.ApiError)
			if apierror.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for cloudip %s to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func testAccCheckBrightboxCloudipExists(n string, cloudip *brightbox.CloudIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudIP ID is set")
		}

		client := testAccProvider.Meta().(*CompositeClient).ApiClient

		// Try to find the CloudIP
		retrieveCloudip, err := client.CloudIP(rs.Primary.ID)

		if err != nil {
			return err
		}

		if retrieveCloudip.Id != rs.Primary.ID {
			return fmt.Errorf("CloudIP not found")
		}

		*cloudip = *retrieveCloudip

		return nil
	}
}

func testAccCheckBrightboxCloudipAttributes(cloudip *brightbox.CloudIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if cloudip.Name != "unmapped" {
			return fmt.Errorf("Bad name: %s", cloudip.Name)
		}
		return nil
	}
}

const testAccCheckBrightboxCloudipConfig_basic = `

resource "brightbox_cloudip" "foobar" {
	name = "unmapped"
}
`

const testAccCheckBrightboxCloudipConfig_empty_name = `

resource "brightbox_cloudip" "foobar" {
	name = ""
}
`

var testAccCheckBrightboxCloudipConfig_mapped = fmt.Sprintf(`

resource "brightbox_cloudip" "foobar" {
	name = "mapped"
	target = "${brightbox_server.boofar.interface}"
}

resource "brightbox_server" "boofar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "map_cip_test"
}
%s`, TestAccBrightboxImageDataSourceConfig_blank_disk)

var testAccCheckBrightboxCloudipConfig_remapped = fmt.Sprintf(`

resource "brightbox_cloudip" "foobar" {
	name = "remapped"
	target = "${brightbox_server.fred.interface}"
}

resource "brightbox_server" "boofar" {
	image = "${data.brightbox_image.foobar.id}"
	name = "map_cip_test"
}

resource "brightbox_server" "fred" {
	image = "${data.brightbox_image.foobar.id}"
	name = "remap_cip_test"
}
%s`, TestAccBrightboxImageDataSourceConfig_blank_disk)
