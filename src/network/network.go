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

const MESSAGE_PORT = "16666"

// Maybe change to not multicast??
const MESSAGE_ADDR = "224.0.0.1:16666"

// Maybe move this to initialization package, but that would require us to import it
const INITIALIZATION_TIMEOUT = 1 * time.Second
const ACK_RETRANSMIT_INTERVAL = 1000 * time.Millisecond //TODO:better name

var cache = newFifoCache()
var pendingAcks = newSafePendingAcks()
var hallRedistributionUpdateCh = make(chan redistributionUpdate)


type messageType int

// TODO: Make not exported??
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

//TODO: max retries? Timeout? So it doesn't send messages for ever if no ack received
// (edge case if we change master for example)

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

	ackCh := make(chan bool, 1)
	pendingAcks.insert(message.m_messageID, ackCh)
	//Always do this when we leave broadcast
	defer pendingAcks.delete(message.m_messageID)

	ticker := time.NewTicker(ACK_RETRANSMIT_INTERVAL)
	defer ticker.Stop()

	for {
		_, err = conn.Write(messageBytes)
		fmt.Println("Broadcasting message: ", message.m_messageID, message.m_messageType, " to ", message.m_receiverID)
		if err != nil {
			fmt.Println("Error writing to UDP connection:", err)
			continue
		}
		select {
		case <-ackCh:
			return

		case update := <-hallRedistributionUpdateCh:
			if message.m_messageID != update.m_messageID && update.m_receiverID == message.m_receiverID {
				fmt.Println("Sending newer order distribution")
				return
			}
			fmt.Println("Sending new hall order distribution, but to another elevator")
		case <-ticker.C:
			continue
		}

	}

}

func generateMessageID(message Message) uint64 {
	timeStamp := uint32(time.Now().Unix())

	data, err := json.Marshal(&message)
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
	assignedOrdersFromMasterCh chan<- map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool, peerConnectedCh chan<- int) {
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

		if !(message.m_receiverID == e.GetID() ||
			(message.m_receiverID == 0 && message.m_senderID != e.GetID())) {
			continue
		}

		if message.m_messageType == Acknowledgement {

			ch, exists := pendingAcks.get(message.m_messageID)

			if exists {
				select {
				case ch <- true:
				default:
				}
			}

			continue
		}

		SendAcknowledgement(message.m_messageID, e.GetID(), message.m_senderID)

		if cache.contains(message.m_messageID) {
			continue
		}

		cache.add(message.m_messageID)

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

			var hallOrderAssignments map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool
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

	hallOrderMessage.m_payload = payload
	hallOrderMessage.m_messageID = generateMessageID(hallOrderMessage)

	fmt.Println("Sending hall order: ", order, " from ", senderID, " to ", receiverId, "message ID is: ", hallOrderMessage.m_messageID)

	go BroadcastMessage(hallOrderMessage)
}

// Inputs a map with elevator id as key and assigned order as value. Should be called by master after running the hall request assignment algorithm
func SendHallOrderRedistribution(globalOrderList map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool, senderID, receiverID int) {
	hallOrderRedistributionMessage := Message{
		m_messageType: HallOrderRedistribution,
		m_senderID:    senderID,
		m_receiverID:  receiverID,
	}

	payload, err := json.Marshal(&globalOrderList)
	if err != nil {
		//Handle error
		return
	}
	//Terminate old broadcasting

	hallOrderRedistributionMessage.m_payload = payload
	hallOrderRedistributionMessage.m_messageID = generateMessageID(hallOrderRedistributionMessage)

	go BroadcastMessage(hallOrderRedistributionMessage)

	hallRedistributionUpdateCh <- redistributionUpdate{
		m_messageID:  hallOrderRedistributionMessage.m_messageID,
		m_receiverID: receiverID,
	}
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

	go BroadcastMessage(worldViewMessage)
}

func SendInitializationMessage(senderID int) {
	initializationMessage := Message{
		m_messageType: Initialization, // Make not exported??
		m_senderID:    senderID,
		m_receiverID:  0, //Send to master
	}

	initializationMessage.m_messageID = generateMessageID(initializationMessage)

	go BroadcastMessage(initializationMessage)
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

	messageAddrSender, err := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)
	//TODO: can we please not use err everywhere? very annoying :(
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

	messageBytes, err := json.Marshal(&acknowledgementMessage)
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

func (m *Message) MarshalJSON() ([]byte, error) {
	type MessageJSON struct {
		MessageType int
		ReceiverID  int
		SenderID    int
		MessageID   uint64
		Payload     json.RawMessage
	}

	return json.Marshal(&MessageJSON{
		MessageType: int(m.m_messageType),
		ReceiverID:  m.m_receiverID,
		SenderID:    m.m_senderID,
		MessageID:   m.m_messageID,
		Payload:     m.m_payload,
	})
}

func (message *Message) UnmarshalJSON(data []byte) error {
	type MessageJSON struct {
		MessageType int
		ReceiverID  int
		SenderID    int
		MessageID   uint64
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
	message.m_messageID = messageJSON.MessageID
	message.m_payload = messageJSON.Payload

	return nil
}
