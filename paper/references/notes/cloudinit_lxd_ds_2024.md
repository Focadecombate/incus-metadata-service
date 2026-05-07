# LXD Datasource — cloud-init Documentation

**Author:** Canonical Ltd.
**Year:** 2024
**Source:** cloud-init documentation
**URL:** https://docs.cloud-init.io/en/latest/reference/datasources/lxd.html

## Summary

Documentation for the LXD datasource, used by cloud-init when running inside LXD/Incus instances. Communicates via Unix socket (`/dev/lxd/sock` or `/dev/incus/sock`) rather than HTTP over the network. Reads configuration from `cloud-init.user-data`, `cloud-init.vendor-data`, `cloud-init.network-config`, and `user.*` keys.

## Key Findings

- Uses Unix domain socket, NOT HTTP over the network
- Socket provided by the LXD/Incus agent running inside the instance
- Timing problem: `ds-identify` and `cloud-init-local` run before the agent starts the socket
- Configuration keys: `cloud-init.user-data`, `cloud-init.vendor-data`, `cloud-init.network-config`
- Falls back to NoCloud if the socket is unavailable

## Pertinent Information for TCC

- Cited in the related work section on Incus/LXD metadata gap
- The timing problem with the socket is the **primary motivation** for building an HTTP metadata service
- Your service provides an alternative path: HTTP network datasource instead of Unix socket
- The same config keys your sync mechanism reads are documented here

## BibTeX Key

`cloudinit_lxd_ds`
