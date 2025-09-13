package statuslist

import (
	"sync"
	"time"
)

// InMemoryStatusListCache provides in-memory caching for status lists
type InMemoryStatusListCache struct {
	mutex  sync.RWMutex
	cache  map[string]*cacheEntry
	stats  CacheStats
	maxSize int
}

type cacheEntry struct {
	list      *StatusList2021
	expiry    time.Time
	lastUsed  time.Time
}

// NewInMemoryStatusListCache creates a new in-memory cache
func NewInMemoryStatusListCache(maxSize int) *InMemoryStatusListCache {
	if maxSize <= 0 {
		maxSize = 100 // Default max size
	}
	
	return &InMemoryStatusListCache{
		cache:   make(map[string]*cacheEntry),
		maxSize: maxSize,
		stats: CacheStats{
			MaxSize: maxSize,
		},
	}
}

// Get retrieves a cached status list
func (c *InMemoryStatusListCache) Get(listID string) (*StatusList2021, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	entry, exists := c.cache[listID]
	if !exists {
		c.stats.Misses++
		c.updateHitRatio()
		return nil, NewStatusListError(ErrorListNotFound, "not in cache")
	}
	
	// Check expiry
	if time.Now().After(entry.expiry) {
		delete(c.cache, listID)
		c.stats.Size--
		c.stats.Misses++
		c.updateHitRatio()
		return nil, NewStatusListError(ErrorListNotFound, "cache entry expired")
	}
	
	// Update last used time
	entry.lastUsed = time.Now()
	c.stats.Hits++
	c.updateHitRatio()
	
	// Return a clone to prevent modification
	return c.cloneStatusList(entry.list), nil
}

// Set stores a status list in cache
func (c *InMemoryStatusListCache) Set(listID string, list *StatusList2021, ttl time.Duration) error {
	if list == nil {
		return NewStatusListError(ErrorInvalidStatusList, "cannot cache nil status list")
	}
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if we need to evict entries
	if len(c.cache) >= c.maxSize {
		c.evictLRU()
	}
	
	// Create cache entry
	entry := &cacheEntry{
		list:     c.cloneStatusList(list),
		expiry:   time.Now().Add(ttl),
		lastUsed: time.Now(),
	}
	
	// Update size if it's a new entry
	if _, exists := c.cache[listID]; !exists {
		c.stats.Size++
	}
	
	c.cache[listID] = entry
	return nil
}

// Invalidate removes a status list from cache
func (c *InMemoryStatusListCache) Invalidate(listID string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if _, exists := c.cache[listID]; exists {
		delete(c.cache, listID)
		c.stats.Size--
	}
	
	return nil
}

// Clear removes all cached status lists
func (c *InMemoryStatusListCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache = make(map[string]*cacheEntry)
	c.stats.Size = 0
	c.stats.Hits = 0
	c.stats.Misses = 0
	c.stats.HitRatio = 0
	
	return nil
}

// Stats returns cache statistics
func (c *InMemoryStatusListCache) Stats() *CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return &CacheStats{
		Hits:     c.stats.Hits,
		Misses:   c.stats.Misses,
		Size:     c.stats.Size,
		MaxSize:  c.stats.MaxSize,
		HitRatio: c.stats.HitRatio,
	}
}

// CleanupExpired removes expired entries from cache
func (c *InMemoryStatusListCache) CleanupExpired() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	expired := make([]string, 0)
	
	for listID, entry := range c.cache {
		if now.After(entry.expiry) {
			expired = append(expired, listID)
		}
	}
	
	for _, listID := range expired {
		delete(c.cache, listID)
		c.stats.Size--
	}
	
	return len(expired)
}

// StartCleanupTimer starts a background timer to clean up expired entries
func (c *InMemoryStatusListCache) StartCleanupTimer(interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	
	go func() {
		for range ticker.C {
			c.CleanupExpired()
		}
	}()
	
	return ticker
}

