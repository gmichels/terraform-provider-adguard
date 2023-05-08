terraform {
  required_providers {
    adguard = {
      source  = "gmichels/adguard"
      version = "0.2.0"
    }
  }
}

# configuration for the provider
provider "adguard" {
  host     = "localhost:8080"
  username = "admin"
  password = "SecretP@ssw0rd"
  scheme   = "http" # defaults to https
  timeout  = 5      # in seconds, defaults to 10
}
