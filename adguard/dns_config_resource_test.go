package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsConfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_dns_config" "test" {
  upstream_dns = ["https://1.1.1.1/dns-query",  "https://1.0.0.1/dns-query"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_dns.#", "2"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_dns.1", "https://1.0.0.1/dns-query"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("adguard_dns_config.test", "id"),
					resource.TestCheckResourceAttrSet("adguard_dns_config.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_dns_config.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_dns_config" "test" {
  upstream_dns = ["https://1.1.1.1/dns-query",  "https://1.0.0.1/dns-query"]
	ratelimit    = 25
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_dns_config.test", "ratelimit", "25"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
