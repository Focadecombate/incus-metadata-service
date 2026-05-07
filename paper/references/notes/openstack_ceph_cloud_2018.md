# Leveraging OpenStack and Ceph for a Controlled-Access Data Cloud

**Authors:** Bollig, E. F.; Allan, G. T.; Lynch, B. J.; Huerta, Y. A.; Mix, M.; Munsell, E. A.; Benson, R. M.; Swartz, B.
**Year:** 2018
**Source:** PEARC '18 (Practice and Experience in Advanced Research Computing), ACM, Pittsburgh
**PDF:** `openstack_ceph_cloud_2018.pdf`

## Summary

Describes the design and implementation of Stratus, a private cloud at the University of Minnesota built on OpenStack Newton + Ceph Luminous storage. Designed to comply with NIH Genomic Data Sharing (GDS) Policy for controlled-access genomic data (dbGaP). Platform uses 20 HPE ProLiant nodes (1120 vCPU cores, 256GB RAM each), 1.5PB Ceph storage. Key security features: two-factor authentication (Duo), full-disk encryption, tiered storage (block volumes, S3 cache, persistent S3, general archive), and strict firewall rules (ingress only between VMs in same project).

## Analysis

Most relevant for the TCC is the explicit description of how Cloud-Init integrates with OpenStack's Nova metadata server. The paper provides a real-world example of the metadata service pattern in a private cloud: Cloud-Init pulls metadata from Nova at first boot (IP, SSH keys), and vendor-data is used to inject security update policies. This is exactly the pattern your TCC implements for Incus — but OpenStack has this built-in via Nova, while Incus does not.

## Key Findings

- "Cloud-Init pulls metadata from the OpenStack Nova metadata server and configures settings at first boot like IP address, user keys, etc."
- Cloud-Init's vendor-data mechanism forces security updates on all MSI-blessed VMs
- Three components for VM security: DiskImage Builder (image creation) + Cloud-Init (first-boot metadata) + Puppet (ongoing configuration)
- VMs run on QEMU+KVM with 2x CPU and 1x memory oversubscription
- Neutron BGP dynamic routing for campus-routable IP addresses
- HPL benchmarks: ~5% virtualization overhead (87% peak performance vs 91% bare metal)
- Subscription model: $1000/year for 16 vCPUs, 2TB volume storage, dbGaP cache access

## Pertinent Information for TCC

**Directly relevant — real-world example of metadata service in private cloud:**

- Demonstrates the three-layer initialization model: image builder -> cloud-init metadata service -> configuration management. Your TCC implements the middle layer for Incus
- The Nova metadata server is the direct equivalent of what you're building. OpenStack has it built-in; Incus does not — this is the gap your TCC fills
- Vendor-data usage for enforcing security policies is a feature your service already supports
- Can cite to show how established private clouds (OpenStack) rely on metadata services, reinforcing that Incus needs one too
- The paper's security model (firewall rules restricting metadata access to same-project VMs) provides design guidance for your service's security considerations

## Suggested BibTeX

```bibtex
@inproceedings{bollig2018leveraging,
  title={Leveraging OpenStack and Ceph for a Controlled-Access Data Cloud},
  author={Bollig, Evan F. and Allan, Graham T. and Lynch, Benjamin J. and Huerta, Yectli A. and Mix, Mathew and Munsell, Edward A. and Benson, Raychel M. and Swartz, Brent},
  booktitle={Proceedings of the Practice and Experience on Advanced Research Computing (PEARC '18)},
  year={2018},
  publisher={ACM},
  address={Pittsburgh, PA, USA},
  doi={10.1145/3219104.3219165}
}
```
