# manage the server configuration
# NOTE: there can only be 1 (one) `adguard_config` resource
# specifying multiple resources will result in errors
resource "adguard_config" "test" {
  filtering = {
    update_interval = 1
  }
  safesearch = {
    enabled  = true
    services = ["bing", "youtube", "google"]
  }
}

