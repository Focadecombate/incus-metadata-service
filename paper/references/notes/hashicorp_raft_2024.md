# hashicorp/raft — Golang implementation of the Raft consensus protocol

**Author:** HashiCorp
**Year:** 2024
**Source:** GitHub repository
**URL:** https://github.com/hashicorp/raft

## Summary

Production-grade Go implementation of the RAFT consensus protocol by HashiCorp. Used in Consul (service mesh), Vault (secrets management), and Nomad (workload orchestrator). Provides an API for leader election, log replication, snapshot/restore, and membership changes.

## Key Findings

- Stable, battle-tested API used in critical infrastructure software
- Supports pluggable log stores (BoltDB, in-memory) and snapshot sinks
- FSM (Finite State Machine) interface: user implements `Apply()` to handle committed log entries
- Automatic leader election and failure detection
- Supports dynamic cluster membership changes (add/remove nodes)

## Pertinent Information for TCC

- Cited in the related work (RAFT subsection) and proposal (consensus section)
- Your service implements the FSM interface to apply sync commands to the local SQLite database
- Uses BoltDB as the RAFT log store for persistence
- Commands: CmdCreateInstance, CmdUpdateInstance, CmdCreateOrUpdateMetadata, CmdCreateOrUpdateUserData, CmdCreateOrUpdateVendorData, CmdCreateOrUpdateNetworkConfig, CmdCreateOrUpdateState

## BibTeX Key

`hashicorp_raft`
