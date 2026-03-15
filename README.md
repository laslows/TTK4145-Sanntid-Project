# Elevator Lab — TTK4145

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Project-Elevator%20Control-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  Distributed and fault-tolerant elevator control software for the TTK4145 elevator lab.
</p>

---

## Introduction

This project was developed for the **TTK4145** elevator lab and implements software for controlling **n elevators across m floors**. The goal is to build a distributed and fault-tolerant system that handles hall calls and cab calls reliably, while continuing to behave sensibly during failures.

The system is built around the core lab requirements: no calls should be lost, button lights represent a service guarantee, and elevators should continue operating as reasonably as possible even if communication is interrupted.

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

<p align="center">
  made for TTK4145
</p>
