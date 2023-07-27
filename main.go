package main

import (
	"flag"
	"log"

	"github.com/brightbox/terraform-provider-brightbox/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
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

	providerServerCreator, err := provider.BrightboxTF5ProviderServerCreatorFactory(version)
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = tf5server.Serve(
		"registry.terraform.io/brightbox/brightbox",
		providerServerCreator,
		opts...,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
}
