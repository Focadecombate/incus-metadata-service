variable "project_id" {
  description = "GCP project ID."
  type        = string
  default     = "mythical-rope-502122-k2"
}

variable "region" {
  description = "GCP region."
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone. Must have capacity for the chosen machine_type."
  type        = string
  default     = "us-central1-a"
}

variable "machine_type" {
  description = <<-EOT
    Machine type. Path A (containers only): e2-standard-4 is cheapest and needs no
    nested virt. Path B (also Incus VMs): use n2-standard-4 and set
    enable_nested_virtualization = true (E2/N2D/T2D do NOT support nested virt).
  EOT
  type        = string
  default     = "e2-standard-4"
}

variable "enable_nested_virtualization" {
  description = "Enable nested virt (needed only to run Incus VMs, not system containers). Requires an N1/N2/C2/M-series machine_type."
  type        = bool
  default     = false
}

variable "boot_disk_gb" {
  description = "Boot disk size in GB."
  type        = number
  default     = 40
}

variable "use_spot" {
  description = "Use a Spot (preemptible) instance to cut cost ~60-70%. Fine for short functional/latency runs; avoid for the long scalability sweep (a preemption corrupts a data point)."
  type        = bool
  default     = false
}

variable "ssh_source_ranges" {
  description = "CIDRs allowed to SSH. Default is GCP IAP's range; connect with `gcloud compute ssh --tunnel-through-iap`. Add your own IP/32 for direct SSH."
  type        = list(string)
  default     = ["35.235.240.0/20"]
}

variable "repo_url" {
  description = "Public git URL of this repo, cloned and built on the VM."
  type        = string
  default     = "https://github.com/Focadecombate/incus-metadata-service.git"
}

variable "repo_ref" {
  description = "Git branch/tag/commit to build (defaults to the fixes branch)."
  type        = string
  default     = "bugfix/cloud-init-spec-compliance"
}

variable "service_port" {
  description = "Port the metadata service listens on (guests reach it at 169.254.169.254:80 via redirect)."
  type        = number
  default     = 8080
}

variable "seed_image_alias" {
  description = "Alias for the published Incus image that has the NoCloud drop-in baked in."
  type        = string
  default     = "mds-ubuntu-2404"
}

variable "labels" {
  description = "Labels applied to the instance."
  type        = map(string)
  default     = { purpose = "incus-metadata-tcc", managed-by = "terraform" }
}
