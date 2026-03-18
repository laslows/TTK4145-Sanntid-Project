# FSM Module

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Module-FSM-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  Finite State Machine (FSM) module for the elevator project
</p>

---

## FSM module

```text

├── fsm/
│   ├── hall_request_assigner/
│   │   ├── ...
│   ├── fsm.go
│   ├── hallRequestAssigner.go
│   ├── masterSlaveFsm.go
│   ├── requests.go
│   └── `README.md`
```
---

This module implements the elevator state machine, master/backup coordination, and applies hall-call assignments.

### What it includes
* **Local FSM** (`fsm.go`): process sensor/button events, manage state transitions (idle, moving, door open), and issue motor/light commands via the driver.
* **Master/backup election** (`masterSlaveFsm.go`): track peer heartbeats, elect a primary elevator, and promote backups when the primary fails.
* **Hall request assignment** (`hallRequestAssigner.go`): collect local state and active requests, call the external assigner, and apply returned assignments to the local order queue.
* **Shared request types** (`requests.go`): defines request structs used throughout the FSM and helper functions for movement logic for own elevator.

### Hall request assigner helper
The `hall_request_assigner/` folder contains a small D program that computes optimal hall-call assignments. It is built via the included `build.sh` script and invoked from Go via JSON. 

> The executable is executed in the code by navigating to the file `"./src/fsm/hall_request_assigner/hall_request_assigner"` in the `hallRequestAssigner`, and built by creating JSONdata for the active requests and running`dmd main.d config.d elevator_algorithm.d elevator_state.d optimal_hall_requests.d d-json/jsonx.d -w -g -ofhall_request_assigner` in `build.sh` for windows machines, and swapping out the `dmd` for `gdc` on arm64 machines (which i quirkily have) or equivalent for Linux.

