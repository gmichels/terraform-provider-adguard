# manage the server configuration
# NOTE: there can only be 1 (one) `adguard_config` resource
# specifying multiple resources will result in errors
resource "adguard_config" "test" {
  filtering = {
    update_interval = 1
  }

  safebrowsing = true

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

    static_leases = [
      {
        mac      = "00:11:22:33:44:55"
        ip       = "192.168.250.20"
        hostname = "test-lease-1"
      },
      {
        mac      = "aa:bb:cc:dd:ee:ff"
        ip       = "192.168.250.30"
        hostname = "test-lease-2"
      }
    ]
  }
}
