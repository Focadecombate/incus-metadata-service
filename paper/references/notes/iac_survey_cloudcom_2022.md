# A Survey on Infrastructure-as-Code Solutions for Cloud Development

**Authors:** Teppan, H.; Fla, L. H.; Jaatun, M. G.
**Year:** 2022
**Source:** IEEE 13th International Conference on Cloud Computing Technology and Science (CloudCom 2022), Bangkok
**PDF:** `iac_survey_cloudcom_2022.pdf`

## Summary

Survey paper examining Infrastructure-as-Code (IaC) solutions for building a self-contained on-premise PaaS ("lab-in-a-box") for a Security Lab. Evaluates Cloud-Native Computing Foundation (CNCF) tools across categories: Provisioning (FCOS + Ignition), Platform (K3s), Orchestration (Kubernetes), App Definition (ArgoCD, Helm, Kustomize), and Observability (Prometheus). Implements a complete GitOps workflow where source code in a Git repository drives both CI (GitLab) and CD (ArgoCD) for Kubernetes deployments on a physical server.

## Analysis

The paper is directly relevant as a peer-reviewed survey of the current IaC landscape. Key observation: the Provisioning category lacks CNCF-mature tools — FCOS with Ignition was selected but acknowledged as insufficient for larger environments. This mirrors the gap your TCC addresses: Incus environments lack proper provisioning/initialization tooling comparable to what cloud providers offer. The paper focuses on Kubernetes-based orchestration, while your work targets the lower-level VM/container initialization layer that Incus is missing.

## Key Findings

- DevOps organizations with highest evolution can deploy changes in <1 hour with <5% failure rate
- GitOps (described 2018) uses Git as single source of truth for both app code and infrastructure
- Provisioning tools vs Configuration Management tools serve different roles: provisioning creates infrastructure; CM manages software on existing infrastructure
- FCOS + Ignition run configuration files once at first boot (similar to cloud-init)
- For small on-premise environments, K3s (40MB binary, 512MB RAM) + GitLab + ArgoCD is sufficient
- 4 of 9 tools discussed are CNCF Graduated (production-ready)

## Pertinent Information for TCC

- Validates the importance of first-boot configuration tools (Ignition, cloud-init) in the IaC ecosystem
- Highlights the provisioning gap for on-premise/private cloud environments — exactly the gap your TCC fills for Incus
- Can cite to contextualize your work within the broader IaC/DevOps landscape
- The GitOps workflow they describe (Git -> CI -> CD -> infrastructure) is the higher-level context in which your metadata service operates: cloud-init reads from your service to configure instances that are then part of such workflows

## Suggested BibTeX

```bibtex
@inproceedings{teppan2022survey,
  title={A Survey on Infrastructure-as-Code Solutions for Cloud Development},
  author={Teppan, H{\aa}kon and Fl{\aa}, Lars Halvdan and Jaatun, Martin Gilje},
  booktitle={2022 IEEE 13th International Conference on Cloud Computing Technology and Science (CloudCom)},
  year={2022},
  organization={IEEE},
  address={Bangkok, Thailand}
}
```
