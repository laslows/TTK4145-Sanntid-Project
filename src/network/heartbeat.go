package network

import (
	"Sanntid/src/elevator"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const HEARTBEAT_PORT = "15555"
const HEARTBEAT_RATE = 500 * time.Millisecond
const HEARTBEAT_ADDR = "224.0.0.1:15555"

func ListenForHeartbeats(elev *elevator.Elevator, changeMasterSlaveCh chan<- bool) {
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

	for {
		//n is number of bytes received, remoteAddr is the address of the sender, err is error :(
		n, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error reading heartbeat:", err)
			//Jump to next iteration of for
			continue
		}

		var heartBeat elevator.Backup

		err = json.Unmarshal(buffer[:n], &heartBeat)

		if err != nil {
			fmt.Println("Error unmarshaling heartbeat:", err)
			continue
		}

		if elev.TryUpdateWorldView(&heartBeat) {
			elev.UpdateWorldView(&heartBeat) //Could also be called from TryUpdateWorldView

			if elev.TryUpdateIsMaster() {
				//THis is true both if we switched from master to slave, and the other
				changeMasterSlaveCh <- elev.GetIsMaster()
			}

			fmt.Printf("Updated worldview with heartbeat from %s received to %s\n", heartBeat.GetID(), elev.GetID())

			// Problem: how do the masters decide what backup to throw away (??)
		}

		// Hvis vi mottar heartbeat fra noen med lavere port og de er master
		// blir vi slave

		//Do stuff with heartbeat now
		//Reset heartbeat-timer

		//fmt.Printf("Heartbeat from %s received to %s\n", heartBeat.GetPort(), elev.GetPort())
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

func heartbeatMonitor() {

	//Make list/dictionary with len(N_elevators) of timers. Reset timer every time worldview is changed
	//Update timer from heartbeatListener. Make channels for this

}
