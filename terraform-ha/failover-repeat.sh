#!/usr/bin/env bash
#
# failover-repeat.sh: repeats the leader-kill re-election measurement RUNS times.
# Run AFTER `terraform apply` and after all nodes report PROVISION_DONE.
#
# Each run: find the leader, start failover-poller.sh on a survivor (it polls the
# leader over the internal network), SIGKILL the leader's service, collect the
# poller's result, wait for the killed node to restart (systemd Restart=on-failure)
# and rejoin, cool down, repeat. Results go to $OUTDIR/failover-runs.csv.
#
# USAGE: ZONE=us-central1-a RUNS=10 OUTDIR=../paper/results/<ts>-ha ./failover-repeat.sh
set -uo pipefail

ZONE="${ZONE:-us-central1-a}"
PORT="${PORT:-8080}"
RUNS="${RUNS:-10}"
COOLDOWN_S="${COOLDOWN_S:-20}"
OUTDIR="${OUTDIR:-./failover-results}"
KEY="${KEY:-}"
SSHK=(); [ -n "$KEY" ] && SSHK=(--ssh-key-file "$KEY")
declare -A IP=([node1]=10.20.0.11 [node2]=10.20.0.12 [node3]=10.20.0.13)
NODES=(node1 node2 node3)
HERE="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$OUTDIR"
CSV="$OUTDIR/failover-runs.csv"
LOG="$OUTDIR/failover-runs.log"
[ -f "$CSV" ] || echo "run,old_leader,new_leader,reelection_s,poll_ms,status" > "$CSV"
log() { echo "[$(date -u +%H:%M:%S)] $*" | tee -a "$LOG"; }

ssh_node() { # node, cmd
  local n="$1"; shift
  gcloud compute ssh "incus-ha-$n" --zone "$ZONE" --tunnel-through-iap "${SSHK[@]}" --quiet \
    --command="$*" 2>/dev/null
}
state_of() { ssh_node "$1" "curl -s -m 3 http://127.0.0.1:$PORT/raft/status" | grep -o '"state":"[A-Za-z]*"' | head -1 | sed 's/.*://; s/"//g'; }

find_leader() {
  for n in "${NODES[@]}"; do
    [ "$(state_of "$n")" = "Leader" ] && { echo "$n"; return 0; }
  done
  return 1
}

wait_stable() { # all three nodes answer, exactly one Leader
  for _ in $(seq 1 60); do
    local leaders=0 answering=0
    for n in "${NODES[@]}"; do
      s=$(state_of "$n")
      [ -n "$s" ] && answering=$((answering+1))
      [ "$s" = "Leader" ] && leaders=$((leaders+1))
    done
    [ "$answering" = 3 ] && [ "$leaders" = 1 ] && return 0
    sleep 3
  done
  return 1
}

log "== Installing poller on all nodes =="
for n in "${NODES[@]}"; do
  gcloud compute scp "$HERE/failover-poller.sh" "incus-ha-$n:/tmp/failover-poller.sh" --zone "$ZONE" --tunnel-through-iap "${SSHK[@]}" --quiet 2>/dev/null
  ssh_node "$n" "sudo install -m 755 /tmp/failover-poller.sh /usr/local/bin/failover-poller.sh && sudo mkdir -p /var/lib/mds/failover"
  ssh_node "$n" "cd /opt/mds && git rev-parse --short HEAD && /usr/local/bin/metadata-service --version 2>/dev/null; uname -r" > "$OUTDIR/node-$n.txt" 2>&1
done

for run in $(seq 1 "$RUNS"); do
  log "== Run $run/$RUNS =="
  wait_stable || { log "cluster not stable before run $run"; echo "$run,,,,,FAIL_unstable" >> "$CSV"; continue; }
  leader=$(find_leader) || { log "no leader"; echo "$run,,,,,FAIL_no_leader" >> "$CSV"; continue; }
  survivors=(); for n in "${NODES[@]}"; do [ "$n" != "$leader" ] && survivors+=("$n"); done
  poller="${survivors[0]}"
  surv_ips=(); for n in "${survivors[@]}"; do surv_ips+=("${IP[$n]}"); done
  log "leader=$leader poller=$poller survivors=${survivors[*]}"

  ssh_node "$poller" "sudo rm -f /var/lib/mds/failover/run-$run.txt; sudo nohup env PORT=$PORT OUT=/var/lib/mds/failover/run-$run.txt /usr/local/bin/failover-poller.sh ${IP[$leader]} ${surv_ips[*]} >/dev/null 2>&1 &"
  # Wait until the poller confirmed the leader healthy.
  for _ in $(seq 1 20); do
    ssh_node "$poller" "grep -q ready /var/lib/mds/failover/run-$run.txt 2>/dev/null" && break
    sleep 1
  done

  log "SIGKILL metadata-service on $leader"
  ssh_node "$leader" "sudo systemctl kill -s SIGKILL metadata-service"

  result=""
  for _ in $(seq 1 90); do
    result=$(ssh_node "$poller" "cat /var/lib/mds/failover/run-$run.txt 2>/dev/null")
    echo "$result" | grep -q DONE && break
    sleep 2
  done
  echo "--- run $run ---" >> "$LOG"; echo "$result" >> "$LOG"
  line=$(echo "$result" | grep '^status=')
  status=$(echo "$line" | sed -n 's/.*status=\([A-Z]*\).*/\1/p')
  new_ip=$(echo "$line" | sed -n 's/.*new_leader=\([0-9.]*\).*/\1/p')
  new_leader=""; for n in "${NODES[@]}"; do [ "${IP[$n]}" = "$new_ip" ] && new_leader="$n"; done
  secs=$(echo "$line" | sed -n 's/.*reelection_s=\([0-9.]*\).*/\1/p')
  pms=$(echo "$line" | sed -n 's/.*poll_ms=\([0-9]*\).*/\1/p')
  echo "$run,$leader,$new_leader,$secs,$pms,${status:-FAIL_no_result}" >> "$CSV"
  log "run $run: $leader -> $new_leader in ${secs:-?} s ($status)"

  # systemd restarts the killed service after 3 s; wait for it to rejoin.
  sleep "$COOLDOWN_S"
done

log "== Done. Results: $CSV =="
cat "$CSV"
