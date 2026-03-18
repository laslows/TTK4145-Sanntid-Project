# src/

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Overview-src/-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  src/ overview for main modules
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
│   │   ├── network.go
│   │   ├── networkHandler.go
│   │   └── README.md
│   ├── orders/
│   ├── timer/
|   └── README.md <---------- YOU'RE HERE!
```
---

The `src` folder contains the modules we described in the Preliminary Design Description (PDD) handed in before we started implementing the elevator systems, which is illustrated in the image below. 

![Modules and module communication](../assets/modulesfinn.drawio%20(1).png "Modules and module communication from PDD")
> **_NOTE_**: This module was created before we started implementing the elevator, which means things are bound to differ slightly from our actual logic.
