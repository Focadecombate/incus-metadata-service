variable "project_id" {
  type    = string
  default = "mythical-rope-502122-k2"
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "zone" {
  type    = string
  default = "us-central1-a"
}

variable "machine_type" {
  description = "Per-node machine type. e2-standard-2 is plenty for a consensus/failover test (no big container load)."
  type        = string
  default     = "e2-standard-2"
}

variable "boot_disk_gb" {
  type    = number
  default = 30
}

variable "ssh_source_ranges" {
  description = "CIDRs allowed to SSH (default: GCP IAP range; use --tunnel-through-iap)."
  type        = list(string)
  default     = ["35.235.240.0/20"]
}

variable "repo_url" {
  type    = string
  default = "https://github.com/Focadecombate/incus-metadata-service.git"
}

variable "repo_ref" {
  type    = string
  default = "bugfix/cloud-init-spec-compliance"
}

variable "service_port" {
  type    = number
  default = 8080
}

variable "raft_port" {
  type    = number
  default = 7000
}

variable "seed_image_alias" {
  type    = string
  default = "mds-ubuntu-2404"
}

# Fixed per-node internal IPs so Raft peer wiring is deterministic and known
# before instance creation (avoids a dependency cycle between the instances).
variable "nodes" {
  description = "Raft nodes: id -> {ip, bootstrap}. Exactly one bootstrap=true."
  type = map(object({
    ip        = string
    bootstrap = bool
  }))
  default = {
    node1 = { ip = "10.20.0.11", bootstrap = true }
    node2 = { ip = "10.20.0.12", bootstrap = false }
    node3 = { ip = "10.20.0.13", bootstrap = false }
  }
}
