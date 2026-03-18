# src/

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Overview-src/-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  `src/` holds the core Go packages for the elevator system implementation.
</p>

---

## Folder overview

```text
│ 
├── src/
│   ├── config/
│   │   └── config.go
│   ├── driver/
│   │   └── driver.go
│   ├── elevator/
│   │   ├── backup.go
│   │   ├── elevator.go
│   │   ├── hardware.go
│   │   └── README.md
│   ├── events/
│   │   └── events.go
│   ├── fsm/
│   │   ├── hall_request_assigner/
│   │   │   ├── ...
│   │   ├── fsm.go
│   │   ├── hallRequestAssigner.go
│   │   ├── masterSlaveFsm.go
│   │   ├── README.md
│   │   └── requests.go
│   ├── initialize/
│   │   └── initialize.go
│   ├── network/
│   │   ├── heartbeat.go
│   │   ├── message.go
│   │   ├── network.go
│   │   ├── networkUtils.go
│   │   └── README.md
│   ├── orders/
│   │   └── orders.go
│   ├── timer/
│   │   └── timer.go
│   └── README.md <---------- YOU'RE HERE!
```

---

## src

This file gives a quick overview of the main packages in `src/`

## Packages without a dedicated README

These packages are small and/or self-contained, so their documentation lives here.

### `config`
- Defines configuration values and constants used across the system.
- Keeps magic numbers centralized (e.g., elevator speed, timer values).

### `driver`
- Wraps the low-level elevator hardware interface (buttons, lights, motor commands, sensors).
- Provides a small API used by higher-level packages like `fsm` and `elevator`.

### `events`
- Defines shared event types used across components (e.g., state changes, order events).

### `initialize`
- Handles startup initialization and restoration of state (e.g., applying backup state, loading active orders).

### `orders`
- Defines the `Order` type used across the system (floor + order source: hall up/down or cab).
- Provides JSON serialization helpers so orders can be sent over the network or persisted.

### `timer`
- Provides a simple timer abstraction used for timeouts and periodic checks.

---

## Main modules - look here for more detail

- [`elevator/README.md`](elevator/README.md) — elevator state machine and backup behavior.
- [`fsm/README.md`](fsm/README.md) — master/slave FSM logic and hall request assignment.
- [`network/README.md`](network/README.md) — messaging, peer discovery, and heartbeat logic.
