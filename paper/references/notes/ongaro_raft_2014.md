# In Search of an Understandable Consensus Algorithm

**Authors:** Ongaro, Diego; Ousterhout, John
**Year:** 2014
**Source:** 2014 USENIX Annual Technical Conference (USENIX ATC 14), pp. 305-319

## Summary

The RAFT consensus algorithm paper, designed as an understandable alternative to Paxos. Decomposes distributed consensus into three sub-problems: leader election, log replication, and safety. Achieves the same guarantees as Paxos but with a clearer structure that makes correct implementation more tractable. One of the most influential systems papers of the 2010s.

## Key Findings

- Leader-based: one leader receives all client requests, replicates to followers
- Leader election via randomized timeouts — simple and effective
- Log replication: leader appends entries and replicates; committed after majority acknowledgement
- Safety guarantee: committed entries are never lost (even with leader changes)
- Tolerates up to ⌊(n-1)/2⌋ failures in a cluster of n nodes
- Designed primarily for understandability — formally proven equivalent to Multi-Paxos

## Pertinent Information for TCC

- **Key reference** for the high-availability module in your service
- Cited in the related work section (RAFT subsection) and proposal (consensus section)
- Your service uses hashicorp/raft which implements this paper's algorithm
- Only the leader syncs from Incus; writes are replicated to followers via RAFT log

## BibTeX Key

`ongaro2014raft`
