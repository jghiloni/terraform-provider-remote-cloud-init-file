package main

import (
	"context"
	"flag"
	"github.com/jghiloni/terraform-provider-remote-cloud-init-file/internal/provider"
	"log"
	"os"
	"os/signal"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		// Also update the tfplugindocs generate command to either remove the
		// -provider-name flag or set its value to the updated provider name.
		Address: "registry.joshghiloni.me/jghiloni/cloud-init-files",
		Debug:   debug,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := providerserver.Serve(ctx, provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
