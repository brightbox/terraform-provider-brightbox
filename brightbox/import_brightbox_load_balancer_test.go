package brightbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccBrightboxLoadBalancer_importBasic(t *testing.T) {
	resourceName := "brightbox_load_balancer.default"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_basic,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBrightboxLoadBalancer_importSsl(t *testing.T) {
	resourceName := "brightbox_load_balancer.default"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBrightboxLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBrightboxLoadBalancerConfig_add_listener,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_pem", "certificate_private_key"},
			},
		},
	})
}
