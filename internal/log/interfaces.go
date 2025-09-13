package log

import (
	"context"
	"encoding/json"
	"time"
)

// TransparencyLog defines the interface for an append-only Merkle tree log
type TransparencyLog interface {
	// AppendLeaves adds new leaves to the log and returns their sequence numbers
	AppendLeaves(ctx context.Context, leaves []Leaf) (*AppendResult, error)
	
	// GetLeafByHash retrieves a leaf by its hash
	GetLeafByHash(ctx context.Context, leafHash []byte) (*Leaf, error)
	
	// GetLeavesByRange retrieves leaves within a sequence number range
	GetLeavesByRange(ctx context.Context, startSeq, endSeq int64) ([]Leaf, error)
	
	// GetInclusionProof returns an audit path proving inclusion of a leaf
	GetInclusionProof(ctx context.Context, leafHash []byte, treeSize int64) (*InclusionProof, error)
	
	// GetConsistencyProof returns proof that two tree states are consistent
	GetConsistencyProof(ctx context.Context, fromSize, toSize int64) (*ConsistencyProof, error)
	
	// GetSignedTreeHead returns the latest signed tree head
	GetSignedTreeHead(ctx context.Context) (*SignedTreeHead, error)
	
	// GetTreeSize returns the current tree size
	GetTreeSize(ctx context.Context) (int64, error)
	
	// Close cleanly shuts down the log
	Close() error
}

// Leaf represents a single entry in the transparency log
type Leaf struct {
	// Sequence number in the log (assigned by log)
	LeafIndex int64 `json:"leaf_index"`
	
	// Hash of the leaf content
	LeafHash []byte `json:"leaf_hash"`
	
	// The actual leaf content (typically event CID + metadata)
	LeafValue []byte `json:"leaf_value"`
	
	// Optional extra data (not included in hash)
	ExtraData []byte `json:"extra_data,omitempty"`
	
	// Timestamp when leaf was appended
	Timestamp time.Time `json:"timestamp"`
}

// AppendResult contains the result of appending leaves
type AppendResult struct {
	// New tree size after append
	TreeSize int64 `json:"tree_size"`
	
	// Root hash of the tree after append
	RootHash []byte `json:"root_hash"`
	
	// Sequence numbers assigned to the appended leaves
	LeafIndexes []int64 `json:"leaf_indexes"`
	
	// Signed tree head for the new state
	SignedTreeHead *SignedTreeHead `json:"signed_tree_head"`
}

// InclusionProof proves that a leaf is included in the tree
type InclusionProof struct {
	// Index of the leaf in the tree
	LeafIndex int64 `json:"leaf_index"`
	
	// Tree size this proof is valid for
	TreeSize int64 `json:"tree_size"`
	
	// Audit path from leaf to root
	AuditPath [][]byte `json:"audit_path"`
}

// ConsistencyProof proves that one tree state is consistent with another
type ConsistencyProof struct {
	// Size of the first tree
	FirstTreeSize int64 `json:"first_tree_size"`
	
	// Size of the second tree
	SecondTreeSize int64 `json:"second_tree_size"`
	
	// Proof path showing consistency
	ProofPath [][]byte `json:"proof_path"`
}

// SignedTreeHead represents a cryptographically signed tree state
type SignedTreeHead struct {
	// Size of the tree
	TreeSize int64 `json:"tree_size"`
	
	// Root hash of the tree
	RootHash []byte `json:"root_hash"`
	
	// Timestamp of this tree head
	Timestamp time.Time `json:"timestamp"`
	
	// Tree ID (for distinguishing multiple logs)
	TreeID int64 `json:"tree_id"`
	
	// Digital signature over the tree head
	Signature []byte `json:"signature"`
	
	// Algorithm used for signature
	SignatureAlgorithm string `json:"signature_algorithm"`
	
	// Key ID that signed this
	KeyID string `json:"key_id"`
}

// EventReference represents a reference to an event stored in the log
type EventReference struct {
	// Content ID of the event
	CID string `json:"cid"`
	
	// Hash of the event content
	ContentHash []byte `json:"content_hash"`
	
	// Event metadata
	Type      string    `json:"type"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Epoch     string    `json:"epoch"`
	Timestamp time.Time `json:"timestamp"`
	
	// Size of the event content
	ContentSize int64 `json:"content_size"`
}

// MarshalJSON implements json.Marshaler
func (e EventReference) MarshalJSON() ([]byte, error) {
	type Alias EventReference
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(e),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (e *EventReference) UnmarshalJSON(data []byte) error {
	type Alias EventReference
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	return json.Unmarshal(data, &aux)
}

// LogStats contains transparency log statistics
type LogStats struct {
	TreeSize      int64     `json:"tree_size"`
	TotalLeaves   int64     `json:"total_leaves"`
	LastAppend    time.Time `json:"last_append"`
	FirstAppend   time.Time `json:"first_append"`
	AppendRate    float64   `json:"append_rate"` // leaves per second
	StorageBytes  int64     `json:"storage_bytes"`
}

// Config holds transparency log configuration
type Config struct {
	// Tree ID for this log instance
	TreeID int64 `json:"tree_id"`
	
	// Signing key configuration
	SigningKey SigningKeyConfig `json:"signing_key"`
	
	// Storage configuration
	Storage StorageConfig `json:"storage"`
	
	// Batching configuration
	Batching BatchConfig `json:"batching"`
	
	// Hash algorithm to use
	HashAlgorithm string `json:"hash_algorithm"` // sha256, sha3-256
}

// SigningKeyConfig configures the signing key for tree heads
type SigningKeyConfig struct {
	// Algorithm (ed25519, ecdsa-p256)
	Algorithm string `json:"algorithm"`
	
	// Private key (PEM format or file path)
	PrivateKey string `json:"private_key"`
	
	// Key ID for identification
	KeyID string `json:"key_id"`
}

// StorageConfig configures log storage
type StorageConfig struct {
	// Backend type (memory, rocksdb, postgres)
	Backend string `json:"backend"`
	
	// Connection string or path
	ConnectionString string `json:"connection_string"`
	
	// Maximum storage size
	MaxStorageSize int64 `json:"max_storage_size"`
}

// BatchConfig configures batching behavior
type BatchConfig struct {
	// Maximum leaves per batch
	MaxBatchSize int `json:"max_batch_size"`
	
	// Maximum time to wait before flushing batch
	MaxBatchDelay time.Duration `json:"max_batch_delay"`
	
	// Enable automatic batching
	EnableBatching bool `json:"enable_batching"`
}

// DefaultConfig returns default transparency log configuration
func DefaultConfig() *Config {
	return &Config{
		TreeID: 1,
		SigningKey: SigningKeyConfig{
			Algorithm:  "ed25519",
			KeyID:      "default",
		},
		Storage: StorageConfig{
			Backend:        "memory",
			MaxStorageSize: 1024 * 1024 * 1024, // 1GB
		},
		Batching: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchDelay:  5 * time.Second,
			EnableBatching: true,
		},
		HashAlgorithm: "sha256",
	}
}