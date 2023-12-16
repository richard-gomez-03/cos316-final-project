package cache

import (
	"container/list"
)

// Fixed-size in-memory cache with last-in first-out eviction
type LIFO struct {
	capacity      int
	usedStorage   int
	evictionStack *list.List
	cache         map[string][]byte
	stats         Stats
}

// Keep track of cache hits and misses
// type Stats struct {
// 	Hits   int
// 	Misses int
// }

// Returns a pointer to a new LIFO with a capacity to store limit bytes
func NewLifo(limit int) *LIFO {
	return &LIFO{
		capacity:      limit,
		usedStorage:   0,
		evictionStack: list.New(),
		cache:         make(map[string][]byte),
		stats:         Stats{},
	}
}

// Returns the maximum number of bytes this LIFO can store
func (lifo *LIFO) MaxStorage() int {
	return lifo.capacity
}

// Returns the number of unused bytes available in this LIFO
func (lifo *LIFO) RemainingStorage() int {
	return (lifo.capacity - lifo.usedStorage)
}

// Returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise.
func (lifo *LIFO) Get(key string) (value []byte, ok bool) {
	value, ok = lifo.cache[key]

	if ok {
		// increment hit counter and return
		lifo.stats.Hits++

		return value, true
	}

	// increment miss counter
	lifo.stats.Misses++

	return nil, false
}

// Removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (lifo *LIFO) Remove(key string) (value []byte, ok bool) {
	value, ok = lifo.cache[key]

	if !ok {
		return nil, false
	}

	// remove key,value pair from map
	delete(lifo.cache, key)

	// remove pair from queue
	for e := lifo.evictionStack.Back(); e != nil; e = e.Prev() {
		if e.Value == key {
			lifo.evictionStack.Remove(e)
			break
		}
	}

	// decrement usedStorage
	lifo.usedStorage -= (len(key) + len(value))

	return value, true
}

// Associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lifo *LIFO) Set(key string, value []byte) bool {
	itemSize := len(key) + len(value)

	// check if item is too big for cache completely
	if itemSize > lifo.capacity {
		return false
	}

	// if key exists, update its value and adjust the used storage
	if val, exists := lifo.cache[key]; exists {
		lifo.usedStorage -= len(val)
	} else {
		// evict most recently added items if necessary (LIFO behavior)
		for lifo.usedStorage+itemSize > lifo.capacity {
			if elem := lifo.evictionStack.Back(); elem != nil {
				evictKey := elem.Value.(string)
				evictValue := lifo.cache[evictKey]
				lifo.usedStorage -= len(evictKey) + len(evictValue)
				delete(lifo.cache, evictKey)
				lifo.evictionStack.Remove(elem)
			}
		}
		// add new key to stack
		lifo.evictionStack.PushBack(key)
	}

	// add or update key-value pair in the cache
	lifo.cache[key] = value
	lifo.usedStorage += itemSize

	return true
}

// Returns the number of bindings in the LIFO
func (lifo *LIFO) Len() int {
	return lifo.evictionStack.Len()
}

// Returns statistics about how many search hits and misses have occurred
func (lifo *LIFO) Stats() *Stats {
	return &lifo.stats
}
