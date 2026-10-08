# Incus — System container and virtual machine manager

**Author:** Linux Containers
**Year:** 2024
**Source:** Official project page
**URL:** https://linuxcontainers.org/incus/

## Summary

Official documentation for Incus, a community-driven system container and virtual machine manager forked from LXD in 2023. Supports both system containers (sharing host kernel) and full virtual machines (via QEMU). Provides cloud-init integration through configuration keys (`cloud-init.user-data`, `cloud-init.vendor-data`, `cloud-init.network-config`) that are delivered to instances via the Incus agent and `/dev/incus/sock` Unix socket.

## Key Findings

- Manages both containers and VMs from a unified API
- cloud-init config keys can be set directly on instances or propagated via profiles
- No native HTTP metadata endpoint — relies on Incus agent socket inside instances
- `security.guestapi` controls access to the guest API socket
- Licensed under Apache 2.0 with no CLA requirement (unlike LXD under Canonical)

## Pertinent Information for TCC

- **Core reference** — your service is built specifically for Incus
- The absence of an HTTP metadata endpoint is the fundamental gap your TCC fills
- Config keys (`cloud-init.user-data`, etc.) are the data sources your sync mechanism reads from
- Cited throughout: introduction, related work (Incus/LXD section), proposal, and conclusion

## BibTeX Key

`incus2024`
