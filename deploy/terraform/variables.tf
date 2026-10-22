# ABOUTME: Inputs for the playground droplet. Every value has a default; override with -var or a gitignored .tfvars file.

variable "region" {
  description = "DigitalOcean region slug. The tutorial greeting reports this location as where the code ran."
  type        = string
  default     = "sfo3"
}

variable "size" {
  description = "Droplet size slug. 4 vCPU / 8 GB fits four concurrent runs plus image builds."
  type        = string
  default     = "s-4vcpu-8gb"
}

variable "droplet_name" {
  description = "Droplet name, also used for the firewall and the Ansible inventory host."
  type        = string
  default     = "temporal-playground"
}

variable "hostname" {
  description = "Record name inside the domain, so the site is <hostname>.<domain>."
  type        = string
  default     = "temporal"
}

variable "domain" {
  description = "A domain already managed in DigitalOcean DNS."
  type        = string
  default     = "egger.codes"
}

variable "ssh_key_names" {
  description = "Names of SSH keys already uploaded to DigitalOcean that may log in as root."
  type        = list(string)
  default     = ["labradillo-dev-box", "Temporal Macbook Pro"]
}

variable "git_ref" {
  description = "Branch, tag, or commit of the playground repository that Ansible deploys."
  type        = string
  default     = "main"
}

variable "tags" {
  description = "Tags applied to the droplet and firewall."
  type        = list(string)
  default     = ["temporal-playground"]
}
