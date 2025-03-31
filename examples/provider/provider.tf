variable "pingdom_api_token" {
  type     = string
  nullable = false
}

provider "pingdom" {
  api_token = var.pingdom_api_token
}
