# Testing Plan — Incus Cloud-Init Metadata Service

Validation protocol for the metadata service on a rented cloud VM. This plan doubles
as the experiment protocol for **Section 4 (Resultados)** of the TCC, whose tables and
graphs are currently `TODO` placeholders. Each experiment below maps to a specific
placeholder in `paper/sections/04_resultados.tex`.

> **Read this first.** Four blocking spec-compliance bugs were found in review (see
> §7). *As the code stands today, a real cloud-init client will not successfully
> consume this datasource.* Run §2 (the smoke test) before anything else — it will
> fail, and that failure is the point. Fix the bugs in §7, then proceed. Reporting
> "it works" without having driven a real cloud-init boot would be reporting a result
> that isn't real.

---

## 1. Test environment

### 1.1 Host choice

| Need | Recommendation |
|---|---|
| Containers only (functional + most perf tests) | **Any** cloud VM — no nested virtualization required. Hetzner CX32 (~€8/mo) or a GCP e2-standard-4 by the hour is plenty. |
| Incus **VMs** too (threats-to-validity item in the paper) | Nested-virt host: **Hetzner CCX23** (dedicated vCPU, ~€24/mo, VMX exposed) or **GCP n2-standard-4** with `--enable-nested-virtualization` (~$0.19/hr, good for short runs). |

The paper's methodology (§4.1) tests **5 / 10 / 25 / 50 system containers** — none of that
needs nested virt. Only add a nested-virt host if you want to defend the "system
containers only" threat-to-validity by also running a couple of Incus VMs.

**Recommended baseline: one Hetzner CCX23, Ubuntu 24.04 LTS.** Record its exact specs
(`lscpu`, `free -h`, `uname -r`, `incus --version`) — these fill the hardware/software
placeholder at `04_resultados.tex:13`.

### 1.2 Provision Incus (Ubuntu 24.04)

```bash
# Zabbly repo (NOT the Ubuntu archive package — it's older)
sudo mkdir -p /etc/apt/keyrings
sudo curl -fsSL https://pkgs.zabbly.com/key.asc -o /etc/apt/keyrings/zabbly.asc
# verify fingerprint: 4EFC 5906 96CB 15B8 7C73  A3AD 82CC 8797 C838 DCFD
sudo sh -c 'cat <<EOF > /etc/apt/sources.list.d/zabbly-incus-stable.sources
Enabled: yes
Types: deb
URIs: https://pkgs.zabbly.com/incus/stable
Suites: noble
Components: main
Signed-By: /etc/apt/keyrings/zabbly.asc
EOF'
sudo apt-get update && sudo apt-get install -y incus incus-tools
sudo incus admin init --minimal      # creates incusbr0
incus --version                      # record for the paper (stable ~7.2, 2026)
```

Docker-on-same-host gotcha: Docker sets the global FORWARD policy to DROP and breaks
Incus bridge traffic. If Docker is present, set `"ip-forward-no-drop": true` in
`/etc/docker/daemon.json`, or don't install Docker on this host.

### 1.3 Deploy the metadata service

```bash
cd metadata-service && go build -o metadata-service ./cmd/server
# env: see internal/config/main.go for the full list
DB_PATH=./metadata.db INCUS_SOCKET=/var/lib/incus/unix.socket ./metadata-service
```

### 1.4 Wire 169.254.169.254 → service

The service listens on a normal port; cloud-init expects the link-local IMDS address.
Use **iptables DNAT on the bridge** (robust for many instances — the OpenStack/EC2
pattern):

```bash
# guest traffic to 169.254.169.254:80 → host service on :8080
sudo iptables -t nat -A PREROUTING -i incusbr0 -d 169.254.169.254 -p tcp --dport 80 \
  -j DNAT --to-destination <incusbr0-host-ip>:8080
```

Verify it coexists with Incus's own nftables: `sudo nft list ruleset`. (Per-instance
Incus **proxy device** is a simpler fallback but scales per-instance:
`incus config device add <inst> mds proxy listen=tcp:169.254.169.254:80 connect=tcp:127.0.0.1:8080 bind=host`.)

### 1.5 Make guests actually select this datasource (critical — see §7 bug #3)

Inside Incus containers, `ds-identify` auto-selects the **LXD datasource** via
`/dev/incus/sock` and never contacts an HTTP service. You must override the image.
Bake an in-image drop-in:

