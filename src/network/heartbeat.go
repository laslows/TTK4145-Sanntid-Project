package network

import (
	"Sanntid/src/elevator"
	"encoding/json"
	"net"
	"time"
)

const HEARTBEAT_ADDR = "224.0.0.1:15555"
const HEARTBEAT_RATE = 10 * time.Millisecond
const HEARTBEAT_TIMEOUT = 250 * time.Millisecond

func ListenForHeartbeats(elev *elevator.Elevator, updateWorldViewCh chan<- elevator.Backup, peerLostCh chan<- int) {

	heartbeatAddrReceiver, _ := net.ResolveUDPAddr("udp4", HEARTBEAT_ADDR)

	conn, _ := net.ListenMulticastUDP("udp4", nil, heartbeatAddrReceiver)
	defer conn.Close()

	buffer := make([]byte, 1024)
	peerLastSeen := make(map[int]time.Time)

	for {
		conn.SetReadDeadline(time.Now().Add(HEARTBEAT_TIMEOUT))
		n, _, err := conn.ReadFromUDP(buffer)

		if err == nil {
			var heartBeat elevator.Backup

			json.Unmarshal(buffer[:n], &heartBeat)

			peerLastSeen[heartBeat.GetID()] = time.Now()

			//TODO: move this
			if elev.TryUpdateWorldView(&heartBeat) {
				updateWorldViewCh <- heartBeat
			}
		}

		for peer, timestamp := range peerLastSeen {
			if time.Since(timestamp) > HEARTBEAT_TIMEOUT {
				delete(peerLastSeen, peer)
				peerLostCh <- peer
			}
		}

	}
}

func BroadcastHeartbeat(e *elevator.Elevator) {
	heartbeatAddrSender, err := net.ResolveUDPAddr("udp4", HEARTBEAT_ADDR)

	conn, err := net.DialUDP("udp", nil, heartbeatAddrSender)

	//TODO: maybe remove this, depending on what studass says
	if err != nil {
		panic("Failed to create UDP connection for heartbeat")
	}
	defer conn.Close()

	ticker := time.NewTicker(HEARTBEAT_RATE)
	defer ticker.Stop()

	for range ticker.C {

		heartbeatPacket, err := json.Marshal(e.GetMyBackup())
		if err != nil {
			continue
		}

		_, err = conn.Write(heartbeatPacket)

		if err != nil {
			continue
		}
	}
}
