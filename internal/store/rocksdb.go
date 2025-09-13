// +build rocksdb

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"
	"github.com/ParichayaHQ/credence/internal/events"
)

// RocksDBStore implements storage interfaces using RocksDB
type RocksDBStore struct {
	config *Config
	db     *grocksdb.DB
	opts   *grocksdb.Options
	
	// Column families
	cfs map[string]*grocksdb.ColumnFamilyHandle
	
	// Read/write options
	readOpts  *grocksdb.ReadOptions
	writeOpts *grocksdb.WriteOptions
	
	// Internal state
	mu     sync.RWMutex
	closed bool
	
	// Statistics
	stats StoreStats
}

// StoreStats tracks database statistics
type StoreStats struct {
	EventsStored     int64     `json:"events_stored"`
	BlobsStored      int64     `json:"blobs_stored"`
	CheckpointsStored int64    `json:"checkpoints_stored"`
	LastActivity     time.Time `json:"last_activity"`
}

// Column family names
const (
	CFDefault     = "default"
	CFEvents      = "events"
	CFIndex       = "index"
	CFCheckpoints = "checkpoints"
	CFStatusLists = "statuslists"
	CFBlobs       = "blobs"
)

// Key prefixes for different data types
const (
	PrefixEvent      = "evt:"
	PrefixIndexTo    = "idx:to:"
	PrefixIndexFrom  = "idx:from:"
	PrefixIndexType  = "idx:type:"
	PrefixCheckpoint = "chk:"
	PrefixStatus     = "status:"
	PrefixBlob       = "blob:"
)

// NewRocksDBStore creates a new RocksDB-backed store
func NewRocksDBStore(config *Config) (*RocksDBStore, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	store := &RocksDBStore{
		config: config,
		cfs:    make(map[string]*grocksdb.ColumnFamilyHandle),
	}
	
	if err := store.open(); err != nil {
		return nil, err
	}
	
	return store, nil
}

// open initializes the RocksDB database
func (s *RocksDBStore) open() error {
	// Create options
	s.opts = grocksdb.NewDefaultOptions()
	s.opts.SetCreateIfMissing(true)
	s.opts.SetCreateIfMissingColumnFamilies(true)
	
	// Apply configuration
	s.applyConfig()
	
	// Define column families
	cfNames := []string{CFDefault, CFEvents, CFIndex, CFCheckpoints, CFStatusLists, CFBlobs}
	cfOpts := make([]*grocksdb.Options, len(cfNames))
	
	for i, name := range cfNames {
		cfOpts[i] = grocksdb.NewDefaultOptions()
		if cfConfig, exists := s.config.RocksDB.ColumnFamilies[name]; exists {
			s.applyCFConfig(cfOpts[i], cfConfig)
		}
	}
	
	// Open database
	db, cfHandles, err := grocksdb.OpenDbColumnFamilies(s.opts, s.config.RocksDB.Path, cfNames, cfOpts)
	if err != nil {
		return ErrDatabase("open", err)
	}
	
	s.db = db
	
	// Store column family handles
	for i, name := range cfNames {
		s.cfs[name] = cfHandles[i]
	}
	
	// Create read/write options
	s.readOpts = grocksdb.NewDefaultReadOptions()
	s.writeOpts = grocksdb.NewDefaultWriteOptions()
	s.writeOpts.SetSync(s.config.RocksDB.SyncWrites)
	
	return nil
}

