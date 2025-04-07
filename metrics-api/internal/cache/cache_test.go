package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCacheBasicOperations(t *testing.T) {
	// Create a cache with default options
	cache := New(DefaultOptions())
	
	// Test Set and Get
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	
	if !found {
		t.Error("Expected to find key1")
	}
	
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
	
	// Test non-existent key
	value, found = cache.Get("nonexistent")
	if found {
		t.Error("Found nonexistent key")
	}
	
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
	
	// Test Has
	if !cache.Has("key1") {
		t.Error("Expected to find key1 with Has")
	}
	
	if cache.Has("nonexistent") {
		t.Error("Found nonexistent key with Has")
	}
	
	// Test Delete
	cache.Delete("key1")
	if cache.Has("key1") {
		t.Error("key1 not deleted")
	}
	
	// Test overwriting
	cache.Set("key2", "value2")
	cache.Set("key2", "new-value2")
	value, _ = cache.Get("key2")
	if value != "new-value2" {
		t.Errorf("Expected new-value2, got %v", value)
	}
	
	// Test Count
	cache.Set("key3", "value3")
	count := cache.Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	
	// Test Flush
	cache.Flush()
	count = cache.Count()
	if count != 0 {
		t.Errorf("Expected count 0 after flush, got %d", count)
	}
}

func TestCacheExpiration(t *testing.T) {
	// Create a cache with no automatic cleanup
	cache := New(Options{
		DefaultExpiration: 50 * time.Millisecond,
		CleanupInterval:   0, // No automatic cleanup
	})
	
	// Test with default expiration
	cache.Set("key1", "value1")
	
	// Test with custom expiration
	cache.SetWithExpiration("key2", "value2", 100*time.Millisecond)
	
	// Test with no expiration
	cache.SetWithExpiration("key3", "value3", 0)
	
	// Wait for key1 to expire, but key2 should still be valid
	time.Sleep(75 * time.Millisecond)
	
	// key1 should be expired
	_, found := cache.Get("key1")
	if found {
		t.Error("key1 should be expired")
	}
	
	// key2 should still be valid
	value, found := cache.Get("key2")
	if !found {
		t.Error("key2 should still be valid")
	}
	if value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}
	
	// Wait for key2 to expire
	time.Sleep(50 * time.Millisecond)
	
	// key2 should now be expired
	_, found = cache.Get("key2")
	if found {
		t.Error("key2 should be expired")
	}
	
	// key3 should never expire
	value, found = cache.Get("key3")
	if !found {
		t.Error("key3 should never expire")
	}
	if value != "value3" {
		t.Errorf("Expected value3, got %v", value)
	}
}

func TestCacheCleanup(t *testing.T) {
	// Create a cache with short cleanup interval
	cache := New(Options{
		DefaultExpiration: 50 * time.Millisecond,
		CleanupInterval:   100 * time.Millisecond,
	})
	
	// Set some items
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, fmt.Sprintf("value%d", i))
	}
	
	// Initial count should be 5
	count := cache.Count()
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
	
	// Wait for items to expire and cleanup to run
	time.Sleep(150 * time.Millisecond)
	
	// After cleanup, count should be 0
	count = cache.Count()
	if count != 0 {
		t.Errorf("Expected count 0 after cleanup, got %d", count)
	}
	
	// Make sure to stop the cleanup goroutine
	cache.StopCleanup()
}

func TestCacheTTL(t *testing.T) {
	// Create a cache
	cache := New(Options{
		DefaultExpiration: 1 * time.Hour,
	})
	
	// Set an item with 1 minute expiration
	cache.SetWithExpiration("key1", "value1", 1*time.Minute)
	
	// Get the TTL
	ttl, found := cache.TTL("key1")
	if !found {
		t.Error("Failed to get TTL for key1")
	}
	
	// TTL should be less than 1 minute (since some time has passed)
	if ttl > 1*time.Minute {
		t.Error("TTL too high")
	}
	
	// TTL should be more than 0
	if ttl <= 0 {
		t.Error("TTL too low")
	}
	
	// Get TTL for non-existent key
	_, found = cache.TTL("nonexistent")
	if found {
		t.Error("Found TTL for nonexistent key")
	}
	
	// Set an item with no expiration
	cache.SetWithExpiration("key2", "value2", 0)
	
	// TTL should be 0
	ttl, found = cache.TTL("key2")
	if !found {
		t.Error("Failed to get TTL for key2")
	}
	if ttl != 0 {
		t.Errorf("Expected TTL 0 for non-expiring item, got %v", ttl)
	}
}

