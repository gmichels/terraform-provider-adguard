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
  upstream_dns        = ["https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"]
  rate_limit          = 30
  cache_ttl_min       = 600
  cache_ttl_max       = 86400
  cache_optimistic    = true
  blocking_mode       = "custom_ip"
  blocking_ipv4       = "1.2.3.4"
  blocking_ipv6       = "fe80::"
  local_ptr_upstreams = ["192.168.0.1", "192.168.0.2"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_dns.#", "2"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_dns.1", "https://1.0.0.1/dns-query"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "rate_limit", "30"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "blocking_mode", "custom_ip"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "blocking_ipv4", "1.2.3.4"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "blocking_ipv6", "fe80::"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "cache_ttl_min", "600"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "cache_ttl_max", "86400"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "cache_optimistic", "true"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "local_ptr_upstreams.#", "2"),
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
  upstream_dns              = ["https://1.1.1.1/dns-query"]
  blocking_mode             = "nxdomain"
  rate_limit                = 25
  edns_cs_enabled           = true
  disable_ipv6              = true
  dnssec_enabled            = true
  cache_size                = 8000000
  upstream_mode             = "load_balance"
  use_private_ptr_resolvers = false
  resolve_clients           = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_dns.#", "1"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_dns.0", "https://1.1.1.1/dns-query"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "rate_limit", "25"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "blocking_mode", "nxdomain"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "blocking_ipv4", ""),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "blocking_ipv6", ""),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "edns_cs_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "disable_ipv6", "true"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "dnssec_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "cache_size", "8000000"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "upstream_mode", "load_balance"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "use_private_ptr_resolvers", "false"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "resolve_clients", "false"),
					resource.TestCheckResourceAttr("adguard_dns_config.test", "local_ptr_upstreams.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
