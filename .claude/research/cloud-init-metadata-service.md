# Research: Cloud-init Metadata Service for Incus

**Date:** 2026-03-13
**Purpose:** Academic TCC (undergraduate thesis) — building a cloud-init metadata service for Incus

---

## Summary

Cloud-init is the industry-standard tool for cloud instance initialization, originally created at Canonical in 2009–2010 by Scott Moser. It operates by fetching configuration data from a "datasource" — which can be a local file, a removable drive, or an HTTP metadata server. The canonical HTTP endpoint for cloud metadata services is `http://169.254.169.254`, a link-local address (RFC 3927) used by AWS, OpenStack, GCE, Azure, and others. Incus (a 2023 community fork of LXD) currently delivers cloud-init data via a Unix socket (`/dev/incus/sock`), but does not expose a standard HTTP metadata service. This gap is the motivation for the TCC project.

---

## 1. Cloud-init: Overview and Boot Stages

### 1.1 Purpose

Cloud-init automates the first-boot configuration of cloud instances, enabling fleet provisioning without manual intervention. It handles:

- Network configuration (hostname, DNS, IP addresses)
- User account creation and SSH key injection
- Package installation and updates
- Arbitrary script execution
- Configuration management integration (Puppet, Chef, Ansible)

**Official documentation:** <https://docs.cloud-init.io/en/latest/explanation/introduction.html>

### 1.2 Boot Stages

Cloud-init runs in four ordered stages:

| Stage | Name | Description |
|-------|------|-------------|
| 1 | `init-local` | Runs before networking. Identifies datasource from local sources (disk, DMI, kernel cmdline). |
| 2 | `init` (network) | Runs after network is up. Fetches metadata, user-data, vendor-data from the datasource. |
| 3 | `modules-config` | Applies configuration modules (users, packages, etc.). |
| 4 | `modules-final` | Runs final scripts (runcmd, final_message, etc.). |

The datasource detection tool `ds-identify` runs early in stage 1 to determine which datasource to use, using DMI data, kernel command line, and filesystem probing.

### 1.3 Data Types Fetched from Datasources

| Type | Description |
|------|-------------|
| `meta-data` | Machine identity: instance-id, hostname, public-keys, network config |
| `user-data` | User-supplied config: cloud-config YAML, shell scripts, MIME multipart |
| `vendor-data` | Cloud-provider-supplied defaults (overrideable by user-data) |
| `network-config` | Network interface configuration (v1 or v2 YAML format) |

**Key constraint:** `instance-id` is the most critical metadata field. cloud-init uses it to determine whether the current boot is a "first boot" (triggering full initialization) or a subsequent boot (skipping already-applied modules).

### 1.4 User-Data Formats

cloud-init supports multiple user-data formats, detected by their first line:

| Format | Identifier | Description |
|--------|-----------|-------------|
| cloud-config | `#cloud-config` | YAML declarative configuration (primary format) |
| Shell script | `#!/bin/...` | Arbitrary shell script |
| Boothook | `#boothook` | Very early execution script |
| MIME multipart | `Content-Type: multipart/mixed` | Combines multiple formats |
| Include | `#include` | References other URLs for user-data |
| Jinja template | `## template: jinja` | Template with instance variable substitution |
| Gzip compressed | (binary) | Any of the above, compressed |

---

## 2. Datasource Types Relevant to This Work

### 2.1 NoCloud Datasource

The NoCloud datasource is the most flexible and is directly relevant to Incus. It fetches configuration files from a base URL or from a labeled filesystem (`CIDATA`).

**HTTP endpoints (when using seedfrom URL):**

```
{seedfrom}/meta-data        # YAML: must contain at minimum instance-id
{seedfrom}/user-data        # Any user-data format
{seedfrom}/vendor-data      # YAML (optional)
{seedfrom}/network-config   # Network config v1 or v2 YAML (optional)
```

**Minimum valid `meta-data` file:**

```yaml
instance-id: my-unique-id-here
local-hostname: myhost
```

