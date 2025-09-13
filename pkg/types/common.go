package types

import (
	"time"

	"github.com/ipfs/go-cid"
)

// Checkpoint represents a system checkpoint with BLS threshold signature
type Checkpoint struct {
	Root      string    `json:"root" validate:"required"`
	TreeID    string    `json:"tree_id" validate:"required"`
	Epoch     int64     `json:"epoch" validate:"required,min=0"`
	Signers   []string  `json:"signers" validate:"required,min=1"`
	Signature string    `json:"threshold_sig" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

// Ruleset represents the scoring algorithm parameters
type Ruleset struct {
	ID          string             `json:"id" validate:"required"`
	Weights     RulesetWeights     `json:"weights" validate:"required"`
	Caps        RulesetCaps        `json:"caps" validate:"required"`
	Vouch       VouchRules         `json:"vouch" validate:"required"`
	Decay       DecayRules         `json:"decay" validate:"required"`
	Diversity   DiversityRules     `json:"diversity" validate:"required"`
	Adjudication AdjudicationRules `json:"adjudication" validate:"required"`
	Signature   string             `json:"signature" validate:"required"`
	ValidFrom   time.Time          `json:"validFrom" validate:"required"`
	TimeLockDays int               `json:"timeLockDays" validate:"min=0"`
}

// RulesetWeights defines the scoring algorithm weights
type RulesetWeights struct {
	Alpha float64 `json:"alpha" validate:"min=0,max=1"` // K factor weight
	Beta  float64 `json:"beta" validate:"min=0,max=1"`  // A factor weight  
	Gamma float64 `json:"gamma" validate:"min=0,max=1"` // V factor weight
	Delta float64 `json:"delta" validate:"min=0,max=1"` // R factor weight
	Tau   float64 `json:"tau" validate:"min=0,max=1"`   // T factor weight
}

// RulesetCaps defines maximum values for each scoring factor
type RulesetCaps struct {
	K float64 `json:"K" validate:"min=0"`
	A float64 `json:"A" validate:"min=0"`
	V float64 `json:"V" validate:"min=0"`
	R float64 `json:"R" validate:"min=0"`
	T float64 `json:"T" validate:"min=0"`
}

// VouchRules defines vouch-related parameters
type VouchRules struct {
	BudgetBase   int     `json:"budget_base" validate:"min=1"`
	BudgetLambda float64 `json:"budget_lambda" validate:"min=0"`
	Aggregation  string  `json:"agg" validate:"oneof=sqrt linear"`
	PerEpoch     string  `json:"per_epoch" validate:"oneof=monthly weekly daily"`
	Bond         BondRules `json:"bond" validate:"required"`
}

// BondRules defines reputation bonding for vouches
type BondRules struct {
	Type         string  `json:"type" validate:"oneof=reputation stake"`
	DecayOnAbuse float64 `json:"decay_on_abuse" validate:"min=0,max=1"`
}

// DecayRules defines time-based decay parameters
type DecayRules struct {
	HalfLifeDays map[string]int `json:"half_life_days" validate:"required"`
}

// DiversityRules defines diversity-related parameters
type DiversityRules struct {
	CommunityOverlapPenalty float64 `json:"community_overlap_penalty" validate:"min=0,max=1"`
	MinClusters            int     `json:"min_clusters" validate:"min=1"`
}

// AdjudicationRules defines adjudication process parameters
type AdjudicationRules struct {
	PoolSize         int `json:"pool_size" validate:"min=3"`
	Quorum          int `json:"quorum" validate:"min=2"`
	AppealWindowDays int `json:"appeal_window_days" validate:"min=1"`
}

// ScoreRecord represents a computed trust score with proofs
type ScoreRecord struct {
	DID        string         `json:"did" validate:"required,did"`
	Context    string         `json:"ctx" validate:"required"`
	Ruleset    RulesetRef     `json:"ruleset" validate:"required"`
	Checkpoint CheckpointRef  `json:"checkpoint" validate:"required"`
	Score      float64        `json:"score" validate:"min=0"`
	Factors    ScoreFactors   `json:"factors" validate:"required"`
	Proofs     ScoreProofs    `json:"proofs" validate:"required"`
	ComputedAt time.Time      `json:"computed_at" validate:"required"`
}

// RulesetRef references a specific ruleset version
type RulesetRef struct {
	ID        string    `json:"id" validate:"required"`
	Hash      string    `json:"hash" validate:"required"`
	ValidFrom time.Time `json:"validFrom" validate:"required"`
}

// CheckpointRef references a specific checkpoint
type CheckpointRef struct {
	Root    string   `json:"root" validate:"required"`
	Epoch   int64    `json:"epoch" validate:"min=0"`
	Signature string `json:"sig" validate:"required"`
	Signers []string `json:"signers" validate:"required,min=1"`
}

// ScoreFactors represents the individual scoring factors
type ScoreFactors struct {
	K string `json:"K"` // KYC/PoP factor (commitment)
	A string `json:"A"` // Additional credentials factor (commitment)
	V string `json:"V"` // Vouch factor (commitment)
	R string `json:"R"` // Reports factor (commitment)
	T string `json:"T"` // Time factor (commitment)
}

// ScoreProofs contains cryptographic proofs for score verification
type ScoreProofs struct {
	Inclusion   []InclusionProof   `json:"inclusion" validate:"required"`
	Consistency []ConsistencyProof `json:"consistency" validate:"required"`
	StatusLists []StatusListProof  `json:"statuslists" validate:"required"`
}

// InclusionProof represents a Merkle inclusion proof
type InclusionProof struct {
	CID  string   `json:"cid" validate:"required"`
	Path []string `json:"path" validate:"required"`
}

// ConsistencyProof represents a Merkle consistency proof
type ConsistencyProof struct {
	OldRoot string   `json:"old" validate:"required"`
	NewRoot string   `json:"new" validate:"required"`
	Path    []string `json:"path" validate:"required"`
}

// StatusListProof represents proof of credential status
type StatusListProof struct {
	Issuer    string `json:"issuer" validate:"required,did"`
	Epoch     int64  `json:"epoch" validate:"min=0"`
	BitmapCID string `json:"bitmapCID" validate:"required"`
}

// ThresholdProof represents a zero-knowledge threshold proof
type ThresholdProof struct {
	Proof      string        `json:"proof" validate:"required"`
	Checkpoint CheckpointRef `json:"checkpoint" validate:"required"`
	Context    string        `json:"context" validate:"required"`
	Threshold  float64       `json:"threshold" validate:"min=0"`
	Nonce      string        `json:"nonce" validate:"required"`
}

// PeerInfo represents information about a peer in the network
type PeerInfo struct {
	ID        string    `json:"id" validate:"required"`
	Addresses []string  `json:"addresses" validate:"required"`
	Protocols []string  `json:"protocols" validate:"required"`
	LastSeen  time.Time `json:"last_seen" validate:"required"`
	Score     float64   `json:"score" validate:"min=0"`
}

// EventReceipt represents a receipt for a submitted event
type EventReceipt struct {
	CID       cid.Cid   `json:"cid"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status" validate:"oneof=queued processed failed"`
	Message   string    `json:"message,omitempty"`
}

// NetworkStatus represents the current state of the network
type NetworkStatus struct {
	ConnectedPeers int       `json:"connected_peers"`
	KnownPeers     int       `json:"known_peers"`
	LastCheckpoint int64     `json:"last_checkpoint"`
	SyncStatus     string    `json:"sync_status" validate:"oneof=synced syncing offline"`
	Timestamp      time.Time `json:"timestamp"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Service    string            `json:"service" validate:"required"`
	Status     string            `json:"status" validate:"oneof=healthy degraded unhealthy"`
	LastCheck  time.Time         `json:"last_check"`
	Metrics    map[string]string `json:"metrics,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
}