func TestCacheGetAllKeys(t *testing.T) {
	// Create a cache
	cache := New(DefaultOptions())
	
	// Set some items
	expectedKeys := []string{"key1", "key2", "key3"}
	for _, key := range expectedKeys {
		cache.Set(key, key)
	}
	
	// Get all keys
	keys := cache.GetAllKeys()
	
	// Check count
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}
	
	// Check each key exists
	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key %s not found", expectedKey)
		}
	}
}

func TestCacheStats(t *testing.T) {
	// Create a cache with stats enabled
	cache := New(Options{
		DefaultExpiration: 1 * time.Hour,
		StatsEnabled:      true,
	})
	
	// Set some items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	
	// Get existing items (hits)
	cache.Get("key1")
	cache.Get("key2")
	cache.Get("key1")
	
	// Get non-existent items (misses)
	cache.Get("nonexistent1")
	cache.Get("nonexistent2")
	
	// Check stats
	stats := cache.GetStats()
	
	if stats.Hits != 3 {
		t.Errorf("Expected 3 hits, got %d", stats.Hits)
	}
	
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
	
	// Reset stats
	cache.ResetStats()
	
	// Check stats again
	stats = cache.GetStats()
	
	if stats.Hits != 0 {
		t.Errorf("Expected 0 hits after reset, got %d", stats.Hits)
	}
	
	if stats.Misses != 0 {
		t.Errorf("Expected 0 misses after reset, got %d", stats.Misses)
	}
	
	// Disable stats
	cache.DisableStats()
	
	// Get an item
	cache.Get("key1")
	
	// Stats shouldn't change
	stats = cache.GetStats()
	
	if stats.Hits != 0 {
		t.Errorf("Expected stats not to change when disabled, got %d hits", stats.Hits)
	}
}

func TestCacheEviction(t *testing.T) {
	// Track evicted keys
	evictedKeys := make([]string, 0)
	evictionCallback := func(key string, value interface{}) {
		evictedKeys = append(evictedKeys, key)
	}
	
	// Create a cache with max items and LRU policy
	cache := New(Options{
		DefaultExpiration: 1 * time.Hour,
		MaxItems:          3,
		EvictionPolicy:    EvictLRU,
		OnEviction:        evictionCallback,
	})
	
	// Add items up to the limit
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// Access key1 to make it most recently used
	cache.Get("key1")
	
	// Add another item, should evict the least recently used (key2 or key3)
	cache.Set("key4", "value4")
	
	// Check count
	count := cache.Count()
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
	
	// Check that either key2 or key3 was evicted
	if len(evictedKeys) != 1 {
		t.Errorf("Expected 1 eviction, got %d", len(evictedKeys))
	}
	
	evictedKey := evictedKeys[0]
	if evictedKey != "key2" && evictedKey != "key3" {
		t.Errorf("Expected key2 or key3 to be evicted, got %s", evictedKey)
	}
	
	// key1 should still be there (was used recently)
	if !cache.Has("key1") {
		t.Error("key1 should still be in cache")
	}
	
	// key4 should be there (just added)
	if !cache.Has("key4") {
		t.Error("key4 should be in cache")
	}
}

func TestCacheConcurrency(t *testing.T) {
	// Create a cache
	cache := New(DefaultOptions())
	
	// Number of goroutines
	numGoroutines := 100
	
	// Wait group for synchronization
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Run concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			// Each goroutine performs a mix of operations
			key := fmt.Sprintf("key%d", id)
			value := fmt.Sprintf("value%d", id)
			
			// Set
			cache.Set(key, value)
			
			// Get own key
			val, found := cache.Get(key)
			if !found || val != value {
				t.Errorf("Goroutine %d: Expected to find %s", id, key)
			}
			
			// Get other keys
			otherKey := fmt.Sprintf("key%d", (id+1)%numGoroutines)
			cache.Get(otherKey) // May or may not find it
			
			// Delete own key
			cache.Delete(key)
			
			// Check own key is gone
			if cache.Has(key) {
				t.Errorf("Goroutine %d: Key %s should be deleted", id, key)
			}
			
			// Set again
			cache.Set(key, value)
		}(i)
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Check final count
	count := cache.Count()
	if count != numGoroutines {
		t.Errorf("Expected final count %d, got %d", numGoroutines, count)
	}
}

