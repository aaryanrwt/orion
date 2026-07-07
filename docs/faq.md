# Orion FAQ

Frequently asked questions regarding the design, scope, and engineering decisions of Orion.

---

## General Questions

### How does Orion differ from SSH loops?
SSH loops require writing scripts to manage collections of IP addresses, handling credentials/keys per machine, configuring hosts manually, and parsing unstructured output. Orion replaces this workflow with a single, fast binary containing native peer discovery, encrypted pairing, and aligned parallel log streaming.

### Is Orion a cloud platform?
No. Orion is a developer utility that runs directly on your local hardware. There are no hosted cloud dashboards, scheduling servers, or external databases involved.

---

## Technical Questions

### What ports does Orion use?
Orion uses ports `8910` and `8911` for peer communication, mutual authentication handshakes, and command output streaming.

### How is security handled?
Orion does not allow anonymous execution. Devices must explicitly pair (`orion join`) and register each other's public cryptographic keys before commands can be processed. All transport connections utilize Mutual TLS (mTLS) to enforce end-to-end encryption, authentication, and traffic integrity.

### Does Orion collect telemetry?
No. Telemetry collection is disabled by default. Orion does not collect execution logs, IP addresses, command histories, or system information.
