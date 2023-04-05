# manage the DNS access list
# NOTE: there can only be 1 (one) `adguard_dns_access` resource
# specifying multiple resources will result in errors
resource "adguard_dns_access" "test" {
  allowed_clients = ["allowed-client", "192.168.200.200"]
}
