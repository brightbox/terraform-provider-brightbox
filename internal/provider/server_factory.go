package provider

import (
	"context"

	sdkprovider "github.com/brightbox/terraform-provider-brightbox/brightbox"
	// "github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

// BrightboxTF5ProviderServerCreatorFactory returns a function that will create a
// Provider Server.
func BrightboxTF5ProviderServerCreatorFactory(name string) (func() tfprotov5.ProviderServer, error) {
	ctx := context.Background()
	providers := []func() tfprotov5.ProviderServer{
		sdkprovider.Provider().GRPCProvider, //SDK Brightbox provider
		// providerserver.NewProtocol5(New(name)), // This provider
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}
