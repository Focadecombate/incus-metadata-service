# Incus cloud-init & LXD datasource — Linux Containers Forum

**Authors:** Gräber, Stéphane and others
**Year:** 2024
**Source:** Linux Containers Forum discussion
**URL:** https://discuss.linuxcontainers.org/t/incus-cloud-init-lxd-datasource/19574

## Summary

Community forum discussion about the cloud-init integration issues with Incus. Stéphane Gräber and community members discuss the timing problem where the LXD datasource fails because the Incus agent socket isn't available early enough in the boot sequence. Proposes solutions including a native Incus datasource for cloud-init and HTTP-based alternatives.

## Key Findings

- The Incus agent starts too late — `ds-identify` runs before `/dev/incus/sock` is available
- This is a race condition: cloud-init's early stages need data before the agent provides the socket
- Proposed solution by Gräber: native Incus datasource that falls back to NoCloud if socket unavailable
- Community interest in an HTTP-based metadata service as a more robust alternative
- The problem is worse for VMs than containers (agent startup takes longer in VMs)

## Pertinent Information for TCC

- **Critical reference** — documents the exact problem your TCC solves
- Cited in the related work (Incus/LXD section) to justify the need for an HTTP metadata service
- Shows that the community recognizes the gap — your work is a direct response
- The proposed fallback to NoCloud HTTP aligns with your implementation approach

## BibTeX Key

`incus_cloudinit_forum`
