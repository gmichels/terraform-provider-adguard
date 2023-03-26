# manage a DNS rewrite rule
resource "adguard_rewrite" "test" {
  domain = "example.com"
  answer = "4.3.2.1"
}
