package network

import "sync"

const FIFO_CAPACITY = 10

type SafePendingAcks struct {
	m_pendingAcks map[uint64]chan bool
	m_mutex       sync.RWMutex
}

func newSafePendingAcks() *SafePendingAcks {
	return &SafePendingAcks{
		m_pendingAcks: make(map[uint64]chan bool),
	}
}

func (p *SafePendingAcks) insert(messageID uint64, ch chan bool) {
	p.m_mutex.Lock()
	defer p.m_mutex.Unlock()

	p.m_pendingAcks[messageID] = ch
}

func (p *SafePendingAcks) get(messageID uint64) (chan bool, bool) {
	p.m_mutex.RLock()
	defer p.m_mutex.RUnlock()

	ch, exists := p.m_pendingAcks[messageID]
	return ch, exists
}

func (p *SafePendingAcks) delete(messageID uint64) {
	p.m_mutex.Lock()
	defer p.m_mutex.Unlock()

	delete(p.m_pendingAcks, messageID)
}

type fifoBuffer struct {
	m_capacity int
	m_queue    []uint64
	m_mutex    sync.RWMutex
}

func newFifoBuffer() *fifoBuffer {
	return &fifoBuffer{
		m_capacity: FIFO_CAPACITY,
		m_queue:    make([]uint64, 0, FIFO_CAPACITY),
	}
}

func (buffer *fifoBuffer) add(messageID uint64) {
	buffer.m_mutex.Lock()
	defer buffer.m_mutex.Unlock()

	if len(buffer.m_queue) >= buffer.m_capacity {
		buffer.m_queue = buffer.m_queue[1:]
	}
	buffer.m_queue = append(buffer.m_queue, messageID)
}

func (buffer *fifoBuffer) contains(messageID uint64) bool {
	buffer.m_mutex.RLock()
	defer buffer.m_mutex.RUnlock()

	for _, id := range buffer.m_queue {
		if id == messageID {
			return true
		}
	}
	return false
}

type redistributionUpdate struct {
    m_messageID  uint64
    m_receiverID int
}
