package brightbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/storagetype"
	"github.com/brightbox/gobrightbox/v2/enums/volumestatus"
	"github.com/brightbox/gobrightbox/v2/enums/volumetype"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var sourceRegexp = regexp.MustCompile(imageRegexp.String() + "|" + volumeRegexp.String())

func TestAccBrightboxVolume_Image(t *testing.T) {
	resourceName := "brightbox_volume.foobar"
	var volume brightbox.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxVolumeConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_label", ""),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_type", ""),
					resource.TestMatchResourceAttr(
						resourceName, "source", sourceRegexp),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "20480"),
					resource.TestCheckResourceAttr(
						resourceName, "serial", fmt.Sprintf("%020d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "source_type", volumetype.Image.String()),
					resource.TestCheckResourceAttr(
						resourceName, "storage_type", storagetype.Network.String()),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxVolumeConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_label", ""),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_type", ""),
					resource.TestMatchResourceAttr(
						resourceName, "source", sourceRegexp),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "20480"),
					resource.TestCheckResourceAttr(
						resourceName, "serial", fmt.Sprintf("%020d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "source_type", volumetype.Image.String()),
					resource.TestCheckResourceAttr(
						resourceName, "storage_type", storagetype.Network.String()),
				),
			},
			{
				Config: testAccCheckBrightboxVolumeConfig_locked(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "locked", "true"),
				),
			},
			{
				Config: testAccCheckBrightboxVolumeConfig_basic(rInt),
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
		},
	})
}

func TestAccBrightboxVolume_rawDefault(t *testing.T) {
	resourceName := "brightbox_volume.foobar"
	var volume brightbox.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxVolumeConfig_rawDefault(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_label", ""),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_type", ""),
					resource.TestCheckResourceAttr(
						resourceName, "source", ""),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "encrypted", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Hello"),
					resource.TestCheckResourceAttr(
						resourceName, "size", "20480"),
					resource.TestMatchResourceAttr(
						resourceName, "serial", volumeRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "source_type", volumetype.Raw.String()),
					resource.TestCheckResourceAttr(
						resourceName, "storage_type", storagetype.Network.String()),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
			{
				Config: testAccCheckBrightboxVolumeConfig_rawMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccBrightboxVolume_raw(t *testing.T) {
	resourceName := "brightbox_volume.foobar"
	var volume brightbox.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxVolumeConfig_rawSized(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_label", ""),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_type", ""),
					resource.TestCheckResourceAttr(
						resourceName, "source", ""),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "61440"),
					resource.TestCheckResourceAttr(
						resourceName, "serial", fmt.Sprintf("%020d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "source_type", volumetype.Raw.String()),
					resource.TestCheckResourceAttr(
						resourceName, "storage_type", storagetype.Network.String()),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
				),
			},
		},
	})
}

func TestAccBrightboxVolume_resize(t *testing.T) {
	resourceName := "brightbox_volume.foobar"
	var volume brightbox.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxVolumeConfig_rawMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "20480"),
				),
			},
			{
				Config: testAccCheckBrightboxVolumeConfig_rawSized(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "61440"),
					resource.TestCheckResourceAttr(
						resourceName, "serial", fmt.Sprintf("%020d", rInt)),
				),
			},
		},
	})
}

func TestAccBrightboxVolume_Formatted(t *testing.T) {
	resourceName := "brightbox_volume.foobar"
	var volume brightbox.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxVolumeConfig_rawFormatted(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_label", ""),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_type", "ext4"),
					resource.TestCheckResourceAttr(
						resourceName, "source", ""),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "61440"),
					resource.TestCheckResourceAttr(
						resourceName, "source_type", volumetype.Raw.String()),
					resource.TestCheckResourceAttr(
						resourceName, "storage_type", storagetype.Network.String()),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "serial", volumeRegexp),
				),
			},
		},
	})
}

