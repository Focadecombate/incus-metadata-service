# Defense in Depth: SSRF Vulnerabilities and EC2 Instance Metadata Service

**Author:** Amazon Web Services
**Year:** 2019
**Source:** AWS Security Blog
**URL:** https://aws.amazon.com/blogs/security/defense-in-depth-open-firewalls-reverse-proxies-ssrf-vulnerabilities-ec2-instance-metadata-service/

## Summary

AWS Security blog post introducing IMDSv2 and explaining the threat model it defends against. Details the five security properties of IMDSv2: PUT request requirement, custom header requirement, IP hop limit of 1, X-Forwarded-For rejection, and instance-specific tokens. Explains how these properties mitigate open firewalls, reverse proxies, and SSRF vulnerabilities.

## Key Findings

- Threat model: SSRF → metadata → credential theft → cloud resource access
- IMDSv2 requires PUT (most SSRF only allows GET)
- Custom TTL header required on PUT (SSRF can't set custom headers)
- Hop limit = 1 on PUT responses prevents multi-hop token forwarding
- X-Forwarded-For presence in PUT causes rejection (blocks proxy attacks)
- Mitigates: open firewalls, open reverse proxies, SSRF, misconfigured layer-3 NAT

## Pertinent Information for TCC

- Cited in the related work (AWS IMDS and security sections) and conclusion (future work)
- Provides the security model your service could adopt in the future
- The five properties are the reference framework for evaluating metadata service security
- Comparison with GCP (header-only) and Azure (header-only) shows IMDSv2 is strongest

## BibTeX Key

`aws_imdsv2_blog`
