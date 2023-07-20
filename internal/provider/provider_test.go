package provider

import (
	"context"
	"testing"

	// sdkprovider "github.com/brightbox/terraform-provider-brightbox/brightbox"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMuxServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"brightbox": func() (tfprotov5.ProviderServer, error) {
				ctx := context.Background()
				providers := []func() tfprotov5.ProviderServer{
					providerserver.NewProtocol5(New("test")), // This provider
					// sdkprovider.Provider().GRPCProvider,      //SDK Brightbox provider
				}

				muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

				if err != nil {
					return nil, err
				}

				return muxServer.ProviderServer(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: "... configuration including simplest data source or managed resource",
			},
		},
	})
}
