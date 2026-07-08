# Security Policy

## Supported Versions

Only the latest release version is actively supported with security updates.

| Version | Supported |
| ------- | --------- |
| v0.1.x  | ✅ Yes    |
| < v0.1  | ❌ No     |

## Reporting a Vulnerability

We take the security of Orion seriously. If you believe you have found a security vulnerability in Orion, please do not report it via GitHub Issues. Instead, please report it responsibly by following these steps:

1. **Email us**: Send a detailed report to `security@orion-infra.org`.
2. **Details to include**:
   - Description of the vulnerability.
   - Step-by-step instructions to reproduce the vulnerability (proof of concept).
   - The impact of the vulnerability.
3. **PGP Encryption**: If you want to encrypt your email, please use our security team's PGP key (available upon request).

We will acknowledge receipt of your vulnerability report within 48 hours and send updates regarding our remediation progress.

## Scope of Security Guarantees

Orion guarantees:
- **Transport Security**: All peer-to-peer traffic on TCP port `8911` is encrypted under Mutual TLS (TLS 1.3).
- **Authentication**: Remote execution paths (`orion run`) require strict certificate verification against local pinned configurations. Unrecognized certificates are immediately rejected.
- **Offline Integrity**: Orion operates without communicating with any cloud orchestration layers, preventing data leakages to third-party endpoints.

Orion does **not** protect against:
- Compromises of the local machine's user configuration directory (e.g. root/administrator access revealing private key PEM blocks).
- Network routing layer configurations that explicitly proxy loopback endpoints to external hosts.