func (c *InMemoryStatusListCache) evictLRU() {
	if len(c.cache) == 0 {
		return
	}
	
	// Find the least recently used entry
	var oldestID string
	var oldestTime time.Time
	first := true
	
	for listID, entry := range c.cache {
		if first || entry.lastUsed.Before(oldestTime) {
			oldestID = listID
			oldestTime = entry.lastUsed
			first = false
		}
	}
	
	// Remove the oldest entry
	if oldestID != "" {
		delete(c.cache, oldestID)
		c.stats.Size--
	}
}

func (c *InMemoryStatusListCache) updateHitRatio() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRatio = float64(c.stats.Hits) / float64(total)
	} else {
		c.stats.HitRatio = 0
	}
}

func (c *InMemoryStatusListCache) cloneStatusList(list *StatusList2021) *StatusList2021 {
	return &StatusList2021{
		Context:      append([]string{}, list.Context...),
		ID:           list.ID,
		Type:         append([]string{}, list.Type...),
		Issuer:       list.Issuer,
		IssuanceDate: list.IssuanceDate,
		CredentialSubject: StatusList{
			ID:           list.CredentialSubject.ID,
			Type:         list.CredentialSubject.Type,
			StatusPurpose: list.CredentialSubject.StatusPurpose,
			EncodedList:  list.CredentialSubject.EncodedList,
		},
		Proof: list.Proof,
	}
}

// NoOpStatusListCache provides a no-op implementation of StatusListCache
type NoOpStatusListCache struct{}

// NewNoOpStatusListCache creates a new no-op cache
func NewNoOpStatusListCache() *NoOpStatusListCache {
	return &NoOpStatusListCache{}
}

// Get always returns cache miss
func (c *NoOpStatusListCache) Get(listID string) (*StatusList2021, error) {
	return nil, NewStatusListError(ErrorListNotFound, "no-op cache always misses")
}

// Set does nothing
func (c *NoOpStatusListCache) Set(listID string, list *StatusList2021, ttl time.Duration) error {
	return nil
}

// Invalidate does nothing
func (c *NoOpStatusListCache) Invalidate(listID string) error {
	return nil
}

// Clear does nothing
func (c *NoOpStatusListCache) Clear() error {
	return nil
}

// Stats returns empty stats
func (c *NoOpStatusListCache) Stats() *CacheStats {
	return &CacheStats{}
}

// TimedStatusListCache wraps another cache with automatic expiration
type TimedStatusListCache struct {
	inner       StatusListCache
	defaultTTL  time.Duration
	maxTTL      time.Duration
}

// NewTimedStatusListCache creates a cache wrapper with automatic expiration
func NewTimedStatusListCache(inner StatusListCache, defaultTTL, maxTTL time.Duration) *TimedStatusListCache {
	if defaultTTL <= 0 {
		defaultTTL = time.Hour
	}
	
	if maxTTL <= 0 {
		maxTTL = 24 * time.Hour
	}
	
	return &TimedStatusListCache{
		inner:      inner,
		defaultTTL: defaultTTL,
		maxTTL:     maxTTL,
	}
}

// Get delegates to inner cache
func (c *TimedStatusListCache) Get(listID string) (*StatusList2021, error) {
	return c.inner.Get(listID)
}

// Set enforces TTL limits
func (c *TimedStatusListCache) Set(listID string, list *StatusList2021, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	
	if ttl > c.maxTTL {
		ttl = c.maxTTL
	}
	
	return c.inner.Set(listID, list, ttl)
}

// Invalidate delegates to inner cache
func (c *TimedStatusListCache) Invalidate(listID string) error {
	return c.inner.Invalidate(listID)
}

// Clear delegates to inner cache
func (c *TimedStatusListCache) Clear() error {
	return c.inner.Clear()
}

// Stats delegates to inner cache
func (c *TimedStatusListCache) Stats() *CacheStats {
	return c.inner.Stats()
}