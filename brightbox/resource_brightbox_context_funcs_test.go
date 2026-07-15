package brightbox

import (
	"context"
	"errors"
	"fmt"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/go-cty/cty"
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

func testAccCheckBrightboxDataSourceID(objectName string, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find %s data source: %s", objectName, n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("%s data source ID not set", objectName)
		}

		return nil
	}
}

func testAccCheckBrightboxObjectExists[O any](
	n string,
	objectName string,
	object *O,
	instance func(*brightbox.Client, context.Context, string) (*O, error),
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No %s ID is set", objectName)
		}
		client := testAccProvider.Meta().(*CompositeClient).APIClient
		retrieveobject, err := instance(client, context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}
		*object = *retrieveobject
		return nil
	}
}

// fullSchemaObject builds a cty.Value of ty's exact object type, defaulting
// every attribute to null and overriding the ones given. Needed because
// cty.ObjectVal requires every attribute of the target type to be present.
// Shared by the per-resource *_snapshots_test.go Plan->Apply tests.
func fullSchemaObject(ty cty.Type, overrides map[string]cty.Value) cty.Value {
	vals := map[string]cty.Value{}
	for k, at := range ty.AttributeTypes() {
		if v, ok := overrides[k]; ok {
			vals[k] = v
		} else {
			vals[k] = cty.NullVal(at)
		}
	}
	return cty.ObjectVal(vals)
}
