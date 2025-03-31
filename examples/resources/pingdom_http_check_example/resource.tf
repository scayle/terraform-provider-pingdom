resource "pingdom_http_check" "this" {
  name      = "Pingdom Terraform Example"
  host      = "google.com"
  frequency = "1m"
  tags      = { name = "name" }
  regions   = ["EU"]
}
