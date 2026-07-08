# Orion Networking & Transport Detection

Orion operates entirely on Local Area Networks (LANs) without requiring external routers, NAT traversal, or cloud relays.

---

## 1. Discovery Engine
Discovery is continuous, passive, and zero-configuration:
* **UDP Heartbeats**: The background daemon broadcasts JSON discovery payloads periodically (every 3 seconds) over UDP port `8910` to `255.255.255.255`.
* **Eviction & Decay**: Discovered peers are held in memory. If no heartbeat is received from a peer for 8 seconds, it is marked as `offline` with a "Last seen" decay timestamp. If it remains inactive for 2 minutes, it is evicted from the discovery cache.
* **Control Port**: Control APIs and remote job execution routes run on TCP port `8911` under strict mTLS.

---

## 2. Link Transport Detection
To prepare the future model scheduler for route optimization, Orion profiles active interfaces to identify the transport link:
* **Loopback**: Localhost connection (under 1ms, high bandwidth).
* **Ethernet**: High-speed, low-latency copper/fiber connection.
* **Wi-Fi**: Wireless local loop.
* **USB**: High-speed local tethered interfaces.
* **Bluetooth**: Short-range personal networks.

Transport detection parses local interfaces matching the IP address used during broadcast, allowing peers to automatically advertise how they are connected.
