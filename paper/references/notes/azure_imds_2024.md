# Azure Instance Metadata Service

**Author:** Microsoft
**Year:** 2024
**Source:** Azure documentation
**URL:** https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service

## Summary

Documentation for Azure's Instance Metadata Service (IMDS), accessible at `169.254.169.254`. Provides VM identity, configuration, network, and scheduled events information. Requires the `Metadata: true` header on all requests as a basic SSRF mitigation.

## Key Findings

- Accessible at `http://169.254.169.254/metadata/instance?api-version=2021-02-01`
- Requires `Metadata: true` header on all requests (rejected without it)
- Provides: compute info, network config, scheduled maintenance events, identity tokens
- API is versioned via query parameter (`api-version=YYYY-MM-DD`)
- Managed identity tokens available for Azure resource access (analogous to AWS IAM roles)

## Pertinent Information for TCC

- Cited in the security subsection and GCP section for header-based SSRF comparison
- Simpler SSRF mitigation than IMDSv2 but stronger than no protection
- Adds to the comparative landscape: AWS (tokens), GCP (flavor header), Azure (metadata header), your service (none yet)

## BibTeX Key

`azure_imds`