// applyConfig applies RocksDB configuration options
func (s *RocksDBStore) applyConfig() {
	cfg := &s.config.RocksDB
	
	s.opts.SetMaxOpenFiles(cfg.MaxOpenFiles)
	s.opts.SetWriteBufferSize(cfg.WriteBufferSize * 1024 * 1024) // Convert MB to bytes
	s.opts.SetMaxWriteBufferNumber(cfg.MaxWriteBufferNumber)
	s.opts.SetMaxBackgroundJobs(cfg.MaxBackgroundJobs)
	s.opts.SetMaxBytesForLevelBase(uint64(cfg.MaxBytesForLevelBase))
	
	// Block cache
	blockCache := grocksdb.NewLRUCache(cfg.BlockCacheSize * 1024 * 1024)
	blockOpts := grocksdb.NewDefaultBlockBasedTableOptions()
	blockOpts.SetBlockCache(blockCache)
	s.opts.SetBlockBasedTableFactory(blockOpts)
	
	// Compression
	switch cfg.CompressionType {
	case "snappy":
		s.opts.SetCompression(grocksdb.SnappyCompression)
	case "lz4":
		s.opts.SetCompression(grocksdb.LZ4Compression)
	case "zstd":
		s.opts.SetCompression(grocksdb.ZSTDCompression)
	default:
		s.opts.SetCompression(grocksdb.NoCompression)
	}
	
	// WAL
	if !cfg.EnableWAL {
		s.opts.SetDisableWAL(true)
	}
}

// applyCFConfig applies column family specific configuration
func (s *RocksDBStore) applyCFConfig(opts *grocksdb.Options, cfg ColumnFamilyConfig) {
	if cfg.WriteBufferSize > 0 {
		opts.SetWriteBufferSize(cfg.WriteBufferSize * 1024 * 1024)
	}
	if cfg.MaxWriteBufferNumber > 0 {
		opts.SetMaxWriteBufferNumber(cfg.MaxWriteBufferNumber)
	}
	
	// Bloom filter
	if cfg.BloomFilterBitsPerKey > 0 {
		blockOpts := grocksdb.NewDefaultBlockBasedTableOptions()
		filter := grocksdb.NewBloomFilter(cfg.BloomFilterBitsPerKey)
		blockOpts.SetFilterPolicy(filter)
		opts.SetBlockBasedTableFactory(blockOpts)
	}
	
	// Compression
	switch cfg.CompressionType {
	case "snappy":
		opts.SetCompression(grocksdb.SnappyCompression)
	case "lz4":
		opts.SetCompression(grocksdb.LZ4Compression)
	case "zstd":
		opts.SetCompression(grocksdb.ZSTDCompression)
	default:
		opts.SetCompression(grocksdb.NoCompression)
	}
}

// Store implements BlobStore.Store
func (s *RocksDBStore) Store(ctx context.Context, data []byte) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return "", ErrClosed
	}
	
	if int64(len(data)) > s.config.BlobStore.MaxBlobSize {
		return "", ErrTooLarge
	}
	
	// Generate CID
	cid, err := events.GenerateCID(data)
	if err != nil {
		return "", ErrDatabase("cid_generation", err)
	}
	
	// Check if already exists
	key := PrefixBlob + cid
	exists, err := s.db.GetCF(s.readOpts, s.cfs[CFBlobs], []byte(key))
	if err != nil {
		return "", ErrDatabaseKey("exists_check", key, err)
	}
	if exists != nil {
		exists.Free()
		return cid, nil // Already exists
	}
	
	// Store blob
	err = s.db.PutCF(s.writeOpts, s.cfs[CFBlobs], []byte(key), data)
	if err != nil {
		return "", ErrDatabaseKey("put", key, err)
	}
	
	s.stats.BlobsStored++
	s.stats.LastActivity = time.Now()
	
	return cid, nil
}

// Get implements BlobStore.Get
func (s *RocksDBStore) Get(ctx context.Context, cid string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	key := PrefixBlob + cid
	value, err := s.db.GetCF(s.readOpts, s.cfs[CFBlobs], []byte(key))
	if err != nil {
		return nil, ErrDatabaseKey("get", key, err)
	}
	defer value.Free()
	
	if !value.Exists() {
		return nil, ErrNotFoundCID(cid)
	}
	
	return value.Data(), nil
}

