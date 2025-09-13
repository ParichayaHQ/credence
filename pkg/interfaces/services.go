package interfaces

import (
	"context"
	"time"

	"github.com/ParichayaHQ/credence/internal/events"
	"github.com/ParichayaHQ/credence/pkg/types"
	"github.com/ipfs/go-cid"
)

// EventService interface for event-related operations
type EventService interface {
	// SubmitEvent submits a new event to the network
	SubmitEvent(ctx context.Context, event *events.Event) (*types.EventReceipt, error)
	
	// GetEvent retrieves an event by CID
	GetEvent(ctx context.Context, cid cid.Cid) (*events.Event, error)
	
	// ValidateEvent validates an event
	ValidateEvent(ctx context.Context, event *events.Event) error
	
	// GetEventsByDID retrieves events involving a specific DID
	GetEventsByDID(ctx context.Context, did string, eventType events.EventType, limit int) ([]*events.Event, error)
}

// StorageService interface for content storage operations
type StorageService interface {
	// Put stores content and returns its CID
	Put(ctx context.Context, data []byte) (cid.Cid, error)
	
	// Get retrieves content by CID
	Get(ctx context.Context, cid cid.Cid) ([]byte, error)
	
	// Has checks if content exists
	Has(ctx context.Context, cid cid.Cid) (bool, error)
	
	// Delete removes content
	Delete(ctx context.Context, cid cid.Cid) error
	
	// Size returns content size
	Size(ctx context.Context, cid cid.Cid) (int64, error)
	
	// Index adds content to searchable index
	Index(ctx context.Context, cid cid.Cid, metadata map[string]interface{}) error
}

// P2PService interface for peer-to-peer networking
type P2PService interface {
	// Start starts the P2P service
	Start(ctx context.Context) error
	
	// Stop stops the P2P service
	Stop(ctx context.Context) error
	
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, data []byte) error
	
	// Subscribe subscribes to a topic
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
	
	// FindProviders finds providers for content
	FindProviders(ctx context.Context, cid cid.Cid) ([]types.PeerInfo, error)
	
	// Connect connects to a peer
	Connect(ctx context.Context, peerID string) error
	
	// GetNetworkStatus returns current network status
	GetNetworkStatus(ctx context.Context) (*types.NetworkStatus, error)
}

// LogService interface for transparency log operations
type LogService interface {
	// AppendEvents appends events to the log
	AppendEvents(ctx context.Context, events []cid.Cid) ([]int64, error)
	
	// GetInclusionProof gets inclusion proof for a CID
	GetInclusionProof(ctx context.Context, cid cid.Cid, treeSize int64) (*types.InclusionProof, error)
	
	// GetConsistencyProof gets consistency proof between two tree sizes
	GetConsistencyProof(ctx context.Context, oldSize, newSize int64) (*types.ConsistencyProof, error)
	
	// GetLatestRoot gets the latest tree root
	GetLatestRoot(ctx context.Context) (string, int64, error)
	
	// VerifyInclusion verifies an inclusion proof
	VerifyInclusion(ctx context.Context, proof *types.InclusionProof, root string, treeSize int64) (bool, error)
}

// ScorerService interface for trust score computation
type ScorerService interface {
	// ComputeScore computes trust score for a DID
	ComputeScore(ctx context.Context, did, context string) (*types.ScoreRecord, error)
	
	// GetScore retrieves cached score for a DID
	GetScore(ctx context.Context, did, context string) (*types.ScoreRecord, error)
	
	// UpdateScores updates scores based on new events
	UpdateScores(ctx context.Context, eventCIDs []cid.Cid) error
	
	// GenerateThresholdProof generates a zero-knowledge threshold proof
	GenerateThresholdProof(ctx context.Context, did, context string, threshold float64, nonce string) (*types.ThresholdProof, error)
	
	// VerifyThresholdProof verifies a threshold proof
	VerifyThresholdProof(ctx context.Context, proof *types.ThresholdProof) (bool, error)
}

// CheckpointService interface for checkpoint management
type CheckpointService interface {
	// CreateCheckpoint creates a new checkpoint
	CreateCheckpoint(ctx context.Context, treeRoot string, treeID string, epoch int64) (*types.Checkpoint, error)
	
	// GetCheckpoint retrieves a checkpoint by epoch
	GetCheckpoint(ctx context.Context, epoch int64) (*types.Checkpoint, error)
	
	// GetLatestCheckpoint gets the most recent checkpoint
	GetLatestCheckpoint(ctx context.Context) (*types.Checkpoint, error)
	
	// VerifyCheckpoint verifies a checkpoint signature
	VerifyCheckpoint(ctx context.Context, checkpoint *types.Checkpoint) (bool, error)
	
	// SubmitPartialSignature submits a partial BLS signature
	SubmitPartialSignature(ctx context.Context, checkpointData []byte, signerIndex int, signature []byte) error
}

// WalletService interface for wallet operations
type WalletService interface {
	// CreateEvent creates and signs a new event
	CreateEvent(ctx context.Context, eventType events.EventType, to, context, epoch string, payload interface{}) (*events.Event, error)
	
	// SignEvent signs an event
	SignEvent(ctx context.Context, event *events.SignableEvent) (*events.Event, error)
	
	// GetPublicKey returns the wallet's public key
	GetPublicKey(ctx context.Context) (string, error)
	
	// GetDID returns the wallet's DID
	GetDID(ctx context.Context) (string, error)
	
	// ImportKey imports a private key from seed
	ImportKey(ctx context.Context, seed []byte) error
	
	// ExportSeed exports the wallet seed (requires auth)
	ExportSeed(ctx context.Context, auth string) ([]byte, error)
}

// RulesService interface for ruleset management
type RulesService interface {
	// GetActiveRuleset returns the currently active ruleset
	GetActiveRuleset(ctx context.Context) (*types.Ruleset, error)
	
	// GetRuleset retrieves a ruleset by ID
	GetRuleset(ctx context.Context, rulesetID string) (*types.Ruleset, error)
	
	// ProposeRuleset proposes a new ruleset (governance)
	ProposeRuleset(ctx context.Context, ruleset *types.Ruleset) error
	
	// ValidateRuleset validates a ruleset
	ValidateRuleset(ctx context.Context, ruleset *types.Ruleset) error
}

// HealthService interface for service health monitoring
type HealthService interface {
	// GetHealth returns the health status of the service
	GetHealth(ctx context.Context) (*types.ServiceHealth, error)
	
	// GetMetrics returns service metrics
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
	
	// IsReady returns true if the service is ready to serve requests
	IsReady(ctx context.Context) (bool, error)
	
	// IsAlive returns true if the service is running
	IsAlive(ctx context.Context) (bool, error)
}

// ConfigService interface for configuration management
type ConfigService interface {
	// Get retrieves a configuration value
	Get(key string) (interface{}, error)
	
	// Set sets a configuration value
	Set(key string, value interface{}) error
	
	// GetString retrieves a string configuration value
	GetString(key string) (string, error)
	
	// GetInt retrieves an integer configuration value
	GetInt(key string) (int, error)
	
	// GetBool retrieves a boolean configuration value
	GetBool(key string) (bool, error)
	
	// GetDuration retrieves a duration configuration value
	GetDuration(key string) (time.Duration, error)
	
	// Watch watches for configuration changes
	Watch(key string) (<-chan interface{}, error)
}