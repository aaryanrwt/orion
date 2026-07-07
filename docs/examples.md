# Orion Examples

This guide details typical execution workflows and command usage in Orion.

---

## 1. Quick Start Loop

Here is the quickest sequence to link two computers and run your first command:

### Initiator Machine (e.g., `macbook`)
Initialize Orion to obtain your unique Device ID:

```bash
$ orion init
✓ Orion initialized successfully.
  Your Device ID: macbook-a10c92fb
```

### Peer Machine (e.g., `sirius`)
Pair with the Macbook:

```bash
$ orion init
✓ Orion initialized successfully.
  Your Device ID: sirius-bb71d293

$ orion join macbook-a10c92fb
Searching nearby systems...
✓ Secure connection established with macbook-a10c92fb
✓ Added device "macbook"
```

---

## 2. Command Execution Patterns

### Stream system kernels across all machines
Ideal for verifying OS versions of connected nodes in your mesh:

```bash
$ orion run uname -a
[macbook]   Darwin macbook 23.4.0 Darwin Kernel Version 23.4.0 ...
[sirius]    Linux sirius 6.8.0-40-generic ...
✓ Command completed successfully across 2 devices
```

### Check hostnames
Verify standard node identifiers are configured correctly:

```bash
$ orion run hostname
[macbook]   macbook
[sirius]    sirius
✓ Command completed successfully across 2 devices
```

### Check current logged-in user
Ensure execution permissions are running under the correct local credentials:

```bash
$ orion run whoami
[macbook]   orion-user
[sirius]    orion-user
✓ Command completed successfully across 2 devices
```

---

## 3. Network Status Verification

Check mesh network stats dynamically:

```bash
$ orion status
Connected   2 devices (2 online)
Ready       Yes
```
