package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMuxServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"brightbox": func() (tfprotov5.ProviderServer, error) {
				providerServerCreator, err := BrightboxTF5ProviderServerCreatorFactory("test")
				if err != nil {
					return nil, err
				}
				return providerServerCreator(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `resource "brightbox_server_group" "default" {
						  name = "Used by the terraform"
						}`,
			},
		},
	})
}
