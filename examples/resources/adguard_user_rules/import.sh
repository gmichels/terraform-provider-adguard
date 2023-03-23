# User rules can be imported by specifying the ID as `1`
# NOTE: there can only be ONE `adguard_user_rules` resource, hence the hardcoded ID
terraform import adguard_user_rules.test "1"
