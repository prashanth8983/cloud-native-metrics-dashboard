// internal/cache/cache.go
package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with expiration
type Item struct {
	Value      interface{}
	Expiration int64
	Created    time.Time
}

// Cache represents a thread-safe cache with expiring items
type Cache struct {
	items       map[string]Item
	mu          sync.RWMutex
	ttl         time.Duration
	maxItems    int
	janitorStop chan bool
}

// New creates a new cache with the specified TTL and maximum items
func New(ttl time.Duration, maxItems int, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		items:       make(map[string]Item),
		ttl:         ttl,
		maxItems:    maxItems,
		janitorStop: make(chan bool),
	}

	// Start the cleanup goroutine if cleanupInterval is positive
	if cleanupInterval > 0 {
		go cache.janitor(cleanupInterval)
	}

	return cache
}

// Set adds an item to the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL adds an item to the cache with a specific TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate expiration time
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	// If we've reached the maximum items, remove the oldest entries
	if c.maxItems > 0 && len(c.items) >= c.maxItems {
		// Find the oldest item
		var oldestKey string
		var oldestTime time.Time

		// Initialize with the first item
		for k, v := range c.items {
			oldestKey = k
			oldestTime = v.Created
			break
		}

		// Find the oldest item
		for k, v := range c.items {
			if v.Created.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.Created
			}
		}

		// Remove oldest item
		delete(c.items, oldestKey)
	}

	// Store the item
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}
}

// Get retrieves an item from the cache
// The second return value indicates whether the key was found
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]Item)
}

// Count returns the number of items in the cache
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Items returns a copy of all unexpired items in the cache
func (c *Cache) Items() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make(map[string]interface{}, len(c.items))
	now := time.Now().UnixNano()

	for k, v := range c.items {
		// Check expiration
		if v.Expiration > 0 && now > v.Expiration {
			continue
		}
		items[k] = v.Value
	}

	return items
}

// Keys returns a slice of all unexpired keys in the cache
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	now := time.Now().UnixNano()

	for k, v := range c.items {
		// Check expiration
		if v.Expiration > 0 && now > v.Expiration {
			continue
		}
		keys = append(keys, k)
	}

	return keys
}

// janitor runs on a separate goroutine and periodically removes expired items
func (c *Cache) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.janitorStop:
			return
		}
	}
}

// DeleteExpired removes all expired items from the cache
func (c *Cache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
}

// Close stops the janitor goroutine
func (c *Cache) Close() {
	close(c.janitorStop)
}

// UpdateTTL updates the default TTL for new items
func (c *Cache) UpdateTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
}

// UpdateMaxItems updates the maximum number of items the cache can hold
func (c *Cache) UpdateMaxItems(maxItems int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxItems = maxItems
}

// Cached creates a caching wrapper for a function that returns a value and error
// The wrapped function will cache the result using the provided key function
func (c *Cache) Cached(keyFn func(args ...interface{}) string, fn func(args ...interface{}) (interface{}, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		// Generate cache key
		key := keyFn(args...)

		// Try to get from cache
		if value, found := c.Get(key); found {
			return value, nil
		}

		// Call the original function
		value, err := fn(args...)
		if err != nil {
			return nil, err
		}

		// Cache the result
		c.Set(key, value)
		return value, nil
	}
}