// Has implements BlobStore.Has
func (s *RocksDBStore) Has(ctx context.Context, cid string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return false, ErrClosed
	}
	
	key := PrefixBlob + cid
	value, err := s.db.GetCF(s.readOpts, s.cfs[CFBlobs], []byte(key))
	if err != nil {
		return false, ErrDatabaseKey("has", key, err)
	}
	defer value.Free()
	
	return value.Exists(), nil
}

// Delete implements BlobStore.Delete
func (s *RocksDBStore) Delete(ctx context.Context, cid string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	key := PrefixBlob + cid
	err := s.db.DeleteCF(s.writeOpts, s.cfs[CFBlobs], []byte(key))
	if err != nil {
		return ErrDatabaseKey("delete", key, err)
	}
	
	return nil
}

// Stats implements BlobStore.Stats
func (s *RocksDBStore) Stats() BlobStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return BlobStats{
		TotalBlobs:   s.stats.BlobsStored,
		TotalBytes:   0, // TODO: Track total bytes
		LastAccessed: s.stats.LastActivity,
	}
}

// StoreEvent implements EventStore.StoreEvent
func (s *RocksDBStore) StoreEvent(ctx context.Context, event *events.Event) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return ErrDatabase("marshal_event", err)
	}
	
	// Generate CID
	cid, err := events.GenerateCID(data)
	if err != nil {
		return ErrDatabase("cid_generation", err)
	}
	
	// Create write batch for atomic update
	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()
	
	// Store event
	eventKey := PrefixEvent + cid
	batch.PutCF(s.cfs[CFEvents], []byte(eventKey), data)
	
	// Create indexes
	if err := s.createEventIndexes(batch, event, cid); err != nil {
		return err
	}
	
	// Write batch
	err = s.db.Write(s.writeOpts, batch)
	if err != nil {
		return ErrDatabase("write_batch", err)
	}
	
	s.stats.EventsStored++
	s.stats.LastActivity = time.Now()
	
	return nil
}

// createEventIndexes creates index entries for an event
func (s *RocksDBStore) createEventIndexes(batch *grocksdb.WriteBatch, event *events.Event, cid string) error {
	epoch := event.Epoch
	eventType := event.Type
	
	// Index by from DID
	if event.From != "" {
		fromKey := fmt.Sprintf("%s%s:%s", PrefixIndexFrom, event.From, epoch)
		s.appendToIndex(batch, s.cfs[CFIndex], fromKey, cid)
	}
	
	// Index by to DID
	if event.To != "" {
		toKey := fmt.Sprintf("%s%s:%s", PrefixIndexTo, event.To, epoch)
		s.appendToIndex(batch, s.cfs[CFIndex], toKey, cid)
	}
	
	// Index by type
	typeKey := fmt.Sprintf("%s%s:%s", PrefixIndexType, eventType, epoch)
	s.appendToIndex(batch, s.cfs[CFIndex], typeKey, cid)
	
	return nil
}

// appendToIndex adds a CID to an index list
func (s *RocksDBStore) appendToIndex(batch *grocksdb.WriteBatch, cf *grocksdb.ColumnFamilyHandle, key, cid string) {
	// Get existing index
	existing, _ := s.db.GetCF(s.readOpts, cf, []byte(key))
	
	var cids []string
	if existing != nil && existing.Exists() {
		json.Unmarshal(existing.Data(), &cids)
		existing.Free()
	}
	
	// Append new CID
	cids = append(cids, cid)
	
	// Serialize and store
	data, _ := json.Marshal(cids)
	batch.PutCF(cf, []byte(key), data)
}

// GetEvent implements EventStore.GetEvent
func (s *RocksDBStore) GetEvent(ctx context.Context, cid string) (*events.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	key := PrefixEvent + cid
	value, err := s.db.GetCF(s.readOpts, s.cfs[CFEvents], []byte(key))
	if err != nil {
		return nil, ErrDatabaseKey("get", key, err)
	}
	defer value.Free()
	
	if !value.Exists() {
		return nil, ErrNotFoundCID(cid)
	}
	
	var event events.Event
	err = json.Unmarshal(value.Data(), &event)
	if err != nil {
		return nil, ErrDatabase("unmarshal_event", err)
	}
	
	return &event, nil
}

