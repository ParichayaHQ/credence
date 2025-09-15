package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
	"github.com/ParichayaHQ/credence/internal/events"
)

// SQLiteStore implements EventStore, CheckpointStore, and StatusListStore using SQLite
type SQLiteStore struct {
	config *Config
	db     *sql.DB
	
	// Internal state
	mu     sync.RWMutex
	closed bool
}

// NewSQLiteStore creates a new SQLite-based store for structured data
func NewSQLiteStore(config *Config) (*SQLiteStore, error) {
	dbPath := filepath.Join(config.RocksDB.Path, "credence.db")
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	
	store := &SQLiteStore{
		config: config,
		db:     db,
	}
	
	// Initialize database schema
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}
	
	return store, nil
}

// initSchema creates the necessary tables
func (s *SQLiteStore) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS events (
			cid TEXT PRIMARY KEY,
			did_from TEXT,
			did_to TEXT,
			event_type TEXT,
			payload_cid TEXT,
			epoch TEXT,
			timestamp DATETIME,
			data TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_events_did_from ON events(did_from);
		CREATE INDEX IF NOT EXISTS idx_events_did_to ON events(did_to);
		CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
		CREATE INDEX IF NOT EXISTS idx_events_epoch ON events(epoch);
		
		CREATE TABLE IF NOT EXISTS checkpoints (
			epoch INTEGER PRIMARY KEY,
			tree_id INTEGER,
			root TEXT,
			tree_size INTEGER,
			timestamp DATETIME,
			signers TEXT,
			signature TEXT,
			metadata TEXT
		);
		
		CREATE TABLE IF NOT EXISTS status_lists (
			issuer TEXT,
			epoch TEXT,
			bitmap_cid TEXT,
			PRIMARY KEY (issuer, epoch)
		);
	`
	
	_, err := s.db.Exec(schema)
	return err
}

// generateEventCID creates a content identifier for an event
func generateEventCID(event *events.Event) string {
	data, _ := json.Marshal(event)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes as simple CID
}

// StoreEvent implements EventStore.StoreEvent
func (s *SQLiteStore) StoreEvent(ctx context.Context, event *events.Event) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	cid := generateEventCID(event)
	
	query := `
		INSERT INTO events (cid, did_from, did_to, event_type, payload_cid, epoch, timestamp, data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = s.db.ExecContext(ctx, query, 
		cid,
		event.From,
		event.To,
		event.Type,
		event.PayloadCID,
		event.Epoch,
		event.IssuedAt,
		string(data),
	)
	
	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}
	
	return nil
}

// GetEvent implements EventStore.GetEvent
func (s *SQLiteStore) GetEvent(ctx context.Context, cid string) (*events.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	query := "SELECT data FROM events WHERE cid = ?"
	
	var data string
	err := s.db.QueryRowContext(ctx, query, cid).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFoundCID(cid)
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	
	var event events.Event
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	
	return &event, nil
}

// GetEventsByDID implements EventStore.GetEventsByDID
func (s *SQLiteStore) GetEventsByDID(ctx context.Context, did string, direction Direction, fromEpoch, toEpoch string) ([]*events.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	var query string
	var args []interface{}
	
	switch direction {
	case DirectionFrom:
		query = "SELECT data FROM events WHERE did_from = ? AND epoch >= ? AND epoch <= ? ORDER BY timestamp"
		args = []interface{}{did, fromEpoch, toEpoch}
	case DirectionTo:
		query = "SELECT data FROM events WHERE did_to = ? AND epoch >= ? AND epoch <= ? ORDER BY timestamp"
		args = []interface{}{did, fromEpoch, toEpoch}
	case DirectionBoth:
		query = "SELECT data FROM events WHERE (did_from = ? OR did_to = ?) AND epoch >= ? AND epoch <= ? ORDER BY timestamp"
		args = []interface{}{did, did, fromEpoch, toEpoch}
	default:
		return nil, fmt.Errorf("invalid direction: %d", direction)
	}
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()
	
	var result []*events.Event
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		
		var event events.Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}
		
		result = append(result, &event)
	}
	
	return result, rows.Err()
}

// GetEventsByType implements EventStore.GetEventsByType
func (s *SQLiteStore) GetEventsByType(ctx context.Context, eventType string, fromEpoch, toEpoch string) ([]*events.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	query := "SELECT data FROM events WHERE event_type = ? AND epoch >= ? AND epoch <= ? ORDER BY timestamp"
	
	rows, err := s.db.QueryContext(ctx, query, eventType, fromEpoch, toEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type: %w", err)
	}
	defer rows.Close()
	
	var result []*events.Event
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		
		var event events.Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}
		
		result = append(result, &event)
	}
	
	return result, rows.Err()
}

