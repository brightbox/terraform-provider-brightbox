package brightbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var accountRe = regexp.MustCompile("acc-.....")
var disktypeRe = regexp.MustCompile("uefi1.img")

func TestAccBrightboxImageDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBrightboxImageDataSource_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagesDataSourceID("data.brightbox_image.foobar"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "source_type", "upload"),
					resource.TestMatchResourceAttr(
						"data.brightbox_image.foobar", "owner", accountRe),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "status", "available"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "locked", "false"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "arch", "x86_64"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "name", "Blank Disk Image"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "description", ""),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "username", ""),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "virtual_size", "0"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "disk_size", "0"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "public", "true"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "compatibility_mode", "false"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "official", "true"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "ancestor_id", ""),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "licence_name", ""),
				),
			},
		},
	})
}

func TestAccBrightboxImageDataSource_ubuntu_xenial_official(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBrightboxImageDataSource_ubuntu_xenial_official,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagesDataSourceID("data.brightbox_image.foobar"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "source_type", "upload"),
					resource.TestMatchResourceAttr(
						"data.brightbox_image.foobar", "owner", accountRe),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "status", "available"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "locked", "false"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "arch", "x86_64"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "name", "ubuntu-xenial-16.04-amd64-server-uefi1"),
					resource.TestMatchResourceAttr(
						"data.brightbox_image.foobar", "description", disktypeRe),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "username", "ubuntu"),
					resource.TestCheckResourceAttrSet(
						"data.brightbox_image.foobar", "virtual_size"),
					resource.TestCheckResourceAttrSet(
						"data.brightbox_image.foobar", "disk_size"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "public", "true"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "compatibility_mode", "false"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "official", "true"),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "ancestor_id", ""),
					resource.TestCheckResourceAttr(
						"data.brightbox_image.foobar", "licence_name", ""),
				),
			},
		},
	})
}

func testAccCheckImagesDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find image data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Image data source ID not set")
		}

		return nil
	}
}

const testAccBrightboxImageDataSource_basic = `
data "brightbox_image" "foobar" {
	name = "^Blank Disk Image$"
	arch = "x86_64"
	official = true
}
`
// Select latest ubuntu with uefi1 image
// Checks name matches partial name
const testAccBrightboxImageDataSource_ubuntu_xenial_official = `
data "brightbox_image" "foobar" {
	name = "^ubuntu-xenial.*server"
	arch = "x86_64"
	official = true
	description = "uefi1.img"
	most_recent = true
}
`
