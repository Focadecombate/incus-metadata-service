locals {
  instance_name = "incus-mds-test"
}

# Dedicated network so the setup is reproducible on a fresh project and doesn't
# depend on the auto-created default network existing.
resource "google_compute_network" "mds" {
  name                    = "incus-mds-net"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "mds" {
  name          = "incus-mds-subnet"
  ip_cidr_range = "10.10.0.0/24"
  region        = var.region
  network       = google_compute_network.mds.id
}

# SSH ingress only. The metadata service is internal to the VM (guests reach it
# over incusbr0 at 169.254.169.254) so no other ingress is required.
resource "google_compute_firewall" "ssh" {
  name          = "incus-mds-allow-ssh"
  network       = google_compute_network.mds.id
  direction     = "INGRESS"
  source_ranges = var.ssh_source_ranges

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

resource "google_compute_instance" "mds" {
  name         = local.instance_name
  machine_type = var.machine_type
  zone         = var.zone
  labels       = var.labels

  advanced_machine_features {
    enable_nested_virtualization = var.enable_nested_virtualization
  }

  dynamic "scheduling" {
    for_each = var.use_spot ? [1] : []
    content {
      provisioning_model          = "SPOT"
      preemptible                 = true
      automatic_restart           = false
      instance_termination_action = "STOP"
    }
  }

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2404-lts-amd64"
      size  = var.boot_disk_gb
      type  = "pd-balanced"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.mds.id
    # Ephemeral external IP for egress (apt, git clone) and direct SSH.
    access_config {}
  }

  metadata_startup_script = templatefile("${path.module}/startup.sh.tftpl", {
    repo_url         = var.repo_url
    repo_ref         = var.repo_ref
    service_port     = var.service_port
    seed_image_alias = var.seed_image_alias
  })

  lifecycle {
    precondition {
      condition     = !var.enable_nested_virtualization || can(regex("^(n1|n2|c2|m1|m2|m3)-", var.machine_type))
      error_message = "enable_nested_virtualization requires an N1/N2/C2/M-series machine_type (not e2/n2d/t2d)."
    }
  }
}