```yaml
# /etc/cloud/cloud.cfg.d/90-nocloud-net.cfg
datasource_list: [ NoCloud ]
datasource:
  NoCloud:
    seedfrom: http://169.254.169.254/configs/
```

Build a reusable image with this baked in and use it for every test instance:
```bash
incus launch images:ubuntu/24.04 seed --ephemeral=false
incus exec seed -- sh -c 'cat > /etc/cloud/cloud.cfg.d/90-nocloud-net.cfg' < 90-nocloud-net.cfg
incus exec seed -- cloud-init clean --logs
incus publish seed --alias mds-ubuntu-2404
incus delete seed --force
```

---

## 2. Smoke test — run this FIRST (blocking gate)

Before any measurement, confirm one instance boots and consumes cloud-init end to end.

```bash
# raw endpoint check from the host (simulates what cloud-init fetches)
for p in meta-data user-data vendor-data network-config; do
  echo "== $p =="; curl -s -o /dev/null -w "%{http_code} %{content_type}\n" \
    -H 'Accept: */*' http://169.254.169.254/configs/$p   # Accept: */* is what cloud-init sends
done
```

**Expected per spec:** `meta-data` 200 YAML, `user-data` 200 (empty OK),
`vendor-data` 200-or-404, `network-config` 200 YAML.
**What you'll actually see today:** `network-config` → **406**, `user-data` → **404**
if no row. Both are §7 blockers. Fix them, then this gate passes.

Then a real boot:
```bash
incus launch mds-ubuntu-2404 t1
incus exec t1 -- cloud-init status --wait --long
incus exec t1 -- cat /var/log/cloud-init.log | grep -i nocloud
```
Success = `status: done` and the NoCloud net datasource in the log. If cloud-init fell
back to the LXD datasource, §1.5 didn't take.

---

## 3. Experiment F — Functional correctness (→ `04_resultados.tex:38,40`)

Four properties, from the paper's §4.2. Fix the design flaw the review flagged: inject
user-data via the **Incus config key** (`cloud-init.user-data`), *not* by writing the DB
directly — the point is to validate the sync path, and a direct DB write bypasses it.

