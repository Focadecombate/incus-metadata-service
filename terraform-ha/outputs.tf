output "nodes" {
  description = "Node name -> internal/external IPs."
  value = {
    for k, inst in google_compute_instance.node : k => {
      internal_ip = inst.network_interface[0].network_ip
      external_ip = inst.network_interface[0].access_config[0].nat_ip
      ssh         = "gcloud compute ssh incus-ha-${k} --zone ${var.zone} --tunnel-through-iap"
    }
  }
}

output "raft_status_check" {
  description = "Check which node is leader (run from any node or via SSH)."
  value       = "for n in ${join(" ", keys(var.nodes))}; do echo -n \"$n: \"; gcloud compute ssh incus-ha-$n --zone ${var.zone} --tunnel-through-iap -- curl -s http://127.0.0.1:${var.service_port}/raft/status; echo; done"
}

output "failover_test_hint" {
  description = "After provisioning, run the failover script."
  value       = "See terraform-ha/README.md — provision, confirm a leader, launch a container on the leader, then kill the leader and measure re-election."
}
