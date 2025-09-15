// +build !rocksdb

package store

import (
	"context"
	"fmt"

	"github.com/ParichayaHQ/credence/internal/events"
)

// StubEventStore provides a no-op implementation for development
type StubEventStore struct{}

func (s *StubEventStore) StoreEvent(ctx context.Context, event *events.Event) error {
	return fmt.Errorf("EventStore not available - build with RocksDB support for full functionality")
}

func (s *StubEventStore) GetEvent(ctx context.Context, cid string) (*events.Event, error) {
	return nil, fmt.Errorf("EventStore not available - build with RocksDB support for full functionality")
}

func (s *StubEventStore) GetEventsByDID(ctx context.Context, did string, direction Direction, fromEpoch, toEpoch string) ([]*events.Event, error) {
	return nil, fmt.Errorf("EventStore not available - build with RocksDB support for full functionality")
}

func (s *StubEventStore) GetEventsByType(ctx context.Context, eventType string, fromEpoch, toEpoch string) ([]*events.Event, error) {
	return nil, fmt.Errorf("EventStore not available - build with RocksDB support for full functionality")
}

func (s *StubEventStore) Close() error {
	return nil
}

// StubCheckpointStore provides a no-op implementation for development
type StubCheckpointStore struct{}

func (s *StubCheckpointStore) StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	return fmt.Errorf("CheckpointStore not available - build with RocksDB support for full functionality")
}

func (s *StubCheckpointStore) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	return nil, fmt.Errorf("CheckpointStore not available - build with RocksDB support for full functionality")
}

func (s *StubCheckpointStore) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	return nil, fmt.Errorf("CheckpointStore not available - build with RocksDB support for full functionality")
}

func (s *StubCheckpointStore) ListCheckpoints(ctx context.Context, fromEpoch, toEpoch int64) ([]*Checkpoint, error) {
	return nil, fmt.Errorf("CheckpointStore not available - build with RocksDB support for full functionality")
}

func (s *StubCheckpointStore) Close() error {
	return nil
}

// StubStatusListStore provides a no-op implementation for development
type StubStatusListStore struct{}

func (s *StubStatusListStore) StoreStatusList(ctx context.Context, issuer, epoch, bitmapCID string) error {
	return fmt.Errorf("StatusListStore not available - build with RocksDB support for full functionality")
}

func (s *StubStatusListStore) GetStatusList(ctx context.Context, issuer, epoch string) (string, error) {
	return "", fmt.Errorf("StatusListStore not available - build with RocksDB support for full functionality")
}

func (s *StubStatusListStore) Close() error {
	return nil
}