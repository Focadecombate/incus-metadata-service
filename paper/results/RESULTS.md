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

## S — Scalability 5→200 (→ 04_resultados.tex:67,69)

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