// GetEventsByDID implements EventStore.GetEventsByDID
func (s *RocksDBStore) GetEventsByDID(ctx context.Context, did string, direction Direction, fromEpoch, toEpoch string) ([]*events.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	var allCIDs []string
	
	// Collect CIDs from relevant indexes
	if direction == DirectionFrom || direction == DirectionBoth {
		cids, err := s.getCIDsFromIndex(PrefixIndexFrom+did, fromEpoch, toEpoch)
		if err != nil {
			return nil, err
		}
		allCIDs = append(allCIDs, cids...)
	}
	
	if direction == DirectionTo || direction == DirectionBoth {
		cids, err := s.getCIDsFromIndex(PrefixIndexTo+did, fromEpoch, toEpoch)
		if err != nil {
			return nil, err
		}
		allCIDs = append(allCIDs, cids...)
	}
	
	// Load events
	events := make([]*events.Event, 0, len(allCIDs))
	for _, cid := range allCIDs {
		event, err := s.GetEvent(ctx, cid)
		if err != nil && !IsNotFound(err) {
			return nil, err
		}
		if event != nil {
			events = append(events, event)
		}
	}
	
	return events, nil
}

// GetEventsByType implements EventStore.GetEventsByType
func (s *RocksDBStore) GetEventsByType(ctx context.Context, eventType string, fromEpoch, toEpoch string) ([]*events.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	// Get CIDs from type index
	cids, err := s.getCIDsFromIndex(PrefixIndexType+eventType, fromEpoch, toEpoch)
	if err != nil {
		return nil, err
	}
	
	// Load events
	events := make([]*events.Event, 0, len(cids))
	for _, cid := range cids {
		event, err := s.GetEvent(ctx, cid)
		if err != nil && !IsNotFound(err) {
			return nil, err
		}
		if event != nil {
			events = append(events, event)
		}
	}
	
	return events, nil
}

// getCIDsFromIndex retrieves CIDs from an index within epoch range
func (s *RocksDBStore) getCIDsFromIndex(prefix, fromEpoch, toEpoch string) ([]string, error) {
	var allCIDs []string
	
	// Create iterator for epoch range
	iterator := s.db.NewIteratorCF(s.readOpts, s.cfs[CFIndex])
	defer iterator.Close()
	
	// Seek to start
	startKey := prefix + ":" + fromEpoch
	iterator.Seek([]byte(startKey))
	
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key().Data())
		
		// Check if still within prefix and epoch range
		if !strings.HasPrefix(key, prefix+":") {
			break
		}
		
		// Extract epoch from key
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}
		epoch := parts[len(parts)-1]
		
		// Check epoch range
		if toEpoch != "" && epoch > toEpoch {
			break
		}
		
		// Parse CID list
		var cids []string
		if err := json.Unmarshal(iterator.Value().Data(), &cids); err == nil {
			allCIDs = append(allCIDs, cids...)
		}
	}
	
	return allCIDs, nil
}

// StoreCheckpoint implements CheckpointStore.StoreCheckpoint
func (s *RocksDBStore) StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	// Serialize checkpoint
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return ErrDatabase("marshal_checkpoint", err)
	}
	
	// Store by epoch
	key := fmt.Sprintf("%s%d", PrefixCheckpoint, checkpoint.Epoch)
	err = s.db.PutCF(s.writeOpts, s.cfs[CFCheckpoints], []byte(key), data)
	if err != nil {
		return ErrDatabaseKey("put", key, err)
	}
	
	s.stats.CheckpointsStored++
	s.stats.LastActivity = time.Now()
	
	return nil
}

