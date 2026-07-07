# Orion Roadmap

This roadmap details the planned engineering steps beyond the v0.1.0-beta release.

---

## Phase 1: Mesh Networking & Transport Layer
* **Secure Mutual TLS (mTLS)**: Implement standard Go TLS configurations utilizing self-signed certificates pinned to cryptographic keypairs generated on `orion init`.
* **Peer Discovery**: Integrate multicast DNS (mDNS) for local network discovery.
* **NAT Traversal**: Integrate rendezvous coordination for traversal across public internet firewalls.

## Phase 2: Command Executions & Security Hardening
* **Process Sandboxing**: Limit command execution scope to specific directories or non-privileged credentials.
* **Input Streaming**: Support piping standard input (`stdin`) to remote processes.
* **Terminal Resize Synchronization**: Propagate terminal window resize events (`SIGWINCH`) to remote command shells.

## Phase 3: Developer Utilities
* **Tagging & Device Groups**: Allow running commands on subsets of devices (e.g. `orion run --tags database uptime`).
* **File Copying**: Securely copy files between paired nodes using Orion's mTLS mesh.
