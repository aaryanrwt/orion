# Orion Getting Started Guide

Orion is a zero-configuration, secure, local-only developer compute mesh utility. It turns the computers you already own into a single, cohesive network of computing power.

---

## 1. Quick Start

### Step 1: Clone and Build
Build the Orion CLI binary from source:
```bash
git clone https://github.com/orion-infra/orion.git
cd orion
go build -o orion ./cmd/orion
```

### Step 2: Initialize Identity
Generate your local device identity keys and configuration files:
```bash
orion init
```
This generates:
* A 256-bit ECDSA P-256 private key and self-signed X.509 certificate.
* A unique, human-friendly device ID (e.g. `ORN-9AF2-81D7`).
* A local JSON database stored in your user configuration directory.

### Step 3: Discover Nearby Devices
List other Orion devices broadcasting on your local network segment:
```bash
orion connect
```
Use the arrow keys to navigate the interactive TUI. Select a discovered peer node and press **Enter** to issue a pairing request.

### Step 4: Accept Connection
On the target machine, run any Orion command (or run `orion respond` manually). The CLI will automatically intercept execution and prompt:
```
Incoming Connection Request
Device:       MacBook
ID:           ORN-8C12
Fingerprint:  8F:E1:12:A2

Accept? [Y] Accept [N] Reject >
```
Upon choosing **Accept**, trust is established, certificates are pinned, and communication is unlocked.

---

## 2. Basic Introspection

### List Trusted Devices
To see all paired nodes, their operating systems, online status, and connection transport type:
```bash
orion devices
```
For detailed metadata including pinned fingerprints and trust timestamps:
```bash
orion devices --verbose
```

### Profile Hardware
Review CPU, RAM, and GPU models of all online cluster nodes:
```bash
orion hardware
```

### List Discovered Models
Audit all downloaded Ollama models across your cluster:
```bash
orion models
```
