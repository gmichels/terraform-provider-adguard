# get a DNS rewrite rule
data "adguard_rewrite" "test" {
  domain = "example.org"
}
