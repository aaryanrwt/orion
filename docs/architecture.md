# Orion Internal Architecture

Orion splits infrastructure into two discrete execution layers:

```
                  ┌──────────────────────────────┐
                  │          Orion CLI           │
                  │   orion devices / run / status│
                  └──────────────┬───────────────┘
                                 │ Loopback HTTP (127.0.0.1:8911)
                  ┌──────────────▼───────────────┐
                  │         Orion Daemon         │
                  │  Event Bus · Discovery Cache │
                  └──────┬───────────────┬───────┘
                         │               │
     UDP Discovery Port  │               │ TCP Control Port (mTLS)
     (0.0.0.0:8910)      │               │ (0.0.0.0:8911)
                  ┌──────▼──────┐ ┌──────▼──────┐
                  │ Remote Peer │ │ Remote Peer │
                  └─────────────┘ └─────────────┘
```

---

## 1. The Decoupled Event Bus
To allow components to scale independently and prevent tight coupling, the daemon implements an in-memory thread-safe `EventBus`:
* **Event Types**: e.g., `EventDeviceOnline`, `EventDeviceOffline`.
* **Subscribers**: The discovery engine, connection managers, and CLI status monitors subscribe to event channels.
* **Publishers**: The UDP listener publishes events immediately upon detecting state transitions.

---

## 2. Asynchronous Hardware Profiling
Orion profiles system specs asynchronously on startup. This prevents blocking daemon execution or delaying CLI startup times, keeping daemon initialization times under 1ms.
* **Windows**: Profiles CPU via CIM Win32, RAM, and GPU specs via `nvidia-smi`.
* **Linux**: Profiles lscpu and nvidia-smi.
* **macOS**: Profiles sysctl and Apple Silicon Metal configurations.
