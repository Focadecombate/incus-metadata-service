#!/usr/bin/env bash
#
# failover-test.sh — Raft leader-failover experiment for the metadata service.
# Run AFTER `terraform apply` in terraform-ha/ and after all nodes report
# PROVISION_DONE. Produces the HA result for the paper (04_resultados.tex).
#
# It: (1) finds the leader, (2) generates data on the leader and verifies it
# replicates to the followers, (3) kills the leader's service and measures
# re-election time on a survivor, (4) verifies the new leader serves the data.
#
# USAGE: ZONE=us-central1-a KEY=~/.ssh/gce_key ./failover-test.sh
set -uo pipefail

ZONE="${ZONE:-us-central1-a}"
KEY="${KEY:-}"
PORT="${PORT:-8080}"
NODES=(node1 node2 node3)
SSHK=(); [ -n "$KEY" ] && SSHK=(--ssh-key-file "$KEY")

ssh_node() { # node, cmd
  local n="$1"; shift
  gcloud compute ssh "incus-ha-$n" --zone "$ZONE" --tunnel-through-iap "${SSHK[@]}" \
    --command="$*" 2>/dev/null | grep -vE "WARNING:|NumPy|performance of the tunnel|please see|^To increase|unable to resolve host|^$"
}

state_of() { # node -> leader|follower|candidate|down
  ssh_node "$1" "curl -s -m 3 http://127.0.0.1:$PORT/raft/status" 2>/dev/null \
    | (grep -o '\"state\":\"[A-Za-z]*\"' || echo down) | head -1 | sed 's/.*://; s/\"//g'
}

echo "== 1. Cluster state =="
LEADER=""
for n in "${NODES[@]}"; do
  s=$(state_of "$n"); echo "  $n: $s"
  [ "$s" = "Leader" ] && LEADER="$n"
done
[ -z "$LEADER" ] && { echo "FAIL: no leader found — cluster did not form"; exit 1; }
echo "  leader = $LEADER"

echo "== 2. Generate data on leader, verify replication =="
ssh_node "$LEADER" "sudo incus launch ${SEED:-mds-ubuntu-2404} ha1 2>/dev/null; sleep 20"
# The instance row replicates via Raft; check each node's DB for it.
for n in "${NODES[@]}"; do
  cnt=$(ssh_node "$n" "sudo sqlite3 /var/lib/mds/metadata.db \"select count(*) from instances where name='ha1' and deleted_at is null;\" 2>/dev/null" | tail -1)
  echo "  $n has ha1 row: ${cnt:-?}"
done

echo "== 3. Kill leader ($LEADER), measure re-election on a survivor =="
SURVIVOR=""; for n in "${NODES[@]}"; do [ "$n" != "$LEADER" ] && SURVIVOR="$n" && break; done
ssh_node "$LEADER" "sudo systemctl stop metadata-service" &
# Poll the survivor locally for the fastest, most accurate timing.
ssh_node "$SURVIVOR" '
  start=$(date +%s.%N)
  for i in $(seq 1 200); do
    s=$(curl -s -m 1 http://127.0.0.1:'"$PORT"'/raft/status | grep -o "\"state\":\"[A-Za-z]*\"" | head -1)
    case "$s" in
      *Leader*) end=$(date +%s.%N); awk -v a=$start -v b=$end "BEGIN{printf \"  re-elected in %.2fs\n\", b-a}"; exit 0;;
    esac
    sleep 0.1
  done
  echo "  no re-election within 20s"; exit 1
'
echo "  survivor polled = $SURVIVOR"

echo "== 4. New leader state + data availability =="
for n in "${NODES[@]}"; do [ "$n" != "$LEADER" ] && echo "  $n: $(state_of "$n")"; done
cnt=$(ssh_node "$SURVIVOR" "sudo sqlite3 /var/lib/mds/metadata.db \"select count(*) from instances where name='ha1' and deleted_at is null;\"" | tail -1)
echo "  $SURVIVOR still has ha1 row: ${cnt:-?}"

echo "== Done. Restore the killed node with: gcloud compute ssh incus-ha-$LEADER --zone $ZONE --tunnel-through-iap -- sudo systemctl start metadata-service =="
