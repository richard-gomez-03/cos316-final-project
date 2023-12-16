package cache

import (
	"math/rand"
	"time"
)

type RR struct {
	capacity    int
	usedStorage int
	cache       map[string][]byte
	keys        []string
	stats       Stats
}

// Create new Random Replacement cache with the given size limit
func NewRR(limit int) *RR {
	rand.Seed(time.Now().UnixNano())
	return &RR{
		capacity:    limit,
		usedStorage: 0,
		cache:       make(map[string][]byte),
		keys:        make([]string, 0),
		stats:       Stats{},
	}
}

// Retrieve a value from the cache
func (rr *RR) Get(key string) (value []byte, ok bool) {
	value, ok = rr.cache[key]
	if ok {
		rr.stats.Hits++
	} else {
		rr.stats.Misses++
	}
	return value, ok
}

// Add new item to the cache, potentially evicting a random item
func (rr *RR) Set(key string, value []byte) bool {
	itemSize := len(key) + len(value)

	// If item is too large for cache, don't add it
	if itemSize > rr.capacity {
		return false
	}

	// Evict random items until there is enough space
	for rr.usedStorage+itemSize > rr.capacity {
		rr.evictRandomItem()
	}

	// Check if updating an existing item
	if _, exists := rr.cache[key]; !exists {
		rr.keys = append(rr.keys, key)
	} else {
		rr.usedStorage -= len(rr.cache[key])
	}

	rr.cache[key] = value
	rr.usedStorage += itemSize

	return true
}

// Remove a random item from the cache
func (rr *RR) evictRandomItem() {
	if len(rr.keys) == 0 {
		return
	}

	// Select a random key and delete it
	randomIndex := rand.Intn(len(rr.keys))
	evictKey := rr.keys[randomIndex]

	// Remove key from slice
	rr.keys[randomIndex] = rr.keys[len(rr.keys)-1]
	rr.keys = rr.keys[:len(rr.keys)-1]

	// Remove key from cache
	rr.usedStorage -= (len(evictKey) + len(rr.cache[evictKey]))
	delete(rr.cache, evictKey)
}

// Remove an item from the cache
func (rr *RR) Remove(key string) (value []byte, ok bool) {
	value, ok = rr.cache[key]
	if !ok {
		return nil, false
	}

	// Find and remove key from the slice
	for i, k := range rr.keys {
		if k == key {
			rr.keys[i] = rr.keys[len(rr.keys)-1]
			rr.keys = rr.keys[:len(rr.keys)-1]
			break
		}
	}

	// Remove key from the cache
	rr.usedStorage -= (len(key) + len(value))
	delete(rr.cache, key)

	return value, true
}

// Return statistics about cache hits and misses
func (rr *RR) Stats() *Stats {
	return &rr.stats
}

// Returns number of items in the cache
func (rr *RR) Len() int {
	return len(rr.keys)
}
