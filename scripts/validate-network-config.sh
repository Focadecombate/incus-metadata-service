#!/usr/bin/env bash
#
# validate-network-config.sh: does cloud-init actually APPLY the network-config
# served by the metadata service? (Threat to validity in 04_resultados.tex:
# the seed image disables cloud-init network management, so the paper's
# experiments only verified that /network-config is served.)
#
# Builds a variant of the seed image with cloud-init network management
# re-enabled. The guest still needs an address to reach the service on first
# boot, so a bootstrap DHCP netplan definition is kept, keyed by interface NAME
# (match: {name: eth0}) under its own id: cloud-init's fallback and any
# macaddress-matched config render to networkd as PermanentMACAddress=, which a
# container veth never matches (interface stays "unmanaged", no DHCP). The id
# sorts after "eth0" so a name-matched served config, if any, takes precedence.
# Then boots:
#   N1: a container with the service's AUTO-GENERATED network-config;
#   N2: a container with an ADMIN-DEFINED network-config set via the Incus key
#       cloud-init.network-config (complete v2 config, dhcp4 + search domain
#       marker), to isolate "delivery path works" from "generated content".
# For each: capture cloud-init status, netplan files, ip addr/route, the served
# config and the relevant cloud-init.log lines, on first boot and after reboot.
#
# USAGE (on the Incus host, as root): OUTDIR=results/<ts> ./validate-network-config.sh
set -uo pipefail

