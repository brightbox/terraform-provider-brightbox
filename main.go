package main

import (
	"context"
	"flag"
	"log"

	sdkprovider "github.com/brightbox/terraform-provider-brightbox/brightbox"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")
	flag.Parse()
	opts := []tf5server.ServeOpt{}

	if *debugFlag {
		opts = append(opts, tf5server.WithManagedDebug())
	}

	providers := []func() tfprotov5.ProviderServer{
		//providerserver.NewProtocol5(provider.New(version)),
		sdkprovider.Provider().GRPCProvider,
	}

	// use the muxer
	muxServer, err := tf5muxserver.NewMuxServer(context.Background(), providers...)
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = tf5server.Serve(
		"registry.terraform.io/brightbox/terraform-provider-brightbox",
		muxServer.ProviderServer,
		opts...,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
}
