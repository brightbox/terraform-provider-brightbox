package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccBrightboxCloudip_importBasic(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importEmptyName(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_empty_name,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importMapped(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_mapped(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importPortMapped(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_port_mapped(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importRemapped(t *testing.T) {
	resourceName := "brightbox_cloudip.foobar"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_remapped(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importMappedDb(t *testing.T) {
	resourceName := "brightbox_cloudip.dbmap"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxCloudipConfig_mappedDb(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxCloudip_importMappedGroup(t *testing.T) {
	resourceName := "brightbox_cloudip.groupmap"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
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

func TestAccBrightboxCloudip_importMappedLb(t *testing.T) {
	resourceName := "brightbox_cloudip.lbmap"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxCloudipDestroy,
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
	target = "${brightbox_server_group.barfoo.id}"
}

%s`, rInt, testAccCheckBrightboxServerConfig_server_group(rInt))
}

func testAccCheckBrightboxCloudipConfig_mappedLb(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_cloudip" "lbmap" {
	name = "bar-%d"
	target = "${brightbox_load_balancer.default.id}"
}

%s`, rInt, testAccCheckBrightboxLoadBalancerConfig_basic)
}

func testAccCheckBrightboxCloudipConfig_mappedDb(rInt int) string {
	return fmt.Sprintf(`

resource "brightbox_cloudip" "dbmap" {
	name = "bar-%d"
	target = "${brightbox_database_server.default.id}"
}

%s`, rInt, testAccCheckBrightboxDatabaseServerConfig_basic("dbmap"))
}
