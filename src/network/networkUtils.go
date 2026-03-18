package network

import "sync"

const FIFO_CAPACITY = 100

type safePendingAcks struct {
	m_pendingAcks map[uint64]chan struct{}
	m_mutex       sync.RWMutex
}

func newSafePendingAcks() *safePendingAcks {
	return &safePendingAcks{
		m_pendingAcks: make(map[uint64]chan struct{}),
	}
}

func (p *safePendingAcks) insert(messageID uint64, ch chan struct{}) {
	p.m_mutex.Lock()
	defer p.m_mutex.Unlock()

	p.m_pendingAcks[messageID] = ch
}

func (p *safePendingAcks) get(messageID uint64) (chan struct{}, bool) {
	p.m_mutex.RLock()
	defer p.m_mutex.RUnlock()

	ch, exists := p.m_pendingAcks[messageID]
	return ch, exists
}

func (p *safePendingAcks) delete(messageID uint64) {
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

type SafeRedistributionCancels struct {
	m_cancels map[int]chan struct{}
	m_mutex   sync.Mutex
}

func newSafeRedistributionCancels() *SafeRedistributionCancels {
	return &SafeRedistributionCancels{
		m_cancels: make(map[int]chan struct{}),
	}
}

func (registry *SafeRedistributionCancels) replace(receiverID int) chan struct{} {
	registry.m_mutex.Lock()
	defer registry.m_mutex.Unlock()

	if oldCancelCh, exists := registry.m_cancels[receiverID]; exists {
		close(oldCancelCh)
	}

	newCancelCh := make(chan struct{})
	registry.m_cancels[receiverID] = newCancelCh
	return newCancelCh
}

func (registry *SafeRedistributionCancels) clearIfCurrent(receiverID int, candidate chan struct{}) {
	registry.m_mutex.Lock()
	defer registry.m_mutex.Unlock()

	current, exists := registry.m_cancels[receiverID]
	if !exists {
		return
	}

	if current == candidate {
		delete(registry.m_cancels, receiverID)
	}
}
