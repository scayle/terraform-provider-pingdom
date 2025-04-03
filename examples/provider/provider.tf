terraform {
  required_providers {
    pingdom = {
      source = "scayle/pingdom"
    }
  }
}

variable "pingdom_api_token" {
  type     = string
  nullable = false
}

provider "pingdom" {
  api_token = var.pingdom_api_token
}
