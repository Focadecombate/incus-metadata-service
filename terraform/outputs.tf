output "instance_name" {
  description = "Name of the created VM."
  value       = google_compute_instance.mds.name
}

output "external_ip" {
  description = "Ephemeral external IP of the VM."
  value       = google_compute_instance.mds.network_interface[0].access_config[0].nat_ip
}

output "ssh_command" {
  description = "SSH via IAP (works with the default firewall)."
  value       = "gcloud compute ssh ${google_compute_instance.mds.name} --zone ${var.zone} --tunnel-through-iap"
}

output "watch_provisioning" {
  description = "Tail the first-boot provisioning log."
  value       = "gcloud compute ssh ${google_compute_instance.mds.name} --zone ${var.zone} --tunnel-through-iap -- tail -f /var/log/mds-provision.log"
}

output "smoke_test" {
  description = "Run after provisioning completes (marker: /var/lib/mds/PROVISION_DONE)."
  value       = "incus launch ${var.seed_image_alias} t1 && incus exec t1 -- cloud-init status --wait --long"
}

output "run_experiments" {
  description = "Capture Section 4 data once the smoke test passes."
  value       = "cd /opt/mds && ./scripts/run-experiments.sh all"
}
