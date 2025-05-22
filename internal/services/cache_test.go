package services

import (
	"sync"
	"testing"
)

func TestLRUCache_Basic(t *testing.T) {
	cache := NewLRUCache(2)
	
	// Test Put and Get
	cache.Put("key1", "value1")
	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Errorf("Expected value1, got %s, found: %v", value, found)
	}
	
	// Test non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Expected key not to be found")
	}
}

func TestLRUCache_Capacity(t *testing.T) {
	cache := NewLRUCache(2)
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3") // Should evict key1
	
	// key1 should be evicted
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be evicted")
	}
	
	// key2 and key3 should still exist
	value, found := cache.Get("key2")
	if !found || value != "value2" {
		t.Errorf("Expected value2, got %s, found: %v", value, found)
	}
	
	value, found = cache.Get("key3")
	if !found || value != "value3" {
		t.Errorf("Expected value3, got %s, found: %v", value, found)
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(2)
	
	cache.Put("key1", "value1")
	cache.Put("key1", "updated_value1") // Update existing key
	
	value, found := cache.Get("key1")
	if !found || value != "updated_value1" {
		t.Errorf("Expected updated_value1, got %s, found: %v", value, found)
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(2)
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	
	// Access key1 to make it most recently used
	cache.Get("key1")
	
	// Add key3, should evict key2 (least recently used)
	cache.Put("key3", "value3")
	
	// key2 should be evicted
	_, found := cache.Get("key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}
	
	// key1 should still exist
	_, found = cache.Get("key1")
	if !found {
		t.Error("Expected key1 to still exist")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	cache := NewLRUCache(100)
	var wg sync.WaitGroup
	
	// Test concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := string(rune('a' + index))
			value := string(rune('A' + index))
			cache.Put(key, value)
		}(i)
	}
	
	// Test concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := string(rune('a' + index))
			cache.Get(key)
		}(i)
	}
	
	wg.Wait()
	// If we get here without deadlock, the test passes
}

func TestLRUCache_EmptyCache(t *testing.T) {
	cache := NewLRUCache(5)
	
	_, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected key not to be found in empty cache")
	}
}

func TestLRUCache_ZeroCapacity(t *testing.T) {
	cache := NewLRUCache(0)
	
	cache.Put("key1", "value1")
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key not to be found with zero capacity")
	}
}