func TestCacheGetItem(t *testing.T) {
	// Create a cache
	cache := New(DefaultOptions())
	
	// Set an item
	cache.Set("key1", "value1")
	
	// Get the item with metadata
	item, found := cache.GetItem("key1")
	
	if !found {
		t.Error("Expected to find key1")
	}
	
	if item.Value != "value1" {
		t.Errorf("Expected value1, got %v", item.Value)
	}
	
	if item.Created == 0 {
		t.Error("Expected created timestamp")
	}
	
	if item.LastAccess == 0 {
		t.Error("Expected last access timestamp")
	}
	
	// Item expiration should match default expiration
	now := time.Now().UnixNano()
	defaultExpiry := now + int64(DefaultOptions().DefaultExpiration)
	allowedDelta := int64(100 * time.Millisecond) // Allow 100ms delta for test timing
	
	if item.Expiration > 0 && (item.Expiration < defaultExpiry-allowedDelta || item.Expiration > defaultExpiry+allowedDelta) {
		t.Errorf("Unexpected expiration: %d vs default %d", item.Expiration, defaultExpiry)
	}
}

func TestCacheOldestEviction(t *testing.T) {
	// Track evicted keys
	evictedKeys := make([]string, 0)
	evictionCallback := func(key string, value interface{}) {
		evictedKeys = append(evictedKeys, key)
	}
	
	// Create a cache with max items and Oldest policy
	cache := New(Options{
		DefaultExpiration: 1 * time.Hour,
		MaxItems:          3,
		EvictionPolicy:    EvictOldest,
		OnEviction:        evictionCallback,
	})
	
	// Add items with delays
	cache.Set("key1", "value1")
	time.Sleep(10 * time.Millisecond)
	
	cache.Set("key2", "value2")
	time.Sleep(10 * time.Millisecond)
	
	cache.Set("key3", "value3")
	time.Sleep(10 * time.Millisecond)
	
	// Add another item, should evict the oldest (key1)
	cache.Set("key4", "value4")
	
	// Check that key1 was evicted
	if len(evictedKeys) != 1 {
		t.Errorf("Expected 1 eviction, got %d", len(evictedKeys))
	}
	
	if len(evictedKeys) > 0 && evictedKeys[0] != "key1" {
		t.Errorf("Expected key1 to be evicted, got %s", evictedKeys[0])
	}
	
	// key1 should be gone
	if cache.Has("key1") {
		t.Error("key1 should not be in cache")
	}
	
	// other keys should be there
	if !cache.Has("key2") || !cache.Has("key3") || !cache.Has("key4") {
		t.Error("key2, key3, and key4 should be in cache")
	}
}

func TestCacheItemExpired(t *testing.T) {
	// Create an item with expiration
	now := time.Now().UnixNano()
	
	// Item that expired 1 minute ago
	expiredItem := Item{
		Value:      "test",
		Expiration: now - int64(time.Minute),
	}
	
	if !expiredItem.Expired() {
		t.Error("Item should be expired")
	}
	
	// Item that expires 1 minute from now
	validItem := Item{
		Value:      "test",
		Expiration: now + int64(time.Minute),
	}
	
	if validItem.Expired() {
		t.Error("Item should not be expired")
	}
	
	// Item with no expiration
	noExpiryItem := Item{
		Value:      "test",
		Expiration: 0,
	}
	
	if noExpiryItem.Expired() {
		t.Error("Item with no expiration should never expire")
	}
}

func TestCacheExpirationWithCleanup(t *testing.T) {
	// Create a cache with cleanup 
	cache := New(Options{
		DefaultExpiration: 50 * time.Millisecond,
		CleanupInterval:   100 * time.Millisecond,
		StatsEnabled:      true,
	})
	
	// Set many items
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, fmt.Sprintf("value%d", i))
	}
	
	// Wait for cleanup to run
	time.Sleep(150 * time.Millisecond)
	
	// Check that all items are gone
	count := cache.Count()
	if count != 0 {
		t.Errorf("Expected count 0 after cleanup, got %d", count)
	}
	
	// Check stats
	stats := cache.GetStats()
	if stats.CleanupRuns == 0 {
		t.Error("No cleanup runs recorded")
	}
	
	if stats.ExpiredDeletions == 0 {
		t.Error("No expired deletions recorded")
	}
	
	// Don't forget to stop the cleanup goroutine
	cache.StopCleanup()
}

// Benchmark Set
func BenchmarkCacheSet(b *testing.B) {
	cache := New(DefaultOptions())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, "value")
	}
}

// Benchmark Get
func BenchmarkCacheGet(b *testing.B) {
	cache := New(DefaultOptions())
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, "value")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000) // Cycle through existing keys
		cache.Get(key)
	}
}

// Benchmark Set and Get with concurrent access
func BenchmarkCacheConcurrent(b *testing.B) {
	cache := New(DefaultOptions())
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			id++
			key := fmt.Sprintf("key%d", id)
			cache.Set(key, "value")
			cache.Get(key)
		}
	})
}