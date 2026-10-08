# AWS IMDSv2 Documentation

**Author:** Amazon Web Services
**Year:** 2024
**Source:** Amazon EC2 Documentation
**URL:** https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html

## Summary

Documentation for IMDSv2, the session-oriented token-based authentication layer for the EC2 metadata service introduced in November 2019. Requires a two-step protocol: (1) PUT request to `/latest/api/token` with TTL header to obtain a session token, then (2) GET requests with the token in `X-aws-ec2-metadata-token` header.

## Key Findings

- PUT request requirement blocks most SSRF attacks (which can only issue GET)
- Custom header requirement (`X-aws-ec2-metadata-token-ttl-seconds`) blocks SSRF that can't set headers
- IP hop limit of 1 on PUT responses prevents token forwarding through network hops
- Presence of `X-Forwarded-For` in PUT request causes rejection — blocks proxy-based attacks
- Token is instance-specific and time-limited (configurable TTL up to 6 hours)

## Pertinent Information for TCC

- Cited in the related work section on AWS IMDS security and in the security subsection
- Referenced in the conclusion as a model for future token-based auth in your service
- The five security properties of IMDSv2 are the gold standard for metadata service security
- Your service currently doesn't implement token auth but documents it as future work

## BibTeX Key

`aws_imdsv2`
