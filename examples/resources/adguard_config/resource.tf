# manage the server configuration
# NOTE: there can only be 1 (one) `adguard_config` resource
# specifying multiple resources will result in errors
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
}
