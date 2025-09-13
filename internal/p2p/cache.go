package p2p

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with TTL
type CacheEntry struct {
	Value     []byte
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// LRUCache is a simple LRU cache with TTL support
type LRUCache struct {
	maxSize  int
	ttl      time.Duration
	entries  map[string]*CacheEntry
	order    []string // LRU order, most recent at the end
	mutex    sync.RWMutex
	cleanup  *time.Ticker
	done     chan struct{}
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
	cache := &LRUCache{
		maxSize: maxSize,
		ttl:     ttl,
		entries: make(map[string]*CacheEntry),
		order:   make([]string, 0, maxSize),
		done:    make(chan struct{}),
	}
	
	// Start cleanup routine if TTL is set
	if ttl > 0 {
		cache.cleanup = time.NewTicker(ttl / 2) // Cleanup twice per TTL period
		go cache.cleanupRoutine()
	}
	
	return cache
}

// Close stops the cache cleanup routine
func (c *LRUCache) Close() {
	if c.cleanup != nil {
		c.cleanup.Stop()
	}
	close(c.done)
}

// Set adds or updates a cache entry
func (c *LRUCache) Set(key string, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Create new entry
	entry := &CacheEntry{
		Value:     make([]byte, len(value)),
		ExpiresAt: time.Now().Add(c.ttl),
	}
	copy(entry.Value, value)
	
	// Check if key already exists
	if _, exists := c.entries[key]; exists {
		// Update existing entry
		c.entries[key] = entry
		c.moveToEnd(key)
		return
	}
	
	// Add new entry
	c.entries[key] = entry
	c.order = append(c.order, key)
	
	// Evict if necessary
	if len(c.entries) > c.maxSize {
		c.evictLRU()
	}
}

// Get retrieves a cache entry
func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	
	// Check if expired
	if entry.IsExpired() {
		c.deleteEntry(key)
		return nil, false
	}
	
	// Move to end (most recently used)
	c.moveToEnd(key)
	
	// Return copy of value
	result := make([]byte, len(entry.Value))
	copy(result, entry.Value)
	return result, true
}

// Has checks if a key exists in the cache
func (c *LRUCache) Has(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return false
	}
	
	return !entry.IsExpired()
}

// Delete removes a cache entry
func (c *LRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.deleteEntry(key)
}

// Size returns the current cache size
func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return len(c.entries)
}

// Clear removes all cache entries
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.entries = make(map[string]*CacheEntry)
	c.order = c.order[:0]
}

// Stats returns cache statistics
func (c *LRUCache) Stats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	expired := 0
	now := time.Now()
	
	for _, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			expired++
		}
	}
	
	return map[string]interface{}{
		"size":        len(c.entries),
		"max_size":    c.maxSize,
		"expired":     expired,
		"ttl_seconds": c.ttl.Seconds(),
	}
}

// moveToEnd moves a key to the end of the LRU order
func (c *LRUCache) moveToEnd(key string) {
	// Find and remove key from current position
	for i, k := range c.order {
		if k == key {
			// Remove from current position
			copy(c.order[i:], c.order[i+1:])
			c.order = c.order[:len(c.order)-1]
			break
		}
	}
	
	// Add to end
	c.order = append(c.order, key)
}

// deleteEntry removes an entry from cache and order
func (c *LRUCache) deleteEntry(key string) {
	// Remove from entries map
	delete(c.entries, key)
	
	// Remove from order slice
	for i, k := range c.order {
		if k == key {
			copy(c.order[i:], c.order[i+1:])
			c.order = c.order[:len(c.order)-1]
			break
		}
	}
}

// evictLRU removes the least recently used entry
func (c *LRUCache) evictLRU() {
	if len(c.order) == 0 {
		return
	}
	
	// Remove oldest entry (first in order)
	oldest := c.order[0]
	c.deleteEntry(oldest)
}

// cleanupRoutine periodically removes expired entries
func (c *LRUCache) cleanupRoutine() {
	for {
		select {
		case <-c.cleanup.C:
			c.cleanupExpired()
		case <-c.done:
			return
		}
	}
}

// cleanupExpired removes all expired entries
func (c *LRUCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	var keysToDelete []string
	
	// Find expired keys
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	// Delete expired keys
	for _, key := range keysToDelete {
		c.deleteEntry(key)
	}
}