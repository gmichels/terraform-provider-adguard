# manage user rules
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
