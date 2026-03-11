package network

const FIFO_CAPACITY = 100

type fifoCache struct {
	capacity int
	queue    []uint64
}

func newFifoCache() *fifoCache {
	return &fifoCache{
		capacity: FIFO_CAPACITY,
		queue:    make([]uint64, 0, FIFO_CAPACITY),
	}
}

func (cache *fifoCache) add(messageID uint64) {
	if len(cache.queue) >= cache.capacity {
		cache.queue = cache.queue[1:]
	}
	cache.queue = append(cache.queue, messageID)
}

func (cache *fifoCache) contains(messageID uint64) bool {
	for _, id := range cache.queue {
		if id == messageID {
			return true
		}
	}
	return false
}



