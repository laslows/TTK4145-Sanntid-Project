package network

import "sync"

const FIFO_CAPACITY = 100

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
