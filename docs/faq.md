# Orion FAQ

### Q1: Does Orion require an internet connection?
No. Orion is designed from the ground up to be offline-first. It discovers and pairs computers over Local Area Networks (LANs) using UDP broadcast/multicast and TCP control sockets.

### Q2: What port does Orion use?
Orion listens on:
* **UDP port 8910**: Used for zero-config heartbeat discovery.
* **TCP port 8911**: Used for mTLS pairings, local loopback control, and job executions.

### Q3: Why is my device showing as Offline?
If a paired device is offline:
1. Ensure the background daemon is running on that device (`orion status`).
2. Verify that your firewall is not blocking TCP port `8911` or UDP port `8910`.
3. Check that both devices are connected to the same network segment (Wi-Fi subnet or Ethernet switch).

### Q4: How is security handled without a cloud CA?
Orion uses **certificate pinning**. During the first pairing handshake, you manually accept the request. This saves the remote node's X.509 certificate PEM in your local config database. Future control actions compare certificates directly, matching them against the pinned version, preventing man-in-the-middle attacks without needing any external Certificate Authorities.
