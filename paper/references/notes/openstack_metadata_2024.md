# OpenStack Metadata Service Documentation

**Author:** OpenStack Foundation
**Year:** 2024
**Source:** OpenStack Nova Documentation
**URL:** https://docs.openstack.org/nova/latest/user/metadata.html

## Summary

Documentation for the OpenStack metadata service, which serves both an EC2-compatible API (`/latest/meta-data/`, `/2009-04-04/`) and a native OpenStack API (`/openstack/latest/meta_data.json`). Tightly coupled with Nova (compute) for data provisioning and Neutron (network) for request routing via metadata proxy.

## Key Findings

- Dual API: EC2-compatible paths for cloud-init EC2 datasource + native JSON API
- Neutron metadata proxy intercepts requests to `169.254.169.254` from VMs and adds instance identification headers (`X-Instance-ID`, `X-Tenant-ID`)
- Cannot run standalone — depends on Nova, Neutron, and Keystone
- Handles overlapping tenant IP addresses via the proxy architecture
- Supports vendor-data at `/openstack/latest/vendor_data.json` and `/vendor_data2.json`

## Pertinent Information for TCC

- Cited in the related work section as the main open-source cloud platform with metadata service
- Key contrast: OpenStack's metadata service cannot run independently, while your service is standalone
- The Neutron proxy architecture solves instance identification differently than your IP-based approach
- Used in the comparison table to show that OpenStack is open-source but tightly coupled

## BibTeX Key

`openstack_metadata`
