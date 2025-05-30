package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttr("data.adguard_config.test", "safebrowsing", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "parental_control", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.#", "7"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.0", "bing"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "safesearch.services.4", "pixabay"),
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
					resource.TestCheckResourceAttr("data.adguard_config.test", "blocked_services_pause_schedule.time_zone", "America/New_York"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "blocked_services_pause_schedule.mon.start", "00:00"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "blocked_services_pause_schedule.mon.end", "23:59"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.bootstrap_dns.#", "4"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.bootstrap_dns.0", "9.9.9.10"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_dns.#", "1"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_dns.0", "https://dns10.quad9.net/dns-query"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.fallback_dns.#", "1"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.fallback_dns.0", "9.9.9.10"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.protection_enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.rate_limit", "20"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.rate_limit_subnet_len_ipv4", "24"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocking_mode", "default"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocking_ipv4", ""),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocking_ipv6", ""),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocked_response_ttl", "10"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.edns_cs_enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.edns_cs_use_custom", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.edns_cs_custom_ip", ""),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.disable_ipv6", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.dnssec_enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_size", "4194304"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_ttl_min", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_ttl_max", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.cache_optimistic", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_mode", "load_balance"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.use_private_ptr_resolvers", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.resolve_clients", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.local_ptr_upstreams.#", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.upstream_timeout", "10"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.allowed_clients.#", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.disallowed_clients.#", "2"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.disallowed_clients.1", "test-client-access-blocked"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocked_hosts.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dns.blocked_hosts.1", "id.server"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.enabled", "false"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.interface", "lo0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.ipv4_settings.gateway_ip", "192.168.200.1"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.ipv4_settings.range_end", "192.168.200.50"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.ipv4_settings.lease_duration", "3600"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.ipv6_settings.lease_duration", "86400"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.leases.#", "3"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.leases.1.hostname", "dynamic-lease-2"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "dhcp.static_leases.#", "0"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "tls.enabled", "true"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "tls.server_name", "TestAdGuardHome"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "tls.port_https", "443"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "tls.port_dns_over_tls", "853"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "tls.certificate_chain", "/opt/adguardhome/ssl/server.crt"),
					resource.TestCheckResourceAttr("data.adguard_config.test", "tls.serve_plain_dns", "true"),
					// Verify internal attributes
					resource.TestCheckResourceAttr("data.adguard_config.test", "id", "placeholder"),
					resource.TestCheckResourceAttrSet("data.adguard_config.test", "last_updated"),
				),
			},
		},
	})
}

// piggy-backing this data source to test the provider in insecure configuration
func TestAccProviderInsecure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfigInsecure + `data "adguard_config" "test" {}`,
				// there is no need to perform checks as the test will fail if unable to connect
			},
		},
	})
}
