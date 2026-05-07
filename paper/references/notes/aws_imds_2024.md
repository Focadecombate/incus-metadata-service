# AWS Instance Metadata Service (IMDS) Documentation

**Author:** Amazon Web Services
**Year:** 2024
**Source:** Amazon EC2 Documentation
**URL:** https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html

## Summary

Documentation for the EC2 Instance Metadata Service, the de-facto standard metadata API that all other cloud providers emulate. Accessible at `169.254.169.254`, exposes a hierarchical HTTP API with versioned paths (`/latest/`, `/2009-04-04/`, etc.) for instance identity, hostname, IP addresses, SSH public keys, IAM credentials, user-data, and network interface information.

## Key Findings

- Established the `169.254.169.254` convention now used universally
- API is organized as a filesystem-like hierarchy: `/latest/meta-data/instance-id`, `/latest/meta-data/hostname`, etc.
- Supports multiple API versions from `1.0` through `2021-03-23`
- Serves user-data at `/latest/user-data` as raw bytes (not interpreted by the service)
- Provides IAM role credentials — the source of the SSRF attack vector
- Also exposes dynamic data at `/latest/dynamic/instance-identity/document`

## Pertinent Information for TCC

- **Defines the metadata API pattern** your service follows (though via NoCloud paths rather than EC2 paths)
- The hierarchical namespace and key-based lookup pattern influenced your `/configs/meta-data/:key` endpoint
- Cited in the introduction to explain what an IMDS is and in the related work for comparison
- Your service implements a subset of the fields (instance-id, hostname, IPs, network, public-keys)

## BibTeX Key

`aws_imds`
