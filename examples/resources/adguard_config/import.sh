# DNS config can be imported by specifying the ID as `1`
# NOTE: there can only be 1 (one) `adguard_dns_config` resource, hence the hardcoded ID
terraform import adguard_config.test "1"
