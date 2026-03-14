package network

import (
	"Sanntid/src/orders"
	"sync"
)

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

type safePendingHallOrders struct {
	m_orders map[uint64]orders.Order
	m_mutex  sync.RWMutex
}

func newSafePendingHallOrders() *safePendingHallOrders {
	return &safePendingHallOrders{
		m_orders: make(map[uint64]orders.Order),
	}
}

func (p *safePendingHallOrders) insert(messageID uint64, order orders.Order) {
	p.m_mutex.Lock()
	defer p.m_mutex.Unlock()

	p.m_orders[messageID] = order
}

func (p *safePendingHallOrders) get(messageID uint64) (orders.Order, bool) {
	p.m_mutex.RLock()
	defer p.m_mutex.RUnlock()

	order, exists := p.m_orders[messageID]
	return order, exists
}

func (p *safePendingHallOrders) delete(messageID uint64) {
	p.m_mutex.Lock()
	defer p.m_mutex.Unlock()

	delete(p.m_orders, messageID)
}

type fifoCache struct {
	m_capacity int
	m_queue    []uint64
	m_mutex    sync.RWMutex
}

func newFifoCache() *fifoCache {
	return &fifoCache{
		m_capacity: FIFO_CAPACITY,
		m_queue:    make([]uint64, 0, FIFO_CAPACITY),
	}
}

func (cache *fifoCache) add(messageID uint64) {
	cache.m_mutex.Lock()
	defer cache.m_mutex.Unlock()

	if len(cache.m_queue) >= cache.m_capacity {
		cache.m_queue = cache.m_queue[1:]
	}
	cache.m_queue = append(cache.m_queue, messageID)
}

func (cache *fifoCache) contains(messageID uint64) bool {
	cache.m_mutex.RLock()
	defer cache.m_mutex.RUnlock()

	for _, id := range cache.m_queue {
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