// StoreCheckpoint implements CheckpointStore.StoreCheckpoint
func (s *SQLiteStore) StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	signersJSON, _ := json.Marshal(checkpoint.Signers)
	metadataJSON, _ := json.Marshal(checkpoint.Metadata)
	
	query := `
		INSERT OR REPLACE INTO checkpoints (epoch, tree_id, root, tree_size, timestamp, signers, signature, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := s.db.ExecContext(ctx, query,
		checkpoint.Epoch,
		checkpoint.TreeID,
		checkpoint.Root,
		checkpoint.TreeSize,
		checkpoint.Timestamp,
		string(signersJSON),
		checkpoint.Signature,
		string(metadataJSON),
	)
	
	if err != nil {
		return fmt.Errorf("failed to store checkpoint: %w", err)
	}
	
	return nil
}

// GetCheckpoint implements CheckpointStore.GetCheckpoint
func (s *SQLiteStore) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	query := "SELECT tree_id, root, tree_size, timestamp, signers, signature, metadata FROM checkpoints WHERE epoch = ?"
	
	var checkpoint Checkpoint
	var signersJSON, metadataJSON string
	
	err := s.db.QueryRowContext(ctx, query, epoch).Scan(
		&checkpoint.TreeID,
		&checkpoint.Root,
		&checkpoint.TreeSize,
		&checkpoint.Timestamp,
		&signersJSON,
		&checkpoint.Signature,
		&metadataJSON,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}
	
	checkpoint.Epoch = epoch
	json.Unmarshal([]byte(signersJSON), &checkpoint.Signers)
	json.Unmarshal([]byte(metadataJSON), &checkpoint.Metadata)
	
	return &checkpoint, nil
}

// GetLatestCheckpoint implements CheckpointStore.GetLatestCheckpoint
func (s *SQLiteStore) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	query := "SELECT epoch, tree_id, root, tree_size, timestamp, signers, signature, metadata FROM checkpoints ORDER BY epoch DESC LIMIT 1"
	
	var checkpoint Checkpoint
	var signersJSON, metadataJSON string
	
	err := s.db.QueryRowContext(ctx, query).Scan(
		&checkpoint.Epoch,
		&checkpoint.TreeID,
		&checkpoint.Root,
		&checkpoint.TreeSize,
		&checkpoint.Timestamp,
		&signersJSON,
		&checkpoint.Signature,
		&metadataJSON,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get latest checkpoint: %w", err)
	}
	
	json.Unmarshal([]byte(signersJSON), &checkpoint.Signers)
	json.Unmarshal([]byte(metadataJSON), &checkpoint.Metadata)
	
	return &checkpoint, nil
}

// ListCheckpoints implements CheckpointStore.ListCheckpoints
func (s *SQLiteStore) ListCheckpoints(ctx context.Context, fromEpoch, toEpoch int64) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return nil, ErrClosed
	}
	
	query := "SELECT epoch, tree_id, root, tree_size, timestamp, signers, signature, metadata FROM checkpoints WHERE epoch >= ? AND epoch <= ? ORDER BY epoch"
	
	rows, err := s.db.QueryContext(ctx, query, fromEpoch, toEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}
	defer rows.Close()
	
	var result []*Checkpoint
	for rows.Next() {
		var checkpoint Checkpoint
		var signersJSON, metadataJSON string
		
		err := rows.Scan(
			&checkpoint.Epoch,
			&checkpoint.TreeID,
			&checkpoint.Root,
			&checkpoint.TreeSize,
			&checkpoint.Timestamp,
			&signersJSON,
			&checkpoint.Signature,
			&metadataJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan checkpoint row: %w", err)
		}
		
		json.Unmarshal([]byte(signersJSON), &checkpoint.Signers)
		json.Unmarshal([]byte(metadataJSON), &checkpoint.Metadata)
		
		result = append(result, &checkpoint)
	}
	
	return result, rows.Err()
}

// StoreStatusList implements StatusListStore.StoreStatusList
func (s *SQLiteStore) StoreStatusList(ctx context.Context, issuer, epoch, bitmapCID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return ErrClosed
	}
	
	query := "INSERT OR REPLACE INTO status_lists (issuer, epoch, bitmap_cid) VALUES (?, ?, ?)"
	
	_, err := s.db.ExecContext(ctx, query, issuer, epoch, bitmapCID)
	if err != nil {
		return fmt.Errorf("failed to store status list: %w", err)
	}
	
	return nil
}

// GetStatusList implements StatusListStore.GetStatusList
func (s *SQLiteStore) GetStatusList(ctx context.Context, issuer, epoch string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.closed {
		return "", ErrClosed
	}
	
	query := "SELECT bitmap_cid FROM status_lists WHERE issuer = ? AND epoch = ?"
	
	var bitmapCID string
	err := s.db.QueryRowContext(ctx, query, issuer, epoch).Scan(&bitmapCID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get status list: %w", err)
	}
	
	return bitmapCID, nil
}

// Close implements the Close method for all interfaces
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		return nil
	}
	
	s.closed = true
	return s.db.Close()
}