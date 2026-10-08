# NoCloud Datasource — cloud-init Documentation

**Author:** Canonical Ltd.
**Year:** 2024
**Source:** cloud-init documentation
**URL:** https://docs.cloud-init.io/en/latest/reference/datasources/nocloud.html

## Summary

Documentation for the NoCloud datasource, the most flexible cloud-init datasource designed for environments without a dedicated cloud platform. Supports fetching configuration from a base URL (`seedfrom`) via HTTP/HTTPS, or from a labeled filesystem (`CIDATA`). Requires minimal metadata: just `instance-id` and `local-hostname`.

## Key Findings

- HTTP endpoints relative to `seedfrom`: `/meta-data`, `/user-data`, `/vendor-data`, `/network-config`
- Minimum valid `meta-data`: `instance-id` and `local-hostname` fields
- Supports HTTP, HTTPS, FTP, FTPS protocols
- Can be configured via kernel command line (`ds=nocloud;s=http://...`) or DMI data
- `dsmode: net` triggers network-based fetching (vs local filesystem)

## Pertinent Information for TCC

- **Critical reference** — your service implements the NoCloud HTTP datasource pattern
- Your `/configs/` prefix maps directly to the `seedfrom` base URL
- The four endpoints your service exposes match exactly: meta-data, user-data, vendor-data, network-config
- Cited in the related work (NoCloud subsection) and proposal (endpoints section)
- The `seedfrom` parameter is configured in Incus profiles to point to your service

## BibTeX Key

`cloudinit_nocloud`
