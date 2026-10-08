#!/usr/bin/env bash
#
# failover-poller.sh: runs ON a surviving Raft node during one failover run.
# Polls the current leader's /raft/status over the internal network every
# POLL_MS until it stops answering (t_down), then polls the surviving nodes
# until one of them reports state=Leader (t_elected). Both timestamps come
# from the same clock, so the measured interval does not depend on SSH latency.
#
# USAGE: failover-poller.sh <leader_ip> <survivor_ip> [<survivor_ip> ...]
# Env:   PORT (default 8080), POLL_MS (default 50), TIMEOUT_S (default 60), OUT (result file)
set -uo pipefail

LEADER_IP="$1"; shift
SURVIVORS=("$@")
PORT="${PORT:-8080}"
POLL_MS="${POLL_MS:-50}"
TIMEOUT_S="${TIMEOUT_S:-60}"
OUT="${OUT:-/var/lib/mds/failover/result.txt}"
mkdir -p "$(dirname "$OUT")"
: > "$OUT"

sleep_poll() { sleep "$(awk -v ms="$POLL_MS" 'BEGIN{printf "%.3f", ms/1000}')"; }
now() { date +%s.%N; }
state_of() { curl -s -m 0.5 "http://$1:$PORT/raft/status" 2>/dev/null | grep -o '"state":"[A-Za-z]*"' | sed 's/.*://; s/"//g'; }

# 1. Confirm the leader is healthy before the kill.
ok=0
for _ in $(seq 1 100); do
  [ "$(state_of "$LEADER_IP")" = "Leader" ] && { ok=1; break; }
  sleep_poll
done
[ "$ok" = 1 ] || { echo "status=FAIL reason=leader_not_healthy" >> "$OUT"; echo DONE >> "$OUT"; exit 1; }
echo "ready" >> "$OUT"

# 2. Wait for the leader to stop answering.
t_start=$(now)
t_down=""
while :; do
  s=$(state_of "$LEADER_IP")
  if [ "$s" != "Leader" ]; then t_down=$(now); break; fi
  if awk -v a="$t_start" -v b="$(now)" -v t="$TIMEOUT_S" 'BEGIN{exit !(b-a>t)}'; then
    echo "status=FAIL reason=leader_never_went_down" >> "$OUT"; echo DONE >> "$OUT"; exit 1
  fi
  sleep_poll
done

# 3. Wait for a survivor to become leader.
t_elected=""; new_leader=""
while :; do
  for ip in "${SURVIVORS[@]}"; do
    if [ "$(state_of "$ip")" = "Leader" ]; then t_elected=$(now); new_leader="$ip"; break 2; fi
  done
  if awk -v a="$t_down" -v b="$(now)" -v t="$TIMEOUT_S" 'BEGIN{exit !(b-a>t)}'; then
    echo "status=FAIL reason=no_reelection_within_${TIMEOUT_S}s t_down=$t_down" >> "$OUT"; echo DONE >> "$OUT"; exit 1
  fi
  sleep_poll
done

elapsed=$(awk -v a="$t_down" -v b="$t_elected" 'BEGIN{printf "%.3f", b-a}')
echo "status=OK old_leader=$LEADER_IP new_leader=$new_leader t_down=$t_down t_elected=$t_elected reelection_s=$elapsed poll_ms=$POLL_MS" >> "$OUT"
echo DONE >> "$OUT"
