# cloud-init Documentation

**Author:** Canonical Ltd. / cloud-init contributors
**Year:** 2024 (ongoing)
**Source:** Official documentation
**URL:** https://cloudinit.readthedocs.io/

## Summary

Official documentation for cloud-init, the industry-standard tool for cloud instance initialization created by Canonical in 2009-2010. Covers the four boot stages (init-local, init, modules-config, modules-final), datasource specification, user-data formats (cloud-config YAML, shell scripts, MIME multipart, Jinja templates, boothooks), vendor-data, and network configuration formats (v1 and v2).

## Key Findings

- cloud-init runs in 4 ordered boot stages with strict timing requirements
- Datasources are the abstraction layer: each cloud provider implements one (EC2, GCE, OpenStack, NoCloud, LXD)
- `instance-id` is the critical metadata field — cloud-init uses it to detect first boot vs subsequent boot
- User-data format is auto-detected by first line (`#cloud-config`, `#!/bin/...`, `#boothook`, etc.)
- Network config v2 is Netplan-based, v1 is cloud-init's own format

## Pertinent Information for TCC

- **Primary reference** for the cloud-init protocol your service implements
- Defines the NoCloud HTTP datasource paths your endpoints follow (`/meta-data`, `/user-data`, `/vendor-data`, `/network-config`)
- Boot stages explain why the Incus agent timing problem exists (ds-identify runs at init-local, before the socket is available)
- Cited throughout the paper: intro, related work, proposal, and endpoints sections

## BibTeX Key

`cloudinit2024`
