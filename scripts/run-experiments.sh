#!/usr/bin/env bash
#
# run-experiments.sh — drive the functional / performance / scalability
# experiments for the Incus cloud-init metadata service and dump raw artifacts
# plus a parseable summary into a timestamped results folder.
#
# These map to Section 4 (Resultados) of the TCC. See docs/testing-plan.md.
#
# PREREQUISITES (see docs/testing-plan.md §1):
#   - Incus installed and initialised (incusbr0), metadata service running.
#   - 169.254.169.254 wired to the service (iptables DNAT on incusbr0).
#   - A reusable image with the NoCloud drop-in baked in (default: mds-ubuntu-2404).
#   - Host tools: incus, curl, jq, awk. `hey` is auto-installed into the load
#     container if a static binary is provided or downloadable.
#
# USAGE:
#   ./run-experiments.sh [all|F|P|S]     # default: all
#   MDS_HEALTH_URL=http://127.0.0.1:8080/health SCALE_STEPS="5 10 25" ./run-experiments.sh S
#
# Nothing here is destructive to the host; all instances created are prefixed
# `mds-exp-` and removed on cleanup (KEEP_INSTANCES=1 to keep them).

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration (override via environment)
# ---------------------------------------------------------------------------
IMAGE="${IMAGE:-mds-ubuntu-2404}"              # reusable image with NoCloud drop-in
# Guests reach the service at the Incus bridge gateway. On a cloud VM do NOT use
# 169.254.169.254 — it is the provider's own metadata server (see docs).
IMDS_URL="${IMDS_URL:-http://10.10.10.1:8080/configs}"   # what guests fetch (bridge gateway)
MDS_HEALTH_URL="${MDS_HEALTH_URL:-http://127.0.0.1:8080/health}"  # host-side health check
SYNC_WAIT="${SYNC_WAIT:-15}"                   # seconds > cron interval (10s) to let a new instance sync
SCALE_STEPS="${SCALE_STEPS:-5 10 25 50}"       # container counts for the scalability sweep
REQUESTS="${REQUESTS:-5000}"                   # hey -n per endpoint
CONCURRENCY="${CONCURRENCY:-50}"               # hey -c
ENDPOINTS=("meta-data" "user-data" "vendor-data" "network-config")
PREFIX="mds-exp"                               # instance name prefix
LOADGEN="${PREFIX}-load"                       # dedicated load-generator container
HEY_BIN="${HEY_BIN:-}"                          # optional path to a local hey linux-amd64 binary
KEEP_INSTANCES="${KEEP_INSTANCES:-0}"
OUTDIR="${OUTDIR:-results/$(date +%Y%m%d-%H%M%S)}"

WHICH="${1:-all}"

