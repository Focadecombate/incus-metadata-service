# GCP Metadata Server Documentation

**Author:** Google Cloud
**Year:** 2024
**Source:** Google Cloud Compute Engine Documentation
**URL:** https://cloud.google.com/compute/docs/metadata/overview

## Summary

Documentation for the Google Compute Engine Metadata Server, accessible at `metadata.google.internal` (resolves to `169.254.169.254`) and organized under `/computeMetadata/v1/`. Requires the mandatory header `Metadata-Flavor: Google` on all requests. Organized in two scopes: project-level metadata (shared across all instances) and instance-level metadata.

## Key Findings

- Uses hostname `metadata.google.internal` rather than raw IP (though both work)
- Mandatory `Metadata-Flavor: Google` header on all requests — SSRF mitigation (simpler than IMDSv2)
- Supports long polling via `?wait_for_change=true` for reactive metadata updates
- Instance metadata at `/computeMetadata/v1/instance/`, project metadata at `/computeMetadata/v1/project/`
- SSH keys stored at `/instance/attributes/ssh-keys`, user-data at `/instance/attributes/user-data`

## Pertinent Information for TCC

- Cited in the related work section on GCP Metadata Server and in the security comparison
- Header-based SSRF mitigation is simpler than IMDSv2 but weaker (some SSRF can set custom headers)
- The long polling feature is relevant as a contrast — your service uses periodic cron-based sync instead
- Comparison point in the table: GCP is proprietary and platform-coupled, your service is standalone

## BibTeX Key

`gcp_metadata`
