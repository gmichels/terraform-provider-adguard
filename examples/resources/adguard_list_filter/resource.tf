# manage a blacklist filter
resource "adguard_list_filter" "test_blacklist" {
  name = "Test Blacklist Filter"
  url  = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_4.txt"
}

# manage a whitelist filter
resource "adguard_list_filter" "test_whitelist" {
  name      = "Test Whitelist Filter"
  url       = "https://adguardteam.github.io/HostlistsRegistry/assets/filter_6.txt"
  enabled   = false
  whitelist = true
}
