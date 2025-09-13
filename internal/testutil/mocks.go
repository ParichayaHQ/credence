package testutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ParichayaHQ/credence/internal/cid"
	"github.com/ParichayaHQ/credence/internal/events"
	"github.com/ParichayaHQ/credence/pkg/types"
	ipfscid "github.com/ipfs/go-cid"
)

// MockStorageService provides an in-memory storage implementation for testing
type MockStorageService struct {
	mu      sync.RWMutex
	content map[string][]byte
	index   map[string]map[string]interface{}
}

// NewMockStorageService creates a new mock storage service
func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		content: make(map[string][]byte),
		index:   make(map[string]map[string]interface{}),
	}
}

func (m *MockStorageService) Put(ctx context.Context, data []byte) (ipfscid.Cid, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Generate CID for the data
	cidGen := cid.NewCIDGenerator()
	c, err := cidGen.GenerateFromBytes(data)
	if err != nil {
		return ipfscid.Undef, err
	}
	
	m.content[c.String()] = make([]byte, len(data))
	copy(m.content[c.String()], data)
	
	return c, nil
}

func (m *MockStorageService) Get(ctx context.Context, c ipfscid.Cid) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data, exists := m.content[c.String()]
	if !exists {
		return nil, ErrContentNotFound
	}
	
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func (m *MockStorageService) Has(ctx context.Context, c ipfscid.Cid) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	_, exists := m.content[c.String()]
	return exists, nil
}

func (m *MockStorageService) Delete(ctx context.Context, c ipfscid.Cid) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.content, c.String())
	delete(m.index, c.String())
	return nil
}

func (m *MockStorageService) Size(ctx context.Context, c ipfscid.Cid) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data, exists := m.content[c.String()]
	if !exists {
		return 0, ErrContentNotFound
	}
	
	return int64(len(data)), nil
}

func (m *MockStorageService) Index(ctx context.Context, c ipfscid.Cid, metadata map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.index[c.String()] = metadata
	return nil
}

// MockEventService provides a mock event service for testing
type MockEventService struct {
	mu     sync.RWMutex
	events map[string]*events.Event
	byDID  map[string][]*events.Event
}

// NewMockEventService creates a new mock event service
func NewMockEventService() *MockEventService {
	return &MockEventService{
		events: make(map[string]*events.Event),
		byDID:  make(map[string][]*events.Event),
	}
}

func (m *MockEventService) SubmitEvent(ctx context.Context, event *events.Event) (*types.EventReceipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Generate CID for the event
	cidGen := cid.NewCIDGenerator()
	canonical, err := events.CanonicalizeEventWithSignature(event)
	if err != nil {
		return nil, err
	}
	
	c, err := cidGen.GenerateFromBytes(canonical)
	if err != nil {
		return nil, err
	}
	
	// Store event
	m.events[c.String()] = event
	
	// Index by DID
	m.byDID[event.From] = append(m.byDID[event.From], event)
	if event.To != "" {
		m.byDID[event.To] = append(m.byDID[event.To], event)
	}
	
	return &types.EventReceipt{
		CID:       c,
		Timestamp: time.Now(),
		Status:    "processed",
	}, nil
}

func (m *MockEventService) GetEvent(ctx context.Context, c ipfscid.Cid) (*events.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	event, exists := m.events[c.String()]
	if !exists {
		return nil, ErrContentNotFound
	}
	
	return event, nil
}

func (m *MockEventService) ValidateEvent(ctx context.Context, event *events.Event) error {
	return events.ValidateEvent(event)
}

func (m *MockEventService) GetEventsByDID(ctx context.Context, did string, eventType events.EventType, limit int) ([]*events.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	allEvents, exists := m.byDID[did]
	if !exists {
		return []*events.Event{}, nil
	}
	
	// Filter by type if specified
	var filtered []*events.Event
	for _, event := range allEvents {
		if eventType == "" || event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	
	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	
	return filtered, nil
}

// MockScorerService provides a mock scorer service for testing
type MockScorerService struct {
	mu     sync.RWMutex
	scores map[string]*types.ScoreRecord
}

// NewMockScorerService creates a new mock scorer service
func NewMockScorerService() *MockScorerService {
	return &MockScorerService{
		scores: make(map[string]*types.ScoreRecord),
	}
}

func (m *MockScorerService) ComputeScore(ctx context.Context, did, context string) (*types.ScoreRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := did + ":" + context
	
	// If score doesn't exist, create a default one
	if _, exists := m.scores[key]; !exists {
		m.scores[key] = CreateTestScoreRecord(did, context, 50.0) // Default score
	}
	
	return m.scores[key], nil
}

func (m *MockScorerService) GetScore(ctx context.Context, did, context string) (*types.ScoreRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	key := did + ":" + context
	score, exists := m.scores[key]
	if !exists {
		return nil, ErrContentNotFound
	}
	
	return score, nil
}

func (m *MockScorerService) UpdateScores(ctx context.Context, eventCIDs []ipfscid.Cid) error {
	// Mock implementation - in reality would process events and update scores
	return nil
}

func (m *MockScorerService) GenerateThresholdProof(ctx context.Context, did, context string, threshold float64, nonce string) (*types.ThresholdProof, error) {
	score, err := m.GetScore(ctx, did, context)
	if err != nil {
		return nil, err
	}
	
	return &types.ThresholdProof{
		Proof:     "mock-zk-proof",
		Context:   context,
		Threshold: threshold,
		Nonce:     nonce,
		Checkpoint: score.Checkpoint,
	}, nil
}

func (m *MockScorerService) VerifyThresholdProof(ctx context.Context, proof *types.ThresholdProof) (bool, error) {
	// Mock verification - always returns true
	return true, nil
}

// SetScore allows tests to set specific scores
func (m *MockScorerService) SetScore(did, context string, score float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := did + ":" + context
	m.scores[key] = CreateTestScoreRecord(did, context, score)
}


// TestErrors defines common test errors
var (
	ErrContentNotFound = fmt.Errorf("content not found")
)

// MockCheckpointService provides mock checkpoint functionality
type MockCheckpointService struct {
	mu          sync.RWMutex
	checkpoints map[int64]*types.Checkpoint
	latest      *types.Checkpoint
}

func NewMockCheckpointService() *MockCheckpointService {
	return &MockCheckpointService{
		checkpoints: make(map[int64]*types.Checkpoint),
	}
}

func (m *MockCheckpointService) CreateCheckpoint(ctx context.Context, treeRoot string, treeID string, epoch int64) (*types.Checkpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	checkpoint := CreateTestCheckpoint(epoch, treeRoot)
	m.checkpoints[epoch] = checkpoint
	m.latest = checkpoint
	
	return checkpoint, nil
}

func (m *MockCheckpointService) GetCheckpoint(ctx context.Context, epoch int64) (*types.Checkpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	checkpoint, exists := m.checkpoints[epoch]
	if !exists {
		return nil, ErrContentNotFound
	}
	
	return checkpoint, nil
}

func (m *MockCheckpointService) GetLatestCheckpoint(ctx context.Context) (*types.Checkpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.latest == nil {
		return nil, ErrContentNotFound
	}
	
	return m.latest, nil
}

func (m *MockCheckpointService) VerifyCheckpoint(ctx context.Context, checkpoint *types.Checkpoint) (bool, error) {
	// Mock verification - always returns true
	return true, nil
}

func (m *MockCheckpointService) SubmitPartialSignature(ctx context.Context, checkpointData []byte, signerIndex int, signature []byte) error {
	// Mock implementation
	return nil
}