terraform {
  required_providers {
    adguard = {
      source  = "gmichels/adguard"
      version = "0.2.0"
    }
  }
}

# configuration for the provider
provider "adguard" {
  host     = "localhost:8080"
  username = "admin"
  password = "SecretP@ssw0rd"
  scheme   = "http" # defaults to https
  timeout  = 5      # in seconds, defaults to 10
}

# resource "adguard_list_filter" "test_blacklist" {
#   name = "Test Blacklist Filtering"
#   url  = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_4.txt"
# }

# resource "adguard_client" "test" {
#   name      = "Test Client Updated"
#   ids       = ["192.168.100.1", "test-clienet"]
#   # upstreams = ["1.2.3.4"]
# }

# resource "adguard_user_rules" "testing" {
#   rules = [
#     "! line 1 bang commen",
# 		"# line 2 respond with 127.0.0.1 for localhost.org (but not for its subdomains)",
# 		"127.0.0.1 localhost.org",
# 		"# line 4 unblock access to unblocked.org and all its subdomains",
# 		"@@||unblocked.org^",
# 		"# line 6 block access to blocked.org and all its subdomains",
# 		"||blocked.org^"
# 	]
# }

# resource "adguard_rewrite" "test" {
#   domain = "example.com"
#   answer = "4.3.2.4"
# }

# resource "adguard_rewrite" "test_new" {
#   domain = "example.org"
#   answer = "example.com"
# }


resource "adguard_dns_config" "test" {
  upstream_dns        = ["https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"]
  rate_limit          = 30
  cache_ttl_min       = 600
  cache_ttl_max       = 86400
  cache_optimistic    = true
  # blocking_mode       = "custom_ip"
  # blocking_ipv4       = "1.2.3.4"
  # blocking_ipv6       = "fe80::"
  local_ptr_upstreams = ["192.168.0.1", "192.168.0.2"]
}


# resource "adguard_dns_config" "test" {
#   upstream_dns              = ["https://1.1.1.1/dns-query"]
#   blocking_mode             = "nxdomain"
#   rate_limit                = 25
#   edns_cs_enabled           = true
#   disable_ipv6              = true
#   dnssec_enabled            = true
#   cache_size                = 8000000
#   upstream_mode             = "load_balance"
#   use_private_ptr_resolvers = false
#   resolve_clients           = false
# }
