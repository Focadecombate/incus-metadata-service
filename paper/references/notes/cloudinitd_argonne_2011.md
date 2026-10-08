# Managing Appliance Launches in Infrastructure Clouds

**Authors:** Bresnahan, J.; Freeman, T.; LaBissoniere, D.; Keahey, K.
**Year:** 2011
**Source:** TG'11 (TeraGrid Conference), ACM, Salt Lake City
**PDF:** `cloudinitd_argonne_2011.pdf`

## Summary

Introduces cloudinit.d, a tool developed at Argonne National Laboratory for launching, configuring, monitoring, and repairing sets of interdependent virtual machines across multiple IaaS clouds. The tool organizes deployments into "launch plans" with hierarchical "run levels" (analogous to UNIX init.d), where services at the same level launch in parallel and higher levels depend on lower ones. Each service has startup (bootpgm), health-check (readypgm), and termination scripts. Developed for the Ocean Observatory Initiative (OOI) to manage complex scientific applications spanning multiple FutureGrid clouds.

## Analysis

This is one of the earliest academic papers addressing the problem of automated, repeatable VM initialization in cloud environments. While cloud-init (the tool) handles single-instance configuration, cloudinit.d addresses multi-instance orchestration with dependency management. The paper identifies requirements that remain relevant: repeatable one-click deployment, coordination of interdependent launches, federated multi-cloud support, and policy-driven repair. Notably, the tool uses SSH-based bootstrapping (no pre-installed agent required) — a lightweight approach similar in philosophy to cloud-init's datasource model.

## Key Findings

- Key requirements: repeatable deployment, interdependent launch coordination, federated multi-cloud, testing/monitoring, policy-driven repair
- Architecture: Launch Plan -> Run Levels -> Services (each with host + scripts)
- Boot process: validate plan -> prestage VMs -> SSH in -> run bootpgm -> propagate dependency info upward
- Comparison: CloudFormation is AWS-only; Puppet/Chef require pre-installed agents; cloudinit.d only needs SSH (sshd)
- Services return dependency documents (JSON key/value) for consumption by higher run levels
- Repair works by restarting failed service + propagating health checks up the dependency chain

## Pertinent Information for TCC

- Early academic work on the same problem space as your TCC: automated initialization of cloud instances
- Your metadata service is complementary: cloudinit.d orchestrates multi-VM launches, while your service provides the per-instance metadata that cloud-init consumes during each VM's boot
- The paper's discussion of "contextualizing" VMs at boot time is exactly what metadata services do — your service enables this for Incus
- Can cite as historical context showing the evolution from SSH-based bootstrapping to metadata-service-based initialization
- The federated/multi-cloud aspect is a direction your "future work" could reference

## Suggested BibTeX

```bibtex
@inproceedings{bresnahan2011managing,
  title={Managing Appliance Launches in Infrastructure Clouds},
  author={Bresnahan, John and Freeman, Tim and LaBissoniere, David and Keahey, Kate},
  booktitle={Proceedings of the 2011 TeraGrid Conference: Extreme Digital Discovery},
  year={2011},
  publisher={ACM},
  address={Salt Lake City, UT, USA},
  doi={10.1145/2016741.2016762}
}
```
