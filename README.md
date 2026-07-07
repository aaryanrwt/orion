# Orion

> Run one command across every computer you own.

Orion is a minimal, secure, and blazing-fast developer utility that lets you execute commands in parallel across multiple machines. Think `ssh` + parallel execution, built with a modern developer experience.

No complex networking setup, no configuration files, and no daemon management. It just works.

```
                  ┌──────────────┐
                  │  Local Host  │
                  └──────┬───────┘
                         │
           ┌─────────────┴─────────────┐
           ▼                           ▼
   ┌───────────────┐           ┌───────────────┐
   │  vega (macOS) │           │ sirius (linux)│
   └───────────────┘           └───────────────┘
```

---

## Why Orion?

* **Zero Infrastructure**: No Kubernetes, no Docker Swarm, no server management.
* **Invisible Networking**: Handles peer discovery, mTLS certificate generation, and firewall traversal transparently.
* **Developer First**: Sub-50ms startup time, semantic colored outputs, and support-oriented diagnostics.
* **Secure by Default**: End-to-end encrypted connection using mutual TLS. No telemetry.

---

## Quick Start

### 1. Installation

Install the Orion CLI on your local and remote machines:

```bash
# macOS / Linux
curl -fsSL https://orion.sh/install.sh | sh

# Windows (PowerShell)
irm https://orion.sh/install.ps1 | iex
```

### 2. Initialize Orion

Initialize Orion on your primary machine:

```bash
$ orion init
✓ Orion initialized successfully.
  Your Device ID: vega-88e2-9b2f
```

### 3. Join a Remote Machine

Start the pairing process on your secondary machine (e.g. `sirius`):

```bash
$ orion join vega-88e2-9b2f
Searching nearby systems...
✓ Secure connection established with vega-88e2-9b2f
✓ Added device "sirius"
```

Verify your devices are connected:

```bash
$ orion devices
DEVICE ID          NAME      OS       STATUS    LATENCY
vega-88e2-9b2f     vega*     darwin   online    -
sirius-a39c-112f   sirius    linux    online    4.2ms
```

### 4. Run Your First Command

Run a command in parallel across all connected devices:

```bash
$ orion run uname -a
[vega]   Darwin vega 23.4.0 Darwin Kernel Version 23.4.0
[sirius] Linux sirius 6.5.0-27-generic #28-Ubuntu SMP
✓ Command completed successfully across 2 devices
```

---

## Command Reference

Orion ships with exactly 8 core commands:

* `orion init`: Initialize local machine credentials and configuration.
* `orion join <device-id>`: Securely pair and trust a new remote device.
* `orion devices`: List all paired devices and their statuses.
* `orion run <command>`: Execute a command across all paired devices in parallel.
* `orion status`: Display a high-level summary of your Orion network.
* `orion doctor`: Diagnose configuration, network, and device connectivity issues.
* `orion version`: Display Orion CLI version details.
* `orion help`: Comprehensive, context-aware command line assistance.

---

## Architecture (Simplified)

Orion utilizes a zero-config peer-to-peer overlay:
1. **Identity**: Every machine generates a unique cryptographic keypair on `orion init`.
2. **Discovery**: Machines discover each other on the local network (via multicast DNS) or through a secure, encrypted rendezvous server if separated by firewalls/NATs.
3. **Execution**: The initiator connects to remote peers, establishes mutual TLS (mTLS) verification, and securely streams command execution inputs/outputs.

---

## FAQ

#### Is Orion a cloud service?
No. Orion is a decentralized developer utility. Commands run directly on your own hardware. Your data never passes through third-party servers.

#### How is this different from SSH loops?
Instead of writing complex shell scripts to loop over `ssh` connections, manage SSH keys, configure hostnames, and handle asynchronous output parsing, Orion provides a single command (`orion run`) that does it all in parallel with beautiful, real-time output grouping.

---

## Troubleshooting

If you encounter issues, run:

```bash
$ orion doctor
```

For detailed troubleshooting steps, see [doctor troubleshooting docs](docs/philosophy.md).

---

## Contributing

Please review [CONTRIBUTING.md](CONTRIBUTING.md) for style guidelines, architecture details, and local development instructions.

## Security

Please read [SECURITY.md](SECURITY.md) to understand Orion's encryption, trust boundaries, and how to report vulnerabilities.

## License

Orion is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for more details.
