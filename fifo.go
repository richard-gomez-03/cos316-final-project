package cache

import (
	"container/list"
)

// An FIFO is a fixed-size in-memory cache with first-in first-out eviction
type FIFO struct {
	// whatever fields you want here
	capacity      int
	usedStorage   int
	evictionQueue *list.List
	cache         map[string][]byte
	stats         Stats
}

// NewFIFO returns a pointer to a new FIFO with a capacity to store limit bytes
func NewFifo(limit int) *FIFO {
	return &FIFO{
		capacity:      limit,
		usedStorage:   0,
		evictionQueue: list.New(),
		cache:         make(map[string][]byte),
		stats:         Stats{},
	}
}

// MaxStorage returns the maximum number of bytes this FIFO can store
func (fifo *FIFO) MaxStorage() int {
	return fifo.capacity
}

// RemainingStorage returns the number of unused bytes available in this FIFO
func (fifo *FIFO) RemainingStorage() int {
	return (fifo.capacity - fifo.usedStorage)
}

// Get returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise.
func (fifo *FIFO) Get(key string) (value []byte, ok bool) {
	value, ok = fifo.cache[key]

	if ok {
		// increment hit counter and return
		fifo.stats.Hits++

		return value, true
	}

	// increment miss counter
	fifo.stats.Misses++

	return nil, false
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (fifo *FIFO) Remove(key string) (value []byte, ok bool) {
	value, ok = fifo.cache[key]

	if !ok {
		return nil, false
	}

	// remove key,value pair from map
	delete(fifo.cache, key)

	// remove pair from queue
	for e := fifo.evictionQueue.Front(); e != nil; e = e.Next() {
		if e.Value == key {
			fifo.evictionQueue.Remove(e)
			break
		}
	}

	// decrement usedStorage
	fifo.usedStorage -= (len(key) + len(value))

	return value, true
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (fifo *FIFO) Set(key string, value []byte) bool {
	itemSize := (len(key) + len(value))

	// check if item is too big for cache completely
	if itemSize > fifo.capacity {
		return false
	}

	if val, exists := fifo.cache[key]; exists {
		fifo.usedStorage -= len(val)

		for (fifo.usedStorage + len(value)) > fifo.capacity {
			if elem := fifo.evictionQueue.Front(); elem != nil {
				key := elem.Value.(string)
				value := fifo.cache[key]
				fifo.usedStorage -= (len(value) + len(key))
				delete(fifo.cache, key)
				fifo.evictionQueue.Remove(elem)
			}
		}

		fifo.usedStorage -= (len(key))

	} else {
		for (fifo.usedStorage + itemSize) > fifo.capacity {
			if elem := fifo.evictionQueue.Front(); elem != nil {
				key := elem.Value.(string)
				value := fifo.cache[key]
				fifo.usedStorage -= (len(value) + len(key))
				delete(fifo.cache, key)
				fifo.evictionQueue.Remove(elem)
			}
		}

		fifo.evictionQueue.PushBack(key)
	}

	fifo.cache[key] = value
	fifo.usedStorage += itemSize

	return true
}

// Len returns the number of bindings in the FIFO.
func (fifo *FIFO) Len() int {
	return fifo.evictionQueue.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (fifo *FIFO) Stats() *Stats {
	return &fifo.stats
}
