# Reference Notes Index

Quick-lookup index for all reference notes, organized by topic.

---

## Cloud-init Core

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [cloudinit_docs_2024.md](cloudinit_docs_2024.md) | `cloudinit2024` | Official docs — boot stages, datasources, user-data formats |
| [cloudinit_nocloud_2024.md](cloudinit_nocloud_2024.md) | `cloudinit_nocloud` | NoCloud datasource — HTTP endpoints our service implements |
| [cloudinit_lxd_ds_2024.md](cloudinit_lxd_ds_2024.md) | `cloudinit_lxd_ds` | LXD datasource — socket-based, timing problem |
| [cloudinit_ec2_ds_2024.md](cloudinit_ec2_ds_2024.md) | `cloudinit_ec2_ds` | EC2 datasource — versioned paths, most widely emulated |
| [cloudinit_netv2_2024.md](cloudinit_netv2_2024.md) | `cloudinit_netv2` | Network Config v2 — Netplan-based format we generate |

## Cloud Provider Metadata Services

| File | BibTeX Key | Provider |
|------|-----------|----------|
| [aws_imds_2024.md](aws_imds_2024.md) | `aws_imds` | AWS — de-facto standard API, 169.254.169.254 |
| [aws_imdsv2_2024.md](aws_imdsv2_2024.md) | `aws_imdsv2` | AWS — session-token auth (IMDSv2) |
| [aws_imdsv2_blog_2019.md](aws_imdsv2_blog_2019.md) | `aws_imdsv2_blog` | AWS — SSRF threat model and IMDSv2 security properties |
| [gcp_metadata_2024.md](gcp_metadata_2024.md) | `gcp_metadata` | GCP — Metadata-Flavor header, long polling |
| [openstack_metadata_2024.md](openstack_metadata_2024.md) | `openstack_metadata` | OpenStack — dual API (EC2-compat + native), Neutron proxy |
| [azure_imds_2024.md](azure_imds_2024.md) | `azure_imds` | Azure — Metadata: true header |

## Incus / LXD

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [incus_2024.md](incus_2024.md) | `incus2024` | Incus official docs — target platform |
| [stgraber_incus_2023.md](stgraber_incus_2023.md) | `stgraber2023` | Gräber's blog post announcing Incus fork |
| [incus_announcement_2023.md](incus_announcement_2023.md) | `incus_announcement` | Official Incus announcement by Aleksa Sarai |
| [incus_cloudinit_forum_2024.md](incus_cloudinit_forum_2024.md) | `incus_cloudinit_forum` | Forum discussion on socket timing problem — core motivation |
| [lxd_docs_2024.md](lxd_docs_2024.md) | `lxd2024` | LXD docs — predecessor architecture |
| [usp_cloud_lxd_2018.md](usp_cloud_lxd_2018.md) | `—` | USP thesis using LXD in private cloud (smart grids) |

## Security

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [mitre_t1552_005_2021.md](mitre_t1552_005_2021.md) | `mitre_t1552_005` | MITRE ATT&CK — metadata API credential theft taxonomy |
| [capital_one_breach_mit.md](capital_one_breach_mit.md) | `—` | MIT case study — SSRF + IMDSv1 → $150M breach |
| [aws_imdsv2_blog_2019.md](aws_imdsv2_blog_2019.md) | `aws_imdsv2_blog` | IMDSv2 security model (also in Cloud Providers above) |

## Distributed Consensus (RAFT)

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [ongaro_raft_2014.md](ongaro_raft_2014.md) | `ongaro2014raft` | RAFT paper — leader election, log replication, safety |
| [hashicorp_raft_2024.md](hashicorp_raft_2024.md) | `hashicorp_raft` | Go implementation used in our service |

## Implementation Technologies

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [gin_framework_2024.md](gin_framework_2024.md) | `gin2024` | Gin HTTP framework for Go |
| [sqlite_2024.md](sqlite_2024.md) | `sqlite2024` | SQLite embedded database |
| [opentelemetry_2024.md](opentelemetry_2024.md) | `opentelemetry2024` | OTel observability framework |
| [prometheus_2024.md](prometheus_2024.md) | `prometheus2024` | Prometheus metrics/monitoring |

