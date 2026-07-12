package consensus

import (
	"context"
	"testing"
	"time"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
)

// newTestDB creates an isolated on-disk sqlite database (one file per node, in a
// temp dir) with the full schema applied, and returns its Queries handle.
func newTestDB(t *testing.T) *db.Queries {
	t.Helper()
	dbPath := t.TempDir() + "/metadata.db"
	cfg := &config.Config{
		Database: &config.DatabaseConfig{
			DBDriver: "sqlite",
			DBSource: dbPath,
		},
	}
	q, err := db.ConnectDB(cfg)
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}
	return q
}

// newTestNode spins up an in-process RaftNode bound to addr with its own temp
// DataDir and its own sqlite database. It returns the node and its db handle so
// tests can assert replication by querying each replica directly.
func newTestNode(t *testing.T, id, addr string, bootstrap bool, peers []string) (*RaftNode, *db.Queries) {
	t.Helper()
	database := newTestDB(t)
	cfg := &RaftConfig{
		NodeID:    id,
		BindAddr:  addr,
		DataDir:   t.TempDir(),
		Peers:     peers,
		Bootstrap: bootstrap,
	}
	node, err := NewRaftNode(cfg, database)
	if err != nil {
		t.Fatalf("failed to create raft node %s: %v", id, err)
	}
	return node, database
}

// waitForLeader polls the given nodes until exactly one reports IsLeader(),
// returning it. Fails the test if no single leader emerges within timeout.
func waitForLeader(t *testing.T, nodes []*RaftNode, timeout time.Duration) *RaftNode {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var leader *RaftNode
		count := 0
		for _, n := range nodes {
			if n.IsLeader() {
				count++
				leader = n
			}
		}
		if count == 1 {
			return leader
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("no single leader elected within %s", timeout)
	return nil
}

// waitForConvergence polls every node's database until all of them contain the
// named instance, proving the create was replicated through the RAFT log.
func waitForConvergence(t *testing.T, dbs []*db.Queries, name, project string, timeout time.Duration) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		converged := true
		for _, q := range dbs {
			if _, err := q.GetInstance(ctx, db.GetInstanceParams{Name: name, Project: project}); err != nil {
				converged = false
				break
			}
		}
		if converged {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("instance %s/%s did not converge to all nodes within %s", name, project, timeout)
}

func strPtr(s string) *string { return &s }

// TestClusterFormationAndFailover validates that a 3-node RAFT cluster forms,
// replicates a write to every node, and re-elects a leader after the current
// leader is killed — all in-process, with no Incus or external processes.
func TestClusterFormationAndFailover(t *testing.T) {
	const (
		addr1 = "127.0.0.1:7801"
		addr2 = "127.0.0.1:7802"
		addr3 = "127.0.0.1:7803"
	)

	// node1 bootstraps the cluster and lists node2/node3 with their real ids and
	// addresses; node2/node3 start without bootstrapping and are caught up by the
	// leader's replicated configuration.
	node1, db1 := newTestNode(t, "node1", addr1, true, []string{"node2=" + addr2, "node3=" + addr3})
	node2, db2 := newTestNode(t, "node2", addr2, false, nil)
	node3, db3 := newTestNode(t, "node3", addr3, false, nil)

	nodes := []*RaftNode{node1, node2, node3}
	dbs := []*db.Queries{db1, db2, db3}

	// Idempotent cleanup: Shutdown is safe to call more than once, so this also
	// covers the leader we deliberately kill mid-test.
	defer func() {
		for _, n := range nodes {
			_ = n.Shutdown()
		}
	}()

	// 1. Cluster formation: exactly one leader.
	leader := waitForLeader(t, nodes, 10*time.Second)
	t.Logf("initial leader elected: %s", leader.config.NodeID)

	// 2. Replication: apply a create through the leader, expect convergence.
	params := db.CreateInstanceParams{
		Name:      "test-instance",
		Project:   "default",
		IpAddress: strPtr("10.0.0.42"),
	}
	if _, err := leader.Apply(CmdCreateInstance, params); err != nil {
		t.Fatalf("failed to apply CmdCreateInstance on leader: %v", err)
	}
	waitForConvergence(t, dbs, "test-instance", "default", 10*time.Second)
	t.Log("instance replicated and converged on all 3 nodes")

	// 3. Failover: kill the leader and expect a new one among the survivors.
	if err := leader.Shutdown(); err != nil {
		t.Fatalf("failed to shut down leader: %v", err)
	}

	var remaining []*RaftNode
	for _, n := range nodes {
		if n != leader {
			remaining = append(remaining, n)
		}
	}

	newLeader := waitForLeader(t, remaining, 10*time.Second)
	if newLeader == leader {
		t.Fatal("expected a new leader distinct from the killed leader")
	}
	t.Logf("re-election succeeded, new leader: %s", newLeader.config.NodeID)
}
