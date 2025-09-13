package consensus

import (
	"crypto/ed25519"
	"time"
)

// Checkpoint represents a signed aggregation of log roots
type Checkpoint struct {
	// Root hash from the transparency log
	Root []byte `json:"root"`
	
	// Epoch identifier (timestamp-based or sequential)
	Epoch int64 `json:"epoch"`
	
	// Tree size at checkpoint creation
	TreeSize int64 `json:"tree_size"`
	
	// Committee members who signed this checkpoint
	Signers []string `json:"signers"`
	
	// BLS threshold signature
	Signature []byte `json:"signature"`
	
	// Timestamp when checkpoint was created
	Timestamp time.Time `json:"timestamp"`
}

// PartialSignature represents a committee member's partial BLS signature
type PartialSignature struct {
	// Committee member DID
	SignerDID string `json:"signer_did"`
	
	// Partial BLS signature
	Signature []byte `json:"signature"`
	
	// Root hash being signed
	Root []byte `json:"root"`
	
	// Epoch being signed
	Epoch int64 `json:"epoch"`
	
	// Timestamp of signature
	Timestamp time.Time `json:"timestamp"`
}

// Committee represents the current checkpoint committee
type Committee struct {
	// Current committee members (DIDs)
	Members []string `json:"members"`
	
	// Threshold for valid checkpoint (t of n)
	Threshold int `json:"threshold"`
	
	// Committee epoch/rotation period
	Epoch int64 `json:"epoch"`
	
	// VRF seed used for selection
	VRFSeed []byte `json:"vrf_seed"`
	
	// Committee term start time
	StartTime time.Time `json:"start_time"`
	
	// Committee term end time
	EndTime time.Time `json:"end_time"`
}

// CommitteeCandidate represents a potential committee member
type CommitteeCandidate struct {
	DID       string         `json:"did"`
	PublicKey ed25519.PublicKey `json:"public_key"`
	VRFProof  []byte         `json:"vrf_proof"`
	Stake     int64          `json:"stake"`
}

// SigningTask represents work for committee members
type SigningTask struct {
	// Root hash to sign
	Root []byte `json:"root"`
	
	// Epoch identifier
	Epoch int64 `json:"epoch"`
	
	// Tree size
	TreeSize int64 `json:"tree_size"`
	
	// Deadline for collecting signatures
	Deadline time.Time `json:"deadline"`
	
	// Current partial signatures collected
	Partials []PartialSignature `json:"partials"`
	
	// Task status
	Status TaskStatus `json:"status"`
}

// TaskStatus represents the state of a signing task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCollecting TaskStatus = "collecting"
	TaskStatusComplete  TaskStatus = "complete"
	TaskStatusFailed    TaskStatus = "failed"
)

// CheckpointorConfig holds configuration for the checkpointor service
type CheckpointorConfig struct {
	// Checkpoint creation interval
	CheckpointInterval time.Duration `json:"checkpoint_interval"`
	
	// Committee rotation period
	RotationPeriod time.Duration `json:"rotation_period"`
	
	// Signature collection timeout
	SignatureTimeout time.Duration `json:"signature_timeout"`
	
	// Committee size
	CommitteeSize int `json:"committee_size"`
	
	// BLS signature threshold (t of n)
	SignatureThreshold int `json:"signature_threshold"`
	
	// HTTP server configuration
	ServerAddress string        `json:"server_address"`
	ServerPort    int           `json:"server_port"`
	ReadTimeout   time.Duration `json:"read_timeout"`
	WriteTimeout  time.Duration `json:"write_timeout"`
}

// DefaultCheckpointorConfig returns default configuration
func DefaultCheckpointorConfig() *CheckpointorConfig {
	return &CheckpointorConfig{
		CheckpointInterval: 10 * time.Minute,
		RotationPeriod:     30 * 24 * time.Hour, // 30 days
		SignatureTimeout:   5 * time.Minute,
		CommitteeSize:      7,
		SignatureThreshold: 5, // 5 of 7
		ServerAddress:      "0.0.0.0",
		ServerPort:         8083,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
	}
}