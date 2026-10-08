# RFC 3927 — Dynamic Configuration of IPv4 Link-Local Addresses

**Authors:** Cheshire, Stuart; Aboba, Bernard; Guttman, Erik
**Year:** 2005
**Source:** IETF RFC 3927
**URL:** https://www.rfc-editor.org/rfc/rfc3927

## Summary

IETF standard defining the mechanism for automatic configuration of IPv4 link-local addresses in the `169.254.0.0/16` range. These addresses are non-routable and scoped to a single network segment, making them suitable for host-to-hypervisor communication within cloud infrastructure.

## Key Findings

- Defines the `169.254.0.0/16` range for automatic link-local address assignment
- Addresses are non-routable — cannot traverse network boundaries (ideal for metadata services)
- Originally designed for zero-configuration networking (devices that need IP addresses without DHCP)
- The specific address `169.254.169.254` became the universal convention for cloud metadata services

## Pertinent Information for TCC

- Cited in the introduction to explain the origin of `169.254.169.254`
- Referenced in the conclusion as the address convention your service should support via future work
- The non-routable property is a natural security boundary for metadata services
- Explains why all cloud providers converged on this address range

## BibTeX Key

`rfc3927`
