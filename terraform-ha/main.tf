resource "google_compute_network" "ha" {
  name                    = "incus-ha-net"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "ha" {
  name          = "incus-ha-subnet"
  ip_cidr_range = "10.20.0.0/24"
  region        = var.region
  network       = google_compute_network.ha.id
}

resource "google_compute_firewall" "ssh" {
  name          = "incus-ha-allow-ssh"
  network       = google_compute_network.ha.id
  direction     = "INGRESS"
  source_ranges = var.ssh_source_ranges
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

# Raft inter-node traffic within the subnet.
resource "google_compute_firewall" "raft" {
  name          = "incus-ha-allow-raft"
  network       = google_compute_network.ha.id
  direction     = "INGRESS"
  source_ranges = ["10.20.0.0/24"]
  allow {
    protocol = "tcp"
    ports    = [tostring(var.raft_port)]
  }
}

resource "google_compute_instance" "node" {
  for_each     = var.nodes
  name         = "incus-ha-${each.key}"
  machine_type = var.machine_type
  zone         = var.zone
  labels       = { purpose = "incus-metadata-ha-tcc", managed-by = "terraform" }

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2404-lts-amd64"
      size  = var.boot_disk_gb
      type  = "pd-balanced"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.ha.id
    network_ip = each.value.ip # deterministic internal IP for Raft peer wiring
    access_config {}           # ephemeral external IP for egress + SSH
  }

  metadata_startup_script = templatefile("${path.module}/startup-ha.sh.tftpl", {
    repo_url         = var.repo_url
    repo_ref         = var.repo_ref
    service_port     = var.service_port
    seed_image_alias = var.seed_image_alias
    node_id          = each.key
    node_ip          = each.value.ip
    raft_port        = var.raft_port
    bootstrap        = each.value.bootstrap
    # Every OTHER node as an "id=ip:port" peer entry.
    peers = join(",", [for k, v in var.nodes : "${k}=${v.ip}:${var.raft_port}" if k != each.key])
  })
}
