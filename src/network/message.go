package network

import (
	"encoding/json"
	"hash/fnv"
	"encoding/binary"
	"time"
)

type messageType int

const (
	HallOrderRequest messageType = iota
	HallOrderRedistribution
	Initialization
	WorldView
	Acknowledgement
)

type message struct {
	m_messageType messageType
	m_senderID    int
	m_receiverID  int
	m_payload     json.RawMessage
	m_messageID   uint64
}

func generateMessageID(message message) uint64 {
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

func (m *message) MarshalJSON() ([]byte, error) {
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

func (m *message) UnmarshalJSON(data []byte) error {
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

	m.m_messageType = messageType(messageJSON.MessageType)
	m.m_receiverID = messageJSON.ReceiverID
	m.m_senderID = messageJSON.SenderID
	m.m_messageID = messageJSON.MessageID
	m.m_payload = messageJSON.Payload

	return nil
}

