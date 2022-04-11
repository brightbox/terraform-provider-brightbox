package brightbox

import (
	"context"
	"errors"
	"fmt"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckBrightboxDestroyBuilder[I any](
	objectName string,
	instance func(*brightbox.Client, context.Context, string) (*I, error),
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*CompositeClient).APIClient

		for _, rs := range s.RootModule().Resources {
			if rs.Type != objectName {
				continue
			}

			// Try to find the Instance
			_, err := instance(client, context.Background(), rs.Primary.ID)

			// Wait

			if err != nil {
				var apierror *brightbox.APIError
				if errors.As(err, &apierror) {
					if apierror.StatusCode != 404 {
						return fmt.Errorf(
							"Error waiting for %s %s to be destroyed: %s",
							strings.TrimPrefix(objectName, "brightbox_"),
							rs.Primary.ID,
							err,
						)
					}
				}
			}
		}

		return nil
	}
}
