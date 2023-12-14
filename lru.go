package cache

import (
	"container/list"
)

// An LRU is a fixed-size in-memory cache with least-recently-used eviction
type LRU struct {
	// whatever fields you want here
	capacity      int
	usedStorage   int
	evictionQueue *list.List
	cache         map[string][]byte
	stats         Stats
}

// NewLRU returns a pointer to a new LRU with a capacity to store limit bytes
func NewLru(limit int) *LRU {
	return &LRU{
		capacity:      limit,
		usedStorage:   0,
		evictionQueue: list.New(),
		cache:         make(map[string][]byte),
		stats:         Stats{},
	}

}

// MaxStorage returns the maximum number of bytes this LRU can store
func (lru *LRU) MaxStorage() int {
	return lru.capacity
}

// RemainingStorage returns the number of unused bytes available in this LRU
func (lru *LRU) RemainingStorage() int {
	return (lru.capacity - lru.usedStorage)
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (lru *LRU) Get(key string) (value []byte, ok bool) {
	value, ok = lru.cache[key]

	if ok {
		// increment hit counter and return
		lru.stats.Hits++

		// remove pair from queue
		for e := lru.evictionQueue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				// find queue element and send to back of queue
				lru.evictionQueue.MoveToBack(e)
				break
			}
		}

		return value, true
	}

	// increment miss counter
	lru.stats.Misses++

	return nil, false
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (lru *LRU) Remove(key string) (value []byte, ok bool) {
	value, ok = lru.cache[key]

	if !ok {
		return nil, false
	}

	// remove key,value pair from map
	delete(lru.cache, key)

	// remove pair from queue
	for e := lru.evictionQueue.Front(); e != nil; e = e.Next() {
		if e.Value == key {
			lru.evictionQueue.Remove(e)
			break
		}
	}

	// decrement usedStorage
	lru.usedStorage -= (len(key) + len(value))

	return value, true
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lru *LRU) Set(key string, value []byte) bool {
	itemSize := (len(key) + len(value))

	// check if item is too big for cache completely
	if itemSize > lru.capacity {
		return false
	}

	// check if key already has value assigned
	if val, exists := lru.cache[key]; exists {
		lru.usedStorage -= len(val)

		// remove
		for (lru.usedStorage + len(value)) > lru.capacity {
			if elem := lru.evictionQueue.Front(); elem != nil {
				key := elem.Value.(string)
				value := lru.cache[key]
				lru.usedStorage -= (len(value) + len(key))
				delete(lru.cache, key)
				lru.evictionQueue.Remove(elem)
			}
		}

		lru.usedStorage -= (len(key))

		// remove pair from queue
		for e := lru.evictionQueue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				lru.evictionQueue.MoveToBack(e)
				break
			}
		}

	} else {
		for (lru.usedStorage + itemSize) > lru.capacity {
			if elem := lru.evictionQueue.Front(); elem != nil {
				key := elem.Value.(string)
				value := lru.cache[key]
				lru.usedStorage -= (len(value) + len(key))
				delete(lru.cache, key)
				lru.evictionQueue.Remove(elem)
			}
		}

		lru.evictionQueue.PushBack(key)
	}

	lru.cache[key] = value
	lru.usedStorage += itemSize

	return true
}

// Len returns the number of bindings in the LRU.
func (lru *LRU) Len() int {
	return lru.evictionQueue.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lru *LRU) Stats() *Stats {
	return &lru.stats
}
