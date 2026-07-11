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

**Code fixes applied; ready for VM validation.** All blocking cloud-init, security,
and reliability issues below have been fixed and the module builds/vets/tests clean
(`go build/vet/test ./...` green; the public serving path now has its first real test
suite). The next step is to run the smoke test in `docs/testing-plan.md` on a VM — a
real cloud-init boot is the only thing that can confirm end-to-end compatibility, so
treat "ready" as "ready to validate", not "validated".

Note for the performance experiments: SQLite is now opened with WAL + a 5 s busy
timeout and `SetMaxOpenConns(1)` (safest — serializes access, no `SQLITE_BUSY`). That
single connection serializes reads too, so the scalability curve may be shaped by
connection serialization as much as by SQLite itself. Decide before measuring whether
to keep `maxconns=1` (matches the paper's "write-serialization" hypothesis) or raise it
under WAL to allow concurrent readers — and report which you used.

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
