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

func BroadcastHeartbeat(e *elevator.Elevator) {
	multicastAddr, _ := net.ResolveUDPAddr("udp4", HEARTBEAT_ADDR)

	conn, _ := net.DialUDP("udp4", nil, multicastAddr)
	defer conn.Close()

	ticker := time.NewTicker(HEARTBEAT_RATE)
	defer ticker.Stop()

	for range ticker.C {

		heartbeatPacket, _ := json.Marshal(e.GetMyBackup())

		conn.Write(heartbeatPacket)

	}
}

func ListenForHeartbeats(tryUpdateWorldViewCh chan<- elevator.Backup, peerLostCh chan<- int) {

	multicastAddr, _ := net.ResolveUDPAddr("udp4", HEARTBEAT_ADDR)

	conn, _ := net.ListenMulticastUDP("udp4", nil, multicastAddr)
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

			tryUpdateWorldViewCh <- heartBeat
		}

		for peer, timestamp := range peerLastSeen {
			if time.Since(timestamp) > HEARTBEAT_TIMEOUT {
				delete(peerLastSeen, peer)
				peerLostCh <- peer
			}
		}

	}
}

