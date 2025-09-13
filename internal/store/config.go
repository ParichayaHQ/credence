package store

import (
	"time"
)

// Config holds configuration for the storage layer
type Config struct {
	RocksDB   RocksDBConfig   `json:"rocksdb"`
	BlobStore BlobStoreConfig `json:"blob_store"`
	Cache     CacheConfig     `json:"cache"`
}

// RocksDBConfig configures RocksDB settings
type RocksDBConfig struct {
	Path string `json:"path"`
	
	// Performance tuning
	MaxOpenFiles         int  `json:"max_open_files"`
	WriteBufferSize      int  `json:"write_buffer_size"`      // MB
	MaxWriteBufferNumber int  `json:"max_write_buffer_number"`
	BlockCacheSize       int  `json:"block_cache_size"`       // MB
	CompactionStyle      int  `json:"compaction_style"`       // 0=level, 1=universal
	EnableWAL            bool `json:"enable_wal"`
	SyncWrites           bool `json:"sync_writes"`
	
	// Compression
	CompressionType string `json:"compression_type"` // none, snappy, lz4, zstd
	
	// Compaction tuning
	MaxBackgroundJobs   int `json:"max_background_jobs"`
	MaxBytesForLevelBase int64 `json:"max_bytes_for_level_base"` // Bytes
	
	// Column Family configs
	ColumnFamilies map[string]ColumnFamilyConfig `json:"column_families"`
}

// ColumnFamilyConfig holds per-CF configuration
type ColumnFamilyConfig struct {
	WriteBufferSize      int    `json:"write_buffer_size"`
	MaxWriteBufferNumber int    `json:"max_write_buffer_number"`
	CompressionType      string `json:"compression_type"`
	BloomFilterBitsPerKey int   `json:"bloom_filter_bits_per_key"`
}

// BlobStoreConfig configures blob storage
type BlobStoreConfig struct {
	Backend string `json:"backend"` // "filesystem", "rocksdb", "s3"
	
	// Filesystem backend
	FSPath string `json:"fs_path"`
	
	// Size limits
	MaxBlobSize   int64 `json:"max_blob_size"`   // bytes
	MaxTotalSize  int64 `json:"max_total_size"`  // bytes
	
	// Cleanup policy
	EnableCleanup     bool          `json:"enable_cleanup"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	MaxAge            time.Duration `json:"max_age"`
	CleanupBatchSize  int           `json:"cleanup_batch_size"`
}

// CacheConfig configures in-memory caching
type CacheConfig struct {
	// Event cache
	EventCacheSize int           `json:"event_cache_size"`
	EventCacheTTL  time.Duration `json:"event_cache_ttl"`
	
	// Blob cache  
	BlobCacheSize int           `json:"blob_cache_size"`
	BlobCacheTTL  time.Duration `json:"blob_cache_ttl"`
	
	// Checkpoint cache
	CheckpointCacheSize int           `json:"checkpoint_cache_size"`
	CheckpointCacheTTL  time.Duration `json:"checkpoint_cache_ttl"`
}

// DefaultConfig returns sensible defaults for storage configuration
func DefaultConfig() *Config {
	return &Config{
		RocksDB: RocksDBConfig{
			Path:                 "./data/rocksdb",
			MaxOpenFiles:         1000,
			WriteBufferSize:      64, // MB
			MaxWriteBufferNumber: 3,
			BlockCacheSize:       128, // MB
			CompactionStyle:      0,   // Level compaction
			EnableWAL:            true,
			SyncWrites:           false, // async for better performance
			CompressionType:      "lz4",
			MaxBackgroundJobs:    4,
			MaxBytesForLevelBase: 256 * 1024 * 1024, // 256MB
			
			ColumnFamilies: map[string]ColumnFamilyConfig{
				"events": {
					WriteBufferSize:       32,
					MaxWriteBufferNumber:  2,
					CompressionType:       "lz4",
					BloomFilterBitsPerKey: 10,
				},
				"index": {
					WriteBufferSize:       16,
					MaxWriteBufferNumber:  2,
					CompressionType:       "lz4",
					BloomFilterBitsPerKey: 15,
				},
				"checkpoints": {
					WriteBufferSize:       8,
					MaxWriteBufferNumber:  1,
					CompressionType:       "zstd",
					BloomFilterBitsPerKey: 10,
				},
				"statuslists": {
					WriteBufferSize:       8,
					MaxWriteBufferNumber:  1,
					CompressionType:       "lz4",
					BloomFilterBitsPerKey: 10,
				},
			},
		},
		
		BlobStore: BlobStoreConfig{
			Backend:           "filesystem",
			FSPath:            "./data/blobs",
			MaxBlobSize:       16 * 1024 * 1024, // 16MB
			MaxTotalSize:      100 * 1024 * 1024 * 1024, // 100GB
			EnableCleanup:     true,
			CleanupInterval:   time.Hour,
			MaxAge:            30 * 24 * time.Hour, // 30 days
			CleanupBatchSize:  1000,
		},
		
		Cache: CacheConfig{
			EventCacheSize:      10000,
			EventCacheTTL:       30 * time.Minute,
			BlobCacheSize:       5000,
			BlobCacheTTL:        15 * time.Minute,
			CheckpointCacheSize: 100,
			CheckpointCacheTTL:  5 * time.Minute,
		},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.RocksDB.Path == "" {
		return ErrInvalidConfig("rocksdb path cannot be empty")
	}
	
	if c.BlobStore.Backend == "filesystem" && c.BlobStore.FSPath == "" {
		return ErrInvalidConfig("filesystem backend requires fs_path")
	}
	
	if c.BlobStore.MaxBlobSize <= 0 {
		return ErrInvalidConfig("max_blob_size must be positive")
	}
	
	if c.Cache.EventCacheSize <= 0 {
		return ErrInvalidConfig("event_cache_size must be positive")
	}
	
	return nil
}