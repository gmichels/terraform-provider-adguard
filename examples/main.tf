terraform {
  required_providers {
    adguard = {
      version = "0.1.0"
      source  = "gmichels/adguard"
    }
  }
}

provider "adguard" {
  host     = "localhost:8080"
  username = "admin"
  password = "SecretP@ssw0rd"
  scheme   = "http" # defaults to https
  timeout  = 5      # in seconds, defaults to 10
}

data "adguard_client" "test" {
  name = "Test Client Data Source"
}

resource "adguard_client" "test" {
  name = "Test Client"
  ids  = ["192.168.100.15", "test-client"]
}
