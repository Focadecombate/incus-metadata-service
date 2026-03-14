# Configuring Incus Instances to Use the Metadata Service

This guide explains how to configure Incus so that cloud-init inside containers
can discover and pull metadata from the service.

## Overview

Cloud-init expects a metadata service at the link-local address `169.254.169.254`,
following the convention used by AWS, GCP, and OpenStack. The setup involves two steps:

1. Making the metadata service reachable at `169.254.169.254` from inside containers
2. Configuring cloud-init to use the correct datasource

## Prerequisites

- Incus installed and running with a managed bridge network (default: `incusbr0`)
- The metadata service running on the host (default port: `8080`)
- Containers using a cloud-init-enabled image (e.g., `images:ubuntu/24.04`)

## Step 1: Route Traffic to the Metadata Service

Add the link-local address to the Incus bridge interface so containers can reach it:

```bash
sudo ip addr add 169.254.169.254/32 dev incusbr0
```

Since cloud-init sends requests to port 80 and the metadata service listens on port 8080,
add a NAT rule to redirect traffic:

```bash
sudo iptables -t nat -A PREROUTING -d 169.254.169.254 -p tcp --dport 80 -j REDIRECT --to-port 8080
```

### Making It Persistent

The commands above are lost on reboot. To persist them:

**systemd-networkd (recommended)**

Create `/etc/systemd/network/99-metadata.network`:

```ini
[Match]
Name=incusbr0

[Address]
Address=169.254.169.254/32
```

Then reload:

```bash
sudo systemctl restart systemd-networkd
```

**iptables persistence (Debian/Ubuntu)**

```bash
sudo apt install iptables-persistent
sudo netfilter-persistent save
```

**nftables alternative**

If your system uses nftables instead of iptables:

```bash
sudo nft add rule ip nat PREROUTING ip daddr 169.254.169.254 tcp dport 80 redirect to :8080
```

## Step 2: Configure cloud-init Datasource

The metadata service serves endpoints under the `/configs/` path prefix:

| Endpoint | Description |
|----------|-------------|
| `/configs/meta-data` | Instance metadata (hostname, IPs, SSH keys) |
| `/configs/user-data` | Cloud-init user-data scripts |
| `/configs/vendor-data` | Vendor-provided configuration |
| `/configs/network-config` | Netplan v2 network configuration |

Cloud-init's **NoCloud** datasource supports fetching configuration over HTTP
when given a seed URL. Set the seed URL in an Incus profile so all instances
inherit it automatically.

### Configure via Incus Profile

```bash
incus profile set default cloud-init.datasource "dsmode: net
seedfrom: http://169.254.169.254/configs/"
```

Or apply it to a specific profile:

```bash
incus profile create metadata-enabled
incus profile set metadata-enabled cloud-init.datasource "dsmode: net
seedfrom: http://169.254.169.254/configs/"
```

Then assign the profile to instances:

```bash
incus launch images:ubuntu/24.04 my-instance --profile default --profile metadata-enabled
```

### Verify from Inside a Container

Launch a test instance and confirm it can reach the metadata service:

```bash
incus launch images:ubuntu/24.04 test-instance
incus exec test-instance -- curl -s http://169.254.169.254/configs/meta-data
```

You should see a JSON response with the instance's metadata (hostname, IP, etc.).

To check that cloud-init used the metadata service during boot:

```bash
incus exec test-instance -- cloud-init status --long
incus exec test-instance -- cat /run/cloud-init/instance-data.json
```

## Troubleshooting

### Container cannot reach 169.254.169.254

Verify the address is on the bridge:

```bash
ip addr show incusbr0 | grep 169.254.169.254
```

Check that the NAT rule exists:

```bash
sudo iptables -t nat -L PREROUTING -n | grep 169.254
```

### cloud-init ignores the metadata service

Check cloud-init logs inside the instance:

```bash
incus exec <instance> -- cat /var/log/cloud-init.log | grep -i datasource
```

Ensure the datasource config is being passed. Verify the instance's cloud-init config:

```bash
incus config show <instance> | grep cloud-init
```

### Metadata returns empty or wrong data

The metadata service identifies instances by their source IP. Confirm the instance's
IP is synced in the service by checking the health endpoint and logs:

```bash
curl http://localhost:8080/health
```

Check service logs for sync errors. The sync cycle runs every 10 seconds by default,
so newly created instances may take up to one cycle to appear.