| ID | Property | Method | Pass criteria |
|---|---|---|---|
| F1 | Metadata isolation | Launch 10 instances concurrently; each fetches `/meta-data`; compare returned `instance-id`/`local-hostname` to the requester | Every instance sees **only its own** id. (This will catch the X-Forwarded-For spoof — add F1b below.) |
| F1b | Identity spoof (security) | From instance A: `curl -H 'X-Forwarded-For: <B-ip>' .../configs/user-data` | Should **not** return B's data. Today it does — §7 bug #4. |
| F2 | user-data execution | Set `cloud-init.user-data` = a `#cloud-config` that `runcmd`s `touch /run/mds-ok`; boot; check file | File exists; `cloud-init status` = done |
| F3 | network-config applied | Provide a v2 config setting a static addr; boot; `ip addr` | Interface matches config (see §7 bug #5 — struct round-trip corrupts v2) |
| F4 | State propagation | Rename/restart an instance; poll `/meta-data` | New state visible within the 10 s sync window |

Record a pass/fail table — that's the §4.2 results table.

---

## 4. Experiment P — Performance / latency (→ `04_resultados.tex:53,55,61`)

Load tool: `hey` or `wrk` (record which — fills `04_resultados.tex:21`). N rounds
(record N).

```bash
hey -n 5000 -c 50 -H 'Accept: */*' http://169.254.169.254/configs/meta-data
```

- **P1 — per-endpoint latency:** run against `/meta-data`, `/user-data`,
  `/network-config`, **and `/vendor-data`** (the paper's §4.3.1 omits vendor-data — add
  it or justify). Report **p50/p95/p99**. This is the §4.3 latency table.
- **P2 — sync latency:** timestamp an Incus create/modify, poll until the record is
  served, record min/avg/max. Fills `04_resultados.tex:61`.

Run P1 while the DB is under concurrent cron writes (every 10 s) — SQLite is opened
with no busy-timeout/WAL (§7 bug #6), so watch for intermittent 500s and report them;
that *is* a finding, not noise.

---

## 5. Experiment S — Scalability 5→50 (→ `04_resultados.tex:67,69`)

For N ∈ {5, 10, 25, 50} containers: launch N, drive steady load, record aggregate
latency + error rate + host CPU/RAM. Plot latency vs N. The paper's stated hypothesis
is SQLite write-serialization degradation — confirm or refute with the curve. This is
the §4.3.2 scalability graph.

---

## 6. Experiment C — Comparison matrix (→ `04_resultados.tex:78,80`)

No cloud VM needed — this is a documentation/feature matrix. Build the table:
metadata fields and supported cloud-init modules (`set_hostname`, `users-groups`,
`write-files`, `runcmd`, network-config) served by this service vs. AWS IMDS / GCP /
OpenStack. Fills the §4.4 comparison table.

## 6b. Experiment HA — Raft failover (MISSING from the paper — add it)

RAFT is a headline feature (§3.6, comparison table "Alta Disp." ✓) but the paper plans
**no HA experiment**. Either add one or qualify the HA claim. Minimum test: 3-node
cluster, kill the leader, measure re-election time and read availability.

> **Warning:** review found the Raft FSM is **non-deterministic** (auto-increment IDs
> generated inside `Apply`) with **no snapshotting** and fabricated peer IDs (§7 bug in
> code review). A multi-node cluster will likely diverge. Get this working before
> promising HA results, or scope the paper's HA claim down to "designed for" rather
> than "evaluated."

---

## 7. Blocking bugs to fix before/around testing

These came out of the code review + spec audit. The **spec bugs** must be fixed or the
functional tests can't pass; the **security bugs** should be fixed or at least measured
and reported as findings.

**Spec-compliance (cloud-init will fail without these):**
1. **`network-config` returns 406 to every cloud-init client.** It gates on
   `Accept: application/yaml`; cloud-init sends `Accept: */*`. Drop the gate, always
   emit YAML. `internal/api/configs/network_config.go:36-42`.
2. **`user-data` 404 kills the whole datasource.** cloud-init's `user-data` fetch isn't
   exception-wrapped — a 404 skips NoCloud entirely (losing meta-data too). Return
   **200 empty** for a known instance with no user-data. `internal/api/configs/userdata.go:19-22`.
3. **Documented discovery doesn't work.** `cloud-init.datasource` is not a valid Incus
   key and containers auto-select the LXD datasource anyway. Use the in-image drop-in
   from §1.5. Fix `docs/cloud-init-setup.md`.
4. **Identity is spoofable + write API is guest-reachable.** `gin.Default()` trusts
   `X-Forwarded-For`; add `router.SetTrustedProxies(nil)`. And `PUT/POST /internal/vendor/*`
   is on the same listener guests can reach at 169.254.169.254 — bind it to a host-only
   listener or require auth. `cmd/server/main.go:55`, `internal/api/internal/routes.go:17-19`.
5. **network-config struct round-trip corrupts netplan v2.** Unmarshalling into
   `NetworkConfig{Version, Ethernets}` drops `bonds`/`bridges`/`vlans`/`mtu`/`set-name`
   and emits junk keys (`version: 0`, empty `match`, `wakeonlan: false`). Serve the
   stored config as opaque YAML, or add `omitempty` everywhere and drop
   `Ethernet.Version`. `pkg/types/network_config.go:69-84`.

**Reliability (will surface during load/scale tests):**
6. SQLite opened with no `_busy_timeout` / WAL — concurrent cron writes + HTTP reads →
   intermittent `SQLITE_BUSY` 500s. `internal/storage/db/connect.go:17`.
7. Instances only soft-deleted; a reused IP can be served a decommissioned instance's
   secrets. No Incus-deletion handler. (Also `idx_instances_ip_address` is non-unique →
   ambiguous identity.)

---

## 8. Deliverables checklist (maps to the thesis)

- [ ] Host + software spec block → `04_resultados.tex:13`
- [ ] Load tool + round count → `04_resultados.tex:21`
- [ ] F1–F4 pass/fail table → `04_resultados.tex:38,40`
- [ ] Per-endpoint p50/p95/p99 table → `04_resultados.tex:53,55`
- [ ] Sync-latency numbers → `04_resultados.tex:61`
- [ ] Scalability graph + analysis → `04_resultados.tex:67,69`
- [ ] Comparison matrix + discussion → `04_resultados.tex:78,80`
- [ ] Discussion + closing synthesis → `04_resultados.tex:6,85,99`
- [ ] (Add) HA failover result, or qualify the HA claim
- [ ] Reconcile abstract/resumo (`main.tex:72,75`) — soften "is evaluated" until data exists, or leave as-is once it does
