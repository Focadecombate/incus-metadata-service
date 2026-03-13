package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/hashicorp/raft"
)

// FSM implements raft.FSM by applying replicated commands to the local database.
type FSM struct {
	db *db.Queries
}

// NewFSM creates a new FSM backed by the given database.
func NewFSM(database *db.Queries) *FSM {
	return &FSM{db: database}
}

// Apply is called by RAFT when a log entry is committed. It deserializes the
// command and applies the corresponding write operation to the local database.
func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		logs.Logger.Error().Err(err).Msg("failed to unmarshal raft command")
		return err
	}

	ctx := context.Background()

	switch cmd.Type {
	case CmdCreateInstance:
		var params db.CreateInstanceParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.CreateInstance(ctx, params)
		if err != nil {
			return err
		}
		return result

	case CmdUpdateInstance:
		var params db.UpdateInstanceParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.UpdateInstance(ctx, params)
		if err != nil {
			return err
		}
		return result

	case CmdCreateOrUpdateMetadata:
		var params db.CreateOrUpdateInstanceMetadataParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.CreateOrUpdateInstanceMetadata(ctx, params)
		if err != nil {
			return err
		}
		return result

	case CmdCreateOrUpdateUserData:
		var params db.CreateOrUpdateInstanceUserDataParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.CreateOrUpdateInstanceUserData(ctx, params)
		if err != nil {
			return err
		}
		return result

	case CmdCreateOrUpdateNetworkConfig:
		var params db.CreateOrUpdateInstanceNetworkConfigParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.CreateOrUpdateInstanceNetworkConfig(ctx, params)
		if err != nil {
			return err
		}
		return result

	case CmdCreateOrUpdateState:
		var params db.CreateOrUpdateInstanceStateParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.CreateOrUpdateInstanceState(ctx, params)
		if err != nil {
			return err
		}
		return result

	case CmdCreateOrUpdateVendorData:
		var params db.CreateOrUpdateInstanceVendorDataParams
		if err := json.Unmarshal(cmd.Data, &params); err != nil {
			return err
		}
		result, err := f.db.CreateOrUpdateInstanceVendorData(ctx, params)
		if err != nil {
			return err
		}
		return result

	default:
		return fmt.Errorf("unknown command type: %d", cmd.Type)
	}
}

// Snapshot returns a no-op snapshot. Full snapshot support is not implemented;
// nodes that fall too far behind must be re-bootstrapped.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &noopSnapshot{}, nil
}

// Restore is a no-op. See Snapshot.
func (f *FSM) Restore(_ io.ReadCloser) error {
	return nil
}

type noopSnapshot struct{}

func (s *noopSnapshot) Persist(_ raft.SnapshotSink) error { return nil }
func (s *noopSnapshot) Release()                          {}