IMAGE="${IMAGE:-mds-ubuntu-2404}"
NETIMG="${NETIMG:-mds-ubuntu-2404-netcfg}"
IMDS_URL="${IMDS_URL:-http://10.10.10.1:8080/configs}"
OUTDIR="${OUTDIR:-results/$(date +%Y%m%d-%H%M%S)}"
PREFIX="mds-exp"
D="$OUTDIR/functional-netcfg"; mkdir -p "$D"
RESULTS_CSV="$OUTDIR/results.csv"
[ -f "$RESULTS_CSV" ] || echo "section,name,verdict,detail" > "$RESULTS_CSV"
record() { echo "$1,$2,$3,\"${4//\"/\'}\"" >> "$RESULTS_CSV"; }
log() { printf '\033[1;34m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }

imds() { incus exec "$1" -- curl -s -H 'Accept: */*' "$IMDS_URL/$2"; }
imds_code() { incus exec "$1" -- curl -s -o /dev/null -w '%{http_code}' -H 'Accept: */*' "$IMDS_URL/$2"; }
wait_synced() { for _ in $(seq 1 20); do [ "$(imds_code "$1" meta-data)" = "200" ] && return 0; sleep 3; done; return 1; }

capture() { # container, tag
  local c="$1" tag="$2" p="$D/$1-$2"
  incus exec "$c" -- cloud-init status --long > "$p.cloud-init-status.txt" 2>&1
  incus exec "$c" -- sh -c 'for f in /etc/netplan/*.yaml; do echo "### $f"; cat "$f"; done' > "$p.netplan-files.txt" 2>&1
  incus exec "$c" -- netplan get > "$p.netplan-get.txt" 2>&1
  incus exec "$c" -- sh -c 'for f in /run/systemd/network/*; do echo "### $f"; cat "$f"; done' > "$p.networkd-files.txt" 2>&1
  incus exec "$c" -- networkctl status eth0 > "$p.networkctl.txt" 2>&1
  incus exec "$c" -- ip -o addr > "$p.ip-addr.txt" 2>&1
  incus exec "$c" -- ip route > "$p.ip-route.txt" 2>&1
  incus exec "$c" -- ip -6 route > "$p.ip6-route.txt" 2>&1
  incus exec "$c" -- resolvectl status > "$p.resolvectl.txt" 2>&1
  incus exec "$c" -- sh -c 'grep -i -E "network|netplan|renderer|apply" /var/log/cloud-init.log' > "$p.cloud-init-log-network.txt" 2>&1
  incus exec "$c" -- sh -c 'grep -i -E "warn|error|traceback" /var/log/cloud-init.log' > "$p.cloud-init-log-warnings.txt" 2>&1
  incus exec "$c" -- sh -c 'journalctl -u systemd-networkd -u netplan\* --no-pager 2>/dev/null | tail -40' > "$p.journal-network.txt" 2>&1
  incus exec "$c" -- sh -c 'curl -s -m 3 -o /dev/null -w "%{http_code}" http://10.10.10.1:8080/health' > "$p.reach-service.txt" 2>&1
  imds "$c" network-config > "$p.served-network-config.yaml" 2>&1
}

log "Building $NETIMG (cloud-init network management enabled) from $IMAGE"
incus image delete "$NETIMG" >/dev/null 2>&1 || true
{
  incus launch "$IMAGE" "${PREFIX}-netseed" >/dev/null
  sleep 5
  incus exec "${PREFIX}-netseed" -- rm -f /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg /etc/netplan/10-dhcp.yaml /etc/netplan/50-cloud-init.yaml
  incus exec "${PREFIX}-netseed" -- tee /etc/netplan/10-bootstrap-dhcp.yaml >/dev/null <<'NP'
network:
  version: 2
  ethernets:
    zz-bootstrap-dhcp:
      match: {name: eth0}
      dhcp4: true
NP
  incus exec "${PREFIX}-netseed" -- chmod 600 /etc/netplan/10-bootstrap-dhcp.yaml
  incus exec "${PREFIX}-netseed" -- sh -c 'ls /etc/cloud/cloud.cfg.d/' > "$D/image-cloud-cfg-d.txt"
  incus exec "${PREFIX}-netseed" -- cloud-init clean --logs
  incus stop "${PREFIX}-netseed"
  incus publish "${PREFIX}-netseed" --alias "$NETIMG" >/dev/null
  incus delete "${PREFIX}-netseed"
}

# ---------------------------------------------------------------------------
# N1: auto-generated network-config
# ---------------------------------------------------------------------------
N1="${PREFIX}-netcfg-auto"
log "N1: boot $N1 from $NETIMG (auto-generated network-config)"
incus delete --force "$N1" >/dev/null 2>&1 || true
incus launch "$NETIMG" "$N1" >/dev/null
wait_synced "$N1" || log "WARN: $N1 did not sync"
incus exec "$N1" -- cloud-init status --wait >/dev/null 2>&1 || true
capture "$N1" boot1
log "N1: reboot and capture again (network datasources apply config on the next boot)"
incus restart "$N1"; sleep 5
incus exec "$N1" -- cloud-init status --wait >/dev/null 2>&1 || true
capture "$N1" boot2
mac=$(incus config get "$N1" volatile.eth0.hwaddr)
ip4=$(incus list "$N1" -c4 --format csv | awk '{print $1}' | head -1)
echo "hwaddr=$mac ipv4=$ip4" > "$D/$N1.identity.txt"
served_addr=$(awk '/addresses:/{f=1;next} f&&/- [0-9]+\./{gsub(/[- ]/,""); print; exit}' "$D/$N1-boot2.served-network-config.yaml")
if [ -n "$served_addr" ] && grep -q -- "$served_addr" "$D/$N1-boot2.netplan-files.txt"; then
  if grep -q "Network File: /run/systemd/network/10-netplan-eth0.network" "$D/$N1-boot2.networkctl.txt"; then
    v=APPLIED; d="served static address $served_addr rendered to /etc/netplan and active on eth0"
  else
    v=RENDERED_NOT_ACTIVE; d="served static address $served_addr rendered to /etc/netplan/50-cloud-init.yaml but networkd did not bind it (see networkctl: match by macaddress -> PermanentMACAddress fails on veth)"
  fi
else
  v=NOT_APPLIED; d="served config (addr '$served_addr') not found in /etc/netplan after reboot"
fi
grep -qi "error\|traceback" "$D/$N1-boot2.cloud-init-log-warnings.txt" && d="$d; cloud-init.log has errors/warnings"
record N N1-netcfg-auto "$v" "$d"
log "N1 -> $v: $d"

# ---------------------------------------------------------------------------
# N2: admin-defined network-config via Incus key (complete v2, dhcp4 + marker)
# ---------------------------------------------------------------------------
N2="${PREFIX}-netcfg-admin"
log "N2: boot $N2 with cloud-init.network-config set on the instance"
cat > "$D/$N2.admin-network-config.yaml" <<'YAML'
version: 2
ethernets:
  eth0:
    dhcp4: true
    nameservers:
      search: [mds-netcfg-test.local]
YAML
incus delete --force "$N2" >/dev/null 2>&1 || true
incus launch "$NETIMG" "$N2" -c cloud-init.network-config="$(cat "$D/$N2.admin-network-config.yaml")" >/dev/null
wait_synced "$N2" || log "WARN: $N2 did not sync"
incus exec "$N2" -- cloud-init status --wait >/dev/null 2>&1 || true
capture "$N2" boot1
incus restart "$N2"; sleep 5
incus exec "$N2" -- cloud-init status --wait >/dev/null 2>&1 || true
capture "$N2" boot2
if grep -q "mds-netcfg-test.local" "$D/$N2-boot2.netplan-files.txt" && grep -q "mds-netcfg-test.local" "$D/$N2-boot2.resolvectl.txt"; then
  v=APPLIED; d="search-domain marker from served config rendered to /etc/netplan and active in resolvectl after reboot"
  grep -q "mds-netcfg-test.local" "$D/$N2-boot1.resolvectl.txt" && d="$d (already on first boot)"
elif grep -q "mds-netcfg-test.local" "$D/$N2-boot2.netplan-files.txt"; then
  v=RENDERED_NOT_ACTIVE; d="marker rendered to /etc/netplan but not active in resolvectl"
else
  v=NOT_APPLIED; d="marker not found in /etc/netplan after reboot"
fi
record N N2-netcfg-admin "$v" "$d"
log "N2 -> $v: $d"

incus delete --force "$N1" "$N2" >/dev/null 2>&1 || true
log "Done. Artifacts in $D"
