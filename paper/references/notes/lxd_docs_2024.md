# LXD Documentation

**Author:** Canonical Ltd.
**Year:** 2024
**Source:** Ubuntu documentation
**URL:** https://documentation.ubuntu.com/lxd/

## Summary

Official documentation for LXD, the original system container and VM manager developed by Canonical. Covers the same architecture Incus inherits: REST API, instance management, profiles, storage, networking, and cloud-init integration. Now maintained by Canonical independently from the community.

## Key Findings

- LXD provides the architectural foundation that Incus inherits
- Cloud-init integration via config keys and agent socket is identical to Incus
- Snap-based distribution model (Canonical's choice that contributed to the fork)
- Documentation covers the guest agent, config drive, and `/dev/lxd/sock` socket

## Pertinent Information for TCC

- Cited in the related work to establish LXD as Incus's predecessor
- Shared architecture means the metadata service gap exists in both LXD and Incus
- LXD documentation helps explain the socket-based cloud-init delivery mechanism

## BibTeX Key

`lxd2024`
