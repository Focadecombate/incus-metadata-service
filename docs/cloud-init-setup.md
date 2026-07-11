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

cloud-init sends its requests to port 80, but the metadata service listens on port
8080. Add a NAT rule on the bridge to forward that traffic (the OpenStack/EC2 pattern,
robust for many instances):

```bash
# guest traffic to 169.254.169.254:80 -> host service on :8080
sudo iptables -t nat -A PREROUTING -i incusbr0 -d 169.254.169.254 -p tcp --dport 80 \
  -j DNAT --to-destination <incusbr0-host-ip>:8080
```

Replace `<incusbr0-host-ip>` with the bridge's own address (`ip addr show incusbr0`).
Verify the rule coexists with Incus's own nftables ruleset: `sudo nft list ruleset`.

For a single host you can instead use a `REDIRECT` rule, which sends the traffic to the
local machine on port 8080 without hardcoding the bridge IP:

```bash
sudo iptables -t nat -A PREROUTING -i incusbr0 -d 169.254.169.254 -p tcp --dport 80 \
  -j REDIRECT --to-port 8080
```

**Per-instance proxy fallback.** Instead of a bridge-wide NAT rule, you can attach an
Incus proxy device to each instance (simpler, but scales per-instance):

```bash
incus config device add <instance> mds proxy \
  listen=tcp:169.254.169.254:80 connect=tcp:127.0.0.1:8080 bind=host
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

## Step 2: Point cloud-init at the Metadata Service

The metadata service serves cloud-init's NoCloud endpoints under the `/configs/` prefix:

| Endpoint | Description |
|----------|-------------|
| `/configs/meta-data` | Instance metadata (instance-id, hostname, IPs, SSH keys) |
| `/configs/user-data` | Cloud-init user-data (cloud-config or shell script) |
| `/configs/vendor-data` | Vendor-provided configuration |
| `/configs/network-config` | Netplan v2 network configuration |

cloud-init fetches these with `Accept: */*`; the service replies with YAML.

> **Important.** There is no Incus config key that points cloud-init at an external HTTP
> datasource. `cloud-init.datasource` is **not** a valid Incus key — Incus accepts only
> `cloud-init.user-data`, `cloud-init.vendor-data`, and `cloud-init.network-config`.
> Worse, inside an Incus container cloud-init's `ds-identify` auto-selects the **LXD**
> datasource over `/dev/incus/sock` and never contacts an HTTP service. To use this
> metadata service you must override the datasource *inside the image* with a baked-in
> drop-in, then publish a reusable image. A stock image will silently ignore the service.

### Containers: bake an in-image datasource drop-in

Create a drop-in that forces the NoCloud datasource and seeds it from the service:

```yaml
# /etc/cloud/cloud.cfg.d/90-nocloud-net.cfg
datasource_list: [ NoCloud ]
datasource:
  NoCloud:
    seedfrom: http://169.254.169.254/configs/
```

Bake it into a reusable image and launch every instance from that image:

```bash
incus launch images:ubuntu/24.04 seed
incus exec seed -- sh -c 'cat > /etc/cloud/cloud.cfg.d/90-nocloud-net.cfg' < 90-nocloud-net.cfg
incus exec seed -- cloud-init clean --logs
incus publish seed --alias mds-ubuntu-2404
incus delete seed --force

incus launch mds-ubuntu-2404 my-instance
```

The trailing slash on `seedfrom` matters: cloud-init appends `meta-data`, `user-data`,
`vendor-data`, and `network-config` to it, producing the endpoint paths above.

### VMs: SMBIOS / kernel cmdline alternative

For Incus **VMs** you can skip the image rebuild and pass the seed via the NoCloud
kernel-cmdline / SMBIOS form instead:

```
ds=nocloud;s=http://169.254.169.254/configs/
```

### Verify from Inside an Instance

Launch an instance from the baked image and confirm it reaches the service:

```bash
incus launch mds-ubuntu-2404 test-instance
incus exec test-instance -- curl -s http://169.254.169.254/configs/meta-data
```

You should see the instance's metadata (instance-id, hostname, IP) returned as YAML.

Confirm cloud-init selected the NoCloud net datasource during boot:

```bash
incus exec test-instance -- cloud-init status --wait --long
incus exec test-instance -- cat /var/log/cloud-init.log | grep -i nocloud
```

`status: done` plus the NoCloud datasource in the log means the drop-in took. If you see
the LXD datasource instead, the image override didn't apply — re-check this step.

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

### cloud-init ignores the metadata service (falls back to the LXD datasource)

Check which datasource cloud-init selected:

```bash
incus exec <instance> -- cat /var/log/cloud-init.log | grep -i datasource
```

If it picked the LXD datasource (or NoCloud seeded from `/dev/incus/sock`) instead of
the net seed, the in-image drop-in from Step 2 isn't present in the image. Confirm it
made it in:

```bash
incus exec <instance> -- cat /etc/cloud/cloud.cfg.d/90-nocloud-net.cfg
```

Re-run `cloud-init clean --logs` before publishing the image if you edited the drop-in
after first boot.

### Metadata returns empty or wrong data

The metadata service identifies instances by their source IP. Confirm the instance's
IP is synced in the service by checking the health endpoint and logs:

```bash
curl http://localhost:8080/health
```

Check service logs for sync errors. The sync cycle runs every 10 seconds by default,
so newly created instances may take up to one cycle to appear.
