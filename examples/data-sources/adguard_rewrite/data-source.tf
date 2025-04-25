# get a DNS rewrite rule
data "adguard_rewrite" "test" {
  domain = "example.org"
  answer = "1.2.3.4"
}
