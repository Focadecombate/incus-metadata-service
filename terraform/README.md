# Terraform — GCP test host for the metadata service

Provisions a GCP VM and takes it **zero-to-ready**: installs Incus (Zabbly),
builds and runs the metadata service as a systemd unit, wires
`169.254.169.254` to it, and publishes the `mds-ubuntu-2404` seed image with the
NoCloud drop-in baked in. Then you run the smoke test and
`scripts/run-experiments.sh` to capture the thesis Section 4 data.

## Prerequisites

- `gcloud` authenticated (`gcloud auth application-default login`) and the
  Compute Engine API enabled on the project (already done for
  `mythical-rope-502122-k2`).
- Terraform >= 1.5. If not installed:
  `sudo snap install terraform --classic` (or via HashiCorp's apt repo).

## Usage

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars   # edit if needed
terraform init
terraform apply            # review the plan, then confirm

# Watch first-boot provisioning (a few minutes: apt + Go build + seed image):
terraform output -raw watch_provisioning | bash

# When /var/lib/mds/PROVISION_DONE exists, SSH in and smoke-test:
terraform output -raw ssh_command | bash
#   incus launch mds-ubuntu-2404 t1
#   incus exec t1 -- cloud-init status --wait --long   # expect: status: done
#   incus exec t1 -- grep -i nocloud /var/log/cloud-init.log

# Capture the experiments:
cd /opt/mds && ./scripts/run-experiments.sh all
# results land in /opt/mds/results/<timestamp>/ — copy them back:
#   gcloud compute scp --recurse --tunnel-through-iap \
#     incus-mds-test:/opt/mds/results ./results --zone us-central1-a
```

## Cost & teardown

`e2-standard-4` is ~$0.13/hr; a full experiment round is a couple of hours.
**Destroy when done** so it stops billing:

```bash
terraform destroy
```

## Path A vs Path B

- **Path A (default):** `e2-standard-4`, containers only — covers all the
  paper's experiments (functional, latency, 5→50 scalability). No nested virt.
- **Path B:** set `machine_type = "n2-standard-4"` and
  `enable_nested_virtualization = true` to also boot Incus **VMs** (for the
  "containers only" threat-to-validity). E2/N2D/T2D can't do nested virt — the
  config has a precondition that enforces this.

## Notes

- SSH is locked to GCP's IAP range by default; connect with
  `--tunnel-through-iap` (the `ssh_command` output does this). Add your own
  `/32` to `ssh_source_ranges` for direct SSH.
- The service authenticates to Incus over the local HTTPS API (`127.0.0.1:8443`)
  with a generated client cert added to Incus trust — the startup script sets
  this up; nothing to do manually.
- `raw.dnsmasq` pushes a DHCP route so guests can actually reach the link-local
  `169.254.169.254` at boot; a nat REDIRECT forwards `:80` to the service port.
- Builds `repo_ref` (default `bugfix/cloud-init-spec-compliance`) — change it to
  `main` once the fixes are merged.
