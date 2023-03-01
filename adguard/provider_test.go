package adguard

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the Adguard Home client is properly configured.
	// It is also possible to use the ADGUARD_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "adguard" {
  host     = "dns-int.michels.link"
  username = "gmichels"
  password = "DJYxNF7Aav8Uzhr90uy8"
  scheme = "https"
}
`
)

// var (
// 	// providerConfig is a shared configuration to combine with the actual
// 	// test configuration so the Adguard Home client is properly configured.
// 	providerConfig = `
// provider "adguard" {
//   host  = "` + os.Getenv("ADGUARD_HOST") + `"
//   username = "` + os.Getenv("ADGUARD_USERNAME") + `"
//   password = "` + os.Getenv("ADGUARD_PASSWORD") + `"
//   scheme = "` + os.Getenv("ADGUARD_SCHEME") + `"
// }
// `
// )

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"adguard": providerserver.NewProtocol6WithError(New()),
	}
)
