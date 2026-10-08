# Section 4 — Measured Results

Captured 2026-07-11 on a real GCP VM. Raw artifacts are in the timestamped
subdirectories here (one per run); the canonical runs are noted below.

## Environment (→ 04_resultados.tex:13)

| Item | Value |
|---|---|
| Host | GCP `e2-standard-4` (`us-central1-a`) |
| CPU | Intel Xeon @ 2.20 GHz, 4 vCPU (2 cores × 2 threads), KVM guest |
| Memory | 15 GiB |
| OS / kernel | Ubuntu 24.04 LTS, kernel 6.17.0-1020-gcp |
| Incus | 6.0.0 (Ubuntu universe) |
| Service commit | `bugfix/cloud-init-spec-compliance` |
| Load tool | `hey`, 5000 requests, concurrency 50 |
| Container image | `mds-ubuntu-2404` (Ubuntu 24.04 cloud + NoCloud drop-in) |
| Addressing | cloud-init NoCloud `seedfrom=http://10.10.10.1:8080/configs/` (Incus bridge gateway) |

## F — Functional correctness (→ 04_resultados.tex:38,40) — run `20260711-233116`

| Test | Result | Evidence |
|---|---|---|
| F1 metadata isolation (3 concurrent instances) | **PASS** | each instance's `meta-data` reports its own `local-hostname` |
| F1b X-Forwarded-For spoof | **PASS** | spoofed request did not return the victim's user-data (`SetTrustedProxies(nil)`) |
| F2 user-data execution | **PASS** | `runcmd` sentinel `/run/mds-ok` created; `cloud-init status` = done |
| F3 network-config delivery | CAPTURED | served as YAML; artifacts saved for manual compare |
| F4 state propagation after restart | **PASS** | metadata served within the 10 s sync window post-restart |

A real cloud-init boot selects `DataSourceNoCloudNet` and provisions from the
service end to end.

## P — Per-endpoint latency (→ 04_resultados.tex:53,55) — run `20260711-235036`

n = 5000, concurrency = 50.

| Endpoint | p50 (ms) | p95 (ms) | p99 (ms) | success |
|---|---|---|---|---|
| `/meta-data` | 13.2 | 53.7 | 83.2 | 5000/5000 |
| `/user-data` | 18.9 | 53.6 | 72.4 | 5000/5000 |
| `/network-config` | 11.4 | 47.2 | 72.4 | 5000/5000 |
| `/vendor-data` | 17.0 | 48.4 | 68.9 | 0/5000 (404 — no vendor data configured; correct) |

**Sync latency** (create → record served): **20.5 s** (one 10 s cron cycle plus
boot/lease time).

## S (superseded) — Scalability 5→200, July run (→ 04_resultados.tex:67,69)

> Superseded by the October re-run below (full raw CSV). Kept for reference.


Load against `/meta-data`, n = 5000, concurrency = 50, at each instance count.
First sweep 5–50 (`20260711-235653`); extended sweep 50–200 same host.

| Instances | p50 (ms) | p95 (ms) | p99 (ms) | errors | mem used (MB) |
|---|---|---|---|---|---|
| 5 | 16.2 | 60.3 | 92.4 | 0.00% | 1152 |
| 10 | 13.1 | 52.7 | 78.2 | 0.00% | 1325 |
| 25 | 12.4 | 51.8 | 77.2 | 0.00% | 1373 |
| 50 | 12.9 | 52.0 | 80.8 | 0.00% | 1733 |
| 100 | 14.4 | 58.6 | 92.1 | 0.00% | 1763 |
| 150 | 12.8 | 53.4 | 82.2 | 0.00% | 1638 |
| 200 | 12.8 | 54.2 | 87.2 | 0.00% | 1711 |

**Finding:** latency stays **flat from 5 to 200 concurrent instances**
(p50 ≈ 12–16 ms, p99 ≈ 77–98 ms) with **0% errors** throughout. This **refutes**
the paper's stated hypothesis of SQLite write-serialization degradation at this
scale: the serving path is read-dominated (`GET` metadata), the periodic sync
writes every 10 s are not a bottleneck, and WAL + a busy timeout absorb the
concurrent read/write mix. State this as a measured bound (no degradation through
200 instances / 4 vCPU) rather than claiming linear degradation.

