# Current State — Incus Cloud-Init Metadata Service

_Last updated: 2026-07-11_

## What this is

A network metadata service for Incus instances: it serves cloud-init
`meta-data` / `user-data` / `vendor-data` / `network-config` over HTTP at the
link-local IMDS address (`169.254.169.254`), backed by SQLite, synced from the
Incus API on a 10 s cron, with OpenTelemetry/Prometheus metrics and an optional
HashiCorp Raft HA mode. It is the subject of the TCC in `paper/`.

## Build / test status (baseline)

- `go build ./...` — **passes**
- `go vet ./...` — **passes**
- `go test ./...` — **passes**, but only `internal/api/internal` (vendor-data admin
  CRUD) has tests. The entire public cloud-init surface, consensus, and dispatcher
  have **zero tests**.

## Readiness verdict

**Validated on GCP — a real cloud-init boot provisions from the service.** All
blocking cloud-init, security, and reliability issues below are fixed, the module
builds/vets/tests clean, and the end-to-end smoke test passes on a GCP VM (see "VM
validation" below). What remains is to run the F/P/S experiments
(`scripts/run-experiments.sh`) and fill in Section 4 of the paper.

Note for the performance experiments: SQLite is now opened with WAL + a 5 s busy
timeout and `SetMaxOpenConns(1)` (safest — serializes access, no `SQLITE_BUSY`). That
single connection serializes reads too, so the scalability curve may be shaped by
connection serialization as much as by SQLite itself. Decide before measuring whether
to keep `maxconns=1` (matches the paper's "write-serialization" hypothesis) or raise it
under WAL to allow concurrent readers — and report which you used.

## VM validation (GCP, 2026-07-11)

Provisioned via `terraform/` on a GCP `e2-standard-4` (Ubuntu 24.04) and driven a
real cloud-init boot end to end. **Smoke test passes:** a container boots, selects
`DataSourceNoCloudNet`, fetches from the service, and `cloud-init status` = `done`
(`meta-data` 200 with the container's own `instance-id`/`local-hostname`,
`user-data` 200). This validates the config fix, Incus HTTPS cert auth, datasource
selection, and addressing together.

Three issues were found and resolved during bring-up — the first is a code bug
(now in the PR); the other two are **findings worth citing in the thesis**:

1. **`LoadConfig()` failed on every startup** (code bug, fixed). The Raft `PEERS`
   env tag was malformed (`env:"PEERS,delimiter=,"`); go-envconfig splits the tag
   on commas, producing an "unknown option" error. Because it fails before the
   logger is initialised, the binary exited 1 with no output. The service could
   never have started as shipped. Fixed + regression test.
2. **`169.254.169.254` collides with the cloud provider's own metadata server**
   on GCP (and AWS/Azure). Binding/routing it on the host broke the VM's DNS.
   Resolution: point cloud-init's `seedfrom` at the **Incus bridge gateway**
   (`http://10.10.10.1:8080/configs/`) — always reachable, no DNAT, no conflict.
   *Implication for the paper: the IMDS `169.254.169.254` convention is not usable
   as-is on public-cloud hosts; a bridge-local address is required there.*
3. **NoCloud-only creates a network bootstrap deadlock.** Incus normally supplies
   network config via the LXD datasource (over the local socket, no network
   needed). Forcing `datasource_list: [NoCloud]` removes that, so the container
   can't get an IP to reach the seed. Resolution: bake an unconditional DHCP
   netplan into the image and disable cloud-init's network management, so the
   interface comes up independent of the datasource.
   *Implication for the paper: a network metadata service cannot be the sole
   datasource without an out-of-band mechanism to bring up the network first.*

All fixes are captured in `terraform/startup.sh.tftpl` (reproducible via
`terraform apply`), so the running VM is no longer a hand-patched one-off.

## Findings and status

Severity → what → status. Detail and file:line live in the fix commits.

### Blocking for cloud-init (fixed)

| # | Finding | Status |
|---|---|---|
| 1 | `network-config` returned 406 to every real client (gated on `Accept: application/yaml`; cloud-init sends `*/*`) | ✅ Accept gate removed; always serves YAML 200 (`network_config.go`) |
| 2 | 404 on `user-data` skipped the whole datasource | ✅ known instance w/ no user-data → 200 empty; only unknown IP → 404 (`userdata.go`) |
| 3 | `network-config` typed round-trip corrupted netplan v2 (`version: 0`, empty `match`) | ✅ generation cleaned (pointer `Match`, `omitempty`); serve is opaque JSON→YAML passthrough |
| 4 | Documented discovery (`cloud-init.datasource` key) didn't exist | ✅ docs rewritten to in-image `90-nocloud-net.cfg` drop-in + SMBIOS for VMs |

### Security (fixed)

| # | Finding | Status |
|---|---|---|
| 5 | Identity spoofable via `X-Forwarded-For` (`gin.Default()` trusts all proxies) | ✅ `SetTrustedProxies(nil)` on both engines (`main.go`) |
| 6 | `PUT/POST /internal/vendor/*` reachable by guests at `169.254.169.254` | ✅ moved to host-only listener `127.0.0.1:8081` (`INTERNAL_ADDR`); public engine no longer mounts it |
| 7 | Soft-deleted instances + IP reuse → decommissioned instance's secrets served | ✅ `reconcileDeletedInstances` soft-deletes DB instances absent from Incus each sync (leader-gated) |

### Reliability (fixed)

| # | Finding | Status |
|---|---|---|
| 8 | SQLite had no `_busy_timeout` / WAL → intermittent `SQLITE_BUSY` 500s | ✅ WAL + `busy_timeout(5000)` + `SetMaxOpenConns(1)` (`connect.go`) — see perf note above |
| 9 | `UpdateVendorData` dropped the `description` column on every update | ✅ Description now passed through (`internal/api/internal/vendordata.go`) |
| 10 | `handleInstanceCreated` was a no-op stub (boot race) | ✅ now fetches the instance and delegates to `handleInstanceSync` |

**Verification:** `go build ./...`, `go vet ./...`, `go test ./...` all pass. New tests
cover the public serving path (meta-data/user-data/vendor-data/network-config) including
the 406→200 and 404→200-empty fixes.

Deferred (not blocking): a **unique** DB index on `ip_address` was not added (finding
from review #9) — reconciliation mitigates the reuse case, but two live instances
sharing an IP still resolves arbitrarily. Add a partial unique index if this matters.

### Deferred — Raft HA (not blocking single-node results)

Raft is **disabled by default** (`RAFT_ENABLED=false`), so none of these block the
VM functional/performance/scalability experiments. They must be addressed before any
HA claim is *evaluated* (vs. merely "designed for") in the paper.

- Non-deterministic FSM: auto-increment IDs generated inside `Apply` diverge across
  nodes; IDs must be assigned by the leader before proposing.
- No snapshot/restore (`noopSnapshot`, `DiscardSnapshotStore`); log grows unbounded,
  lagging nodes can't recover.
- Fabricated peer IDs (`peer-0`) never match real `RAFT_NODE_ID`s → cluster won't form.
- Follower write path bypasses the leader (`handleInstanceSync` has no leader guard).

**Decision:** keep Raft off for the thesis experiments; scope the paper's HA claim to
"designed for HA" unless a follow-up implements the above and runs a failover test.

## Thesis (`paper/`) status

- **Section 4 (Resultados) is empty** — 14 `TODO` placeholders, no data. The
  experiments in `docs/testing-plan.md` are written to fill exactly these.
- Abstract/resumo claim the evaluation was already done — reconcile once data exists.
- Internal contradiction on `169.254.169.254` (implemented in §3.4 vs. future work in §5).
- Overclaim: §3.2 says full CRUD "per instance" (no DELETE; vendor-data keyed by name).
- README overclaims real-time Incus event-stream sync (code is cron-only) — **Agent D**.
- Cleanup: leftover files from another paper (`sections/test.tex`,
  `sections/04_resultados/` subtree, all `figures/*`), 4 uncited bib entries, a few
  grammar/agreement fixes (line-referenced in the review).

## Next steps

1. Land Agents A–D; `go build/vet/test` green; new tests for the serving path pass.
2. Follow `docs/testing-plan.md`: provision VM → smoke test (gate) → F/P/S experiments.
3. Fill Section 4 with the measured tables/graphs; reconcile the abstract.
4. (Optional) Implement the Raft fixes and add the HA failover experiment.
