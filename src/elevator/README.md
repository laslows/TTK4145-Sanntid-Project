# Elevator module

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Project-Elevator%20Module-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  Elevator/Backup module for the elevator project
</p>

---

## Elevator module

```text

├── elevator/
│   ├── backup.go
│   ├── elevator.go
│   └── `README.md`
```
---

The elevator module contains the core state representation and synchronization logic for each elevator instance.

It has two main pieces:

- **Elevator state:** tracks current floor, direction, active requests, and behavior.
- **Backup state:** a serializable snapshot of elevator state used for network synchronization and fault tolerance.

The module also maintains a **worldview**: a shared view of all elevators' backups used for master election and order distribution.

---

### Overview

- `elevator.go`: elevator state machine helpers, request tracking, and world-view management
- `backup.go`: backup (serialized state) struct + JSON serialization helpers

---

> Quick notes

> The *master elevator* is elected based on the highest ID in the worldview.
> Backups store the last known state of peers and can restore local state after reconnection.
> The backup version number is used to resolve newer vs older state.