// GetCheckpoint implements CheckpointStore.GetCheckpoint
func (s *RocksDBStore) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	key := fmt.Sprintf("%s%d", PrefixCheckpoint, epoch)
	value, err := s.db.GetCF(s.readOpts, s.cfs[CFCheckpoints], []byte(key))
	if err != nil {
		return nil, ErrDatabaseKey("get", key, err)
	}
	defer value.Free()
	
	if !value.Exists() {
		return nil, ErrNotFoundEpoch(fmt.Sprintf("%d", epoch))
	}
	
	var checkpoint Checkpoint
	err = json.Unmarshal(value.Data(), &checkpoint)
	if err != nil {
		return nil, ErrDatabase("unmarshal_checkpoint", err)
	}
	
	return &checkpoint, nil
}

// GetLatestCheckpoint implements CheckpointStore.GetLatestCheckpoint
func (s *RocksDBStore) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	// Iterate backwards from latest
	iterator := s.db.NewIteratorCF(s.readOpts, s.cfs[CFCheckpoints])
	defer iterator.Close()
	
	// Seek to end of checkpoint prefix
	prefix := PrefixCheckpoint
	iterator.SeekForPrev([]byte(prefix + "~")) // '~' is after numbers
	
	for ; iterator.Valid(); iterator.Prev() {
		key := string(iterator.Key().Data())
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		
		var checkpoint Checkpoint
		if err := json.Unmarshal(iterator.Value().Data(), &checkpoint); err == nil {
			return &checkpoint, nil
		}
	}
	
	return nil, ErrNotFound
}

// ListCheckpoints implements CheckpointStore.ListCheckpoints
func (s *RocksDBStore) ListCheckpoints(ctx context.Context, fromEpoch, toEpoch int64) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	var checkpoints []*Checkpoint
	
	iterator := s.db.NewIteratorCF(s.readOpts, s.cfs[CFCheckpoints])
	defer iterator.Close()
	
	startKey := fmt.Sprintf("%s%d", PrefixCheckpoint, fromEpoch)
	iterator.Seek([]byte(startKey))
	
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key().Data())
		if !strings.HasPrefix(key, PrefixCheckpoint) {
			break
		}
		
		var checkpoint Checkpoint
		if err := json.Unmarshal(iterator.Value().Data(), &checkpoint); err != nil {
			continue
		}
		
		if checkpoint.Epoch > toEpoch {
			break
		}
		
		checkpoints = append(checkpoints, &checkpoint)
	}
	
	return checkpoints, nil
}

// StoreStatusList implements StatusListStore.StoreStatusList
func (s *RocksDBStore) StoreStatusList(ctx context.Context, issuer, epoch, bitmapCID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	key := fmt.Sprintf("%s%s:%s", PrefixStatus, issuer, epoch)
	err := s.db.PutCF(s.writeOpts, s.cfs[CFStatusLists], []byte(key), []byte(bitmapCID))
	if err != nil {
		return ErrDatabaseKey("put", key, err)
	}
	
	return nil
}

// GetStatusList implements StatusListStore.GetStatusList
func (s *RocksDBStore) GetStatusList(ctx context.Context, issuer, epoch string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return "", ErrClosed
	}
	
	key := fmt.Sprintf("%s%s:%s", PrefixStatus, issuer, epoch)
	value, err := s.db.GetCF(s.readOpts, s.cfs[CFStatusLists], []byte(key))
	if err != nil {
		return "", ErrDatabaseKey("get", key, err)
	}
	defer value.Free()
	
	if !value.Exists() {
		return "", ErrNotFoundEpoch(epoch)
	}
	
	return string(value.Data()), nil
}

// Close closes the RocksDB store
func (s *RocksDBStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		return nil
	}
	
	// Close column families
	for _, cf := range s.cfs {
		cf.Destroy()
	}
	
	// Close options
	if s.readOpts != nil {
		s.readOpts.Destroy()
	}
	if s.writeOpts != nil {
		s.writeOpts.Destroy()
	}
	if s.opts != nil {
		s.opts.Destroy()
	}
	
	// Close database
	if s.db != nil {
		s.db.Close()
	}
	
	s.closed = true
	return nil
}