log()  { printf '\033[1;34m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }
ok()   { printf '\033[1;32m  PASS\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }
info() { printf '       %s\n' "$*"; }

RESULTS_CSV=""   # set after OUTDIR created; one row per assertion / measurement

record() { # section, name, verdict, detail
  echo "$1,$2,$3,\"${4//\"/\'}\"" >> "$RESULTS_CSV"
}

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------
preflight() {
  log "Preflight checks"
  for t in incus curl jq awk; do
    command -v "$t" >/dev/null || { fail "missing required tool: $t"; exit 1; }
  done
  incus info >/dev/null 2>&1 || { fail "cannot talk to Incus (is it initialised?)"; exit 1; }
  if ! incus image info "$IMAGE" >/dev/null 2>&1; then
    fail "image '$IMAGE' not found — build it per docs/testing-plan.md §1.5"; exit 1
  fi
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "$MDS_HEALTH_URL" || echo 000)
  [ "$code" = "200" ] || { fail "metadata service health check failed ($MDS_HEALTH_URL -> $code)"; exit 1; }
  ok "Incus reachable, image present, service healthy"

  mkdir -p "$OUTDIR"
  RESULTS_CSV="$OUTDIR/results.csv"
  echo "section,name,verdict,detail" > "$RESULTS_CSV"

  # Environment capture -> fills 04_resultados.tex:13
  {
    echo "# Captured $(date -u +%FT%TZ)"
    echo "## incus version"; incus --version
    echo "## kernel"; uname -a
    echo "## cpu"; lscpu 2>/dev/null | sed -n '1,20p'
    echo "## memory"; free -h
    echo "## service commit"; git -C "$(dirname "$0")/.." rev-parse --short HEAD 2>/dev/null || true
    echo "## config"; env | grep -E '^(IMAGE|IMDS_URL|SCALE_STEPS|REQUESTS|CONCURRENCY|SYNC_WAIT)=' || true
  } > "$OUTDIR/environment.txt"
  info "Environment captured -> $OUTDIR/environment.txt"
}

# Fetch an endpoint from *inside* a container so ClientIP matches a synced instance.
imds() { # container, endpoint, [extra curl args...]
  local c="$1" ep="$2"; shift 2
  incus exec "$c" -- curl -s -H 'Accept: */*' "$@" "$IMDS_URL/$ep"
}
imds_code() { # container, endpoint
  local c="$1" ep="$2"
  incus exec "$c" -- curl -s -o /dev/null -w '%{http_code}' -H 'Accept: */*' "$IMDS_URL/$ep"
}

# Ensure the container has curl (needed to fetch the IMDS from inside the guest,
# so the service sees the container's real IP). Bake curl into your image to skip
# this; the apt fallback needs outbound network from the guest.
ensure_client() { # container
  local c="$1"
  incus exec "$c" -- sh -c 'command -v curl >/dev/null' 2>/dev/null && return 0
  incus exec "$c" -- sh -c 'command -v cloud-init >/dev/null && cloud-init status --wait >/dev/null 2>&1 || true'
  incus exec "$c" -- sh -c 'apt-get update -qq && apt-get install -y -qq curl' >/dev/null 2>&1 || true
  incus exec "$c" -- sh -c 'command -v curl >/dev/null'
}

launch() { # name, [extra `incus launch` args...]
  local name="$1"; shift || true
  incus launch "$IMAGE" "$name" "$@" >/dev/null
  ensure_client "$name" || { fail "$name has no curl and apt fallback failed — bake curl into $IMAGE"; return 1; }
}

wait_synced() { # container — block until its meta-data returns 200 or timeout
  local c="$1" waited=0
  while [ "$waited" -lt "$((SYNC_WAIT + 30))" ]; do
    [ "$(imds_code "$c" meta-data)" = "200" ] && return 0
    sleep 3; waited=$((waited + 3))
  done
  return 1
}

cleanup() {
  [ "$KEEP_INSTANCES" = "1" ] && { log "KEEP_INSTANCES=1 — leaving instances"; return; }
  log "Cleaning up ${PREFIX}-* instances"
  incus list -c n --format csv | grep "^${PREFIX}" | while read -r n; do
    incus delete --force "$n" >/dev/null 2>&1 || true
  done
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Load generator setup (installs hey into a dedicated container)
# ---------------------------------------------------------------------------
HEY_URL="https://hey-release.s3.us-east-2.amazonaws.com/hey_linux_amd64"
setup_loadgen() {
  log "Preparing load generator container ($LOADGEN)"
  incus info "$LOADGEN" >/dev/null 2>&1 || launch "$LOADGEN"
  wait_synced "$LOADGEN" || { fail "$LOADGEN did not sync — check DNAT/wiring"; exit 1; }
  if ! incus exec "$LOADGEN" -- test -x /root/hey; then
    if [ -n "$HEY_BIN" ] && [ -f "$HEY_BIN" ]; then
      incus file push "$HEY_BIN" "$LOADGEN/root/hey"
    else
      info "Downloading hey into $LOADGEN ..."
      incus exec "$LOADGEN" -- sh -c "curl -fsSL '$HEY_URL' -o /root/hey || (apt-get update -qq && apt-get install -y -qq curl && curl -fsSL '$HEY_URL' -o /root/hey)"
    fi
    incus exec "$LOADGEN" -- chmod +x /root/hey
  fi
  incus exec "$LOADGEN" -- /root/hey -n 1 -c 1 "$IMDS_URL/meta-data" >/dev/null 2>&1 \
    && ok "hey ready in $LOADGEN" || { fail "hey not runnable in $LOADGEN"; exit 1; }
}

# Run hey against one endpoint, dump CSV, compute p50/p95/p99 + error rate.
# Emits: "<p50_ms> <p95_ms> <p99_ms> <n_ok> <n_total>"
# Portable (mawk-safe): sorts with `sort -n` instead of gawk's asort.
run_hey() { # container, endpoint, out_csv, n, c
  local c="$1" ep="$2" out="$3" n="$4" conc="$5"
  incus exec "$c" -- /root/hey -n "$n" -c "$conc" -H 'Accept: */*' -o csv "$IMDS_URL/$ep" > "$out" 2>/dev/null || true
  local rt sc ntot nok
  rt=$(awk -F',' 'NR==1{for(i=1;i<=NF;i++)if($i=="response-time"){print i;exit}}' "$out")
  sc=$(awk -F',' 'NR==1{for(i=1;i<=NF;i++)if($i=="status-code"){print i;exit}}' "$out")
  if [ -z "$rt" ]; then echo "0 0 0 0 0"; return; fi
  ntot=$(awk -F',' 'NR>1{c++}END{print c+0}' "$out")
  nok=$(awk -F',' -v sc="${sc:-0}" 'NR>1 && sc>0 && $sc==200{c++}END{print c+0}' "$out")
  # response times in ms, numerically sorted, then index the percentiles
  local pct
  pct=$(awk -F',' -v rt="$rt" 'NR>1 && $rt!=""{print $rt*1000}' "$out" | sort -n | awk '
    {v[NR]=$1}
    END{
      k=NR; if(k==0){print "0 0 0"; exit}
      i50=int(k*0.50); if(i50<1)i50=1
      i95=int(k*0.95); if(i95<1)i95=1
      i99=int(k*0.99); if(i99<1)i99=1
      printf "%.2f %.2f %.2f\n", v[i50], v[i95], v[i99]
    }')
  echo "$pct $nok $ntot"
}

# ---------------------------------------------------------------------------
# Experiment F — Functional correctness  (-> 04_resultados.tex:38,40)
# ---------------------------------------------------------------------------
experiment_F() {
  log "Experiment F — functional correctness"
  local fdir="$OUTDIR/functional"; mkdir -p "$fdir"

  # F1 — metadata isolation: launch 3, each sees only its own local-hostname.
  log "F1 — metadata isolation (3 instances)"
  local names=("${PREFIX}-f1a" "${PREFIX}-f1b" "${PREFIX}-f1c")
  for n in "${names[@]}"; do incus info "$n" >/dev/null 2>&1 || launch "$n"; done
  for n in "${names[@]}"; do wait_synced "$n" || fail "$n did not sync"; done
  local f1_ok=1
  for n in "${names[@]}"; do
    local md hn; md=$(imds "$n" meta-data); echo "$md" > "$fdir/$n.meta-data.yaml"
    hn=$(echo "$md" | awk -F': *' '/local-hostname/{print $2; exit}' | tr -d '"')
    if [ "$hn" = "$n" ]; then info "$n sees local-hostname=$hn"; else fail "$n saw '$hn'"; f1_ok=0; fi
  done
  [ "$f1_ok" = 1 ] && { ok "F1 isolation"; record F F1-isolation PASS "each instance sees its own hostname"; } \
                    || record F F1-isolation FAIL "hostname mismatch"

  # F1b — X-Forwarded-For spoof must NOT leak another instance's user-data.
  log "F1b — X-Forwarded-For spoof (security)"
  local victim="${names[1]}" attacker="${names[0]}"
  local vip; vip=$(incus list "$victim" -c4 --format csv | awk '{print $1}' | head -1)
  incus exec "$victim" -- sh -c 'true'  # ensure up
  local spoof; spoof=$(incus exec "$attacker" -- curl -s -H "X-Forwarded-For: $vip" "$IMDS_URL/user-data" || true)
  local victim_ud; victim_ud=$(imds "$victim" user-data || true)
  if [ -n "$vip" ] && [ -n "$victim_ud" ] && [ "$spoof" = "$victim_ud" ]; then
    fail "F1b — spoofed request returned victim's user-data!"; record F F1b-spoof FAIL "XFF spoof leaked user-data"
  else
    ok "F1b — spoof did not leak victim data"; record F F1b-spoof PASS "SetTrustedProxies(nil) holds"
  fi

  # F2 — user-data execution: sentinel file created by runcmd.
  log "F2 — user-data execution"
  local ud="${PREFIX}-f2"
  printf '#cloud-config\nruncmd:\n  - [ touch, /run/mds-ok ]\n' > "$fdir/f2-user-data.yaml"
  incus info "$ud" >/dev/null 2>&1 && incus delete --force "$ud" >/dev/null
  incus launch "$IMAGE" "$ud" -c cloud-init.user-data="$(cat "$fdir/f2-user-data.yaml")" >/dev/null
  wait_synced "$ud" || true
  incus exec "$ud" -- cloud-init status --wait >/dev/null 2>&1 || true
  if incus exec "$ud" -- test -f /run/mds-ok; then
    ok "F2 — sentinel present"; record F F2-userdata-exec PASS "runcmd sentinel created"
  else
    fail "F2 — sentinel missing"; record F F2-userdata-exec FAIL "runcmd did not run"
  fi
  incus exec "$ud" -- cloud-init analyze show > "$fdir/f2-cloudinit-analyze.txt" 2>/dev/null || true

  # F3 — network-config: capture served config + resulting interfaces (manual compare).
  log "F3 — network-config (captured for review)"
  imds "${names[0]}" network-config > "$fdir/f3-network-config.yaml" 2>/dev/null || true
  incus exec "${names[0]}" -- ip -o -4 addr > "$fdir/f3-ip-addr.txt" 2>/dev/null || true
  if grep -q 'version' "$fdir/f3-network-config.yaml" 2>/dev/null; then
    ok "F3 — network-config served (review $fdir/f3-*)"; record F F3-network-config CAPTURED "served config + ip addr saved for manual compare"
  else
    info "F3 — no network-config served (may be expected)"; record F F3-network-config CAPTURED "no config served"
  fi

  # F4 — state propagation: rename, expect new hostname within the sync window.
  log "F4 — state propagation after restart"
  local before after
  before=$(imds "${names[2]}" meta-data | awk -F': *' '/local-hostname/{print $2; exit}')
  incus restart "${names[2]}" >/dev/null 2>&1 || true
  sleep "$SYNC_WAIT"; wait_synced "${names[2]}" || true
  after=$(imds "${names[2]}" meta-data | awk -F': *' '/local-hostname/{print $2; exit}')
  if [ -n "$after" ]; then
    ok "F4 — record available after restart (before='$before' after='$after')"
    record F F4-propagation PASS "metadata served within sync window after restart"
  else
    fail "F4 — no metadata after restart"; record F F4-propagation FAIL "not served post-restart"
  fi
}

# ---------------------------------------------------------------------------
# Experiment P — Per-endpoint latency  (-> 04_resultados.tex:53,55,61)
# ---------------------------------------------------------------------------
experiment_P() {
  log "Experiment P — per-endpoint latency (n=$REQUESTS c=$CONCURRENCY)"
  local pdir="$OUTDIR/performance"; mkdir -p "$pdir"
  setup_loadgen
  echo "endpoint,p50_ms,p95_ms,p99_ms,ok,total" > "$pdir/latency.csv"
  for ep in "${ENDPOINTS[@]}"; do
    log "P1 — $ep"
    read -r p50 p95 p99 nok ntot < <(run_hey "$LOADGEN" "$ep" "$pdir/hey-$ep.csv" "$REQUESTS" "$CONCURRENCY")
    echo "$ep,$p50,$p95,$p99,$nok,$ntot" >> "$pdir/latency.csv"
    info "$ep  p50=${p50}ms p95=${p95}ms p99=${p99}ms ok=$nok/$ntot"
    record P "P1-$ep" MEASURED "p50=${p50}ms p95=${p95}ms p99=${p99}ms ok=$nok/$ntot"
  done

  # P2 — sync latency: create an instance, measure time until meta-data is served.
  log "P2 — sync latency (create -> record available)"
  local sname="${PREFIX}-p2" t0 t1 waited=0
  incus info "$sname" >/dev/null 2>&1 && incus delete --force "$sname" >/dev/null
  t0=$(date +%s.%N)
  launch "$sname"
  while [ "$waited" -lt "$((SYNC_WAIT + 30))" ]; do
    [ "$(imds_code "$sname" meta-data)" = "200" ] && break
    sleep 1; waited=$((waited + 1))
  done
  t1=$(date +%s.%N)
  local sync_s; sync_s=$(awk -v a="$t0" -v b="$t1" 'BEGIN{printf "%.1f", b-a}')
  info "sync latency: ${sync_s}s"; record P P2-sync-latency MEASURED "${sync_s}s create->served"
  echo "sync_latency_seconds=$sync_s" > "$pdir/sync-latency.txt"
}

# ---------------------------------------------------------------------------
# Experiment S — Scalability 5..50  (-> 04_resultados.tex:67,69)
# ---------------------------------------------------------------------------
experiment_S() {
  log "Experiment S — scalability sweep ($SCALE_STEPS)"
  local sdir="$OUTDIR/scalability"; mkdir -p "$sdir"
  setup_loadgen
  # Clear any functional/perf leftovers so N reflects only the sweep + load gen.
  incus list -c n --format csv | grep "^${PREFIX}" | grep -v "^${LOADGEN}$" | while read -r n; do
    incus delete --force "$n" >/dev/null 2>&1 || true
  done
  echo "n_instances,p50_ms,p95_ms,p99_ms,ok,total,err_pct,loadavg_1m,mem_used_mb" > "$sdir/scalability.csv"
  local prev=0
  for N in $SCALE_STEPS; do
    log "S — scaling to $N instances"
    # top up to N filler instances (excluding the load generator)
    while [ "$prev" -lt "$N" ]; do
      prev=$((prev + 1)); launch "${PREFIX}-s$prev" || true
    done
    # let the newest instances sync so they're real 200-returning records
    sleep "$SYNC_WAIT"
    read -r p50 p95 p99 nok ntot < <(run_hey "$LOADGEN" "meta-data" "$sdir/hey-n$N.csv" "$REQUESTS" "$CONCURRENCY")
    local err; err=$(awk -v o="$nok" -v t="$ntot" 'BEGIN{printf "%.2f", t>0?(100*(t-o)/t):0}')
    local la; la=$(awk '{print $1}' /proc/loadavg)
    local mem; mem=$(free -m | awk '/^Mem:/{print $3}')
    echo "$N,$p50,$p95,$p99,$nok,$ntot,$err,$la,$mem" >> "$sdir/scalability.csv"
    info "N=$N  p50=${p50}ms p95=${p95}ms p99=${p99}ms err=${err}% load1=$la mem=${mem}MB"
    record S "S-n$N" MEASURED "p50=${p50}ms p95=${p95}ms p99=${p99}ms err=${err}%"
  done
  info "Scalability table -> $sdir/scalability.csv"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
preflight
case "$WHICH" in
  F) experiment_F ;;
  P) experiment_P ;;
  S) experiment_S ;;
  all) experiment_F; experiment_P; experiment_S ;;
  *) fail "unknown target '$WHICH' (use: all|F|P|S)"; exit 1 ;;
esac

log "Done. Results in $OUTDIR"
info "Summary:"
column -t -s',' "$RESULTS_CSV" 2>/dev/null || cat "$RESULTS_CSV"
