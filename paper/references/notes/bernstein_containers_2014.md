# Containers and Cloud: From LXC to Docker to Kubernetes

**Author:** Bernstein, David
**Year:** 2014
**Source:** IEEE Cloud Computing, vol. 1, no. 3, pp. 81-84
**Publisher:** IEEE

## Summary

Early survey paper tracing the evolution of container technology from LXC (Linux Containers) through Docker to Kubernetes. Explains how containers provide OS-level virtualization with near-native performance, lower overhead than VMs, and faster startup times. Discusses the shift from heavyweight VMs to lightweight containers for cloud workloads.

## Key Findings

- LXC provided the foundation: cgroups + namespaces for process isolation
- Docker added standardized packaging (images), layered filesystem, and developer experience
- Kubernetes emerged for container orchestration at scale
- Containers vs VMs: containers share kernel (faster, lighter) but less isolation
- System containers (LXC/LXD) vs application containers (Docker) serve different purposes

## Pertinent Information for TCC

- Provides historical context for containerization technology
- LXC -> LXD -> Incus lineage establishes the technology genealogy of your target platform
- The distinction between system containers (what Incus manages) and application containers (Docker) is relevant to explain why Incus needs cloud-init support

## BibTeX Key

`bernstein2014containers`
