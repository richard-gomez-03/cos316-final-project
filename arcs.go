package cache

import (
	"container/list"
)

// An ARC is a fixed-size in-memory cache with adaptive replacement caching
type ARC struct {
	cacheCapacity  int
	t1Used, t2Used int
	b1Used, b2Used int

	p int

	t1Queue, t2Queue *list.List
	cacheT1, cacheT2 map[string][]byte

	b1Queue, b2Queue *list.List
	cacheB1, cacheB2 map[string][]byte

	stats Stats
}

// type Stats struct {
// 	Hits   int
// 	Misses int
// }

// NewARC returns a pointer to a new ARC with a capacity of limit bytes
func NewArc(limit int) *ARC {
	pCap := limit / 2

	return &ARC{
		cacheCapacity: limit,

		t1Used: 0,
		t2Used: 0,
		b1Used: 0,
		b2Used: 0,

		p: pCap,

		t1Queue: list.New(),
		t2Queue: list.New(),
		cacheT1: make(map[string][]byte),
		cacheT2: make(map[string][]byte),

		b1Queue: list.New(),
		b2Queue: list.New(),
		cacheB1: make(map[string][]byte),
		cacheB2: make(map[string][]byte),

		stats: Stats{},
	}

}

// MaxStorage returns the maximum number of bytes the ARC can store
func (arc *ARC) MaxStorage() int {
	return arc.cacheCapacity
}

// FullRemainingStorage returns the number of unused bytes available in this ARC
func (arc *ARC) FullRemainingStorage() int {
	return arc.cacheCapacity - (arc.t1Used + arc.t2Used)
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (arc *ARC) Get(key string) (value []byte, ok bool) {

	// check if key exists in recently used queue (t1)
	value, ok = arc.cacheT1[key]

	if ok {
		arc.stats.Hits++

		for e := arc.t1Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				// moving key instance
				arc.t1Queue.Remove(e)
				delete(arc.cacheT1, key)
				arc.t1Used -= (len(key) + len(value))

				arc.t2Queue.PushFront(key)
				arc.t2Used += (len(key) + len(value))
				arc.cacheT2[key] = value
				return value, true
			}
		}
	}

	// check if key exists in frequenly used queue (t2)
	value, ok = arc.cacheT2[key]

	// if found, move to front to queue to indicate freqently used
	if ok {
		arc.stats.Hits++

		for e := arc.t2Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.t2Queue.MoveToFront(e)
				return value, true
			}
		}
	}

	// check if key was recently evicted from t1
	value, ok = arc.cacheB1[key]

	if ok {
		arc.stats.Misses++

		// formula to calculate change in p
		delta := 1
		if arc.b2Used > arc.b1Used {
			delta = arc.b2Used / max(arc.b1Used, 1)
		}
		arc.p = min(arc.cacheCapacity, arc.p+delta)

		return nil, false
	}

	// check if key was recently evicted from t2
	value, ok = arc.cacheB2[key]

	if ok {
		arc.stats.Misses++

		// formula to calculate change in p
		delta := 1
		if arc.b1Used > arc.b2Used {
			delta = arc.b1Used / max(arc.b2Used, 1)
		}
		arc.p = max(0, arc.p-delta)

		return nil, false
	}

	// increment miss counter
	arc.stats.Misses++

	return nil, false
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (arc *ARC) Remove(key string) (value []byte, ok bool) {
	// Check and remove from t1
	if value, ok = arc.cacheT1[key]; ok {
		delete(arc.cacheT1, key)
		arc.t1Used -= (len(key) + len(value))
		for e := arc.t1Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.t1Queue.Remove(e)
				break
			}
		}
		return value, true
	}

	// Check and remove from t2
	if value, ok = arc.cacheT2[key]; ok {
		delete(arc.cacheT2, key)
		arc.t2Used -= (len(key) + len(value))
		for e := arc.t2Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.t2Queue.Remove(e)
				break
			}
		}
		return value, true
	}

	// Check and remove from b1
	if value, ok = arc.cacheB1[key]; ok {
		delete(arc.cacheB1, key)
		arc.b1Used -= (len(key) + len(value))
		for e := arc.b1Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.b1Queue.Remove(e)
				break
			}
		}
		return nil, false
	}

	// Check and remove from b2
	if value, ok = arc.cacheB2[key]; ok {
		delete(arc.cacheB2, key)
		arc.b2Used -= (len(key) + len(value))
		for e := arc.b2Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.b2Queue.Remove(e)
				break
			}
		}
		return nil, false
	}

	return nil, false
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (arc *ARC) Set(key string, value []byte) bool {
	itemSize := (len(key) + len(value))

	// check if item is too big for cache completely
	if itemSize > arc.cacheCapacity {
		return false
	}

	// check if item is in t1
	val, ok := arc.cacheT1[key]

	if ok {
		// moving from t1 to t2
		arc.t1Used -= len(val) + len(key)

		for e := arc.t1Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.t1Queue.Remove(e)
				break
			}
		}

		delete(arc.cacheT1, key)

		// adding the key-value pair into t2
		arc.cacheT2[key] = value
		arc.t2Queue.PushFront(key)
		arc.t2Used += itemSize
		return true
	}

	// check if item is in t2
	val2, ok := arc.cacheT2[key]

	if ok {
		arc.t2Used -= (len(val2) + len(key))
		arc.t2Used += itemSize

		// setting new value
		arc.cacheT2[key] = value

		// moving the key-value pair to front
		for e := arc.t1Queue.Front(); e != nil; e = e.Next() {
			if e.Value == key {
				arc.t2Queue.MoveToFront(e)
				break
			}
		}

		return true
	}

	// Make room for the new item
	for (arc.t1Used + arc.t2Used + itemSize) > arc.cacheCapacity {
		// the replacement and sizing feature checks the following
		// 1. if the queue has at least 1 element
		// 2. if the the t1 portion of the queue is full (tracked by arc.p)
		if (arc.t1Queue.Len() > 0) && (len(arc.cacheB1) > 0 && arc.t1Used == arc.p || arc.t1Used > arc.p) {

			// last item will get evicted
			e := arc.t1Queue.Back()

			//
			if e != nil {
				evictedKey := e.Value.(string)
				evictedValue := arc.cacheT1[evictedKey]
				delete(arc.cacheT1, evictedKey)
				arc.t1Queue.Remove(e)
				arc.t1Used -= len(evictedValue) + len(evictedKey)

				// Move to b1
				arc.b1Queue.PushFront(evictedKey)
				arc.cacheB1[evictedKey] = evictedValue
			}
		} else {
			// Evict from t2
			e := arc.t2Queue.Back()
			if e != nil {
				evictedKey := e.Value.(string)
				evictedValue := arc.cacheT2[evictedKey]
				delete(arc.cacheT2, evictedKey)
				arc.t2Queue.Remove(e)
				arc.t2Used -= len(evictedValue) + len(evictedKey)

				// Move to b2
				arc.b2Queue.PushFront(evictedKey)
				arc.cacheB2[evictedKey] = evictedValue
			}
		}
	}

	// Add the new item to t1
	arc.cacheT1[key] = value
	arc.t1Queue.PushFront(key)
	arc.t1Used += itemSize
	return true
}

// Len returns the number of bindings in the ARC.
func (arc *ARC) Len() int {
	return (arc.t1Queue.Len() + arc.t2Queue.Len())
}

// Stats returns statistics about how many search hits and misses have occurred.
func (arc *ARC) Stats() *Stats {
	return &arc.stats
}
