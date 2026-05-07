# EC2 Datasource — cloud-init Documentation

**Author:** Canonical Ltd.
**Year:** 2024
**Source:** cloud-init documentation
**URL:** https://docs.cloud-init.io/en/latest/reference/datasources/ec2.html

## Summary

Documentation for the EC2 datasource in cloud-init. Defines how cloud-init discovers and communicates with EC2-compatible metadata services at `169.254.169.254`. This is the most widely emulated datasource — OpenStack, CloudStack, and other providers implement EC2-compatible endpoints.

## Key Findings

- Uses versioned paths: `/latest/meta-data/`, `/latest/user-data`, etc.
- Auto-detected via DMI data or kernel command line
- Supports IMDSv2 token-based authentication
- Most widely supported datasource across cloud providers
- OpenStack serves EC2-compatible endpoints alongside its native API

## Pertinent Information for TCC

- Cited in the related work (Incus/LXD section) as one of the datasources Incus can't use
- Your service uses NoCloud HTTP instead of EC2 paths — simpler but less compatible
- The EC2 datasource's versioned path convention (`/latest/`, `/2009-04-04/`) is noted as a feature your service doesn't implement

## BibTeX Key

`cloudinit_ec2_ds`
