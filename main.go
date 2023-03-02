package main

import (
	"context"

	"github.com/gmichels/terraform-provider-adguard/adguard"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// provider documentation generation
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name adguard

func main() {
	providerserver.Serve(context.Background(), adguard.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/gmichels/adguard",
	})
}
