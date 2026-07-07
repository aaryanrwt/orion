# Orion CLI Philosophy

This document outlines the core values and product design principles behind Orion. When designing new features, commands, or behaviors, evaluate them against these principles.

---

## 1. Hide Engineering Complexity
Orion is built on complex distributed systems primitives—mTLS, overlay networking, peer discovery, and concurrent task scheduling. However, the user should never see these terms.
* **No protocol jargon**: Do not print "gRPC", "mTLS", "overlay", "Raft", or "handshake".
* **Magical experience**: The CLI should make connecting machines feel as seamless as a local loopback.

## 2. Fast Startup & Low Overhead
A developer utility must feel instantaneous.
* Target command invocation to rendering output under **50ms**.
* Keep dependencies lightweight.
* Do not block startup on slow network calls; perform discovery or synchronization asynchronously or only when explicitly requested.

## 3. Support-Oriented Diagnostics (`orion doctor`)
When things break, the tool must act as the support engineer.
* The first response to any issue should be "Please run `orion doctor`".
* The doctor report must be self-contained, human-readable, and highlight exactly how to fix the issue.

## 4. Helpful Errors (What, Why, Try)
An error message should never just be a Go stack trace or a generic "connection refused".
* **What**: Clear summary of what failed.
* **Why**: The most likely root causes.
* **Try**: Actionable steps or commands the user can run right now to resolve the issue.

## 5. Security by Default
We do not sacrifice security for simplicity.
* **Explicit pairing**: Devices must explicitly trust each other before communication is permitted.
* **Encrypted communication**: All traffic must be encrypted end-to-end.
* **No telemetry**: Do not collect user data, commands, or network details. We respect developer privacy.

## 6. Consistent Output (Visual Discipline)
The terminal interface must look like a premium developer tool, not an engineering console.
* **Semantic Colors**: Green for Success, Yellow for Warning, Red for Error, Blue for Action, Gray for Secondary. Never use color as decoration.
* **NO_COLOR**: Respect the `NO_COLOR` standard.
* **Grid Alignment**: Always format lists, tables, and streams with precise margins and alignment.
