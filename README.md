# Elevator Lab — TTK4145

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Project-Elevator%20Control-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  Distributed and fault-tolerant elevator control software for the TTK4145 elevator lab.
</p>
<!-- FOR BEST VIEWING PLEASURE READ ON GITHUB REPO PAGE -->
---

# Introduction

This project was developed for the **TTK4145** elevator lab and implements software for controlling **n elevators across m floors**. The goal is to build a distributed and fault-tolerant system that handles hall calls and cab calls reliably, while continuing to behave sensibly during failures.

The system is built around the core lab requirements: no calls should be lost, button lights represent a service guarantee, door should function, and elevators should continue operating as reasonably as possible even if communication is interrupted. These assumptions were given in the project description:
> 1. There is always at least one elevator that is not in a failure state
>    - I.e. there is always at least one elevator that can serve calls
>    - "No failure" includes the door obstruction: At least one elevator will be able to close its doors
> 2. Cab call redundancy with a single elevator or a disconnected elevator is not required
>    - Given assumption 1, a system containing only one elevator is assumed to be unable to fail
>    - In a system containing more than one elevator, a disconnected elevator will not have more failures
> 3. No network partitioning: There will never be a situation where there are multiple sets of two or more elevators with no connection between them
>    - Note that this needs 4 or more elevators to become applicable, which we will not test anyway
> 
>[From the TTK4145 course github](https://github.com/TTK4145/Project)
---

## Executable

The program starts one elevator instance and connects it to the simulator through a chosen port.

Run the program with:

```bash
go run main.go -port=15657
```

The default port is `15657`, but additional instances can be started on other ports.

Start the simulator with:

```bash
./SimElevatorServer --port 15657
```

### Example

```bash
./SimElevatorServer --port 15657
./SimElevatorServer --port 15658

go run main.go -port=15657
go run main.go -port=15658
```

> Each elevator instance should use the same port number as its corresponding simulator instance.

For the physichal elevator, its run using:
```bash
./elevatorserver

go run main.go
```

---

### Example: Netimpair

To simulate netimpair and package loss, run the executable

```bash
sudo netimpair -p 12345,23456,34567 --loss 25
```
> which gives a package loss of 25% to the ports 12345, 23456, and 34567
>
---

## Folder Structure

<p align="center">
  Overview of the project structure.
</p>

---

```text
TTK4145-SANNTID-PROJECT
├── elevator-server/
├── Project-resources/
├── simulator/
│   ├── src/
│   ├── SimElevatorServer
│   ├── SimElevatorServer.o
│   └── simulator.con
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
│   │   │   ├── 
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
│   └── timer/
├── go.mod
├── main.go
├── namingConventions.txt
└── README.md <---------- YOU'RE HERE!
```

---

### Notes

* `main.go` is the entry point of the project.
* `src/` contains the main application logic.
* `simulator/` contains the handed out simulator and related files.

---

## Main Modules

The project is divided into a few main modules located in the [`src/`](src/README.md) folder. The most important parts are the [`elevator logic`](src/elevator/README.md), the [`finite-state machine (FSM)`](src/fsm/README.md), and the [`network logic`](src/network/README.md).

The `elevator` module defines the elevator state and backup data structures. The `fsm` module contains the primary-backup logic between master and slave instances, handles requests, and integrates the hall request assigner executable used in the project. The `network` module is responsible for communication between elevators and for sharing state across the system.


<p align="center">
  made for TTK4145
</p>
