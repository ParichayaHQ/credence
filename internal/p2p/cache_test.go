package p2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUCache(t *testing.T) {
	t.Run("BasicSetGet", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		// Test set and get
		cache.Set("key1", []byte("value1"))
		data, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, []byte("value1"), data)

		// Test non-existent key
		_, found = cache.Get("nonexistent")
		assert.False(t, found)
	})

	t.Run("LRUEviction", func(t *testing.T) {
		cache := NewLRUCache(2, time.Hour)
		defer cache.Close()

		// Fill cache to capacity
		cache.Set("key1", []byte("value1"))
		cache.Set("key2", []byte("value2"))

		// Both should be present
		_, found := cache.Get("key1")
		assert.True(t, found)
		_, found = cache.Get("key2")
		assert.True(t, found)

		// Add third item, should evict key1 (least recently used)
		cache.Set("key3", []byte("value3"))

		// key1 should be evicted
		_, found = cache.Get("key1")
		assert.False(t, found)

		// key2 and key3 should still be present
		_, found = cache.Get("key2")
		assert.True(t, found)
		_, found = cache.Get("key3")
		assert.True(t, found)
	})

	t.Run("LRUOrdering", func(t *testing.T) {
		cache := NewLRUCache(2, time.Hour)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))
		cache.Set("key2", []byte("value2"))

		// Access key1 to make it most recently used
		_, found := cache.Get("key1")
		assert.True(t, found)

		// Add key3, should evict key2 (now least recently used)
		cache.Set("key3", []byte("value3"))

		// key2 should be evicted, key1 should remain
		_, found = cache.Get("key1")
		assert.True(t, found)
		_, found = cache.Get("key2")
		assert.False(t, found)
		_, found = cache.Get("key3")
		assert.True(t, found)
	})

	t.Run("TTLExpiry", func(t *testing.T) {
		cache := NewLRUCache(10, 100*time.Millisecond)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))

		// Should be present immediately
		_, found := cache.Get("key1")
		assert.True(t, found)

		// Wait for expiry
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		_, found = cache.Get("key1")
		assert.False(t, found)
	})

	t.Run("Update", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))
		cache.Set("key1", []byte("updated_value1"))

		data, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, []byte("updated_value1"), data)

		// Size should still be 1
		assert.Equal(t, 1, cache.Size())
	})

	t.Run("Has", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))

		assert.True(t, cache.Has("key1"))
		assert.False(t, cache.Has("nonexistent"))
	})

	t.Run("Delete", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))
		cache.Set("key2", []byte("value2"))

		assert.True(t, cache.Has("key1"))
		cache.Delete("key1")
		assert.False(t, cache.Has("key1"))
		assert.True(t, cache.Has("key2"))

		assert.Equal(t, 1, cache.Size())
	})

	t.Run("Clear", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))
		cache.Set("key2", []byte("value2"))
		assert.Equal(t, 2, cache.Size())

		cache.Clear()
		assert.Equal(t, 0, cache.Size())
		assert.False(t, cache.Has("key1"))
		assert.False(t, cache.Has("key2"))
	})

	t.Run("Stats", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))
		cache.Set("key2", []byte("value2"))

		stats := cache.Stats()
		assert.Equal(t, 2, stats["size"])
		assert.Equal(t, 3, stats["max_size"])
		assert.Equal(t, 0, stats["expired"])
		assert.Equal(t, time.Hour.Seconds(), stats["ttl_seconds"])
	})

	t.Run("CleanupExpired", func(t *testing.T) {
		cache := NewLRUCache(10, 50*time.Millisecond)
		defer cache.Close()

		cache.Set("key1", []byte("value1"))
		cache.Set("key2", []byte("value2"))

		// Wait for expiry and cleanup
		time.Sleep(150 * time.Millisecond)

		// Force cleanup by checking stats
		stats := cache.Stats()
		
		// Items should be cleaned up eventually
		assert.Equal(t, 0, stats["size"])
	})

	t.Run("DataIsolation", func(t *testing.T) {
		cache := NewLRUCache(3, time.Hour)
		defer cache.Close()

		original := []byte("original")
		cache.Set("key1", original)

		// Modify original slice
		original[0] = 'X'

		// Retrieved data should be unchanged
		data, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, []byte("original"), data)

		// Modify retrieved data
		data[0] = 'Y'

		// Cache should still have original data
		data2, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, []byte("original"), data2)
	})
}

func TestCacheEntry(t *testing.T) {
	t.Run("IsExpired", func(t *testing.T) {
		// Not expired
		entry := &CacheEntry{
			Value:     []byte("test"),
			ExpiresAt: time.Now().Add(time.Hour),
		}
		assert.False(t, entry.IsExpired())

		// Expired
		entry.ExpiresAt = time.Now().Add(-time.Hour)
		assert.True(t, entry.IsExpired())

		// Exactly at expiry time (should be expired)
		entry.ExpiresAt = time.Now()
		time.Sleep(time.Millisecond) // Ensure we're past the expiry time
		assert.True(t, entry.IsExpired())
	})
}

// Benchmark tests
func BenchmarkCacheSet(b *testing.B) {
	cache := NewLRUCache(1000, time.Hour)
	defer cache.Close()

	data := []byte("benchmark data")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%1000))
		cache.Set(key, data)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewLRUCache(1000, time.Hour)
	defer cache.Close()

	// Pre-populate cache
	data := []byte("benchmark data")
	for i := 0; i < 1000; i++ {
		key := "key" + string(rune(i))
		cache.Set(key, data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%1000))
		cache.Get(key)
	}
}

func BenchmarkCacheMixed(b *testing.B) {
	cache := NewLRUCache(1000, time.Hour)
	defer cache.Close()

	data := []byte("benchmark data")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%1000))
		if i%3 == 0 {
			cache.Set(key, data)
		} else {
			cache.Get(key)
		}
	}
}