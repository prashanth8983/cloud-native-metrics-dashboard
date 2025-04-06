// internal/cache/cache_test.go
package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_Set_Get(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Test setting and getting a value
	c.Set("key1", "value1")
	value, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1, but it was not found")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test getting a non-existent key
	_, found = c.Get("key2")
	if found {
		t.Error("Did not expect to find key2, but it was found")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Test setting with a short TTL
	c.SetWithTTL("key1", "value1", 100*time.Millisecond)
	
	// Verify the item exists before expiration
	value, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1 before expiration, but it was not found")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Wait for the item to expire
	time.Sleep(200 * time.Millisecond)

	// Verify the item is now gone
	_, found = c.Get("key1")
	if found {
		t.Error("Expected key1 to be expired, but it was found")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Set a value
	c.Set("key1", "value1")
	
	// Verify it exists
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1, but it was not found")
	}

	// Delete it
	c.Delete("key1")

	// Verify it's gone
	_, found = c.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted, but it was found")
	}

	// Delete a non-existent key (should not panic)
	c.Delete("key2")
}

func TestCache_Clear(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Set multiple values
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	// Verify they exist
	if count := c.Count(); count != 3 {
		t.Errorf("Expected 3 items, got %d", count)
	}

	// Clear the cache
	c.Clear()

	// Verify it's empty
	if count := c.Count(); count != 0 {
		t.Errorf("Expected 0 items after clear, got %d", count)
	}
}

func TestCache_Count(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Initial count should be 0
	if count := c.Count(); count != 0 {
		t.Errorf("Expected initial count to be 0, got %d", count)
	}

	// Add items and check count
	c.Set("key1", "value1")
	if count := c.Count(); count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}

	c.Set("key2", "value2")
	if count := c.Count(); count != 2 {
		t.Errorf("Expected count to be 2, got %d", count)
	}

	// Delete an item and check count
	c.Delete("key1")
	if count := c.Count(); count != 1 {
		t.Errorf("Expected count to be 1 after delete, got %d", count)
	}

	// Clear and check count
	c.Clear()
	if count := c.Count(); count != 0 {
		t.Errorf("Expected count to be 0 after clear, got %d", count)
	}
}

func TestCache_Items(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Initial items should be empty
	items := c.Items()
	if len(items) != 0 {
		t.Errorf("Expected initial items to be empty, got %d items", len(items))
	}

	// Add items and check
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	
	items = c.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	
	if items["key1"] != "value1" {
		t.Errorf("Expected items[key1] to be value1, got %v", items["key1"])
	}
	
	if items["key2"] != "value2" {
		t.Errorf("Expected items[key2] to be value2, got %v", items["key2"])
	}

	// Add item with expired TTL
	c.SetWithTTL("key3", "value3", 100*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	
	// Should not include expired items
	items = c.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 items (excluding expired), got %d", len(items))
	}
	
	if _, found := items["key3"]; found {
		t.Error("Expected key3 to be excluded as it's expired")
	}
}

func TestCache_Keys(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Initial keys should be empty
	keys := c.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected initial keys to be empty, got %d keys", len(keys))
	}

	// Add items and check
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	
	keys = c.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
	
	// Check keys are present (order is not guaranteed)
	foundKey1 := false
	foundKey2 := false
	for _, k := range keys {
		if k == "key1" {
			foundKey1 = true
		}
		if k == "key2" {
			foundKey2 = true
		}
	}
	
	if !foundKey1 {
		t.Error("Expected to find key1 in keys, but it was not found")
	}
	
	if !foundKey2 {
		t.Error("Expected to find key2 in keys, but it was not found")
	}

	// Add item with expired TTL
	c.SetWithTTL("key3", "value3", 100*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	
	// Should not include expired keys
	keys = c.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys (excluding expired), got %d", len(keys))
	}
	
	for _, k := range keys {
		if k == "key3" {
			t.Error("Expected key3 to be excluded as it's expired")
		}
	}
}

func TestCache_DeleteExpired(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Add both normal and soon-to-expire items
	c.Set("key1", "value1")
	c.SetWithTTL("key2", "value2", 100*time.Millisecond)
	
	// Wait for key2 to expire
	time.Sleep(200 * time.Millisecond)
	
	// Before calling DeleteExpired, key2 should still be in the internal map
	// but shouldn't be retrievable via Get
	if c.Count() != 2 {
		t.Errorf("Expected count of 2 before DeleteExpired, got %d", c.Count())
	}
	
	_, found := c.Get("key2")
	if found {
		t.Error("Expected key2 to be expired, but it was found")
	}
	
	// Call DeleteExpired
	c.DeleteExpired()
	
	// Now the internal count should match the actual count
	if c.Count() != 1 {
		t.Errorf("Expected count of 1 after DeleteExpired, got %d", c.Count())
	}
}