**Supported protocols:** HTTP, HTTPS, FTP, FTPS

**Documentation:** <https://docs.cloud-init.io/en/latest/reference/datasources/nocloud.html>

### 2.2 EC2/AWS Datasource

The original and most widely emulated datasource. Uses versioned URL paths under `http://169.254.169.254`.

**Supported API versions:** `1.0`, `2007-01-19`, `2007-03-01`, `2007-08-29`, `2007-10-10`, `2007-12-15`, `2008-02-01`, `2008-09-01`, `2009-04-04` (baseline used by most implementations), through `2021-03-23`.

**Key endpoints:**

```
GET /                                    # Lists supported API versions
GET /latest/meta-data/                   # Lists available metadata keys
GET /latest/meta-data/instance-id        # Instance unique identifier
GET /latest/meta-data/hostname           # Instance hostname
GET /latest/meta-data/local-hostname     # Private hostname
GET /latest/meta-data/local-ipv4         # Private IPv4 address
GET /latest/meta-data/public-ipv4        # Public IPv4 address
GET /latest/meta-data/public-keys/       # SSH public keys list
GET /latest/meta-data/public-keys/0/openssh-key  # First SSH key
GET /latest/meta-data/placement/availability-zone
GET /latest/meta-data/ami-id
GET /latest/meta-data/instance-type
GET /latest/meta-data/security-groups
GET /latest/meta-data/iam/               # IAM roles/credentials
GET /latest/meta-data/tags/instance      # EC2 instance tags (if enabled)
GET /latest/user-data                    # User-supplied data
GET /latest/dynamic/instance-identity/document  # JSON instance identity
```

**IMDSv2 (session-oriented):**

```
PUT /latest/api/token                    # Create session token (TTL in header)
  Header: X-aws-ec2-metadata-token-ttl-seconds: 21600
GET /latest/meta-data/...               # Use token in header
  Header: X-aws-ec2-metadata-token: {token}
```

**Documentation:** <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html>
**cloud-init EC2 datasource:** <https://docs.cloud-init.io/en/latest/reference/datasources/ec2.html>

### 2.3 OpenStack/Config Drive Datasource

OpenStack implements both its own native API and an EC2-compatible API.

**OpenStack native API:**

```
GET /openstack                           # Lists API versions
GET /openstack/latest/meta_data.json     # Nova instance metadata (JSON)
GET /openstack/latest/user_data          # User-supplied data
GET /openstack/latest/network_data.json  # Neutron network config (JSON)
GET /openstack/latest/vendor_data.json   # Provider data (JSON)
GET /openstack/latest/vendor_data2.json  # Alternative vendor data
```

**EC2-compatible API (also served by OpenStack):**

```
GET /2009-04-04/meta-data/
GET /2009-04-04/user-data
```

**Versioned paths for OpenStack native API:** `2012-08-10`, `2013-04-04`, `2015-10-15`, `2016-06-30`, `2016-10-06`, `2017-02-22`, `2018-08-27`, `latest`

**Architecture note:** OpenStack uses a "metadata proxy" in Neutron network namespaces, which intercepts requests to `169.254.169.254` from VMs and adds instance-identification headers (`X-Instance-ID`, `X-Tenant-ID`) before forwarding to the Nova metadata API. This solves the ambiguity of overlapping tenant IP addresses.

**Documentation:**

- <https://docs.openstack.org/nova/latest/user/metadata.html>
- <https://docs.cloud-init.io/en/latest/reference/datasources/openstack.html>

### 2.4 Google Compute Engine Datasource

GCE uses a distinct base URL and requires a mandatory header for all requests.

**Base URL:** `http://metadata.google.internal/computeMetadata/v1/`

**Required header:** `Metadata-Flavor: Google` (all requests without this header are rejected)

**Key paths:**

