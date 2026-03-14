package network

import (
	"Sanntid/src/elevator"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const HEARTBEAT_PORT = "15555"
const HEARTBEAT_ADDR = "224.0.0.1:15555"
const HEARTBEAT_RATE = 15 * time.Millisecond
const HEARTBEAT_TIMEOUT = 500 * time.Millisecond

func ListenForHeartbeats(elev *elevator.Elevator, updateWorldViewCh chan<- elevator.Backup, peerLostCh chan<- int) {
	//heartbeatAddrReceiver, err := net.ResolveUDPAddr("udp", ":" + HEARTBEAT_PORT)
	heartbeatAddrReceiver, err := net.ResolveUDPAddr("udp4", HEARTBEAT_ADDR)

	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	//conn, err := net.ListenUDP("udp", heartbeatAddrReceiver)
	conn, err := net.ListenMulticastUDP("udp4", nil, heartbeatAddrReceiver)

	if err != nil {
		fmt.Println("Error listening for heartbeats:", err)
		return
	}
	defer conn.Close()

	//Buffer to read incoming heartbeats into
	buffer := make([]byte, 1024)

	lastSeen := make(map[int]time.Time)

	for {
		conn.SetReadDeadline(time.Now().Add(HEARTBEAT_TIMEOUT))
		n, _, err := conn.ReadFromUDP(buffer)

		if err == nil {
			if !elev.GetConnectedToNetwork() {
				continue
			}

			var heartBeat elevator.Backup

			json.Unmarshal(buffer[:n], &heartBeat)

			lastSeen[heartBeat.GetID()] = time.Now()

			if elev.TryUpdateWorldView(&heartBeat) {
				updateWorldViewCh <- heartBeat

				// Problem: how do the masters decide what backup to throw away (??)
			}
		}

		for peer, timestamp := range lastSeen {
			if time.Since(timestamp) > HEARTBEAT_TIMEOUT {
				delete(lastSeen, peer)
				peerLostCh <- peer
			}
		}

	}
}

func BroadcastHeartbeat(e *elevator.Elevator) {
	//heartbeatAddrSender, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+HEARTBEAT_PORT)
	heartbeatAddrSender, err := net.ResolveUDPAddr("udp4", HEARTBEAT_ADDR)

	if err != nil {
		fmt.Println("Error resolving multicast address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, heartbeatAddrSender)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(HEARTBEAT_RATE)
	defer ticker.Stop()

	for range ticker.C {

		if !e.GetConnectedToNetwork() {
			continue
		}

		heartbeatPacket, err := json.Marshal(e.GetMyBackup())
		if err != nil {
			fmt.Println("Error marshaling heartbeat:", err)
			continue
		}

		_, err = conn.Write(heartbeatPacket)

		if err != nil {
			fmt.Println("Error sending heartbeat:", err)
			continue
		}
	}
}
