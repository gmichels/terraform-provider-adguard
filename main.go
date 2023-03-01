package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/gmichels/terraform-provider-adguard/adguard"
)

func main() {
	providerserver.Serve(context.Background(), adguard.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/gmichels/adguard",
	})
}
