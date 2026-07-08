# Orion Network Protocol Specification

This document details the serialization formats, network port handshakes, and wire protocols utilized by Orion v0.1.0-beta.

---

## 1. Discovery Packet Schema (UDP port 8910)
Broadcasted as a serialized JSON object to `255.255.255.255:8910` every 3 seconds:
```json
{
  "device_id": "ORN-9AF2-81D7",
  "device_name": "My-Host",
  "os": "windows",
  "port": 8911,
  "fingerprint": "8F:E1:12:A2",
  "transport": "Wi-Fi",
  "hardware": {
    "cpu": "13th Gen Intel Core i5",
    "ram": "16 GB",
    "gpu": "NVIDIA GeForce RTX 4050",
    "vram": "6.0 GB",
    "cuda": "Yes",
    "os": "windows",
    "arch": "amd64",
    "driver_version": "576.xx",
    "cuda_version": "12.9"
  },
  "ollama_installed": true,
  "ollama_version": "0.1.48",
  "models": [
    {
      "name": "llama3:latest",
      "size": 4661234567,
      "quantization": "Q4_K_M",
      "status": "Idle"
    }
  ],
  "protocol": 1,
  "version": "0.1.0-beta",
  "capabilities": {
    "ollama": true,
    "cuda": true,
    "gpu": true,
    "rocm": false,
    "metal": false,
    "benchmark": true,
    "hardware": true
  }
}
```

---

## 2. Pairing Handshake Flow (TCP port 8911)
To establish trust:
1. **Request**: Sender issues a `POST /remote/pair` carrying `PairRequest` payload:
   ```json
   {
     "device_id": "ORN-9AF2-81D7",
     "device_name": "My-Host",
     "os": "windows",
     "fingerprint": "8F:E1:12:A2",
     "transport": "Wi-Fi",
     "certificate": "-----BEGIN CERTIFICATE-----\n...",
     "protocol": 1,
     "version": "0.1.0-beta",
     "capabilities": { ... }
   }
   ```
2. **Compatibility Assertions**: The recipient verifies that `"protocol"` matches `1`. If mismatched, it rejects with `400 Bad Request`.
3. **Interactive Consent**: The recipient blocks the connection and prompts the user.
4. **Response**: If accepted, it responds with `PairResponse` (carrying its cert and capabilities):
   ```json
   {
     "approved": true,
     "device_id": "ORN-8C12-A109",
     "device_name": "Target-Host",
     "os": "macos",
     "fingerprint": "8F:E1:12:A2",
     "certificate": "-----BEGIN CERTIFICATE-----\n...",
     "protocol": 1,
     "version": "0.1.0-beta",
     "capabilities": { ... }
   }
   ```
5. **Trust Storage**: Both parties save each other's certificate PEM, alias, and trust timestamps in their local configuration database.
