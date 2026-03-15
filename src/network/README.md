# Network module

<p align="center">
  <img src="https://img.shields.io/badge/Course-TTK4145-f4a6c1?style=for-the-badge" alt="TTK4145 badge" />
  <img src="https://img.shields.io/badge/Project-Network%20Module-e88cab?style=for-the-badge" alt="Project badge" />
  <img src="https://img.shields.io/badge/Language-Go-f8d7e4?style=for-the-badge" alt="Go badge" />
</p>

<p align="center">
  Network module for the elevator project
</p>

---

## Network module

```text
│   
├── src/
│   │ 
│   ├── network/
│   │   ├── heartbeat.go
│   │   ├── network.go
│   │   ├── networkHandler.go
│   │   └── `README.md`
```
---

The network module is divided into three parts: 
* `heatbeat`, which uses Multicast UDP to listen to and broadcast heartbeats (`I'm alive!`) with a read deadline for the heartbeats to arrive to recipients. If the heartbeat-signal from an elevator doesn't make the deadline, the peer will be marked as lost and disconnected from the elevator system.
* `network`, which handles the messages sent, their contents, replying with acknowledgements and updating the systems worldview.
* and `networkHandler`, which uses mutex locks to make safe pendingacks and creates FIFOcaches.