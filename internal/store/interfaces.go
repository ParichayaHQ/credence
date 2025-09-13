package store

import (
	"context"
	"time"

	"github.com/ParichayaHQ/credence/internal/events"
)

// BlobStore defines the interface for storing content-addressed blobs
type BlobStore interface {
	// Store saves a blob by its CID and returns the CID
	Store(ctx context.Context, data []byte) (string, error)
	
	// Get retrieves a blob by CID
	Get(ctx context.Context, cid string) ([]byte, error)
	
	// Has checks if a blob exists
	Has(ctx context.Context, cid string) (bool, error)
	
	// Delete removes a blob (rarely used, mainly for cleanup)
	Delete(ctx context.Context, cid string) error
	
	// Stats returns storage statistics
	Stats() BlobStats
	
	// Close cleanly shuts down the store
	Close() error
}

// EventStore defines the interface for storing and indexing events
type EventStore interface {
	// StoreEvent persists an event and creates indexes
	StoreEvent(ctx context.Context, event *events.Event) error
	
	// GetEvent retrieves an event by CID
	GetEvent(ctx context.Context, cid string) (*events.Event, error)
	
	// GetEventsByDID returns events by DID (from/to) within epoch range
	GetEventsByDID(ctx context.Context, did string, direction Direction, fromEpoch, toEpoch string) ([]*events.Event, error)
	
	// GetEventsByType returns events by type within epoch range
	GetEventsByType(ctx context.Context, eventType string, fromEpoch, toEpoch string) ([]*events.Event, error)
	
	// Close cleanly shuts down the store
	Close() error
}

// CheckpointStore defines storage for transparency log checkpoints
type CheckpointStore interface {
	// StoreCheckpoint saves a checkpoint
	StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error
	
	// GetCheckpoint retrieves a checkpoint by epoch
	GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error)
	
	// GetLatestCheckpoint returns the most recent checkpoint
	GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error)
	
	// ListCheckpoints returns checkpoints in epoch range
	ListCheckpoints(ctx context.Context, fromEpoch, toEpoch int64) ([]*Checkpoint, error)
	
	// Close cleanly shuts down the store
	Close() error
}

// StatusListStore defines storage for VC revocation status lists
type StatusListStore interface {
	// StoreStatusList saves a status list bitmap
	StoreStatusList(ctx context.Context, issuer, epoch, bitmapCID string) error
	
	// GetStatusList retrieves a status list
	GetStatusList(ctx context.Context, issuer, epoch string) (string, error)
	
	// Close cleanly shuts down the store
	Close() error
}

// Direction indicates query direction for event indexing
type Direction int

const (
	DirectionFrom Direction = iota // Events from this DID
	DirectionTo                    // Events to this DID
	DirectionBoth                  // Events from or to this DID
)

// BlobStats contains storage statistics
type BlobStats struct {
	TotalBlobs   int64     `json:"total_blobs"`
	TotalBytes   int64     `json:"total_bytes"`
	LastAccessed time.Time `json:"last_accessed"`
}

// Checkpoint represents a transparency log checkpoint
type Checkpoint struct {
	TreeID    int64             `json:"tree_id"`
	Root      string            `json:"root"`
	Epoch     int64             `json:"epoch"`
	TreeSize  int64             `json:"tree_size"`
	Timestamp time.Time         `json:"timestamp"`
	Signers   []string          `json:"signers"`
	Signature string            `json:"signature"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// FullNodeStore combines all storage interfaces for a full node
type FullNodeStore interface {
	BlobStore
	EventStore
	CheckpointStore
	StatusListStore
}