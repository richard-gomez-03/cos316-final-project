package cache

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

// Test creation of new LIFO cache
func TestNewLifo(t *testing.T) {
	cache := NewLifo(1024)

	if cache == nil {
		t.Errorf("NewLifo() = nil, want non-nil")
	}
	if cache.MaxStorage() != 1024 {
		t.Errorf("MaxStorage() = %d, want 1024", cache.MaxStorage())
	}
}

// Test setting and getting values in LIFO cache
func TestLifoSetAndGet(t *testing.T) {
	cache := NewLifo(1024)

	key := "test_key"
	value := []byte("test_value")

	cache.Set(key, value)

	got, ok := cache.Get(key)
	if !ok || !bytes.Equal(got, value) {
		t.Errorf("Get(%s) = %v, %t; want %v, true", key, got, ok, value)
	}
}

// Test LIFO eviction policy
func TestLifoEvictionPolicy(t *testing.T) {
	cache := NewLifo(10)

	cache.Set("key1", []byte("val1"))
	cache.Set("key2", []byte("val2"))
	cache.Set("key3", []byte("val3")) // Should evict key1

	_, ok := cache.Get("key1")
	if ok {
		t.Errorf("Get(key1) should be false after eviction")
	}
}

// Test performance of LIFO cache with various sizes: 1KB, 1MB, 10MB, 32MB, 64MB, 500MB, and 1GB
func TestLIFOPerformance(t *testing.T) {
	sizes := []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewLifo(size)

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

// Test memory usage of LIFO cache with various sizes: 1KB, 1MB, 10MB, 32MB, 64MB
func TestLIFOMemoryUsage(t *testing.T) {
	for _, size := range []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024} {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewLifo(size)

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

// Test hit rate of LIFO cache of various sizes
func TestLIFOHitRate(t *testing.T) {
	sizes := []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewLifo(size)
			totalOps := 0
			hits := 0

			// Populate cache
			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("key%d", i)
				value := []byte(fmt.Sprintf("value%d", i))
				cache.Set(key, value)
				totalOps++
			}

			// Test hit rate
			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("key%d", rand.Intn(1000)) // Accessing random keys
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
