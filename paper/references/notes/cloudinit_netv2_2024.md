# Networking Config Version 2 — cloud-init Documentation

**Author:** Canonical Ltd.
**Year:** 2024
**Source:** cloud-init documentation
**URL:** https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v2.html

## Summary

Documentation for the Networking Config Version 2 format, based on Netplan (also used by Ubuntu's network configuration). Defines network interfaces, addresses, routes, DNS, bonds, bridges, and VLANs in a declarative YAML format under the `ethernets`, `bonds`, `bridges`, and `vlans` keys.

## Key Findings

- Based on Netplan specification — same format used by Ubuntu's native network config
- Version 2 is the modern format; version 1 is cloud-init's legacy format
- Supports match by MAC address, driver, or name for interface identification
- Supports DHCP4/DHCP6, static addresses, routes, DNS, and MTU
- Backends: systemd-networkd or NetworkManager

## Pertinent Information for TCC

- Cited in the proposal section — your service generates network config in this format
- Your `NetworkConfig` Go struct mirrors the v2 YAML structure (ethernets, match by MAC, addresses)
- When `cloud-init.network-config` is set in Incus, it's parsed as v2 YAML
- Auto-generated config matches interfaces by MAC address from the Incus instance state

## BibTeX Key

`cloudinit_netv2`