> Note on methodology: SQLite was opened with `SetMaxOpenConns(1)` (WAL +
> `busy_timeout(5000)`). Report this in the methodology — it serializes access and
> is the relevant knob if a follow-up wants to probe the write-contention limit.

## HA — Raft failover (→ new HA subsection)

3-node cluster on GCP (`terraform-ha/`, 3 × e2-standard-2, deterministic internal
IPs, node1 bootstrap). Requires the Raft fixes (peer-id parsing; reconcile routed
through the log). Validated in-process first by `internal/consensus` cluster test.

| Property | Result |
|---|---|
| Cluster formation | **1 Leader (node1) + 2 Followers**, all agreeing on the leader address |
| Replication | an instance created on the leader appeared in **all 3 nodes'** databases via the Raft log |
| Leader-kill re-election | **2.40 s** (SIGKILL node1 → node2 elected), consistent with the ~1 s heartbeat + ~1 s election timeouts |
| Node rejoin | a killed node restarts and rejoins as a Follower, catching up from the log |

### Topology matters — two runs

**Run 1 (per-node Incus — wrong topology):** each node pointed at its own Incus. On
failover the new leader synced *its* (different) Incus and `reconcileDeletedInstances`
soft-deleted the failed leader's instances (it treats "absent from my Incus" as
"deleted"). This is a topology mistake in the test, not a consensus bug.

