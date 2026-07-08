# Orion Feature & Architecture Roadmap

Orion's development plan is structured in sequential releases, ensuring each phase establishes a stable foundation for the next.

---

## v0.1.0-beta — Secure Mesh Foundation (Current Release)
Focuses entirely on building the secure P2P mesh network:
*   Secure identity generation (ECDSA P-256 + self-signed X.509 certificates).
*   Certificate pinning and strict Mutual TLS (mTLS).
*   LAN UDP heartbeats and passive offline eviction/decay caching.
*   Automatic console command interception for connection approvals.
*   Zero-inference hardware and model observability tools.

---

## v0.2.0-beta — Distributed Inference MVP
Introduces distributed model execution over the trusted mesh:
*   Integration with **llama.cpp RPC backend**.
*   **Pipeline parallelism (layer sharding)**: split weight matrices across laptops proportional to VRAM capacity.
*   Dynamic scheduling based on transport link telemetry (latencies, bandwidth, handshake metrics).
*   Master node coordinates computation graph dispatching activations via mTLS channels.

---

## v0.3.0-beta — Cluster Visibility & Resilience
Adds visibility tools and dropout handling:
*   Local web dashboard showing cluster topology maps, per-device specs, VRAM usage, and active layer placements.
*   Heartbeat state machine handles node dropouts mid-generation, triggering automatic layer re-placement on online devices.

---

## v0.4.0-beta — Data Parallel serving & Continuous Batching
Maximizes cluster throughput:
*   Exposes OpenAI-compatible `/v1/chat/completions` API endpoint from the mesh.
*   Continuous micro-batching scheduler pipelines multiple concurrent queries to avoid pipeline bubble stalls.
