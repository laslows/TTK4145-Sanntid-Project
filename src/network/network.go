package network

import (
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
	hallOrder messageType = iota
)

type Message struct {
	m_messageType messageType
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
		m_messageType: hallOrder,
	}

	payload, err := json.Marshal(order)
	if err != nil {
		//Handle error
		return
	}

	hallOrderMessage.m_payload = payload
	BroadcastMessage(hallOrderMessage)
}
