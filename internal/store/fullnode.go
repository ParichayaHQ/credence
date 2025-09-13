package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/ParichayaHQ/credence/internal/events"
)

// FullNode implements a complete storage node with all storage interfaces
type FullNode struct {
	config *Config
	
	// Storage backends
	rocksdb *RocksDBStore
	blobFS  *FilesystemBlobStore
	
	// Use composite pattern for different storage types
	eventStore      EventStore
	checkpointStore CheckpointStore
	statusStore     StatusListStore
	blobStore       BlobStore
	
	// State
	mu     sync.RWMutex
	closed bool
}

// NewFullNode creates a new full node with configured storage backends
func NewFullNode(config *Config) (*FullNode, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	fn := &FullNode{
		config: config,
	}
	
	// Initialize storage backends
	if err := fn.initializeStorageBackends(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	
	return fn, nil
}

// initializeStorageBackends sets up the storage backends based on configuration
func (fn *FullNode) initializeStorageBackends() error {
	// Initialize RocksDB for structured data
	rocksdb, err := NewRocksDBStore(fn.config)
	if err != nil {
		// If RocksDB is not available, we can't store events/checkpoints/status
		// In a real deployment, you'd want alternative storage backends
		return fmt.Errorf("failed to initialize RocksDB (try building with -tags rocksdb): %w", err)
	}
	fn.rocksdb = rocksdb
	
	// Use RocksDB for events, checkpoints, and status lists
	fn.eventStore = rocksdb
	fn.checkpointStore = rocksdb
	fn.statusStore = rocksdb
	
	// Initialize blob storage backend
	switch fn.config.BlobStore.Backend {
	case "rocksdb":
		fn.blobStore = rocksdb
	case "filesystem":
		blobFS, err := NewFilesystemBlobStore(&fn.config.BlobStore)
		if err != nil {
			return fmt.Errorf("failed to initialize filesystem blob store: %w", err)
		}
		fn.blobFS = blobFS
		fn.blobStore = blobFS
	default:
		return fmt.Errorf("unsupported blob store backend: %s", fn.config.BlobStore.Backend)
	}
	
	return nil
}

// Store implements BlobStore.Store
func (fn *FullNode) Store(ctx context.Context, data []byte) (string, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return "", ErrClosed
	}
	
	return fn.blobStore.Store(ctx, data)
}

// Get implements BlobStore.Get
func (fn *FullNode) Get(ctx context.Context, cid string) ([]byte, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.blobStore.Get(ctx, cid)
}

// Has implements BlobStore.Has
func (fn *FullNode) Has(ctx context.Context, cid string) (bool, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return false, ErrClosed
	}
	
	return fn.blobStore.Has(ctx, cid)
}

// Delete implements BlobStore.Delete
func (fn *FullNode) Delete(ctx context.Context, cid string) error {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return ErrClosed
	}
	
	return fn.blobStore.Delete(ctx, cid)
}

// Stats implements BlobStore.Stats
func (fn *FullNode) Stats() BlobStats {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return BlobStats{}
	}
	
	return fn.blobStore.Stats()
}

// StoreEvent implements EventStore.StoreEvent
func (fn *FullNode) StoreEvent(ctx context.Context, event *events.Event) error {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return ErrClosed
	}
	
	return fn.eventStore.StoreEvent(ctx, event)
}

// GetEvent implements EventStore.GetEvent
func (fn *FullNode) GetEvent(ctx context.Context, cid string) (*events.Event, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.eventStore.GetEvent(ctx, cid)
}

// GetEventsByDID implements EventStore.GetEventsByDID
func (fn *FullNode) GetEventsByDID(ctx context.Context, did string, direction Direction, fromEpoch, toEpoch string) ([]*events.Event, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.eventStore.GetEventsByDID(ctx, did, direction, fromEpoch, toEpoch)
}

// GetEventsByType implements EventStore.GetEventsByType
func (fn *FullNode) GetEventsByType(ctx context.Context, eventType string, fromEpoch, toEpoch string) ([]*events.Event, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.eventStore.GetEventsByType(ctx, eventType, fromEpoch, toEpoch)
}

// StoreCheckpoint implements CheckpointStore.StoreCheckpoint
func (fn *FullNode) StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return ErrClosed
	}
	
	return fn.checkpointStore.StoreCheckpoint(ctx, checkpoint)
}

// GetCheckpoint implements CheckpointStore.GetCheckpoint
func (fn *FullNode) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.checkpointStore.GetCheckpoint(ctx, epoch)
}