```
GET /computeMetadata/v1/instance/         # Instance metadata root
GET /computeMetadata/v1/instance/id       # Unique numeric instance ID
GET /computeMetadata/v1/instance/name     # Instance name
GET /computeMetadata/v1/instance/hostname # Instance FQDN
GET /computeMetadata/v1/instance/zone     # Zone (full resource path)
GET /computeMetadata/v1/instance/attributes/ssh-keys   # SSH public keys
GET /computeMetadata/v1/instance/attributes/user-data  # User data
GET /computeMetadata/v1/project/          # Project-level metadata
```

**HTTPS option:** GCE also provides an HTTPS endpoint using client certificates for additional security.

**Documentation:**

- <https://cloud.google.com/compute/docs/metadata/overview>
- <https://docs.cloud-init.io/en/latest/reference/datasources/gce.html>

### 2.5 LXD Datasource (and its relationship with Incus)

The LXD datasource is the current mechanism cloud-init uses when running inside LXD/Incus instances.

**Key distinction:** It does NOT use HTTP over the network. Instead, it communicates via a Unix domain socket.

**Socket path:** `/dev/lxd/sock` (LXD) or `/dev/incus/sock` (Incus)

**How it works:** cloud-init makes HTTP GET requests over the Unix socket to retrieve configuration data. The socket exposes instance configuration as versioned HTTP routes.

**Configuration keys (Incus):**

```
cloud-init.user-data    # Highest priority: user configuration
cloud-init.vendor-data  # Cloud-operator defaults
cloud-init.network-config  # Network v1 or v2 YAML
user.<any-key>          # Custom key-value pairs
```

**Incus-specific issue:** A timing problem exists where `ds-identify` and `cloud-init-local` run before `/dev/incus/sock` becomes available (since `incus-agent` starts late in the boot). This is the primary motivation for building an HTTP-based metadata service as an alternative approach.

**Proposed solution (from Stéphane Graber):** A native Incus datasource for cloud-init (shim around the LXD one) that falls back to NoCloud if the socket is unavailable.

**Documentation:**

- <https://docs.cloud-init.io/en/latest/reference/datasources/lxd.html>
- <https://linuxcontainers.org/incus/docs/main/cloud-init/>
- <https://discuss.linuxcontainers.org/t/incus-cloud-init-lxd-datasource/19574>

---

## 3. The 169.254.169.254 Address

The IP address `169.254.169.254` is from the IPv4 link-local range `169.254.0.0/16`, reserved by:

- **RFC 3927** — "Dynamic Configuration of IPv4 Link-Local Addresses" (May 2005)
- **RFC 6890** — "Special-Purpose IP Address Registries" (April 2013)

Link-local addresses are non-routable (scoped to a single network segment), making them ideal for host-to-hypervisor communication inside cloud infrastructure. Cloud providers configure routing rules so that requests to `169.254.169.254` from VMs are intercepted by the hypervisor or a network namespace proxy.

This address is used by: AWS, OpenStack, GCE (via `metadata.google.internal` which resolves to it), Azure, DigitalOcean, CloudStack, and nearly all IaaS providers.

**References:**

- <https://www.baeldung.com/linux/cloud-ip-meaning>
- <https://www.middlewareinventory.com/blog/link-local-ip-address-aws-imds-169-254-169-254/>

---

## 4. Network Configuration Formats

### 4.1 Networking Config Version 1

A cloud-init-specific format using a list of configuration types under the `network` key:

```yaml
version: 1
config:
  - type: physical
    name: eth0
    mac_address: "00:11:22:33:44:55"
    subnets:
      - type: dhcp4
  - type: nameserver
    address: [8.8.8.8, 8.8.4.4]
    search: [example.com]
```

Device types: `physical`, `bond`, `bridge`, `vlan`, `nameserver`, `route`

**Documentation:** <https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v1.html>

### 4.2 Networking Config Version 2

