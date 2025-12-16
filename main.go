package main

import (
	"context"
	"log"

	"github.com/dmachard/terraform-provider-http-client/httpclient"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), httpclient.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/dmachard/httpclient",
	})
	if err != nil {
		log.Fatalf("error serving provider: %s", err)
	}
}
