package cache

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

// Test creation of new RR cache
func TestNewRR(t *testing.T) {
	cache := NewRR(1024)

	if cache == nil {
		t.Errorf("NewRR() = nil, want non-nil")
	}
	if cache.capacity != 1024 {
		t.Errorf("capacity = %d, want 1024", cache.capacity)
	}
}

// Test setting and getting values in RR cache
func TestRRSetAndGet(t *testing.T) {
	cache := NewRR(1024)

	key := "test_key"
	value := []byte("test_value")

	cache.Set(key, value)

	got, ok := cache.Get(key)
	if !ok || !bytes.Equal(got, value) {
		t.Errorf("Get(%s) = %v, %t; want %v, true", key, got, ok, value)
	}
}

// Test RR eviction policy
func TestRREvictionPolicy(t *testing.T) {
	cache := NewRR(10)

	cache.Set("key1", []byte("val1"))
	cache.Set("key2", []byte("val2"))
	cache.Set("key3", []byte("val3")) // This may evict any of the previous keys

	if cache.Len() > cache.capacity {
		t.Errorf("Cache size after eviction should not exceed capacity")
	}
}

// Test performance of RR cache with varoius sizes: 1KB, 1MB, 10MB, 32MB, 64MB, 500MB, and 1GB
func TestRRPerformance(t *testing.T) {
	sizes := []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewRR(size)

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

// Test memory usage of RR cache with various sizes
func TestRRMemoryUsageWithDifferentSizes(t *testing.T) {
	for _, size := range []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024} {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewRR(size)

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

// Tets hit rate of RR cache for various sizes
func TestRRHitRate(t *testing.T) {
	sizes := []int{1024, 1024 * 1024, 10 * 1024 * 1024, 32 * 1024 * 1024, 64 * 1024 * 1024}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			cache := NewRR(size)
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
