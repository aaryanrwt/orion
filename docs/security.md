# Orion Security Architecture

Orion guarantees **offline-first security by default**. It is designed to work in zero-trust environments (such as public Wi-Fi networks at hackathons or shared office networks) without depending on a centralized internet PKI or authority.

---

## 1. Cryptographic Identity
Every Orion node permanently owns its cryptographic identity:
* **Keypair**: ECDSA P-256 (standard NIST curve).
* **Certificate**: A self-signed X.509 certificate carrying the node's unique Device ID (e.g. `ORN-9AF2-81D7`) as its Common Name.
* **Fingerprint**: The first 4 bytes of the certificate's SHA-256 hash, formatted as a hexadecimal verifier (e.g. `8F:E1:12:A2`).

---

## 2. Mutual TLS & Certificate Pinning
All P2P control traffic between remote nodes is encrypted and validated using **Strict Mutual TLS (mTLS)**:
1. **Trust-on-First-Use (TOFU)**: During the pairing handshake, devices exchange X.509 certificates and fingerprints. The user must manually approve the incoming request on the target computer.
2. **Pinned Certificates**: Once paired, each device's full X.509 certificate PEM is permanently stored in the local `config.json`.
3. **Strict Validation**: During subsequent TLS connection handshakes, each node validates that the certificate presented by the remote client matches the pinned certificate in its database. Unknown or altered certificates are rejected.
4. **No CA Chain**: Since certificates are pinned directly, Orion requires no external Certificate Authorities (CAs) and operates fully offline.

---

## 3. Remote Execution Constraints
The remote job runner (`orion run`) is designed to run arbitrary shell commands on trusted nodes. To secure this:
* Only strict, mTLS-validated connections from trusted peers can invoke `/remote/run`.
* Revoked trust (removing a device via `orion remove`) instantly invalidates the pinned certificate, causing the daemon to drop any connection requests from that peer during the TLS handshake.
* Exposing raw RPC or unvalidated endpoints is prevented.
