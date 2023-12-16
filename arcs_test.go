package cache

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

// Test creation of new ARC cache
func TestNewARC(t *testing.T) {
	cache := NewArc(1024)

	if cache == nil {
		t.Errorf("NewArc() = nil, want non-nil")
	}
	if cache.MaxStorage() != 1024 {
		t.Errorf("MaxStorage() = %d, want 1024", cache.MaxStorage())
	}
}

// Test setting and getting values in ARC cache
func TestARCSetAndGet(t *testing.T) {
	cache := NewArc(1024)

	key := "test_key"
	value := []byte("test_value")

	cache.Set(key, value)

	got, ok := cache.Get(key)
	if !ok || !bytes.Equal(got, value) {
		t.Errorf("Get(%s) = %v, %t; want %v, true", key, got, ok, value)
	}
}

// Test eviction policy
func TestARCEvictionPolicy(t *testing.T) {
	cache := NewArc(10)

	cache.Set("key1", []byte("val1"))
	cache.Set("key2", []byte("val2"))
	cache.Set("key3", []byte("val3"))

	_, ok := cache.Get("key1")
	if ok {
		t.Errorf("Get(key1) should be false after eviction")
	}
}

// Test ARC cache performance with various sizes
func TestARCPerformance(t *testing.T) {
	sizes := []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewArc(size)

			start := time.Now()

			for i := 0; i < 10000; i++ {
				key := fmt.Sprintf("key%d", i)
				value := []byte(fmt.Sprintf("value%d", i))
				cache.Set(key, value)
			}

			duration := time.Since(start)
			t.Logf("Size: %d, Performance test took %s", size, duration)
		})
	}
}

// Test memory usage for ARC cache with various sizes
func TestARCMemoryUsage(t *testing.T) {
	for _, size := range []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024} {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewArc(size)

			var m runtime.MemStats
			measureMemory := func() uint64 {
				runtime.GC()
				runtime.ReadMemStats(&m)
				return m.Alloc
			}

			before := measureMemory()

			for i := 0; i < 10000; i++ {
				key := fmt.Sprintf("key%d", i)
				value := []byte(fmt.Sprintf("value%d", i))
				cache.Set(key, value)
			}

			after := measureMemory()

			t.Logf("Size: %d, Memory usage before: %d, after: %d", size, before, after)
		})
	}
}

// Test hit rate of ARC cache with various sizes
func TestARCHitRate(t *testing.T) {
	sizes := []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewArc(size)
			totalOps := 0
			hits := 0

			// Populate cache and perform get operations
			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("key%d", i)
				value := []byte(fmt.Sprintf("value%d", i))
				cache.Set(key, value)
				totalOps++
			}

			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("key%d", rand.Intn(1000))
				if _, ok := cache.Get(key); ok {
					hits++
				}
				totalOps++
			}

			hitRate := float64(hits) / float64(totalOps)
			t.Logf("Size: %d, Hit Rate: %.2f", size, hitRate)
		})
	}
}
