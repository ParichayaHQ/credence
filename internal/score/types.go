package score

import (
	"time"
)

// Score represents a computed trust score for an identity
type Score struct {
	DID       string    `json:"did"`
	Context   string    `json:"context"`
	Value     float64   `json:"value"`
	Epoch     int64     `json:"epoch"`
	Timestamp time.Time `json:"timestamp"`
	
	// Component breakdown
	Components ScoreComponents `json:"components"`
	
	// Metadata
	ComputedBy string `json:"computed_by"`
	Version    string `json:"version"`
}

// ScoreComponents breaks down the score into its constituent factors
type ScoreComponents struct {
	K float64 `json:"k"` // PoP/KYC factor
	A float64 `json:"a"` // Attestations factor
	V float64 `json:"v"` // Vouches factor (with sqrt aggregation)
	R float64 `json:"r"` // Reports factor (negative)
	T float64 `json:"t"` // Time factor
}

// ScoreFactors defines the weights for each component in the scoring algorithm
type ScoreFactors struct {
	Alpha float64 `json:"alpha"` // Weight for K factor
	Beta  float64 `json:"beta"`  // Weight for A factor
	Gamma float64 `json:"gamma"` // Weight for V factor
	Delta float64 `json:"delta"` // Weight for R factor (subtracted)
	Tau   float64 `json:"tau"`   // Weight for T factor
}

// DefaultScoreFactors returns the default factor weights
func DefaultScoreFactors() ScoreFactors {
	return ScoreFactors{
		Alpha: 0.3, // PoP/KYC weight
		Beta:  0.2, // Attestations weight
		Gamma: 0.3, // Vouches weight
		Delta: 0.1, // Reports weight (penalty)
		Tau:   0.1, // Time weight
	}
}

// VouchData represents a vouch relationship for scoring
type VouchData struct {
	FromDID   string    `json:"from_did"`
	ToDID     string    `json:"to_did"`
	Context   string    `json:"context"`
	Strength  float64   `json:"strength"` // q_ij in the formula
	Timestamp time.Time `json:"timestamp"`
	Epoch     int64     `json:"epoch"`
}

// AttestationData represents an attestation for scoring
type AttestationData struct {
	DID              string    `json:"did"`
	Context          string    `json:"context"`
	Type             string    `json:"type"` // employment, education, etc.
	IssuerDID        string    `json:"issuer_did"`
	IssuerReputation float64   `json:"issuer_reputation"`
	Weight           float64   `json:"weight"`
	Timestamp        time.Time `json:"timestamp"`
	Epoch            int64     `json:"epoch"`
}

// ReportData represents a report against an identity for scoring
type ReportData struct {
	ReporterDID string    `json:"reporter_did"`
	ReportedDID string    `json:"reported_did"`
	Context     string    `json:"context"`
	Severity    float64   `json:"severity"` // 0.0 to 1.0
	Adjudicated bool      `json:"adjudicated"`
	Upheld      bool      `json:"upheld"`
	Timestamp   time.Time `json:"timestamp"`
	Epoch       int64     `json:"epoch"`
}

// KYCData represents KYC/PoP credential data for scoring
type KYCData struct {
	DID        string    `json:"did"`
	Context    string    `json:"context"`
	Type       string    `json:"type"` // kyc, pop, etc.
	Level      int       `json:"level"` // trust level 1-5
	IssuerDID  string    `json:"issuer_did"`
	Weight     float64   `json:"weight"`
	Timestamp  time.Time `json:"timestamp"`
	Epoch      int64     `json:"epoch"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// TimeData represents time-based scoring factors
type TimeData struct {
	DID           string    `json:"did"`
	Context       string    `json:"context"`
	FirstActivity time.Time `json:"first_activity"`
	LastActivity  time.Time `json:"last_activity"`
	ActivityCount int64     `json:"activity_count"`
	Epoch         int64     `json:"epoch"`
}

// VouchBudget represents spending limits for vouching
type VouchBudget struct {
	DID            string  `json:"did"`
	Context        string  `json:"context"`
	Epoch          int64   `json:"epoch"`
	TotalBudget    float64 `json:"total_budget"`
	SpentBudget    float64 `json:"spent_budget"`
	RemainingBudget float64 `json:"remaining_budget"`
	ReputationBond float64 `json:"reputation_bond"`
}

// ScoreConfig holds configuration for score computation
type ScoreConfig struct {
	// Factor weights
	Factors ScoreFactors `json:"factors"`
	
	// Vouch configuration
	VouchCap          float64 `json:"vouch_cap"`           // Maximum vouch weight cap
	VouchDecayRate    float64 `json:"vouch_decay_rate"`    // Exponential decay rate
	VouchHalfLife     int64   `json:"vouch_half_life"`     // Half-life in epochs
	
	// Report configuration
	ReportDecayRate   float64 `json:"report_decay_rate"`   // Report decay rate
	ReportHalfLife    int64   `json:"report_half_life"`    // Report half-life in epochs
	
	// Time configuration
	TimeMaxGrowth     float64 `json:"time_max_growth"`     // Maximum time factor growth
	TimeInactivityDecay float64 `json:"time_inactivity_decay"` // Decay rate for inactivity
	
	// Budget configuration
	BaseBudget        float64 `json:"base_budget"`         // Base vouch budget per epoch
	BudgetMultiplier  float64 `json:"budget_multiplier"`   // Score-based budget multiplier
	
	// Diversity configuration
	DiversityPenalty  float64 `json:"diversity_penalty"`   // Penalty for low diversity
	CommunityThreshold float64 `json:"community_threshold"` // Threshold for community detection
	
	// Anti-collusion
	CollusionThreshold float64 `json:"collusion_threshold"` // Dense subgraph threshold
	CollusionPenalty   float64 `json:"collusion_penalty"`   // Penalty for collusion
}

// DefaultScoreConfig returns default configuration
func DefaultScoreConfig() *ScoreConfig {
	return &ScoreConfig{
		Factors:             DefaultScoreFactors(),
		VouchCap:            100.0,
		VouchDecayRate:      0.1,
		VouchHalfLife:       10,
		ReportDecayRate:     0.05,
		ReportHalfLife:      20,
		TimeMaxGrowth:       50.0,
		TimeInactivityDecay: 0.02,
		BaseBudget:          10.0,
		BudgetMultiplier:    0.1,
		DiversityPenalty:    0.2,
		CommunityThreshold:  0.7,
		CollusionThreshold:  0.8,
		CollusionPenalty:    0.5,
	}
}

// ScoreProof provides cryptographic proof of score computation
type ScoreProof struct {
	Score     *Score    `json:"score"`
	InputHash string    `json:"input_hash"`   // Hash of all input data
	Signature []byte    `json:"signature"`    // Signature over score and input hash
	PublicKey []byte    `json:"public_key"`   // Public key for verification
	Algorithm string    `json:"algorithm"`    // Signing algorithm
	Timestamp time.Time `json:"timestamp"`    // Proof timestamp
}