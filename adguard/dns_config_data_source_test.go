package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsConfigDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "adguard_dns_config" "test" { }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "bootstrap_dns.#", "4"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "bootstrap_dns.0", "9.9.9.10"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "upstream_dns.#", "1"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "upstream_dns.0", "https://dns10.quad9.net/dns-query"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "upstream_dns_file", ""),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "protection_enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "dhcp_available", "false"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "ratelimit", "20"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "blocking_mode", "default"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "blocking_ipv4", ""),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "blocking_ipv6", ""),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "edns_cs_enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "disable_ipv6", "false"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "dnssec_enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "cache_size", "4194304"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "cache_ttl_min", "0"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "cache_ttl_max", "0"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "cache_optimistic", "false"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "upstream_mode", ""),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "use_private_ptr_resolvers", "true"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "resolve_clients", "true"),
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "local_ptr_upstreams.#", "0"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.adguard_dns_config.test", "id", "placeholder"),
				),
			},
		},
	})
}
