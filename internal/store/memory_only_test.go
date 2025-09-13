package store

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test just the filesystem store without RocksDB dependencies
func TestFilesystemBlobStore(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "credence_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	config := &BlobStoreConfig{
		Backend:      "filesystem",
		FSPath:       tmpDir,
		MaxBlobSize:  1024 * 1024, // 1MB
		MaxTotalSize: 10 * 1024 * 1024, // 10MB
		EnableCleanup: false, // Disable for tests
	}
	
	store, err := NewFilesystemBlobStore(config)
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	t.Run("StoreAndGet", func(t *testing.T) {
		testData := []byte("Hello, World!")
		
		// Store blob
		cid, err := store.Store(ctx, testData)
		require.NoError(t, err)
		assert.NotEmpty(t, cid)
		
		// Verify it exists
		exists, err := store.Has(ctx, cid)
		require.NoError(t, err)
		assert.True(t, exists)
		
		// Retrieve blob
		retrieved, err := store.Get(ctx, cid)
		require.NoError(t, err)
		assert.Equal(t, testData, retrieved)
		
		// Check stats
		stats := store.Stats()
		assert.Equal(t, int64(1), stats.TotalBlobs)
		assert.True(t, stats.TotalBytes > 0)
	})
	
	t.Run("DuplicateStore", func(t *testing.T) {
		testData := []byte("Duplicate test")
		
		// Store same data twice
		cid1, err := store.Store(ctx, testData)
		require.NoError(t, err)
		
		cid2, err := store.Store(ctx, testData)
		require.NoError(t, err)
		
		// Should return same CID
		assert.Equal(t, cid1, cid2)
		
		// Should still exist
		exists, err := store.Has(ctx, cid1)
		require.NoError(t, err)
		assert.True(t, exists)
	})
	
	t.Run("Delete", func(t *testing.T) {
		testData := []byte("Delete me")
		
		cid, err := store.Store(ctx, testData)
		require.NoError(t, err)
		
		// Verify exists
		exists, err := store.Has(ctx, cid)
		require.NoError(t, err)
		assert.True(t, exists)
		
		// Delete
		err = store.Delete(ctx, cid)
		require.NoError(t, err)
		
		// Verify deleted
		exists, err = store.Has(ctx, cid)
		require.NoError(t, err)
		assert.False(t, exists)
		
		// Get should return not found
		_, err = store.Get(ctx, cid)
		assert.True(t, IsNotFound(err))
	})
	
	t.Run("TooLarge", func(t *testing.T) {
		// Create data larger than max size
		testData := make([]byte, config.MaxBlobSize+1)
		
		_, err := store.Store(ctx, testData)
		assert.True(t, IsTooLarge(err))
	})
	
	t.Run("NotFound", func(t *testing.T) {
		fakeCID := "QmNonExistentCID123"
		
		exists, err := store.Has(ctx, fakeCID)
		require.NoError(t, err)
		assert.False(t, exists)
		
		_, err = store.Get(ctx, fakeCID)
		assert.True(t, IsNotFound(err))
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("StoreError", func(t *testing.T) {
		err := &StoreError{
			Op:  "test_op",
			Err: ErrNotFound,
			CID: "test-cid",
		}
		
		assert.Contains(t, err.Error(), "test_op")
		assert.Contains(t, err.Error(), "not found")
		assert.Contains(t, err.Error(), "test-cid")
		
		assert.True(t, IsNotFound(err))
		assert.False(t, IsExists(err))
	})
	
	t.Run("ErrorClassification", func(t *testing.T) {
		// Direct errors
		assert.True(t, IsNotFound(ErrNotFound))
		assert.True(t, IsExists(ErrExists))
		assert.True(t, IsTooLarge(ErrTooLarge))
		
		// Wrapped errors
		wrappedNotFound := ErrNotFoundCID("test")
		assert.True(t, IsNotFound(wrappedNotFound))
		
		// Non-matching errors
		assert.False(t, IsNotFound(ErrExists))
		assert.False(t, IsExists(ErrNotFound))
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := DefaultConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("InvalidConfigs", func(t *testing.T) {
		// Empty RocksDB path
		config := DefaultConfig()
		config.RocksDB.Path = ""
		err := config.Validate()
		assert.Error(t, err)
		
		// Invalid blob backend
		config = DefaultConfig()
		config.BlobStore.Backend = "filesystem"
		config.BlobStore.FSPath = ""
		err = config.Validate()
		assert.Error(t, err)
		
		// Invalid sizes
		config = DefaultConfig()
		config.BlobStore.MaxBlobSize = 0
		err = config.Validate()
		assert.Error(t, err)
		
		config = DefaultConfig()
		config.Cache.EventCacheSize = 0
		err = config.Validate()
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkFilesystemBlobStore(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "credence_bench_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	config := &BlobStoreConfig{
		Backend:      "filesystem",
		FSPath:       tmpDir,
		MaxBlobSize:  1024 * 1024,
		MaxTotalSize: 100 * 1024 * 1024,
		EnableCleanup: false,
	}
	
	store, err := NewFilesystemBlobStore(config)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()
	
	ctx := context.Background()
	testData := make([]byte, 1024) // 1KB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	b.Run("Store", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Make each blob unique by appending counter
			data := append(testData, byte(i))
			_, err := store.Store(ctx, data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	// Pre-store some blobs for Get benchmark
	var cids []string
	for i := 0; i < 1000; i++ {
		data := append(testData, byte(i))
		cid, _ := store.Store(ctx, data)
		cids = append(cids, cid)
	}
	
	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cid := cids[i%len(cids)]
			_, err := store.Get(ctx, cid)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("Has", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cid := cids[i%len(cids)]
			_, err := store.Has(ctx, cid)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}