Based on the Netplan format (also used by Ubuntu's network configuration):

```yaml
version: 2
ethernets:
  eth0:
    dhcp4: true
    match:
      macaddress: "00:11:22:33:44:55"
    set-name: eth0
```

Supports networkd and NetworkManager backends.

**Documentation:** <https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v2.html>

---

## 5. Security: Metadata Service Vulnerabilities and IMDSv2

### 5.1 SSRF Attack Vector

Server-Side Request Forgery (SSRF) against metadata services is a critical cloud security issue. An attacker who can make a web application issue arbitrary HTTP requests can:

1. Send a request to `http://169.254.169.254/latest/meta-data/iam/security-credentials/`
2. Retrieve temporary IAM credentials from the metadata response
3. Use those credentials to access cloud resources

**Capital One breach (2019):** The most prominent real-world case. A SSRF vulnerability allowed retrieval of IAM credentials from the metadata service, exposing data of 100+ million customers.

**MITRE ATT&CK:** Technique T1552.005 — "Unsecured Credentials: Cloud Instance Metadata API"
URL: <https://attack.mitre.org/techniques/T1552/005/>

### 5.2 IMDSv2: Session-Oriented Token Authentication (AWS, 2019)

AWS introduced IMDSv2 in November 2019 to defend against SSRF and related attacks.

**How it works (two-step protocol):**

```
# Step 1: Create session token (PUT request)
PUT /latest/api/token HTTP/1.1
X-aws-ec2-metadata-token-ttl-seconds: 21600

# Response: token string (e.g., "TOKEN_VALUE")

# Step 2: Use token in metadata requests
GET /latest/meta-data/instance-id HTTP/1.1
X-aws-ec2-metadata-token: TOKEN_VALUE
```

**Security properties of IMDSv2:**

1. Requires a PUT request (most SSRF vulnerabilities only enable GET)
2. Requires a custom request header (`X-aws-ec2-metadata-token-ttl-seconds`) in the PUT — SSRF cannot set custom headers
3. IP-level hop limit of 1 on PUT responses — prevents token forwarding through network hops
4. `X-Forwarded-For` in the PUT request causes rejection — blocks proxy-based attacks
5. Token is instance-specific — cannot be used from other instances

**Threats mitigated:** Open firewalls, open reverse proxies, SSRF vulnerabilities, misconfigured layer-3 NAT

**Other provider header-based defenses:**

- **Azure IMDS:** Requires header `Metadata: true` on all requests
- **GCE IMDS:** Requires header `Metadata-Flavor: Google` on all requests
- Both of these mitigate simple SSRF (which cannot set custom headers) but are weaker than IMDSv2's session token

**References:**

- <https://aws.amazon.com/blogs/security/defense-in-depth-open-firewalls-reverse-proxies-ssrf-vulnerabilities-ec2-instance-metadata-service/>
- <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html>
- <https://d1.awsstatic.com/events/reinvent/2019/Security_best_practices_for_the_Amazon_EC2_instance_metadata_service_SEC310.pdf>

---

## 6. Incus: Background and Architecture

### 6.1 History

- **LXD** was created at Canonical as a system container/VM manager built on top of LXC.
- In **July 2023**, Canonical announced it was withdrawing LXD from the Linux Containers community project to develop it in-house.
- In **August 2023**, Aleksa Sarai forked LXD as **Incus**, with backing from the original LXD core team: Stéphane Graber (former LXD lead), Christian Brauner, Serge Hallyn, and Tycho Andersen.
- The fork was adopted under the Linux Containers umbrella.
- First stable release: **Incus 0.1** (October 2023).
- **Goal:** Provide a fully community-led alternative to Canonical's LXD, licensed under Apache 2.0, with no CLA requirement.

**Announcement:** <https://linuxcontainers.org/incus/announcement/>
**GitHub:** <https://github.com/lxc/incus>

### 6.2 Incus Features Relevant to Cloud-init

Incus manages both **system containers** (lightweight, share host kernel) and **virtual machines** (full isolation via QEMU).

For VMs, Incus provides:

- `incus-agent` — a guest agent running inside the VM
- `/dev/incus/sock` — a Unix socket exposing instance metadata and config
- `security.guestapi` — enables/disables the guest API socket

**Current cloud-init integration:** Via the LXD datasource, which reads from `/dev/incus/sock`.

**Problem:** The `incus-agent` starts too late in the VM boot sequence, causing a race condition with `cloud-init-local` and `ds-identify`, which run before the socket is available.

**References:**

- <https://linuxcontainers.org/incus/docs/main/cloud-init/>
- <https://discuss.linuxcontainers.org/t/incus-cloud-init-lxd-datasource/19574>

---

## 7. Bibliography and Academic References

### 7.1 Primary Technical Documentation

| # | Title | Author/Publisher | Year | URL | Relevance |
|---|-------|-----------------|------|-----|-----------|
| 1 | cloud-init Documentation | Canonical / cloud-init contributors | 2009–present | <https://docs.cloud-init.io/en/latest/> | Authoritative reference for datasource API, boot stages, user-data formats |
| 2 | EC2 Instance Metadata Service documentation | Amazon Web Services | 2006–present | <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html> | Defines the de-facto standard metadata API paths and IMDSv1/v2 specification |
| 3 | OpenStack Nova Metadata Service | OpenStack Foundation | 2012–present | <https://docs.openstack.org/nova/latest/user/metadata.html> | Documents OpenStack metadata API, EC2-compat paths, proxy architecture |
| 4 | Google Compute Engine Metadata | Google | 2012–present | <https://cloud.google.com/compute/docs/metadata/overview> | GCE metadata architecture, Metadata-Flavor header requirement |
| 5 | NoCloud Datasource | cloud-init contributors | ongoing | <https://docs.cloud-init.io/en/latest/reference/datasources/nocloud.html> | HTTP endpoints for NoCloud (closest to the implemented service) |
| 6 | LXD Datasource | cloud-init contributors | ongoing | <https://docs.cloud-init.io/en/latest/reference/datasources/lxd.html> | Current LXD/Incus integration mechanism |
| 7 | Incus Documentation: cloud-init | Linux Containers | 2023–present | <https://linuxcontainers.org/incus/docs/main/cloud-init/> | Incus cloud-init configuration options and current limitations |
| 8 | Networking Config Version 1 | cloud-init contributors | ongoing | <https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v1.html> | Network config format specification |
| 9 | Networking Config Version 2 | cloud-init contributors | ongoing | <https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v2.html> | Network config format specification (Netplan-based) |
| 10 | Vendor-Data | cloud-init contributors | ongoing | <https://docs.cloud-init.io/en/latest/explanation/vendordata.html> | Vendor-data specification and use cases |

### 7.2 Security References

| # | Title | Author/Publisher | Year | URL | Relevance |
|---|-------|-----------------|------|-----|-----------|
| 11 | Add defense in depth against open firewalls, reverse proxies, and SSRF vulnerabilities with enhancements to the EC2 Instance Metadata Service | AWS Security Blog | 2019 | <https://aws.amazon.com/blogs/security/defense-in-depth-open-firewalls-reverse-proxies-ssrf-vulnerabilities-ec2-instance-metadata-service/> | Introduces IMDSv2 and explains the SSRF threat model |
| 12 | Security best practices for the Amazon EC2 instance metadata service (SEC310) | AWS re:Invent | 2019 | <https://d1.awsstatic.com/events/reinvent/2019/Security_best_practices_for_the_Amazon_EC2_instance_metadata_service_SEC310.pdf> | AWS re:Invent presentation on IMDSv2 security model |
| 13 | T1552.005: Unsecured Credentials — Cloud Instance Metadata API | MITRE ATT&CK | 2021 | <https://attack.mitre.org/techniques/T1552/005/> | Authoritative taxonomy of metadata API exploitation techniques |
| 14 | Get the full benefits of IMDSv2 and disable IMDSv1 across your AWS infrastructure | AWS Security Blog | 2021 | <https://aws.amazon.com/blogs/security/get-the-full-benefits-of-imdsv2-and-disable-imdsv1-across-your-aws-infrastructure/> | IMDSv2 migration guidance and security rationale |
| 15 | Understanding Azure IMDS Risks and Protections | Mohsen Akhavan | 2024 | <https://mohsenakhavan.com/understanding-azure-imds-instance-metadata-service-risks-and-protections/> | Azure IMDS security model (Metadata: true header) |

### 7.3 Incus/LXD References

| # | Title | Author/Publisher | Year | URL | Relevance |
|---|-------|-----------------|------|-----|-----------|
| 16 | Incus: A new fork of Canonical's LXD 'containervisor' | The Register / Liam Proven | 2023 | <https://www.theregister.com/2023/08/04/incus_lxd_fork/> | News report on Incus fork history |
| 17 | Linux Containers Forks LXD Project As "Incus" | Phoronix / Michael Larabel | 2023 | <https://www.phoronix.com/news/Linux-Containers-LXD-Incus> | Fork announcement coverage |
| 18 | Incus Announcement | Linux Containers / Aleksa Sarai | August 2023 | <https://linuxcontainers.org/incus/announcement/> | Official announcement of the Incus fork |
| 19 | Incus cloud-init & LXD datasource (forum discussion) | Linux Containers Forum / Stéphane Graber et al. | 2024 | <https://discuss.linuxcontainers.org/t/incus-cloud-init-lxd-datasource/19574> | Community discussion on the need for a native Incus datasource and HTTP metadata service |
| 20 | GitHub: lxc/incus | Linux Containers | 2023–present | <https://github.com/lxc/incus> | Incus source code and issue tracker |

### 7.4 Infrastructure as Code and Automated Provisioning

| # | Title | Author/Publisher | Year | URL | Relevance |
|---|-------|-----------------|------|-----|-----------|
| 21 | Provisioning the Last Mile with Cloud-Init (Chapter 5) | Kief Morris — O'Reilly "Infrastructure as Code Cookbook" | 2017 | <https://www.oreilly.com/library/view/infrastructure-as-code/9781786464910/ch05.html> | Book chapter on cloud-init as an IaC provisioning tool |
| 22 | End-to-End Automation in Cloud Infrastructure Provisioning | ResearchGate / Academia.edu | 2017 | <https://www.researchgate.net/publication/318040301_End-to-End_Automation_in_Cloud_Infrastructure_Provisioning> | Academic paper on automated cloud provisioning pipelines |
| 23 | Static Analysis of Infrastructure as Code: a Survey | arXiv | 2022 | <https://arxiv.org/pdf/2206.10344> | Survey of IaC practices and tools; provides academic context for automated provisioning |
| 24 | A Survey on Infrastructure-as-Code Solutions for Cloud Development | ResearchGate | 2022 | <https://www.researchgate.net/publication/366980891_A_Survey_on_Infrastructure-as-Code_Solutions_for_Cloud_Development> | Academic survey of IaC tools and approaches |
| 25 | Infrastructure as Code (IaC) and Its Role in Achieving DevOps Goals | ResearchGate | 2024 | <https://www.researchgate.net/publication/379075919_Infrastructure_as_Code_IaC_and_Its_Role_in_Achieving_DevOps_Goals> | IaC role in DevOps; context for automated instance provisioning |

### 7.5 Cloud Computing Fundamentals

| # | Title | Author/Publisher | Year | URL | Relevance |
|---|-------|-----------------|------|-----|-----------|
| 26 | The history of cloud-init and virtualization in Ubuntu | Soren Hansen (Medium) | 2021 | <https://sorenisanerd.medium.com/the-history-of-cloud-init-and-virtualization-in-ubuntu-604d33c3275c> | Historical context: cloud-init created at Canonical 2009–2010 by Scott Moser |
| 27 | RFC 3927: Dynamic Configuration of IPv4 Link-Local Addresses | IETF / Stuart Cheshire, Bernard Aboba, Erik Guttman | 2005 | <https://www.rfc-editor.org/rfc/rfc3927> | Defines the 169.254.0.0/16 address range used by all metadata services |
| 28 | Security Infrastructure for On-demand Provisioned Cloud Infrastructure Services | IEEE Cloud Computing | 2012 | <https://ieeexplore.ieee.org/document/6133151/> | Academic paper on security services in on-demand cloud provisioning |
| 29 | On-demand provisioning of workflow middleware and services into the cloud | ACM/Springer Computing | 2016 | <https://dl.acm.org/doi/abs/10.1007/s00607-016-0521-x> | Academic paper on on-demand cloud service provisioning |
| 30 | OpenStack Metadata Service: Metadata and cloud-init (OpenStack Summit talk) | OpenStack Foundation | 2017 | <https://www.openstack.org/videos/sydney-2017/metadata-and-cloud-init> | Conference talk on OpenStack metadata service and cloud-init integration |

### 7.6 Additional Technical References

| # | Title | Author/Publisher | Year | URL | Relevance |
|---|-------|-----------------|------|-----|-----------|
| 31 | OpenStack Metadata API and OVN | OpenStack Neutron contributors | ongoing | <https://docs.openstack.org/neutron/latest/contributor/internals/ovn/metadata_api.html> | How Neutron proxies metadata requests; relevant to proxy architecture patterns |
| 32 | Configuring and Managing cloud-init for RHEL 9 | Red Hat | 2022–present | <https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/configuring_and_managing_cloud-init_for_rhel_9/> | Enterprise Linux cloud-init configuration reference |
| 33 | About VM metadata (GCE) | Google Cloud | ongoing | <https://cloud.google.com/compute/docs/metadata/overview> | Complete GCE metadata service reference |

---

## 8. Feature Matrix: Metadata Endpoints by Provider

| Endpoint / Feature | AWS EC2 | OpenStack | GCE | NoCloud HTTP | Incus (socket) |
|-------------------|---------|-----------|-----|--------------|----------------|
| Base address | 169.254.169.254 | 169.254.169.254 | metadata.google.internal | configurable | /dev/incus/sock |
| Versioned paths | Yes (`/2009-04-04/`, `/latest/`) | Yes (`/2012-08-10/`, `/latest/`) | Yes (`/v1/`) | No | N/A |
| `instance-id` | `/meta-data/instance-id` | `meta_data.json`.uuid | `/instance/id` | `meta-data` file | via config key |
| `hostname` | `/meta-data/hostname` | `meta_data.json`.hostname | `/instance/hostname` | `meta-data` file | auto-generated |
| `user-data` | `/user-data` | `/user_data` | `/instance/attributes/user-data` | `user-data` file | cloud-init.user-data |
| `vendor-data` | N/A | `/vendor_data.json` | N/A | `vendor-data` file | cloud-init.vendor-data |
| `network-config` | N/A (uses DHCP) | `network_data.json` | N/A | `network-config` file | cloud-init.network-config |
| SSH keys | `/public-keys/` | `meta_data.json`.public_keys | `/instance/attributes/ssh-keys` | `meta-data` file | (user accounts module) |
| Token auth | IMDSv2 (PUT token) | Shared secret (proxy) | `Metadata-Flavor: Google` | None | Unix socket (implicit) |
| Instance identity doc | `/dynamic/instance-identity/document` | No | No | No | No |

---

## 9. Key Design Decisions for the TCC Implementation

Based on the research, a cloud-init metadata service for Incus should:

1. **Serve HTTP on 169.254.169.254:80** — the universally expected address; requires a routing rule inside instances pointing this address to the host.

2. **Support the EC2 API paths** — the most widely compatible format; used by cloud-init's EC2 datasource (or NoCloud with `seedfrom`). Minimum: `/latest/meta-data/instance-id`, `/latest/meta-data/hostname`, `/latest/user-data`.

3. **Support the NoCloud HTTP format** — simpler alternative; paths at `/meta-data`, `/user-data`, `/vendor-data`, `/network-config` relative to a configurable base URL.

4. **Include `instance-id` from Incus** — must be stable and unique per instance; cloud-init uses it to detect first boot. The Incus instance UUID is an ideal source.

5. **Serve `user-data` as raw text** — cloud-init reads it as-is; the service should not parse or modify it. It may be cloud-config YAML, a shell script, or MIME multipart.

6. **Serve `network-config` for network setup** — allows cloud-init to configure static IPs, bonds, VLANs, and DNS inside the instance.

7. **Serve `vendor-data`** — allows the Incus operator to inject default configuration across all instances via profiles.

8. **Populate SSH public keys** — via `/latest/meta-data/public-keys/` (EC2 format) or in `meta-data.public_keys` (OpenStack format).

9. **Consider security** — for a private infrastructure service (no internet exposure), token auth may be optional, but the design should document the risk. IMDSv2-style token auth is the best practice.

10. **Instance identification** — the service must identify which Incus instance is making the request. Options: source IP lookup (unreliable with NAT), per-instance routing (each instance gets a unique IP), or routing via a virtual network interface with instance metadata in the request path.

---

## Sources Index

- <https://docs.cloud-init.io/en/latest/> — cloud-init official documentation
- <https://docs.cloud-init.io/en/latest/reference/datasources/nocloud.html> — NoCloud datasource
- <https://docs.cloud-init.io/en/latest/reference/datasources/lxd.html> — LXD datasource
- <https://docs.cloud-init.io/en/latest/reference/datasources/ec2.html> — EC2 datasource
- <https://docs.cloud-init.io/en/latest/reference/datasources/gce.html> — GCE datasource
- <https://docs.cloud-init.io/en/latest/reference/datasources/openstack.html> — OpenStack datasource
- <https://docs.cloud-init.io/en/latest/explanation/vendordata.html> — Vendor-data
- <https://docs.cloud-init.io/en/latest/explanation/format.html> — User-data formats
- <https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v1.html> — Network config v1
- <https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v2.html> — Network config v2
- <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html> — AWS IMDS
- <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html> — IMDSv2
- <https://aws.amazon.com/blogs/security/defense-in-depth-open-firewalls-reverse-proxies-ssrf-vulnerabilities-ec2-instance-metadata-service/> — IMDSv2 security blog
- <https://docs.openstack.org/nova/latest/user/metadata.html> — OpenStack Nova metadata
- <https://docs.openstack.org/nova/latest/admin/metadata-service.html> — OpenStack metadata admin
- <https://docs.openstack.org/neutron/latest/contributor/internals/ovn/metadata_api.html> — Neutron metadata proxy
- <https://cloud.google.com/compute/docs/metadata/overview> — GCE metadata
- <https://linuxcontainers.org/incus/docs/main/cloud-init/> — Incus cloud-init docs
- <https://linuxcontainers.org/incus/announcement/> — Incus announcement
- <https://github.com/lxc/incus> — Incus source code
- <https://discuss.linuxcontainers.org/t/incus-cloud-init-lxd-datasource/19574> — Incus cloud-init forum discussion
- <https://attack.mitre.org/techniques/T1552/005/> — MITRE ATT&CK T1552.005
- <https://www.rfc-editor.org/rfc/rfc3927> — RFC 3927 (link-local addresses)
- <https://sorenisanerd.medium.com/the-history-of-cloud-init-and-virtualization-in-ubuntu-604d33c3275c> — cloud-init history
- <https://arxiv.org/pdf/2206.10344> — Static Analysis of IaC survey
- <https://www.researchgate.net/publication/366980891_A_Survey_on_Infrastructure-as-Code_Solutions_for_Cloud_Development> — IaC survey
- <https://www.researchgate.net/publication/318040301_End-to-End_Automation_in_Cloud_Infrastructure_Provisioning> — Cloud provisioning paper
- <https://www.researchgate.net/publication/379075919_Infrastructure_as_Code_IaC_and_Its_Role_in_Achieving_DevOps_Goals> — IaC and DevOps paper
- <https://ieeexplore.ieee.org/document/6133151/> — IEEE: Security in on-demand cloud provisioning
- <https://www.openstack.org/videos/sydney-2017/metadata-and-cloud-init> — OpenStack Summit talk
