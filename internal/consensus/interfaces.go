package consensus

import (
	"context"
	"time"
)

// Checkpointor defines the main checkpoint committee service interface
type Checkpointor interface {
	// Start the checkpointor service
	Start(ctx context.Context) error
	
	// Stop the checkpointor service
	Stop(ctx context.Context) error
	
	// Get latest checkpoint
	GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error)
	
	// Get checkpoint by epoch
	GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error)
	
	// Submit partial signature (committee members)
	SubmitPartialSignature(ctx context.Context, partial *PartialSignature) error
	
	// Get current signing task (committee members)
	GetCurrentTask(ctx context.Context) (*SigningTask, error)
	
	// Force checkpoint creation (admin/testing)
	ForceCheckpoint(ctx context.Context) (*Checkpoint, error)
}

// CommitteeManager handles committee selection and rotation
type CommitteeManager interface {
	// Get current committee
	GetCurrentCommittee(ctx context.Context) (*Committee, error)
	
	// Get next committee (during rotation period)
	GetNextCommittee(ctx context.Context) (*Committee, error)
	
	// Select new committee using VRF
	SelectCommittee(ctx context.Context, vrfSeed []byte, candidates []CommitteeCandidate) (*Committee, error)
	
	// Rotate to next committee
	RotateCommittee(ctx context.Context) error
	
	// Check if member is in current committee
	IsMember(ctx context.Context, memberDID string) (bool, error)
}

// BLSAggregator handles BLS signature aggregation
type BLSAggregator interface {
	// Verify partial signature
	VerifyPartialSignature(partial *PartialSignature, memberPublicKey []byte) error
	
	// Aggregate partial signatures into threshold signature
	AggregateSignatures(partials []PartialSignature, threshold int) ([]byte, error)
	
	// Verify aggregated signature
	VerifyAggregatedSignature(checkpoint *Checkpoint, committeePublicKeys [][]byte) error
}

// VRFProvider handles VRF operations for committee selection
type VRFProvider interface {
	// Generate VRF proof for committee selection
	GenerateProof(privateKey []byte, seed []byte) ([]byte, []byte, error)
	
	// Verify VRF proof
	VerifyProof(publicKey []byte, seed []byte, proof []byte, output []byte) error
	
	// Get current VRF seed (from drand or other source)
	GetCurrentSeed(ctx context.Context) ([]byte, error)
}

// LogNodeClient defines interface to transparency log
type LogNodeClient interface {
	// Get latest signed tree head
	GetLatestSTH(ctx context.Context) (*SignedTreeHead, error)
	
	// Get signed tree head by size
	GetSTH(ctx context.Context, treeSize int64) (*SignedTreeHead, error)
}

// P2PPublisher defines interface for publishing checkpoints
type P2PPublisher interface {
	// Publish checkpoint to gossip network
	PublishCheckpoint(ctx context.Context, checkpoint *Checkpoint) error
	
	// Subscribe to checkpoint announcements
	SubscribeCheckpoints(ctx context.Context) (<-chan *Checkpoint, error)
}

// CheckpointStore defines storage interface for checkpoints
type CheckpointStore interface {
	// Store checkpoint
	StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error
	
	// Get checkpoint by epoch
	GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error)
	
	// Get latest checkpoint
	GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error)
	
	// List checkpoints in range
	ListCheckpoints(ctx context.Context, startEpoch, endEpoch int64) ([]*Checkpoint, error)
	
	// Store committee
	StoreCommittee(ctx context.Context, committee *Committee) error
	
	// Get current committee
	GetCurrentCommittee(ctx context.Context) (*Committee, error)
	
	// Store signing task
	StoreSigningTask(ctx context.Context, task *SigningTask) error
	
	// Get current signing task
	GetCurrentSigningTask(ctx context.Context) (*SigningTask, error)
	
	// Update signing task with partial signature
	AddPartialSignature(ctx context.Context, taskID string, partial *PartialSignature) error
}

// SignedTreeHead represents a log's signed tree head
type SignedTreeHead struct {
	TreeSize     int64     `json:"tree_size"`
	RootHash     []byte    `json:"root_hash"`
	Timestamp    time.Time `json:"timestamp"`
	Signature    []byte    `json:"signature"`
	SignerKeyID  string    `json:"signer_key_id"`
}

// CheckpointorService combines all interfaces for the main service
type CheckpointorService interface {
	Checkpointor
	CommitteeManager
}