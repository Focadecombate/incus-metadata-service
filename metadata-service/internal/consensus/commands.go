package consensus

import "encoding/json"

// CommandType identifies the type of write operation to replicate.
type CommandType uint8

const (
	CmdCreateInstance CommandType = iota
	CmdUpdateInstance
	CmdCreateOrUpdateMetadata
	CmdCreateOrUpdateUserData
	CmdCreateOrUpdateNetworkConfig
	CmdCreateOrUpdateState
	CmdCreateOrUpdateVendorData
)

// Command represents a write operation to be replicated via RAFT.
type Command struct {
	Type CommandType `json:"type"`
	Data json.RawMessage `json:"data"`
}
