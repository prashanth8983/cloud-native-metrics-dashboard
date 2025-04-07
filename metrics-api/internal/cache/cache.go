package cache

import (
	"fmt"
	"sync"
	"time"
)

// Item represents a cached item with value and expiration time
type Item struct {
	Value      interface{}
	Expiration int64
	Created    int64
	LastAccess int64
}

// Expired checks if the item has expired
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// keyExpiration is used for sorting during eviction
type keyExpiration struct {
	key   string
	value int64 // This value depends on the eviction policy
}

// Cache is a thread-safe in-memory cache with expiration
type Cache struct {
	items             map[string]Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
	maxItems          int
	evictionPolicy    EvictionPolicy
	onEviction        func(string, interface{})
	statsEnabled      bool
	stats             Stats
}

// EvictionPolicy defines strategies for removing items when the cache is full
type EvictionPolicy string

const (
	// EvictLRU evicts the least recently used items
	EvictLRU EvictionPolicy = "LRU"
	// EvictLFU evicts the least frequently used items
	EvictLFU EvictionPolicy = "LFU"
	// EvictOldest evicts the oldest items
	EvictOldest EvictionPolicy = "OLDEST"
)

// Stats tracks cache usage statistics
type Stats struct {
	Hits             int64
	Misses           int64
	Evictions        int64
	CleanupRuns      int64
	ExpiredDeletions int64
}

// Options contains configuration for a new cache
type Options struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
	MaxItems          int
	EvictionPolicy    EvictionPolicy
	OnEviction        func(string, interface{})
	StatsEnabled      bool
}

// DefaultOptions returns default cache options
func DefaultOptions() Options {
	return Options{
		DefaultExpiration: 5 * time.Minute,
		CleanupInterval:   10 * time.Minute,
		MaxItems:          0, // No limit
		EvictionPolicy:    EvictLRU,
		OnEviction:        nil,
		StatsEnabled:      false,
	}
}

// New creates a new cache with the specified options
func New(options Options) *Cache {
	cache := &Cache{
		items:             make(map[string]Item),
		defaultExpiration: options.DefaultExpiration,
		cleanupInterval:   options.CleanupInterval,
		stopCleanup:       make(chan bool),
		maxItems:          options.MaxItems,
		evictionPolicy:    options.EvictionPolicy,
		onEviction:        options.OnEviction,
		statsEnabled:      options.StatsEnabled,
	}

	// Start cleanup goroutine if interval is positive
	if options.CleanupInterval > 0 {
		go cache.startCleanupTimer()
	}

	return cache
}

// Set adds an item to the cache with the default expiration
func (c *Cache) Set(key string, value interface{}) error {
	return c.SetWithExpiration(key, value, c.defaultExpiration)
}

// SetWithExpiration adds an item to the cache with a specific expiration
func (c *Cache) SetWithExpiration(key string, value interface{}, duration time.Duration) error {
	var expiration int64

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if cache is full and eviction is needed
	if c.maxItems > 0 && len(c.items) >= c.maxItems && c.items[key].Value == nil {
		if err := c.evict(1); err != nil {
			return err
		}
	}

	// Get the current time in nanoseconds
	now := time.Now().UnixNano()

	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
		Created:    now,
		LastAccess: now,
	}

	return nil
}

// Get retrieves an item from the cache
// The second return value indicates whether the key was found
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	if !found {
		c.mu.RUnlock()
		if c.statsEnabled {
			c.incrementMisses()
		}
		return nil, false
	}

	// Check if the item has expired
	if item.Expired() {
		c.mu.RUnlock()
		if c.statsEnabled {
			c.incrementMisses()
		}
		return nil, false
	}

	c.mu.RUnlock()

	// Update last access time
	c.mu.Lock()
	if it, found := c.items[key]; found {
		it.LastAccess = time.Now().UnixNano()
		c.items[key] = it
	}
	c.mu.Unlock()

	if c.statsEnabled {
		c.incrementHits()
	}

	return item.Value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, found := c.items[key]; found && c.onEviction != nil {
		c.onEviction(key, item.Value)
	}

	delete(c.items, key)
}

// Flush removes all items from the cache
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Call eviction callback if provided
	if c.onEviction != nil {
		for k, v := range c.items {
			c.onEviction(k, v.Value)
		}
	}

	c.items = make(map[string]Item)
}

// Count returns the number of items in the cache
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Has checks if a key exists in the cache and is not expired
func (c *Cache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return false
	}

	// Check if the item has expired
	return !item.Expired()
}

// GetAllKeys returns all keys in the cache
func (c *Cache) GetAllKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}

	return keys
}