**Run 2 (shared Incus — Option A, correct):** all three nodes pointed at one shared
Incus (`node1`'s, exposed on the internal network with cross-node cert trust).

| Step | Result |
|---|---|
| Launch instance `ha2` on the shared Incus | replicated to **all 3 nodes** (live row on node1/2/3) |
| Kill the leader's *service* (Incus stays up) | node3 elected leader |
| New leader runs a sync+reconcile cycle | **`ha2` still present on both survivors — no prune** |

So under the correct HA topology (**N stateless replicas over one shared Incus**),
data stays available across a failover and the reconciliation is correct. State this
topology in the paper. The per-host/edge alternative would need source-scoped
reconciliation; the `source_node` column now records ownership as groundwork for it.
Consensus itself (formation, replication, 2.4 s re-election, rejoin) is sound.

## HA — Repeated leader-kill re-election (→ 04_resultados.tex, tab:ha) — run `20261007-172104-ha`

Cluster re-provisioned with `terraform-ha/` (3 × e2-standard-2, service commit
`098f792`, kernel 7.0.0-1011-gcp, per-node Incus). Ten repetitions via
`terraform-ha/failover-repeat.sh`: the current leader's process is SIGKILLed and
a survivor node runs `failover-poller.sh`, polling the leader's `/raft/status`
over the internal network every 50 ms; re-election time = first failed leader
poll → first survivor reporting `Leader`. Raw: `failover-runs.csv`,
`failover-runs.log`, `environment-node{1,2,3}.txt`.

| run | old → new leader | re-election (s) |
|---|---|---|
| 1 | node1 → node3 | 2.287 |
| 2 | node3 → node2 | 1.913 |
| 3 | node2 → node1 | 1.785 |
| 4 | node1 → node2 | 2.362 |
| 5 | node2 → node1 | 2.309 |
| 6 | node1 → node2 | 1.463 |
| 7 | node2 → node1 | 1.599 |
| 8 | node1 → node3 | 2.172 |
| 9 | node3 → node2 | 2.654 |
| 10 | node2 → node3 | 1.840 |

**n = 10, mean 2.04 s, median 2.04 s, min 1.46 s, max 2.65 s, sd 0.38 s.**
Consistent with the original single run (2.40 s) and with the 1 s heartbeat +
1 s randomized election timeout. Only re-election time was measured in these
runs; data availability after failover was verified in the original shared-Incus
run above.

## N — Does cloud-init consume /network-config? (→ 04_resultados.tex, F3) — run `20261007-202837`

Same host type as above (`e2-standard-4`, Ubuntu 24.04, Incus 6.0.0, service
commit `098f792`); guest: cloud-init 26.1-0ubuntu1~24.04.1, netplan.io 1.1.2.
Script: `scripts/validate-network-config.sh`. A variant of the seed image
re-enables cloud-init network management (a bootstrap DHCP netplan matched by
interface *name* is kept so the guest can reach the service on first boot).
Raw: `functional-netcfg/` (cloud-init status, /etc/netplan, networkd files,
networkctl, ip addr/route, resolvectl, cloud-init.log excerpts, served config).

| Test | Result | Evidence |
|---|---|---|
| N1 auto-generated network-config | **NOT CONSUMED** | cloud-init `status: done` via `DataSourceNoCloudNet`, but the service access log shows only `/meta-data`, `/user-data`, `/vendor-data` requests from the guest; `/network-config` was never requested |
| N2 admin network-config (`cloud-init.network-config` key, dhcp4 + search-domain marker) | applied, **but not via the service** | marker present in `/etc/netplan/50-cloud-init.yaml` and `resolvectl`; it came from the local seed `/var/lib/cloud/seed/nocloud-net/network-config` that Incus renders from the same key |

**Root cause:** `cloudinit/util.py::read_seeded()` (cloud-init 26.1) fetches
`meta-data`, `user-data` and `vendor-data` from the `seedfrom` URL and hardcodes
`network = None`; NoCloud reads `network-config` only from local seeds. So the
service's `/network-config` endpoint is served correctly but is not part of the
NoCloud-over-HTTP flow. Incus itself writes a local NoCloud seed
(`meta-data`, `user-data`, `vendor-data`, `network-config` v1 DHCP by default)
into every container.

**Side findings on the auto-generated content** (relevant if any consumer
applied it): `match: macaddress` renders to networkd `PermanentMACAddress=`,
which a container veth never matches (interface left "unmanaged", no DHCP:
observed with cloud-init's own fallback config); no gateway/nameservers; IPv6
address emitted without prefix length.

**Operational note:** after `incus restart`, `/network-config` (and the other
endpoints) return 404 for the instance until the next 10 s sync cycle rewrites
its IP; cloud-init's retries cover this window.

**Infra note:** the 200-container sweep filled the default 40 GB boot disk at
~95 containers (`dir` storage pool copies the full image per container);
`terraform/variables.tf` now defaults `boot_disk_gb` to 250.

## S — Scalability 5→200, re-run with full raw CSV (→ tab:escalabilidade, fig:escalabilidade) — run `20261007-211639`

Same procedure (`scripts/run-experiments.sh S`, `SCALE_STEPS="5 10 25 50 100 150 200"`,
n = 5000, c = 50 against `/meta-data`), host `e2-standard-4`, kernel
7.0.0-1011-gcp, Incus 6.0.0, service commit `098f792`, 250 GB boot disk.
Raw: `scalability/scalability.csv` + `hey-n*.csv` (per-request), `environment.txt`, `sweep.log`.

| Instances | p50 (ms) | p95 (ms) | p99 (ms) | errors | mem used (MB) |
|---|---|---|---|---|---|
| 5 | 19.3 | 77.7 | 117.4 | 0.00% | 1216 |
| 10 | 14.5 | 58.0 | 87.6 | 0.00% | 1404 |
| 25 | 12.7 | 52.5 | 78.5 | 0.00% | 2053 |
| 50 | 12.6 | 53.5 | 78.5 | 0.00% | 3147 |
| 100 | 14.3 | 60.0 | 94.9 | 0.00% | 4715 |
| 150 | 14.9 | 69.1 | 124.8 | 0.00% | 4909 |
| 200 | 13.2 | 55.6 | 85.4 | 0.00% | 5171 |

Same conclusion as July: no upward trend in latency with container count, 0%
errors through 200. p99 is noisier than in July (117 ms at n=5, 125 ms at
n=150, 85 ms at n=200). Host memory grows ~20 MB per container (container
userspace, not the service) and differs from the July column, which was
captured on kernel 6.17. The `sweep.log` contains 101 "has no curl" lines:
they come from `incus exec` racing a just-launched container (PID not yet
available); all 200 containers were created and synced (service log shows 200
"creating new instance" entries, no launch failures).
