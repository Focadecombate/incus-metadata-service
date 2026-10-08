# DevOps: Introducing Infrastructure-as-Code

**Authors:** Artač, Matej; Borovšak, Tadej; Di Nitto, Elisabetta; Guerriero, Michele; Tamburri, Damian A.
**Year:** 2017
**Source:** IEEE/ACM 39th International Conference on Software Engineering Companion (ICSE-C), pp. 497-498
**DOI:** 10.1109/ICSE-C.2017.162

## Summary

Short paper introducing Infrastructure-as-Code as a DevOps practice. Defines IaC as managing and provisioning infrastructure through machine-readable configuration files rather than manual processes. Discusses the benefits of version control, repeatability, and automated testing for infrastructure definitions.

## Key Findings

- IaC treats infrastructure definitions as software artifacts — versioned, tested, reviewed
- Enables DevOps practices: CI/CD for infrastructure, not just application code
- Key principle: infrastructure should be reproducible from code alone
- Reduces human error in environment setup and configuration

## Pertinent Information for TCC

- Cited alongside guerriero2019iac in the related work closing paragraph
- Provides the theoretical foundation for why automated provisioning (via cloud-init + metadata service) matters
- Your service is a building block in the IaC pipeline: cloud-init reads from it to configure instances declaratively

## BibTeX Key

`artac2017iac`
