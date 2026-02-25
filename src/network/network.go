package network

import (
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"encoding/json"
	"fmt"
	"net"
)

const MESSAGE_PORT = "16666"

// Maybe change to not multicast??
const MESSAGE_ADDR = "224.0.0.1:16666"

type messageType int

const (
	hallOrderRequest messageType = iota
	hallOrderAssignment
	motorStop
	orderRedistribution
	backup
)

type Message struct {
	m_messageType messageType
	m_payload     json.RawMessage
}

type assignedOrderMessage struct {
	m_ID int 
	m_order orders.Order
}

type motorStopMessage struct {
	m_ID int
	m_hasMotorstop bool
}

func BroadcastMessage(message Message) {
	//Send message to multicast address
	messageAddrSender, err := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)

	if err != nil {
		fmt.Println("Error resolving multicast address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, messageAddrSender)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return
	}
	defer conn.Close()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}

	_, err = conn.Write(messageBytes)
	if err != nil {
		fmt.Println("Error writing to UDP connection:", err)
		return
	}

}

func SendHallOrderToMaster(order orders.Order) {
	hallOrderMessage := Message{
		m_messageType: hallOrderRequest,
	}

	payload, err := json.Marshal(order)
	if err != nil {
		//Handle error
		return
	}

	hallOrderMessage.m_payload = payload
	BroadcastMessage(hallOrderMessage)
}

func SendBackupToRestoredElevator(b elevator.Backup) {
	backupMessage := Message{
		m_messageType: backup,
	}

	payload, err := json.Marshal(b)
	if err != nil {
		//Handle error
		return
	}

	backupMessage.m_payload = payload
	BroadcastMessage(backupMessage)
}

func ListenForMessages(e *elevator.Elevator, hallButtonCh chan<- orders.Order) {
	//heartbeatAddrReceiver, err := net.ResolveUDPAddr("udp", ":" + HEARTBEAT_PORT)
	messageAddrReceiver, err := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)

	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	//conn, err := net.ListenUDP("udp", heartbeatAddrReceiver)
	conn, err := net.ListenMulticastUDP("udp4", nil, messageAddrReceiver)

	if err != nil {
		fmt.Println("Error listening for messages:", err)
		return
	}
	defer conn.Close()

	//Buffer to read incoming messages into
	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}

		fmt.Println("Received message of length", n, "bytes")

		var message Message

		err = json.Unmarshal(buffer[:n], &message)

		if err != nil {
			fmt.Println("Error unmarshaling message:", err)
			continue
		}

		switch message.m_messageType {
		case hallOrderRequest:
			var hallOrderRequest orders.Order
			err = json.Unmarshal(message.m_payload, &hallOrderRequest)

			if err != nil {
				fmt.Println("Error unmarshaling hall order:", err)
				continue
			}
			//Handle hall order. Use cost function.

			if e.GetIsMaster() {
				hallButtonCh <- hallOrderRequest
				fmt.Printf("Received hall order request: %+v\n", hallOrderRequest)
			} else {
				fmt.Println("Received hall order request, but I am not master. Ignoring.")
			}

		case hallOrderAssignment:
			var hallOrder orders.Order
			err = json.Unmarshal(message.m_payload, &hallOrder)

			if err != nil {
				fmt.Println("Error unmarshaling hall order assignment:", err)
				continue
			}


			//Received hall order from master. Add to local queue.
		}


	}

	}
	
