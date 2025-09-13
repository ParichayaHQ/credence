// +build !rocksdb

package store

import (
	"context"
	"fmt"

	"github.com/ParichayaHQ/credence/internal/events"
)

// RocksDBStore stub implementation when RocksDB is disabled
type RocksDBStore struct {
	closed bool
}

func NewRocksDBStore(config *Config) (*RocksDBStore, error) {
	return nil, fmt.Errorf("RocksDB support not compiled in - use build tag 'rocksdb' to enable")
}

func (s *RocksDBStore) Store(ctx context.Context, data []byte) (string, error) {
	return "", fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) Get(ctx context.Context, cid string) ([]byte, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) Has(ctx context.Context, cid string) (bool, error) {
	return false, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) Delete(ctx context.Context, cid string) error {
	return fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) Stats() BlobStats {
	return BlobStats{}
}

func (s *RocksDBStore) StoreEvent(ctx context.Context, event *events.Event) error {
	return fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) GetEvent(ctx context.Context, cid string) (*events.Event, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) GetEventsByDID(ctx context.Context, did string, direction Direction, fromEpoch, toEpoch string) ([]*events.Event, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) GetEventsByType(ctx context.Context, eventType string, fromEpoch, toEpoch string) ([]*events.Event, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	return fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) ListCheckpoints(ctx context.Context, fromEpoch, toEpoch int64) ([]*Checkpoint, error) {
	return nil, fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) StoreStatusList(ctx context.Context, issuer, epoch, bitmapCID string) error {
	return fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) GetStatusList(ctx context.Context, issuer, epoch string) (string, error) {
	return "", fmt.Errorf("RocksDB not available")
}

func (s *RocksDBStore) Close() error {
	return nil
}