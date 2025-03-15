terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.digitalocean_token
}

resource "digitalocean_droplet" "cheap_droplet" {
  name   = "minimal-droplet"
  region = "fra1"               # Frankfurt region
  size   = "s-1vcpu-512mb-10gb" # Cheapest plan (1 vCPU, 512mb RAM)
  image  = "ubuntu-24-10-x64"   # Ubuntu image

  backups = false

  ssh_keys = [var.ssh_key] # Add your SSH key for access

  tags = ["web", "dev"]
}

output "droplet_ip" {
  value = digitalocean_droplet.cheap_droplet.ipv4_address
}
