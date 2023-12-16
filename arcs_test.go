package cache

import (
	"bytes"
	"os"
	"testing"
)

// TestNewARC tests the creation of a new ARC.
func TestNewARC(t *testing.T) {
	arc := NewArc(100)
	if arc == nil {
		t.Error("Expected NewArc to create a new ARC, got nil")
	}
	if arc.MaxStorage() != 100 {
		t.Errorf("Expected MaxStorage to be 100, got %d", arc.MaxStorage())
	}
}

// TestSetAndGet tests setting and getting values in the ARC.
func TestSetAndGet(t *testing.T) {
	arc := NewArc(100)
	key := "testKey"
	value := []byte("testValue")

	success := arc.Set(key, value)
	if !success {
		t.Errorf("Failed to set value for key %s", key)
	}

	retrievedValue, ok := arc.Get(key)
	if !ok {
		t.Errorf("Failed to get value for key %s", key)
	}
	if !bytes.Equal(retrievedValue, value) {
		t.Errorf("Expected value %s, got %s", string(value), string(retrievedValue))
	}
}

// TestEviction tests the eviction policy of the ARC.
func TestEviction(t *testing.T) {
	arc := NewArc(10) // Small size to test eviction
	arc.Set("key1", []byte("val1"))
	arc.Set("key2", []byte("val2"))
	arc.Set("key3", []byte("val3")) // This should cause an eviction

	_, ok := arc.Get("key1")
	if ok {
		t.Error("Expected key1 to be evicted")
	}
}

// TestRemove tests the Remove function.
func TestRemove(t *testing.T) {
	arc := NewArc(100)
	key := "testKey"
	value := []byte("testValue")
	arc.Set(key, value)

	removedValue, ok := arc.Remove(key)
	if !ok {
		t.Errorf("Failed to remove value for key %s", key)
	}
	if !bytes.Equal(removedValue, value) {
		t.Errorf("Expected value %s, got %s", string(value), string(removedValue))
	}

	_, ok = arc.Get(key)
	if ok {
		t.Error("Expected value to be removed from the cache")
	}
}

// TestLen tests the Len function.
func TestLen(t *testing.T) {
	arc := NewArc(100)
	arc.Set("key1", []byte("val1"))
	arc.Set("key2", []byte("val2"))

	if arc.Len() != 2 {
		t.Errorf("Expected length to be 2, got %d", arc.Len())
	}
}

// TestStats tests the Stats function.
func TestStats(t *testing.T) {
	arc := NewArc(100)
	arc.Get("nonexistent")
	stats := arc.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestMain(m *testing.M) {
	// setup if needed
	code := m.Run()
	// teardown if needed
	os.Exit(code)
}
