# Orion Troubleshooting

Orion is built to diagnose its own issues using the `orion doctor` command. However, if you run into common failures, use this guide to identify and resolve them.

---

## 1. Diagnostics with `orion doctor`

Before checking logs or configurations, always run:

```bash
$ orion doctor
```

It acts as the first-line support engineer, checking configuration permissions, local ports, and peer connectivity.

---

## 2. Common Scenarios

### Scenario A: Configuration file missing
* **Symptom**: `doctor` reports `Configuration file missing` or commands fail with `orion is not initialized`.
* **Why**: Orion has not been initialized on the machine, or the configuration file was deleted.
* **Try**: Run `orion init` to generate a fresh identity and configuration.

### Scenario B: Device is reported offline
* **Symptom**: `devices` shows `offline` or `doctor` reports a specific device is offline.
* **Why**: The remote device has lost network connectivity, or the background service on the remote device has crashed or is not running.
* **Try**:
  1. Verify the remote device is connected to the network.
  2. Start the Orion agent on the remote device.
  3. Ensure firewalls are not blocking ports `8910` and `8911`.

### Scenario C: Unused / Stale Devices listed
* **Symptom**: Unnecessary devices appear in `orion devices` list.
* **Why**: You paired machines that are no longer in use.
* **Try**: In version 0.1.0-beta, you can clean up paired devices by manually removing the entry in the `config.json` file (indicated in the `orion doctor` report).

---

## 3. Getting Help

When opening an issue on GitHub, please always attach:
1. The output of `orion version`
2. The complete printout of `orion doctor`
3. The name of the Operating System for both the initiator and remote targets.
