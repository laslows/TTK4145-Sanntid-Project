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
>[Taken from the TTK4145 course github](https://github.com/TTK4145/Project) (klasbo, krlampe, 2023)
---

## Executable

The program starts one elevator instance and connects it to the simulator or physical elevator system through a chosen port.

Run the program with:

```bash
go run main.go -port=15657 -id=1
```

The default port is `15657`, but additional instances can be started on other ports, and the ID chooses who will become master, where ID > 0.

Start the simulator with:

```bash
./SimElevatorServer --port 15657
```

### Example

```bash
./SimElevatorServer --port 15657
./SimElevatorServer --port 15658

go run main.go -port=15657 -id=1
go run main.go -port=15658 -id=2
```

> Each elevator instance should use the same port number as its corresponding simulator instance.

For the physichal elevator, it is run using:
```bash
./elevatorserver

go run main.go -id=1
```

---

### Example: Netimpair

To simulate netimpair and package loss, run the executable

```bash
sudo netimpair -p 15555,16666 --loss 25
```
> which gives a package loss of 25% to the ports 15555, and 16666
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
│   │   ├── message.go
│   │   ├── network.go
│   │   ├── networkUtils.go
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

### Our solution

Our system is built as a **hybrid Primary-Backup + Peer-to-Peer architecture** to satisfy the lab requirements for reliability, fault tolerance, and no lost requests.

- **Primary/Backup coordination**
  - Each elevator instance runs the same code and participates in a distributed group.
  - The **master (primary)** is elected dynamically: the alive elevator with the highest ID becomes master.
  - The master is responsible for assigning *hall calls* to elevators using the outsourced hall-request assigner algorithm.
  - **Backups (slaves)** continuously receive state updates from the master (and each other) so they can take over immediately if the master fails.

- **Peer-to-peer state sharing**
  - Each elevator keeps a local **worldview**: a map of every elevator’s last known state (position, direction, button/door status, pending requests, etc.).
  - These states are serialized into a **backup snapshot** and exchanged over the network, allowing peers to detect failures and recover smoothly.

- **Reliability and failure handling**
  - We use UDP multicast heartbeats to detect disconnected elevators quickly.
  - When a peer is marked lost, its pending requests are re-assigned by the current master so no hall calls are lost.
  - Each elevator persists its last known state in memory and uses version numbers to resolve stale updates.

- **Request assignment**
  - Hall calls are assigned using the course-provided `hall_request_assigner` executable (written in D) that computes an optimal assignment based on current elevator states and pending requests.
  - The master gathers the current worldview, serializes it to JSON, runs the assigner, and then distributes the resulting plan to all elevators.

This hybrid approach gives us the simplicity of a primary–backup system (easy master failover, consistent assignment decisions) while still keeping the overall system **decentralized** (every elevator can become master and every elevator has the same view of the world).

---

## Main Modules

The system is split into cohesive modules under `src/`, each with a clear responsibility. The three most important modules are:

- **`elevator/`** (state & backups): tracks per-elevator state, active orders, and provides the serialization/replication primitives used by the distributed algorithm.
- **`fsm/`** (master/backup coordination): orchestrates the elevator state machine, performs master election, detects failures, and drives request assignment via the hall request assigner.
- **`network/`** (messaging & heartbeat): provides reliable message delivery, peer discovery, and failure detection so the system can stay consistent across multiple running instances.

### Key modules (high level)

#### [`src/elevator`](src/elevator/README.md)
- Contains the core elevator model (position, direction, door status, request queues).
- Implements **backup snapshots**: serializable state used to keep peers in sync and to recover after disconnects.
- Maintains the **worldview** (the combined global state seen by each elevator) for decision-making and master election.

#### [`src/fsm`](src/fsm/README.md)
- Manages the elevator finite-state machine: reading sensors, updating driver outputs, and handling transitions.
- Runs the **Primary/Backup protocol**: electing a master, promoting backups, and reacting to dropped peer heartbeats.
- Integrates the external `hall_request_assigner` executable to compute optimal hall-call assignments.
- Handles request routing: assigns hall requests to elevators and ensures cab requests are served locally.

#### [`src/network`](src/network/README.md)
- Sends and receives messages between all elevator instances.
- Tracks message acknowledgements to provide **at-least-once delivery semantics**.
- Broadcasts and listens for **heartbeat packets** to detect peer liveliness and drive failure recovery.
- Maintains FIFO ordering for messages so state updates are applied consistently.

### Supporting modules

These smaller modules provide supporting functionality used by the core three:

- `src/config`: command-line flags and runtime configuration parsing.
- `src/driver`: abstracts the output interface to the simulator/physical elevator (motors, lights, buttons, sensors).
- `src/events`: internal event types and pub/sub patterns for communicating between modules.
- `src/initialize`: startup initialization logic (connecting to the simulator, setting initial state, starting goroutines).
- `src/orders`: local ordering logic and bookkeeping for cab/hall requests that belong to a single elevator.
- `src/timer`: timeouts and periodic task scheduling (e.g., door timing, retry loops).

For more details, each module contains its own [`README.md`](src/README.md).

<p align="center">
  made for TTK4145
</p>
