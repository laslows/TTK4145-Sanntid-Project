package network

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net"
	"time"
)

//TODO: mutex

const MESSAGE_PORT = "16666"

// Maybe change to not multicast??
const MESSAGE_ADDR = "224.0.0.1:16666"

// Maybe move this to initialization package, but that would require us to import it
const INITIALIZATION_TIMEOUT = 1 * time.Second

const ACKNOWLEDGEMENT_TIMEOUT = 10 * time.Millisecond //TODO:better name

var pendingAcks = make(map[uint64]chan bool)
var cache = newFifoCache()

type messageType int

// Make not exported??
const (
	HallOrderRequest messageType = iota
	HallOrderRedistribution
	Initialization
	WorldView
	Acknowledgement
)

type Message struct {
	m_messageType messageType
	m_senderID    int
	m_receiverID  int
	m_payload     json.RawMessage
	m_messageID   uint64
}

func BroadcastMessage(message Message, newHallOrderDistributionCh <-chan uint64) {
	//Send message to multicast address
	messageAddrSender, err := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)

	//if id = myId break
	// must check is even value on channel, and that we are in correct case
	messageID := message.m_messageID

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

	ackCh := make(chan bool)
	pendingAcks[message.m_messageID] = ackCh

	ticker := time.NewTicker(ACKNOWLEDGEMENT_TIMEOUT)
	defer ticker.Stop()

	for range ticker.C {
		_, err = conn.Write(messageBytes)
		fmt.Println("Broadcasting message: ", message.m_messageID)
		if err != nil {
			fmt.Println("Error writing to UDP connection:", err)
			continue
		}
		select {
		case <-ackCh:
			delete(pendingAcks, message.m_messageID)
			return
		case newID := <-newHallOrderDistributionCh:
			if messageID != newID {
				delete(pendingAcks, message.m_messageID)
				return
			}
		}
	}
}

func SendHallOrder(order orders.Order, senderID, receiverId int) {
	hallOrderMessage := Message{
		m_messageType: HallOrderRequest,
		m_senderID:    senderID,
		m_receiverID:  receiverId,
	}

	payload, err := json.Marshal(&order)
	if err != nil {
		//Handle error
		return
	}

	fmt.Println("Sending hall order: ", order, " from ", senderID, " to ", receiverId)

	hallOrderMessage.m_payload = payload
	hallOrderMessage.m_messageID = generateMessageID(hallOrderMessage)

	fmt.Println("Sending hall order: ", order, " from ", senderID, " to ", receiverId)

	go BroadcastMessage(hallOrderMessage, nil)
}

// Inputs a map with elevator id as key and assigned order as value. Should be called by master after running the hall request assignment algorithm
func SendHallOrderRedistribution(orderList [config.N_FLOORS][config.N_BUTTONS - 1]bool, senderID, receiverID int) {
	hallOrderRedistributionMessage := Message{
		m_messageType: HallOrderRedistribution,
		m_senderID:    senderID,
		m_receiverID:  receiverID,
	}

	payload, err := json.Marshal(&orderList)
	if err != nil {
		//Handle error
		return
	}
	messageIdChannel := make(chan uint64)
	//Terminate old broadcasting

	hallOrderRedistributionMessage.m_payload = payload
	hallOrderRedistributionMessage.m_messageID = generateMessageID(hallOrderRedistributionMessage)

	go BroadcastMessage(hallOrderRedistributionMessage, messageIdChannel)
	messageIdChannel <- hallOrderRedistributionMessage.m_messageID
}

func SendWorldView(worldView [config.N_ELEVATORS]*elevator.Backup, senderID, receiverId int) {
	worldViewMessage := Message{
		m_messageType: WorldView,
		m_senderID:    senderID,
		m_receiverID:  receiverId,
	}

	payload, err := json.Marshal(worldView)
	if err != nil {
		//Handle error
		return
	}

	worldViewMessage.m_payload = payload
	worldViewMessage.m_messageID = generateMessageID(worldViewMessage)

	go BroadcastMessage(worldViewMessage, nil)
}

