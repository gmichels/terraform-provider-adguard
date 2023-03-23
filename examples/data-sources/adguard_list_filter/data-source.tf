# get a blacklist filter
data "adguard_list_filter" "test_blacklist" {
  name = "AdGuard DNS filter"
}

# get a whitelist filter
data "adguard_list_filter" "test_whitelist" {
  name      = "Test Allow List"
  whitelist = true
}
