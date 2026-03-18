# Network module

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Module-Network-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  Network module for the elevator project
</p>

---

## Network module

```text
├── network/
│   ├── heartbeat.go
│   ├── message.go
│   ├── network.go
│   ├── networkUtils.go
│   └── `README.md`
```
---

The network module is divided into three parts: 
* `heartbeat`, which uses Multicast UDP to listen to and broadcast heartbeats (`I'm alive!`) with a read deadline for the heartbeats to arrive to recipients. If a heartbeat from an elevator is missed, the peer is marked as lost and removed from the system.
* `network`, which handles sending/receiving messages, tracking acknowledgements, and updating the global worldview.
* `networkHandler`, which uses mutex locks to safely track pending acknowledgements and maintains FIFO caches for message ordering.

---

## Which is important for...

- **Fault detection:** Heartbeats are the primary mechanism for detecting disconnected elevators.
- **Reliable messaging:** The network layer ensures messages are acknowledged and re-sent when needed.
- **Worldview consistency:** Each elevator keeps a local copy of others' state; the network module keeps that view up to date.

---

## Overview

- `heartbeat.go`: responsible for discovery + liveness tracking
- `network.go`: message serialization/deserialization + ack tracking
- `networkHandler.go`: concurrency-safe buffering + FIFO caches for ordered delivery