func SendInitializationMessage(senderID int) {
	initializationMessage := Message{
		m_messageType: Initialization, // Make not exported??
		m_senderID:    senderID,
		m_receiverID:  0, //Send to master
	}

	initializationMessage.m_messageID = generateMessageID(initializationMessage)

	go BroadcastMessage(initializationMessage, nil)
}

func SendAcknowledgement(messageID uint64, senderID, receiverID int) {
	acknowledgementMessage := Message{
		m_messageType: Acknowledgement,
		m_senderID:    senderID,
		m_receiverID:  receiverID,
		m_messageID:   messageID,
	}

	payload, err := json.Marshal(messageID)
	if err != nil {
		//Handle error
		return
	}

	acknowledgementMessage.m_payload = payload

	go BroadcastMessage(acknowledgementMessage, nil)
}

func generateMessageID(message Message) uint64 {
	timeStamp := uint32(time.Now().Unix())

	data, err := json.Marshal(message)
	if err != nil {
		return 0
	}

	buffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(buffer, timeStamp)

	hash := fnv.New64a()
	hash.Write(data)
	hash.Write(buffer)

	return hash.Sum64()
}

func ListenForMessages(e *elevator.Elevator, hallButtonCh chan<- orders.Order,
	assignedOrdersFromMasterCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, peerConnectedCh chan<- int) {
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

		if !(message.m_receiverID == e.GetID() || message.m_receiverID == 0) {
			continue
		}

		if cache.contains(message.m_messageID) {
			continue
		}

		cache.add(message.m_messageID)

		if message.m_messageType == Acknowledgement {
			ch, exists := pendingAcks[message.m_messageID]

			if exists {
				ch <- true
			}

			continue
		}

		SendAcknowledgement(message.m_messageID, e.GetID(), message.m_senderID)
		//if in cache: send act + continue
		//If not, send act and do code under

		//Come here if ID == my ID or if receiverID is 0

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

		case HallOrderRedistribution:

			if e.GetIsMaster() {
				continue
			}

			var hallOrderAssignments [config.N_FLOORS][config.N_BUTTONS - 1]bool
			err = json.Unmarshal(message.m_payload, &hallOrderAssignments)

			if err != nil {
				fmt.Println("Error unmarshaling hall order assignment:", err)
				continue
			}

			assignedOrdersFromMasterCh <- hallOrderAssignments

		case Initialization:

			peerConnectedCh <- message.m_senderID

		}

	}

}

func TryListenForWorldView() ([config.N_ELEVATORS]*elevator.Backup, bool) {

	var worldView [config.N_ELEVATORS]*elevator.Backup

	messageAddrReceiver, err := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return worldView, false
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, messageAddrReceiver)
	if err != nil {
		fmt.Println("Error listening for messages:", err)
		return worldView, false
	}
	defer conn.Close()

	//Buffer to read incoming messages into
	buffer := make([]byte, 1024)

	conn.SetReadDeadline(time.Now().Add(INITIALIZATION_TIMEOUT))

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			//Timeout reached

			//maybe should make peer receive worldView even if not new peer. We could get some weird behaviour where you become master and then slave
			// immediately after, depending on what heartbeats you receive first

			fmt.Println("Didn't receive worldView, I am probably first peer.")
			return worldView, false
		}

		var message Message
		err = json.Unmarshal(buffer[:n], &message)
		if err != nil {
			fmt.Println("Error unmarshaling message:", err)
			continue
		}

		if message.m_messageType != WorldView {
			continue
		}

		err = json.Unmarshal(message.m_payload, &worldView)
		if err != nil {
			fmt.Println("Error unmarshaling world view:", err)
			continue
		}

		fmt.Println("Got worldView.")
		return worldView, true
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