func TestAccBrightboxVolume_FormattedLabelled(t *testing.T) {
	resourceName := "brightbox_volume.foobar"
	var volume brightbox.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(),
		CheckDestroy:      testAccCheckBrightboxVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxVolumeConfig_rawFormattedLabel(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrightboxObjectExists(
						resourceName,
						"Volume",
						&volume,
						(*brightbox.Client).Volume,
					),
					testAccCheckBrightboxVolumeAttributes(&volume),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_label", "123456789012"),
					resource.TestCheckResourceAttr(
						resourceName, "filesystem_type", "xfs"),
					resource.TestCheckResourceAttr(
						resourceName, "source", ""),
					resource.TestMatchResourceAttr(
						resourceName, "image", imageRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "description", ""),
					resource.TestCheckResourceAttr(
						resourceName, "size", "61440"),
					resource.TestCheckResourceAttr(
						resourceName, "source_type", volumetype.Raw.String()),
					resource.TestCheckResourceAttr(
						resourceName, "storage_type", storagetype.Network.String()),
					resource.TestCheckResourceAttr(
						resourceName, "locked", "false"),
					resource.TestMatchResourceAttr(
						resourceName, "serial", volumeRegexp),
				),
			},
		},
	})
}

func testAccCheckBrightboxVolumeConfig_locked(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	serial = "%020d"
	locked = true
	size = 20480
}

%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_ubuntu_latest_official)

}

func testAccCheckBrightboxVolumeConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	image = data.brightbox_image.foobar.id
	name = "foo-%d"
	serial = "%020d"
	size = 20480
}

%s`, rInt, rInt, TestAccBrightboxImageDataSourceConfig_ubuntu_latest_official)

}

func testAccCheckBrightboxVolumeConfig_rawDefault(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	name = "foo-%d"
	description = "Hello"
	size = 20480
}

`, rInt)

}

func testAccCheckBrightboxVolumeConfig_rawMinimal(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	name = "foo-%d"
	size = 20480
}

`, rInt)

}

func testAccCheckBrightboxVolumeConfig_rawSized(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	name = "foo-%d"
	serial = "%020d"
	size = 61440
}

`, rInt, rInt)

}

func testAccCheckBrightboxVolumeConfig_rawFormatted(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	name = "foo-%d"
	size = 61440
	filesystem_type = "ext4"
}

`, rInt)

}

func testAccCheckBrightboxVolumeConfig_rawFormattedLabel(rInt int) string {
	return fmt.Sprintf(`
resource "brightbox_volume" "foobar" {
	name = "foo-%d"
	size = 61440
	filesystem_type = "xfs"
	filesystem_label = "123456789012"
}

`, rInt)

}

func testAccCheckBrightboxVolumeAttributes(volume *brightbox.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if !imageRegexp.MatchString(volume.Image.ID) {
			return fmt.Errorf("Bad image id: %s", volume.Image.ID)
		}

		return nil
	}
}

func testAccCheckBrightboxVolumeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CompositeClient).APIClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "brightbox_volume" {
			continue
		}

		// Try to find the Volume
		volume, err := client.Volume(context.Background(), rs.Primary.ID)

		// Wait

		if err != nil {
			var apierror *brightbox.APIError
			if errors.As(err, &apierror) {
				if apierror.StatusCode != 404 {
					return fmt.Errorf(
						"Error waiting for volume (%s) to be destroyed: %s",
						rs.Primary.ID, err)
				}
			}
		} else if volume.Status != volumestatus.Deleted &&
			volume.Status != volumestatus.Deleting {
			return fmt.Errorf(
				"Volume %s not in deleted state. Status is %s", rs.Primary.ID, volume.Status)
		}
	}

	return nil
}

func init() {
	resource.AddTestSweepers("volume", &resource.Sweeper{
		Name: "volume",
		F: func(_ string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, errs := obtainCloudClient()
			if errs != nil {
				return fmt.Errorf(errs[0].Summary)
			}
			objects, err := client.APIClient.Volumes(ctx)
			if err != nil {
				return err
			}
			for _, object := range objects {
				if object.Status != volumestatus.Detached {
					continue
				}
				if isTestName(object.Name) {
					log.Printf("[INFO] removing %s named %s", object.ID, object.Name)
					if _, err := client.APIClient.UnlockVolume(ctx, object.ID); err != nil {
						log.Printf("error unlocking %s during sweep: %s", object.ID, err)
					}
					if _, err := client.APIClient.DestroyVolume(ctx, object.ID); err != nil {
						log.Printf("error destroying %s during sweep: %s", object.ID, err)
					}
				}
			}
			return nil
		},
	})
}
