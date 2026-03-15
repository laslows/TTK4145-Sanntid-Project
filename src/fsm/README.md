# FSM Module

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Project-FSM%20Module-e88cab?style=for-the-badge" alt="Project badge" />
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

The FSM module contains the logic for controlling the elevators hardware, as well as the logic for deciding which elevators are the Primary in the system and which are Backups. It's therefore also what controls the alogorithm for request assignment, given in the handed out code and executable in `hall_request_assigner/` from the `TTK4145` course github-repo and applied in our `hallRequestAssigner`.

Its main functionalities are to:
* Keeps the elevator state machine in sync with buttons/sensors and driver outputs.
* Elects a master/primary elevator (and promotes backups(slaves) if the master fails).
* Uses the external hall-request assigner to compute optimal hall-call assignments.

The executable is executed in the code by navigating to the file `"./src/fsm/hall_request_assigner/hall_request_assigner"` in the `hallRequestAssigner`, and built by creating JSONdata for the active requests and running
`dmd main.d config.d elevator_algorithm.d elevator_state.d optimal_hall_requests.d d-json/jsonx.d -w -g -ofhall_request_assigner` in `build.sh` for windows machines, and swapping out the `dmd` for `gdc` on arm64 machines (which i quirkily have) or equivalent for Linux.

![Primary Backup](../../assets/primaryBackup.drawio%20(1).png "Primary and Backup functionality")
> **_NOTE_**: This module was created before we started implementing the elevator, which means things are bound to differ slightly from our actual logic.
