# ABOUTME: One DigitalOcean droplet for the Temporal tutorial playground, its firewall, and its DNS record.
# ABOUTME: Also writes the Ansible inventory so `just deploy` knows where to go. State is local and gitignored.

terraform {
  required_version = ">= 1.5"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

# The token comes from the DIGITALOCEAN_TOKEN environment variable.
provider "digitalocean" {}

data "digitalocean_ssh_key" "operator" {
  for_each = toset(var.ssh_key_names)
  name     = each.value
}

resource "digitalocean_droplet" "playground" {
  name       = var.droplet_name
  image      = "ubuntu-24-04-x64"
  region     = var.region
  size       = var.size
  ssh_keys   = [for key in data.digitalocean_ssh_key.operator : key.id]
  monitoring = true
  ipv6       = true
  tags       = var.tags
}

# Only SSH and the web ports are open; Codapi, Docker, and the gateway's own ports stay private.
resource "digitalocean_firewall" "playground" {
  name        = "${var.droplet_name}-web"
  droplet_ids = [digitalocean_droplet.playground.id]
  tags        = var.tags

  dynamic "inbound_rule" {
    for_each = ["22", "80", "443"]
    content {
      protocol         = "tcp"
      port_range       = inbound_rule.value
      source_addresses = ["0.0.0.0/0", "::/0"]
    }
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

resource "digitalocean_record" "playground" {
  domain = var.domain
  type   = "A"
  name   = var.hostname
  value  = digitalocean_droplet.playground.ipv4_address
  ttl    = 300
}

locals {
  fqdn = "${var.hostname}.${var.domain}"
}

resource "local_file" "ansible_inventory" {
  filename        = "${path.module}/../ansible/inventory/hosts.yml"
  file_permission = "0644"
  content = yamlencode({
    all = {
      hosts = {
        (var.droplet_name) = {
          ansible_host        = digitalocean_droplet.playground.ipv4_address
          ansible_user        = "root"
          playground_hostname = local.fqdn
          playground_git_ref  = var.git_ref
        }
      }
    }
  })
}
