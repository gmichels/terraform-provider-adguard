package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsAccessResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_dns_access" "test" {
  allowed_clients = ["allowed-client", "192.168.200.200"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_dns_access.test", "allowed_clients.#", "2"),
					resource.TestCheckResourceAttr("adguard_dns_access.test", "allowed_clients.1", "192.168.200.200"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("adguard_dns_access.test", "id"),
					resource.TestCheckResourceAttrSet("adguard_dns_access.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_dns_access.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_dns_access" "test" {
  disallowed_clients = ["blocked-client", "172.16.0.0/16"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_dns_access.test", "allowed_clients.#", "0"),
					resource.TestCheckResourceAttr("adguard_dns_access.test", "disallowed_clients.#", "2"),
					resource.TestCheckResourceAttr("adguard_dns_access.test", "disallowed_clients.1", "172.16.0.0/16"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
