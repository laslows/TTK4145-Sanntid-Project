package network

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const MESSAGE_ADDR = "224.0.0.1:16666"

const INITIALIZATION_TIMEOUT = 1000 * time.Millisecond
const RETRY_BROADCAST_RATE = 10 * time.Millisecond
const BROADCAST_TIMEOUT = 2500 * time.Millisecond

var g_pendingAcks = newSafePendingAcks()
var g_hallRedistributionUpdateCh = make(chan redistributionUpdate, 1)

func BroadcastMessage(senderID, receiverID int, messageType messageType, payload json.RawMessage) {
	message := message{
		m_messageType: messageType,
		m_senderID:    senderID,
		m_receiverID:  receiverID,
		m_payload:     payload,
	}

	message.m_messageID = generateMessageID(message)

	if message.m_messageType == HallOrderRedistribution {
		g_hallRedistributionUpdateCh <- redistributionUpdate{
			m_messageID:  message.m_messageID,
			m_receiverID: message.m_receiverID,
		}
	}

	messageBytes, _ := json.Marshal(&message)

	ackCh := make(chan struct{}, 1)
	g_pendingAcks.insert(message.m_messageID, ackCh)
	defer g_pendingAcks.delete(message.m_messageID)

	retryTicker := time.NewTicker(RETRY_BROADCAST_RATE)
	defer retryTicker.Stop()
	broadcastTimeout := time.NewTicker(BROADCAST_TIMEOUT)
	defer broadcastTimeout.Stop()

	multicastAddr, _ := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)

	conn, _ := net.DialUDP("udp4", nil, multicastAddr)
	defer conn.Close()

	for {
		conn.Write(messageBytes)

		select {
		case <-ackCh:
			return

		//TODO: this does not work
		case update := <-g_hallRedistributionUpdateCh:
			if message.m_messageID != update.m_messageID && update.m_receiverID == message.m_receiverID {
				fmt.Println("Sending newer order distribution")
				return
			}
			fmt.Println("Extracted my own message")
		case <-retryTicker.C:
			continue
		case <-broadcastTimeout.C:
			fmt.Println("Broadcast timeout reached")
			//Send orders back to master, so master can resend
			return
		}
	}
}

func ListenForMessages(e *elevator.Elevator, hallButtonCh chan<- orders.Order,
	assignedOrdersFromMasterCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, peerConnectedCh chan<- int) {

	var messageBuffer = newFifoBuffer()

	multicastAddr, _ := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)
	conn, _ := net.ListenMulticastUDP("udp4", nil, multicastAddr)
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, _, _ := conn.ReadFromUDP(buffer)

		var incomingMessage message

		json.Unmarshal(buffer[:n], &incomingMessage)

		if !(incomingMessage.m_receiverID == e.GetID() ||
			(incomingMessage.m_receiverID == 0 && incomingMessage.m_senderID != e.GetID())) {
			continue
		}

		if incomingMessage.m_messageType == Acknowledgement {

			ch, exists := g_pendingAcks.get(incomingMessage.m_messageID)

			if exists {
				select {
				case ch <- struct{}{}:
				default:
				}
			}

			continue
		}
		SendAcknowledgement(incomingMessage.m_messageID, e.GetID(), incomingMessage.m_senderID)

		if messageBuffer.contains(incomingMessage.m_messageID) {
			continue
		}
		messageBuffer.add(incomingMessage.m_messageID)

		switch incomingMessage.m_messageType {
		case HallOrderRequest:
			var hallOrderRequest orders.Order
			json.Unmarshal(incomingMessage.m_payload, &hallOrderRequest)

			hallButtonCh <- hallOrderRequest

		case HallOrderRedistribution:

			if e.GetIsMaster() {
				continue
			}

			var hallOrderAssignments [config.N_FLOORS][config.N_BUTTONS - 1]bool
			json.Unmarshal(incomingMessage.m_payload, &hallOrderAssignments)

			assignedOrdersFromMasterCh <- hallOrderAssignments

		case Initialization:

			peerConnectedCh <- incomingMessage.m_senderID

		}
	}
}

func TryListenForWorldView() ([config.N_ELEVATORS]*elevator.Backup, bool) {

	var worldView [config.N_ELEVATORS]*elevator.Backup

	messageAddrReceiver, _ := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)
	conn, _ := net.ListenMulticastUDP("udp4", nil, messageAddrReceiver)
	conn.SetReadDeadline(time.Now().Add(INITIALIZATION_TIMEOUT))
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			//Timeout reached

			//maybe should make peer receive worldView even if not new peer. We could get some weird behaviour where you become master and then slave
			// immediately after, depending on what heartbeats you receive first

			fmt.Println("Didn't receive worldView, I am probably first peer.")
			return worldView, false
		}

		var message message
		json.Unmarshal(buffer[:n], &message)

		if message.m_messageType != WorldView {
			continue
		}

		json.Unmarshal(message.m_payload, &worldView)

		fmt.Println("Got worldView.")
		return worldView, true
	}

}

func SendHallOrder(order orders.Order, senderID, receiverId int) {
	payload, _ := json.Marshal(&order)

	fmt.Println("Sending hall order: ", order, " from ", senderID, " to ", receiverId)

	go BroadcastMessage(senderID, receiverId, HallOrderRequest, payload)
}

func SendHallOrderRedistribution(orderList [config.N_FLOORS][config.N_BUTTONS - 1]bool, senderID, receiverID int) {
	payload, _ := json.Marshal(&orderList)
	go BroadcastMessage(senderID, receiverID, HallOrderRedistribution, payload)
}

func SendWorldView(worldView [config.N_ELEVATORS]*elevator.Backup, senderID, receiverId int) {
	payload, _ := json.Marshal(worldView)

	go BroadcastMessage(senderID, receiverId, WorldView, payload)
}

func SendInitializationMessage(senderID int) {
	go BroadcastMessage(senderID, 0, Initialization, nil)
}

func SendAcknowledgement(messageID uint64, senderID, receiverID int) {
	acknowledgementMessage := message{
		m_messageType: Acknowledgement,
		m_senderID:    senderID,
		m_receiverID:  receiverID,
		m_messageID:   messageID,
	}

	payload, _ := json.Marshal(messageID)

	acknowledgementMessage.m_payload = payload

	multicastAddr, _ := net.ResolveUDPAddr("udp4", MESSAGE_ADDR)

	conn, _ := net.DialUDP("udp4", nil, multicastAddr)
	defer conn.Close()

	messageBytes, _ := json.Marshal(&acknowledgementMessage)

	conn.Write(messageBytes)
}
