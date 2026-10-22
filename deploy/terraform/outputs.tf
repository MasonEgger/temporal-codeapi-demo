# ABOUTME: What the operator needs after apply: where the droplet is and what the site is called.

output "ipv4" {
  description = "Public IPv4 address of the droplet."
  value       = digitalocean_droplet.playground.ipv4_address
}

output "fqdn" {
  description = "Hostname the site answers on once DNS propagates."
  value       = local.fqdn
}

output "url" {
  description = "Where to open the playground."
  value       = "https://${local.fqdn}/"
}
