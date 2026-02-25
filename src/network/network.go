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
	HallOrderRequest messageType = iota
	HallOrderAssignment
	MotorStop
	OrderRedistribution
	Backup
)

type Message struct {
	m_messageType messageType
	m_senderID    int
	m_receiverID  int
	m_payload     json.RawMessage
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

	messageBytes, err := json.Marshal(&message)
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

func SendHallOrder(order orders.Order, senderID, receiverId int, messageType messageType) {
	hallOrderMessage := Message{
		m_messageType: messageType,
		m_senderID:    senderID,
		m_receiverID:  receiverId,
	}

	payload, err := json.Marshal(&order)
	if err != nil {
		//Handle error
		return
	}

	hallOrderMessage.m_payload = payload
	BroadcastMessage(hallOrderMessage)
}

// Motpart i onfloorarrival
func SendMotorStopMessage(senderID, receiverId int, motorStopped bool) {
	motorStopMessage := Message{
		m_messageType: MotorStop,
		m_senderID:    senderID,
		m_receiverID:  receiverId,
	}

	payload, err := json.Marshal(&motorStopped)
	if err != nil {
		//Handle error
		return
	}

	motorStopMessage.m_payload = payload
	BroadcastMessage(motorStopMessage)
}

func SendBackupToRestoredElevator(b *elevator.Backup) {
	backupMessage := Message{
		m_messageType: Backup,
	}

	payload, err := json.Marshal(b)
	if err != nil {
		//Handle error
		return
	}

	backupMessage.m_payload = payload
	BroadcastMessage(backupMessage)
}

func ListenForMessages(e *elevator.Elevator, hallButtonCh chan<- orders.Order,
	assignedOrderCh chan<- orders.Order) {
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

		var message Message

		err = json.Unmarshal(buffer[:n], &message)

		if err != nil {
			fmt.Println("Error unmarshaling message:", err)
			continue
		}

		if message.m_receiverID != elevator.GetIPandPortAsInt(e.GetIP(), e.GetPort()) {
			continue
		}

		switch message.m_messageType {
		case HallOrderRequest:
			var hallOrderRequest orders.Order
			err = json.Unmarshal(message.m_payload, &hallOrderRequest)

			if err != nil {
				fmt.Println("Error unmarshaling hall order:", err)
				continue
			}
			//Handle hall order. Use cost function.

			hallButtonCh <- hallOrderRequest

		case HallOrderAssignment:

			var hallOrderAssignment orders.Order
			err = json.Unmarshal(message.m_payload, &hallOrderAssignment)

			if err != nil {
				fmt.Println("Error unmarshaling hall order assignment:", err)
				continue
			}

			assignedOrderCh <- hallOrderAssignment
			//Received hall order from master. Add to local queue.
		case MotorStop:

			var motorStopped bool
			err = json.Unmarshal(message.m_payload, &motorStopped)

			if err != nil {
				fmt.Println("Error unmarshaling motor stop message:", err)
				continue
			}

			fmt.Println("Received motor stop message:", motorStopped, "from", message.m_senderID)

		}

	}

}

func (m *Message) MarshalJSON() ([]byte, error) {
	type MessageJSON struct {
		MessageType int
		ReceiverID  int
		SenderID    int
		Payload     json.RawMessage
	}

	return json.Marshal(&MessageJSON{
		MessageType: int(m.m_messageType),
		ReceiverID:  m.m_receiverID,
		SenderID:    m.m_senderID,
		Payload:     m.m_payload,
	})
}

func (message *Message) UnmarshalJSON(data []byte) error {
	type MessageJSON struct {
		MessageType int
		ReceiverID  int
		SenderID    int
		Payload     json.RawMessage
	}

	var messageJSON MessageJSON
	err := json.Unmarshal(data, &messageJSON)
	if err != nil {
		return err
	}

	message.m_messageType = messageType(messageJSON.MessageType)
	message.m_receiverID = messageJSON.ReceiverID
	message.m_senderID = messageJSON.SenderID
	message.m_payload = messageJSON.Payload

	return nil
}
