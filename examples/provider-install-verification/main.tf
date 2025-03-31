terraform {
  required_providers {
    pingdom = {
      source = "hashicorp.com/local/pingdom"
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

resource "pingdom_http_check" "this" {
  name = "Terraform Provider Test"
  host = "google.com"
  frequency = "1m"
  contact_ids = [data.pingdom_contact.this.id]
  tags = { generated-by = "terraform" }
  regions = ["EU"]
}

data "pingdom_contact"  "this" {
  name = "Terraform Contact Test"
}
