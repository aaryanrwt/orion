# Security Policy

Security is a core, non-negotiable tenet of Orion. We design for explicit user authorization and end-to-end encryption.

## Security Model

1. **Mutual Trust & Pairing**: Orion does not allow anonymous execution. A device can only execute commands on another device if they have completed the pairing flow (`orion join`) and explicitly trusted each other's cryptographic identity.
2. **Encrypted Communication**: All communication between nodes is secured using Mutual TLS (mTLS). This ensures:
   * Confidentiality (e.g. no eavesdropping on network traffic).
   * Authenticity (only explicitly paired devices can connect).
   * Integrity (messages cannot be tampered with in transit).
3. **No Centralized Storage**: Your identities, keys, and device lists are stored locally on your machines. Orion does not use cloud databases or centralized coordinates for command routing.
4. **No Telemetry**: We do not collect or transmit command history, node details, network addresses, or error logs to any external server.
5. **Least Privilege**: Executed commands run with the permissions of the user running the Orion background process. We strongly recommend running Orion under a non-root developer user account.

## Reporting a Vulnerability

If you discover a security vulnerability in Orion, please do not open a public GitHub issue. Instead, report it privately to our security team.

### Vulnerability Report Process

1. Email your report to **security@orion.sh** (in a real-world project, replace this with a secure contact address).
2. Include a detailed description of the vulnerability, including:
   * The components involved.
   * Steps to reproduce the issue (PoC).
   * Potential impact.
3. We will acknowledge receipt of your vulnerability report within 48 hours and work with you to coordinate a patch and public disclosure.
