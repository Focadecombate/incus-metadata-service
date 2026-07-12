package consensus

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

// RaftConfig holds the configuration for the RAFT consensus node.
type RaftConfig struct {
	// NodeID is a unique identifier for this node in the cluster.
	NodeID string `env:"NODE_ID,default=node1"`
	// BindAddr is the address for RAFT inter-node communication (e.g. "localhost:7000").
	BindAddr string `env:"BIND_ADDR,default=localhost:7000"`
	// DataDir is the directory where RAFT stores its log and snapshots.
	DataDir string `env:"DATA_DIR,default=raft-data"`
	// Peers is a comma-separated list of peer entries in "id=host:port" form
	// (e.g. "node2=10.0.0.2:7000,node3=10.0.0.3:7000"). Each entry's id must be
	// the peer's RAFT_NODE_ID so the bootstrap configuration matches real nodes.
	// Leave empty for single-node mode.
	Peers []string `env:"PEERS"`
	// Bootstrap indicates whether this node should bootstrap a new cluster.
	Bootstrap bool `env:"BOOTSTRAP,default=false"`
}

// RaftNode wraps a hashicorp/raft instance and provides helper methods.
type RaftNode struct {
	Raft   *raft.Raft
	fsm    *FSM
	config *RaftConfig
}

// NewRaftNode creates and starts a new RAFT node.
func NewRaftNode(cfg *RaftConfig, database *db.Queries) (*RaftNode, error) {
	logger := logs.Logger.With().Str("component", "raft").Logger()
	logger.Info().
		Str("node_id", cfg.NodeID).
		Str("bind_addr", cfg.BindAddr).
		Msg("Initializing RAFT node")

	// Create data directory
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create raft data dir: %w", err)
	}

	// RAFT configuration
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.NodeID)
	raftCfg.HeartbeatTimeout = 1000 * time.Millisecond
	raftCfg.ElectionTimeout = 1000 * time.Millisecond
	raftCfg.LeaderLeaseTimeout = 500 * time.Millisecond
	raftCfg.CommitTimeout = 500 * time.Millisecond

	// FSM
	fsm := NewFSM(database)

	// Log store and stable store (BoltDB)
	boltStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create bolt store: %w", err)
	}

	// Snapshot store
	snapshotStore := raft.NewDiscardSnapshotStore()

	// Transport
	addr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bind address: %w", err)
	}

	transport, err := raft.NewTCPTransport(cfg.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create RAFT instance
	r, err := raft.NewRaft(raftCfg, fsm, boltStore, boltStore, snapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	node := &RaftNode{
		Raft:   r,
		fsm:    fsm,
		config: cfg,
	}

	// Bootstrap cluster if configured
	if cfg.Bootstrap {
		servers := []raft.Server{
			{
				ID:      raft.ServerID(cfg.NodeID),
				Address: raft.ServerAddress(cfg.BindAddr),
			},
		}

		// Add peers as voters. Each entry carries the peer's real node id so the
		// bootstrap configuration matches the peers' actual LocalIDs.
		for _, peer := range cfg.Peers {
			id, addr, found := strings.Cut(peer, "=")
			if !found || id == "" || addr == "" {
				return nil, fmt.Errorf("invalid peer entry %q: expected format id=host:port", peer)
			}
			servers = append(servers, raft.Server{
				ID:      raft.ServerID(id),
				Address: raft.ServerAddress(addr),
			})
		}

		future := r.BootstrapCluster(raft.Configuration{Servers: servers})
		if err := future.Error(); err != nil && err != raft.ErrCantBootstrap {
			return nil, fmt.Errorf("failed to bootstrap cluster: %w", err)
		}

		logger.Info().Int("servers", len(servers)).Msg("Cluster bootstrap attempted")
	}

	logger.Info().Msg("RAFT node initialized")
	return node, nil
}

// IsLeader returns true if this node is the current RAFT leader.
func (n *RaftNode) IsLeader() bool {
	return n.Raft.State() == raft.Leader
}

// LeaderAddr returns the address of the current leader.
func (n *RaftNode) LeaderAddr() string {
	addr, _ := n.Raft.LeaderWithID()
	return string(addr)
}

// Apply proposes a command to the RAFT cluster. Only the leader can apply.
// The command is replicated to all nodes and applied via the FSM.
func (n *RaftNode) Apply(cmdType CommandType, data interface{}) (interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command data: %w", err)
	}

	cmd := Command{
		Type: cmdType,
		Data: dataBytes,
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	future := n.Raft.Apply(cmdBytes, 5*time.Second)
	if err := future.Error(); err != nil {
		return nil, fmt.Errorf("failed to apply raft command: %w", err)
	}

	response := future.Response()
	if err, ok := response.(error); ok {
		return nil, err
	}

	return response, nil
}

// Shutdown gracefully shuts down the RAFT node.
func (n *RaftNode) Shutdown() error {
	return n.Raft.Shutdown().Error()
}
