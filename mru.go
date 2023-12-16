package cache

import (
	"container/list"
)

// Fixed-size in-memory cache with most-recently-used eviction
type MRU struct {
	capacity      int
	usedStorage   int
	evictionQueue *list.List
	cache         map[string][]byte
	stats         Stats
}

// Returns a pointer to a new MRU with a capacity to store limit bytes
func NewMRU(limit int) *MRU {
	return &MRU{
		capacity:      limit,
		usedStorage:   0,
		evictionQueue: list.New(),
		cache:         make(map[string][]byte),
		stats:         Stats{},
	}

}

// Returns the maximum number of bytes this MRU can store
func (mru *MRU) MaxStorage() int {
	return mru.capacity
}

// Returns the number of unused bytes available in this MRU
func (mru *MRU) RemainingStorage() int {
	return (mru.capacity - mru.usedStorage)
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (mru *MRU) Get(key string) (value []byte, ok bool) {
	value, ok = mru.cache[key]

	if ok {
		// increment hit counter and return
		mru.stats.Hits++

		// remove pair from queue
		for e := mru.evictionQueue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				mru.evictionQueue.MoveToFront(e)
				break
			}
		}

		return value, true
	}

	// increment miss counter
	mru.stats.Misses++

	return nil, false
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (mru *MRU) Remove(key string) (value []byte, ok bool) {
	value, ok = mru.cache[key]

	if !ok {
		return nil, false
	}

	// remove key,value pair from map
	delete(mru.cache, key)

	// remove pair from queue
	for e := mru.evictionQueue.Front(); e != nil; e = e.Next() {
		if e.Value == key {
			mru.evictionQueue.Remove(e)
			break
		}
	}

	// decrement usedStorage
	mru.usedStorage -= (len(key) + len(value))

	return value, true
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (mru *MRU) Set(key string, value []byte) bool {
	itemSize := len(key) + len(value)

	if itemSize > mru.capacity {
		return false
	}

	// Check if we are updating an existing key
	existingValue, exists := mru.cache[key]
	if exists {
		// Adjust usedStorage for the old value
		mru.usedStorage -= len(existingValue)
	} else {
		// New key, check if eviction is needed
		for mru.usedStorage+itemSize > mru.capacity {
			// Evict the most recently used item (the front of the queue)
			if elem := mru.evictionQueue.Front(); elem != nil {
				evictKey := elem.Value.(string)
				evictValue := mru.cache[evictKey]
				mru.usedStorage -= (len(evictKey) + len(evictValue))
				delete(mru.cache, evictKey)
				mru.evictionQueue.Remove(elem)
			}
		}
		// Add the new key to the front of the queue
		mru.evictionQueue.PushFront(key)
	}

	// Add or update the value in the cache and adjust usedStorage
	mru.cache[key] = value
	mru.usedStorage += itemSize

	return true
}

// Len returns the number of bindings in the MRU.
func (mru *MRU) Len() int {
	return mru.evictionQueue.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (mru *MRU) Stats() *Stats {
	return &mru.stats
}
