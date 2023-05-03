package adguard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccConfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "adguard_config" "test" {
	filtering = {
		update_interval = 1
	}
	safesearch = {
		enabled  = true
		services = ["bing", "youtube", "google"]
	}
	querylog = {
		enabled             = false
		interval            = 8
		anonymize_client_ip = true
		ignored             = ["test2.com", "example2.com"]
	}
	stats = {
		enabled  = false
		interval = 2
		ignored  = ["test3.net", "example4.com"]
	}
	blocked_services = ["youtube", "pinterest"]
	dns = {
		upstream_dns        = ["https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"]
		rate_limit          = 30
		cache_ttl_min       = 600
		cache_ttl_max       = 86400
		cache_optimistic    = true
		blocking_mode       = "custom_ip"
		blocking_ipv4       = "1.2.3.4"
		blocking_ipv6       = "fe80::"
		local_ptr_upstreams = ["192.168.0.1", "192.168.0.2"]
		allowed_clients     = ["allowed-client", "192.168.200.200"]
	}
	dhcp = {
		interface = "eth1"
		ipv4_settings = {
			gateway_ip     = "192.168.250.1"
			subnet_mask    = "255.255.255.0"
			range_start    = "192.168.250.10"
			range_end      = "192.168.250.100"
			lease_duration = 7200
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.update_interval", "1"),
					resource.TestCheckResourceAttr("adguard_config.test", "safebrowsing.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "parental_control.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.services.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.services.1", "google"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.interval", "8"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.anonymize_client_ip", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.0", "example2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.1", "test2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.interval", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.0", "example4.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.1", "test3.net"),
					resource.TestCheckResourceAttr("adguard_config.test", "blocked_services.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "blocked_services.0", "pinterest"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.upstream_dns.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.upstream_dns.1", "https://1.0.0.1/dns-query"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.rate_limit", "30"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.blocking_mode", "custom_ip"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.blocking_ipv4", "1.2.3.4"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.blocking_ipv6", "fe80::"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.cache_ttl_min", "600"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.cache_ttl_max", "86400"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.cache_optimistic", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.local_ptr_upstreams.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.allowed_clients.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.allowed_clients.1", "allowed-client"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.interface", "eth1"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.gateway_ip", "192.168.250.1"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.range_start", "192.168.250.10"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.range_end", "192.168.250.100"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.lease_duration", "7200"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("adguard_config.test", "id"),
					resource.TestCheckResourceAttrSet("adguard_config.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "adguard_config.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in AdGuard Home,
				// therefore there is no value for it during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "adguard_config" "test" {
	filtering = {
		update_interval = 72
	}
	querylog = {
		ignored = ["test2.com", "example2.com", "abc2.com"]
	}
	stats = {
		ignored = ["test9.com", "example15.com", "abc5.com"]
	}
	dns = {
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
		disallowed_clients        = ["blocked-client", "172.16.0.0/16"]
	}
	dhcp = {
		interface = "eth1"
		ipv4_settings = {
			gateway_ip     = "192.168.250.1"
			subnet_mask    = "255.255.255.0"
			range_start    = "192.168.250.20"
			range_end      = "192.168.250.90"
			lease_duration = 14400
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "filtering.update_interval", "72"),
					resource.TestCheckResourceAttr("adguard_config.test", "safebrowsing.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "parental_control.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.enabled", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "safesearch.services.#", "6"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.interval", "2160"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.anonymize_client_ip", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.0", "abc2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.1", "example2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "querylog.ignored.2", "test2.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.interval", "24"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.#", "3"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.0", "abc5.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.1", "example15.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "stats.ignored.2", "test9.com"),
					resource.TestCheckResourceAttr("adguard_config.test", "blocked_services.#", "0"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.upstream_dns.#", "1"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.upstream_dns.0", "https://1.1.1.1/dns-query"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.rate_limit", "25"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.blocking_mode", "nxdomain"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.blocking_ipv4", ""),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.blocking_ipv6", ""),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.edns_cs_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.disable_ipv6", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.dnssec_enabled", "true"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.cache_size", "8000000"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.upstream_mode", "load_balance"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.use_private_ptr_resolvers", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.resolve_clients", "false"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.local_ptr_upstreams.#", "0"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.allowed_clients.#", "0"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.disallowed_clients.#", "2"),
					resource.TestCheckResourceAttr("adguard_config.test", "dns.disallowed_clients.1", "blocked-client"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.range_start", "192.168.250.20"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.range_end", "192.168.250.90"),
					resource.TestCheckResourceAttr("adguard_config.test", "dhcp.ipv4_settings.lease_duration", "14400"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
