# Terraform — 3-node Raft HA cluster + failover test

Provisions three GCP VMs, each running the metadata service with Raft consensus
enabled, wired into one cluster (deterministic internal IPs so peer config is
known ahead of time). Then `failover-test.sh` produces the HA result for the
paper: cluster formation, replication, leader-kill re-election time, and read
availability on the survivors.

> Requires the Raft fixes on the `bugfix/cloud-init-spec-compliance` branch
> (peer-id parsing, reconcile-through-Raft). The `internal/consensus` integration
> test validates cluster formation/failover locally before you spend on VMs.

## Usage

```bash
cd terraform-ha
terraform init
terraform apply            # 3 × e2-standard-2

# Wait for all nodes to finish provisioning (watch any node):
gcloud compute ssh incus-ha-node1 --zone us-central1-a --tunnel-through-iap -- \
  'while [ ! -f /var/lib/mds/PROVISION_DONE ]; do sleep 5; done; echo done'

# Confirm the cluster formed (one Leader, two Follower):
terraform output -raw raft_status_check | bash

# Run the failover experiment:
ZONE=us-central1-a ./failover-test.sh          # uses your default gcloud SSH key
# or: ZONE=us-central1-a KEY=/path/to/key ./failover-test.sh

terraform destroy          # when done
```

## What the test measures (→ paper HA section)

1. **Formation** — exactly one Leader across the three `/raft/status` endpoints.
2. **Replication** — a container launched on the leader produces an `instances`
   row that appears in all three nodes' databases (via the Raft log).
3. **Failover** — stop the leader's service; poll a survivor's `/raft/status`
   locally and report the re-election time (Raft election timeout is ~1 s).
4. **Availability** — the surviving new leader still holds the replicated data.

## Notes

- `e2-standard-2` per node is enough — this test exercises consensus, not big
  container load. Bump `machine_type` if you also want scale on each node.
- Nodes use fixed internal IPs (`10.20.0.11-13`); `node1` is the bootstrap node.
- Raft port (7000) is open only within the subnet; SSH only via IAP.
- Cost: 3 × ~$0.07/hr. Destroy when finished.