// GetItem returns an item from the cache along with its metadata
// The second return value indicates whether the key was found
func (c *Cache) GetItem(key string) (Item, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		if c.statsEnabled {
			c.incrementMisses()
		}
		return Item{}, false
	}

	// Check if the item has expired
	if item.Expired() {
		if c.statsEnabled {
			c.incrementMisses()
		}
		return Item{}, false
	}

	if c.statsEnabled {
		c.incrementHits()
	}

	return item, true
}

// TTL returns the time to live for a cached item
// The second return value indicates whether the key was found
func (c *Cache) TTL(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return 0, false
	}

	// If item doesn't expire
	if item.Expiration == 0 {
		return 0, true
	}

	// Check if the item has expired
	now := time.Now().UnixNano()
	if now > item.Expiration {
		return 0, false
	}

	// Calculate remaining time
	return time.Duration(item.Expiration - now), true
}

// startCleanupTimer starts a timer to periodically clean up expired items
func (c *Cache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// deleteExpired deletes all expired items from the cache
func (c *Cache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.statsEnabled {
		c.stats.CleanupRuns++
	}

	var expiredCount int

	for k, v := range c.items {
		// Delete if expired
		if v.Expired() {
			// Call eviction callback if provided
			if c.onEviction != nil {
				c.onEviction(k, v.Value)
			}
			delete(c.items, k)
			expiredCount++
		}
	}

	if c.statsEnabled && expiredCount > 0 {
		c.stats.ExpiredDeletions += int64(expiredCount)
	}
}

// StopCleanup stops the cleanup goroutine
func (c *Cache) StopCleanup() {
	c.stopCleanup <- true
}

// evict removes items from the cache based on the eviction policy
func (c *Cache) evict(count int) error {
	if count <= 0 {
		return nil
	}

	if len(c.items) == 0 {
		return nil
	}

	// Make sure we don't evict more items than we have
	if count > len(c.items) {
		count = len(c.items)
	}

	// Prepare a slice of items to find eviction candidates
	candidates := make([]keyExpiration, 0, len(c.items))

	// Populate candidates based on eviction policy
	switch c.evictionPolicy {
	case EvictLRU:
		// Evict least recently used items (by last access time)
		for k, v := range c.items {
			candidates = append(candidates, keyExpiration{k, v.LastAccess})
		}
		// Sort candidates by last access time (oldest first)
		sortByValue(candidates, true)

	case EvictOldest:
		// Evict oldest items (by creation time)
		for k, v := range c.items {
			candidates = append(candidates, keyExpiration{k, v.Created})
		}
		// Sort candidates by creation time (oldest first)
		sortByValue(candidates, true)

	case EvictLFU:
		// Not implemented yet, fall back to LRU
		return fmt.Errorf("EvictLFU policy not implemented, please use EvictLRU or EvictOldest")

	default:
		return fmt.Errorf("unknown eviction policy: %s", c.evictionPolicy)
	}

	// Evict the required number of items
	for i := 0; i < count && i < len(candidates); i++ {
		key := candidates[i].key
		if item, found := c.items[key]; found {
			// Call eviction callback if provided
			if c.onEviction != nil {
				c.onEviction(key, item.Value)
			}
			delete(c.items, key)

			if c.statsEnabled {
				c.stats.Evictions++
			}
		}
	}

	return nil
}

// Helper function to sort key-expiration pairs
func sortByValue(candidates []keyExpiration, ascending bool) {
	if ascending {
		// Sort by value ascending (oldest/least accessed first)
		for i := 0; i < len(candidates)-1; i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[i].value > candidates[j].value {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}
	} else {
		// Sort by value descending (newest/most accessed first)
		for i := 0; i < len(candidates)-1; i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[i].value < candidates[j].value {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}
	}
}

// GetStats returns cache statistics
func (c *Cache) GetStats() Stats {
	if !c.statsEnabled {
		return Stats{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// incrementHits increments the hit counter
func (c *Cache) incrementHits() {
	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()
}

// incrementMisses increments the miss counter
func (c *Cache) incrementMisses() {
	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
}

// EnableStats enables statistics collection
func (c *Cache) EnableStats() {
	c.mu.Lock()
	c.statsEnabled = true
	c.mu.Unlock()
}

// DisableStats disables statistics collection
func (c *Cache) DisableStats() {
	c.mu.Lock()
	c.statsEnabled = false
	c.mu.Unlock()
}

// ResetStats resets all statistics counters
func (c *Cache) ResetStats() {
	c.mu.Lock()
	c.stats = Stats{}
	c.mu.Unlock()
}