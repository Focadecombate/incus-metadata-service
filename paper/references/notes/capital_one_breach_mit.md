# A Case Study of the Capital One Data Breach

**Authors:** Novaes Neto, G.; Madnick, S.; Moraes, A.; Borges, N.
**Year:** 2020
**Source:** MIT CISL Working Paper / Journal of Information Technology Teaching Cases
**PDF:** `capital_one_breach_mit_2020.pdf`

## Summary

Academic case study analyzing the 2019 Capital One data breach, where an attacker exploited a Server-Side Request Forgery (SSRF) vulnerability in a misconfigured Web Application Firewall (ModSecurity) running on AWS EC2. The SSRF allowed the attacker to query the AWS Instance Metadata Service (IMDS) at `169.254.169.254`, retrieving temporary IAM credentials from a role named "WAF-Role". With those credentials, the attacker accessed S3 buckets containing 106 million customer records (100M US, 6M Canada). The breach occurred on March 22-23, 2019 but was only detected on July 19, 2019. Capital One shares fell 15% and damages exceeded $150 million.

## Analysis

The paper uses the Three Lines of Defense governance framework to analyze the organizational failures that enabled the breach. It demonstrates that the technical vulnerability (IMDSv1's unauthenticated GET-based access) was compounded by governance failures in configuration management, monitoring, and incident response. The attack chain was: SSRF -> IMDS credential theft -> S3 data exfiltration. Notably, AWS introduced IMDSv2 with session-based tokens (PUT requests, hop-limit restrictions, mandatory headers) specifically in response to this class of attack.

## Key Findings

- IMDSv1 allowed any process on an EC2 instance to access credentials via simple GET to `169.254.169.254`
- SSRF + IMDSv1 = credential theft without needing code execution on the host
- IMDSv2 mitigations: session tokens via PUT, `X-aws-ec2-metadata-token-ttl-seconds` header, hop-limit of 1
- The attack required no sophisticated techniques — just exploitation of well-understood vulnerabilities
- Detection took ~4 months despite active logging

## Pertinent Information for TCC

**Directly relevant to your security section (Seguranca em Servicos de Metadados):**

- Provides a concrete, well-documented case of metadata service exploitation with real-world impact ($150M+)
- Your metadata service for Incus should consider the security lessons: IP-based identification alone is vulnerable if an attacker can make requests from the instance's IP
- The comparison between IMDSv1 (unauthenticated GET) and IMDSv2 (session tokens) maps directly to your security discussion
- Reinforces why your service's approach of identifying instances by source IP should include additional safeguards (e.g., not exposing sensitive credentials, network isolation)
- Can be cited in your comparative table to show real-world consequences of insecure metadata service design

## Suggested BibTeX

```bibtex
@article{novaes2020capital,
  title={A Case Study of the Capital One Data Breach},
  author={Novaes Neto, Germ{\'a}n and Madnick, Stuart and Moraes, Amador and Borges, Natasha},
  journal={MIT Sloan School Working Paper},
  number={2020-07},
  year={2020},
  institution={MIT Cybersecurity Interdisciplinary Systems Laboratory (CISL)}
}
```