func TestCache_MaxItems(t *testing.T) {
	// Create a cache with max 3 items
	c := New(5*time.Minute, 3, 0)

	// Add items up to the limit
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	
	// Verify they all exist
	if count := c.Count(); count != 3 {
		t.Errorf("Expected 3 items, got %d", count)
	}
	
	// Add another item, should evict the oldest one (key1)
	c.Set("key4", "value4")
	
	// Verify count is still 3
	if count := c.Count(); count != 3 {
		t.Errorf("Expected 3 items after exceeding max, got %d", count)
	}
	
	// Verify key1 was evicted
	_, found := c.Get("key1")
	if found {
		t.Error("Expected key1 to be evicted, but it was found")
	}
	
	// Verify newer keys still exist
	for _, key := range []string{"key2", "key3", "key4"} {
		_, found := c.Get(key)
		if !found {
			t.Errorf("Expected %s to exist, but it was not found", key)
		}
	}
}

func TestCache_UpdateTTL(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Set the initial value
	c.Set("key1", "value1")
	
	// Verify it exists
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1, but it was not found")
	}
	
	// Update TTL to a shorter value
	c.UpdateTTL(100 * time.Millisecond)
	
	// Set a new value with the updated TTL
	c.Set("key2", "value2")
	
	// Verify both keys exist
	_, found = c.Get("key1")
	if !found {
		t.Error("Expected to find key1, but it was not found")
	}
	
	_, found = c.Get("key2")
	if !found {
		t.Error("Expected to find key2, but it was not found")
	}
	
	// Wait for the new TTL to expire
	time.Sleep(200 * time.Millisecond)
	
	// key1 should still exist (it uses the old TTL), but key2 should be gone
	_, found = c.Get("key1")
	if !found {
		t.Error("Expected to find key1, but it was not found")
	}
	
	_, found = c.Get("key2")
	if found {
		t.Error("Expected key2 to be expired, but it was found")
	}
}

func TestCache_UpdateMaxItems(t *testing.T) {
	// Create a cache with max 10 items
	c := New(5*time.Minute, 10, 0)

	// Add 5 items
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		c.Set(key, fmt.Sprintf("value%d", i))
	}
	
	// Verify all 5 exist
	if count := c.Count(); count != 5 {
		t.Errorf("Expected 5 items, got %d", count)
	}
	
	// Update max items to 3
	c.UpdateMaxItems(3)
	
	// Add another item, should evict the oldest ones
	c.Set("key5", "value5")
	
	// Verify count is now 3
	if count := c.Count(); count != 3 {
		t.Errorf("Expected 3 items after updating max, got %d", count)
	}
	
	// Verify newer keys still exist
	for i := 3; i <= 5; i++ {
		key := fmt.Sprintf("key%d", i)
		_, found := c.Get(key)
		if !found {
			t.Errorf("Expected %s to exist, but it was not found", key)
		}
	}
}

func TestCache_Janitor(t *testing.T) {
	// Create a cache with a janitor that runs every 100ms
	c := New(5*time.Minute, 100, 100*time.Millisecond)
	defer c.Close()

	// Add an item with a short TTL
	c.SetWithTTL("key1", "value1", 50*time.Millisecond)
	
	// Verify it exists
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1, but it was not found")
	}
	
	// Wait for the janitor to clean up the expired item
	time.Sleep(200 * time.Millisecond)
	
	// Verify internal count is updated
	if count := c.Count(); count != 0 {
		t.Errorf("Expected 0 items after janitor cleanup, got %d", count)
	}
}

func TestCache_Cached(t *testing.T) {
	c := New(5*time.Minute, 100, 0)

	// Function to cache
	callCount := 0
	fn := func(args ...interface{}) (interface{}, error) {
		callCount++
		if len(args) > 0 {
			return fmt.Sprintf("result-%v", args[0]), nil
		}
		return "default-result", nil
	}

	// Key function
	keyFn := func(args ...interface{}) string {
		if len(args) > 0 {
			return fmt.Sprintf("key-%v", args[0])
		}
		return "default-key"
	}

	// Create cached function
	cachedFn := c.Cached(keyFn, fn)

	// First call should execute the function
	result, err := cachedFn("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "result-test" {
		t.Errorf("Expected result-test, got %v", result)
	}
	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d", callCount)
	}

	// Second call with same args should use cache
	result, err = cachedFn("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "result-test" {
		t.Errorf("Expected result-test, got %v", result)
	}
	if callCount != 1 {
		t.Errorf("Expected function to still be called once (cache hit), got %d", callCount)
	}

	// Call with different args should execute the function again
	result, err = cachedFn("other")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "result-other" {
		t.Errorf("Expected result-other, got %v", result)
	}
	if callCount != 2 {
		t.Errorf("Expected function to be called twice, got %d", callCount)
	}
}