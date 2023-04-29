### this resource has been DEPRECATED and will be removed in a future release
### Use the `dns` block in the `adguard_config` resource instead
# manage the DNS configuration
# NOTE: there can only be 1 (one) `adguard_dns_config` resource
# specifying multiple resources will result in errors
resource "adguard_dns_config" "test" {
  upstream_dns        = ["https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"]
  rate_limit          = 30
  cache_ttl_min       = 600
  cache_ttl_max       = 86400
  cache_optimistic    = true
  local_ptr_upstreams = ["192.168.0.1", "192.168.0.2"]
}