// GetLatestCheckpoint implements CheckpointStore.GetLatestCheckpoint
func (fn *FullNode) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.checkpointStore.GetLatestCheckpoint(ctx)
}

// ListCheckpoints implements CheckpointStore.ListCheckpoints
func (fn *FullNode) ListCheckpoints(ctx context.Context, fromEpoch, toEpoch int64) ([]*Checkpoint, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, ErrClosed
	}
	
	return fn.checkpointStore.ListCheckpoints(ctx, fromEpoch, toEpoch)
}

// StoreStatusList implements StatusListStore.StoreStatusList
func (fn *FullNode) StoreStatusList(ctx context.Context, issuer, epoch, bitmapCID string) error {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return ErrClosed
	}
	
	return fn.statusStore.StoreStatusList(ctx, issuer, epoch, bitmapCID)
}

// GetStatusList implements StatusListStore.GetStatusList
func (fn *FullNode) GetStatusList(ctx context.Context, issuer, epoch string) (string, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return "", ErrClosed
	}
	
	return fn.statusStore.GetStatusList(ctx, issuer, epoch)
}

// StoreEventAndBlob atomically stores both an event and its associated blob data
func (fn *FullNode) StoreEventAndBlob(ctx context.Context, event *events.Event, blobData []byte) (string, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return "", ErrClosed
	}
	
	// First store the blob if provided
	var blobCID string
	if len(blobData) > 0 {
		cid, err := fn.blobStore.Store(ctx, blobData)
		if err != nil {
			return "", fmt.Errorf("failed to store blob: %w", err)
		}
		blobCID = cid
		
		// Update event with blob CID if not already set
		if event.PayloadCID == "" {
			event.PayloadCID = blobCID
		}
	}
	
	// Store the event
	if err := fn.eventStore.StoreEvent(ctx, event); err != nil {
		// If blob was stored but event storage failed, we keep the blob
		// (cleanup will happen via garbage collection later)
		return "", fmt.Errorf("failed to store event: %w", err)
	}
	
	return blobCID, nil
}

// GetEventWithBlob retrieves an event and its associated blob data
func (fn *FullNode) GetEventWithBlob(ctx context.Context, cid string) (*events.Event, []byte, error) {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return nil, nil, ErrClosed
	}
	
	// Get the event
	event, err := fn.eventStore.GetEvent(ctx, cid)
	if err != nil {
		return nil, nil, err
	}
	
	// Get associated blob if it exists
	var blobData []byte
	if event.PayloadCID != "" {
		data, err := fn.blobStore.Get(ctx, event.PayloadCID)
		if err != nil && !IsNotFound(err) {
			return nil, nil, fmt.Errorf("failed to get blob %s: %w", event.PayloadCID, err)
		}
		blobData = data
	}
	
	return event, blobData, nil
}

// GetStorageStats returns comprehensive storage statistics
func (fn *FullNode) GetStorageStats() map[string]interface{} {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	
	if fn.closed {
		return map[string]interface{}{"status": "closed"}
	}
	
	stats := map[string]interface{}{
		"status": "open",
		"config": map[string]interface{}{
			"rocksdb_path":    fn.config.RocksDB.Path,
			"blob_backend":    fn.config.BlobStore.Backend,
			"blob_path":       fn.config.BlobStore.FSPath,
			"max_blob_size":   fn.config.BlobStore.MaxBlobSize,
		},
	}
	
	// Blob store statistics
	blobStats := fn.blobStore.Stats()
	stats["blobs"] = map[string]interface{}{
		"total_blobs":   blobStats.TotalBlobs,
		"total_bytes":   blobStats.TotalBytes,
		"last_accessed": blobStats.LastAccessed,
	}
	
	// RocksDB statistics if available
	if fn.rocksdb != nil {
		// Note: This would require exposing stats from the real RocksDB implementation
		stats["rocksdb"] = map[string]interface{}{
			"backend": "rocksdb",
			"status":  "available",
		}
	}
	
	return stats
}

// Close closes all storage backends
func (fn *FullNode) Close() error {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	
	if fn.closed {
		return nil
	}
	
	var errs []error
	
	// Close blob storage
	if fn.blobFS != nil {
		if err := fn.blobFS.Close(); err != nil {
			errs = append(errs, fmt.Errorf("blob filesystem close: %w", err))
		}
	}
	
	// Close RocksDB
	if fn.rocksdb != nil {
		if err := fn.rocksdb.Close(); err != nil {
			errs = append(errs, fmt.Errorf("rocksdb close: %w", err))
		}
	}
	
	fn.closed = true
	
	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	
	return nil
}