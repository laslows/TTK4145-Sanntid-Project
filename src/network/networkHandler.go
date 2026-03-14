package network

import "sync"

const FIFO_CAPACITY = 10

type SafePendingAcks struct {
	m_pendingAcks map[uint64]chan bool
	m_mutex       sync.RWMutex
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

type fifoCache struct {
	capacity int
	queue    []uint64
	mutex    sync.RWMutex
}

func newFifoCache() *fifoCache {
	return &fifoCache{
		capacity: FIFO_CAPACITY,
		queue:    make([]uint64, 0, FIFO_CAPACITY),
	}
}

func (cache *fifoCache) add(messageID uint64) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if len(cache.queue) >= cache.capacity {
		cache.queue = cache.queue[1:]
	}
	cache.queue = append(cache.queue, messageID)
}

func (cache *fifoCache) contains(messageID uint64) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	for _, id := range cache.queue {
		if id == messageID {
			return true
		}
	}
	return false
}
