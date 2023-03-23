# manage user rules
# NOTE: there can only be 1 (one) `adguard_user_rules` resource
# specifying multiple resources will result in errors
resource "adguard_user_rules" "test" {
  rules = [
    "! line 1 bang comment",
    "# line 2 respond with 127.0.0.1 for localhost.org (but not for its subdomains)",
    "127.0.0.1 localhost.org",
    "# line 4 unblock access to unblocked.org and all its subdomains",
    "@@||unblocked.org^",
    "# line 6 block access to blocked.org and all its subdomains",
    "||blocked.org^"
  ]
}
