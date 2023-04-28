package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccConfigDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "adguard_config" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.adguard_config.test", "filtering.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "filtering.update_interval", "24"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safebrowsing.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "parental_control.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.#", "6"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.0", "bing"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.4", "yandex"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.anonymize_client_ip", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.interval", "4"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.0", "abc.com"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.1", "example.com"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "querylog.ignored.2", "test.com"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "stats.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "stats.interval", "8"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "stats.ignored.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "stats.ignored.0", "domain1.com"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "stats.ignored.1", "ignored.net"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "stats.ignored.2", "test3.zyx"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "blocked_services.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "blocked_services.1", "instagram"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.bootstrap_dns.#", "4"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.bootstrap_dns.0", "9.9.9.10"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_dns.#", "1"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_dns.0", "https://dns10.quad9.net/dns-query"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.rate_limit", "20"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocking_mode", "default"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocking_ipv4", ""),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocking_ipv6", ""),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.edns_cs_enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.disable_ipv6", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.dnssec_enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_size", "4194304"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_ttl_min", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_ttl_max", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_optimistic", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_mode", "load_balance"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.use_private_ptr_resolvers", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.resolve_clients", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.local_ptr_upstreams.#", "0"),
					// Verify internal attributes
					resource.TestCheckResourceAttr("data.adguard_config.test", "id", "placeholder"),
					resource.TestCheckResourceAttrSet("data.adguard_config.test", "last_updated"),
				),
			},
		},
	})
}