## Cloud Computing Fundamentals

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [mell_nist_cloud_2011.md](mell_nist_cloud_2011.md) | `mell2011nist` | NIST cloud definition — 5 characteristics, 3 models |
| [bernstein_containers_2014.md](bernstein_containers_2014.md) | `bernstein2014containers` | Containers evolution: LXC → Docker → Kubernetes |
| [shi_edge_computing_2016.md](shi_edge_computing_2016.md) | `shi2016` | Edge computing vision — lightweight infra automation |
| [rfc3927_linklocal_2005.md](rfc3927_linklocal_2005.md) | `rfc3927` | RFC 3927 — 169.254.0.0/16 link-local addresses |

## Infrastructure as Code

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [guerriero_iac_2019.md](guerriero_iac_2019.md) | `guerriero2019iac` | IaC adoption study — 44 practitioners, 11 companies |
| [artac_iac_devops_2017.md](artac_iac_devops_2017.md) | `artac2017iac` | IaC as DevOps practice — definitions and principles |
| [iac_survey_cloudcom_2022.md](iac_survey_cloudcom_2022.md) | `—` | IaC survey — CNCF tools, GitOps, provisioning gap |
| [ufrgs_iac_provisioning_tcc.md](ufrgs_iac_provisioning_tcc.md) | `—` | UFRGS TCC on IaC provisioning |

## Brazilian Academic Works

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [usp_cloud_lxd_2018.md](usp_cloud_lxd_2018.md) | `—` | USP — LXD + OpenNebula for smart grids |
| [ifrn_cloud_comparison_tcc.md](ifrn_cloud_comparison_tcc.md) | `—` | IFRN — cloud platform comparison |
| [ufrgs_iac_provisioning_tcc.md](ufrgs_iac_provisioning_tcc.md) | `—` | UFRGS — IaC provisioning |
| [ufrj_edge_automation_2024.md](ufrj_edge_automation_2024.md) | `—` | UFRJ — edge infrastructure automation |

## Other

| File | BibTeX Key | Description |
|------|-----------|-------------|
| [cloudinitd_argonne_2011.md](cloudinitd_argonne_2011.md) | `—` | Argonne — cloudinit.d orchestration tool (2011) |
| [openstack_ceph_cloud_2018.md](openstack_ceph_cloud_2018.md) | `—` | OpenStack + Ceph private cloud |

---

## By Paper Section

**Introdução:** `cloudinit2024`, `aws_imds`, `rfc3927`, `incus2024`, `stgraber2023`, `openstack_metadata`, `shi2016`

**Trabalhos Relacionados — cloud-init:** `cloudinit2024`, `cloudinit_nocloud`, `cloudinit_lxd_ds`, `cloudinit_ec2_ds`

**Trabalhos Relacionados — Provedores:** `aws_imds`, `aws_imdsv2`, `gcp_metadata`, `openstack_metadata`, `azure_imds`

**Trabalhos Relacionados — Incus/LXD:** `incus2024`, `stgraber2023`, `lxd2024`, `incus_cloudinit_forum`, `cloudinit_lxd_ds`

**Trabalhos Relacionados — RAFT:** `ongaro2014raft`, `hashicorp_raft`

**Trabalhos Relacionados — Segurança:** `mitre_t1552_005`, `aws_imdsv2_blog`, `gcp_metadata`, `azure_imds`

**Trabalhos Relacionados — IaC:** `guerriero2019iac`, `artac2017iac`

**Proposta:** `incus2024`, `cloudinit2024`, `cloudinit_nocloud`, `cloudinit_netv2`, `gin2024`, `sqlite2024`, `ongaro2014raft`, `hashicorp_raft`, `opentelemetry2024`, `prometheus2024`, `aws_imds`, `gcp_metadata`

**Conclusão:** `rfc3927`, `aws_imdsv2`, `mitre_t1552_005`
