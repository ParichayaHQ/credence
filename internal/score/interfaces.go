package score

import (
	"context"
)

// Engine defines the interface for trust score computation
type Engine interface {
	// ComputeScore computes the trust score for a DID in a given context
	ComputeScore(ctx context.Context, did, context string, epoch int64) (*Score, error)
	
	// ComputeScores computes scores for multiple DIDs efficiently
	ComputeScores(ctx context.Context, dids []string, context string, epoch int64) ([]*Score, error)
	
	// RecomputeScore recomputes a score with fresh data
	RecomputeScore(ctx context.Context, did, context string, epoch int64) (*Score, error)
	
	// GetFactors returns the current factor breakdown for a score
	GetFactors(ctx context.Context, did, context string, epoch int64) (*ScoreComponents, error)
	
	// ValidateScore validates a computed score against input data
	ValidateScore(ctx context.Context, score *Score) error
	
	// GetProof generates a cryptographic proof for a score
	GetProof(ctx context.Context, score *Score) (*ScoreProof, error)
	
	// VerifyProof verifies a cryptographic proof
	VerifyProof(ctx context.Context, proof *ScoreProof) error
}

// DataProvider defines the interface for accessing scoring input data
type DataProvider interface {
	// GetVouches returns all vouches for a DID up to a given epoch
	GetVouches(ctx context.Context, did, context string, maxEpoch int64) ([]*VouchData, error)
	
	// GetAttestations returns all attestations for a DID up to a given epoch
	GetAttestations(ctx context.Context, did, context string, maxEpoch int64) ([]*AttestationData, error)
	
	// GetReports returns all reports against a DID up to a given epoch
	GetReports(ctx context.Context, did, context string, maxEpoch int64) ([]*ReportData, error)
	
	// GetKYCData returns KYC/PoP data for a DID up to a given epoch
	GetKYCData(ctx context.Context, did, context string, maxEpoch int64) ([]*KYCData, error)
	
	// GetTimeData returns time-based activity data for a DID
	GetTimeData(ctx context.Context, did, context string, maxEpoch int64) (*TimeData, error)
	
	// GetScore returns a previously computed score if available
	GetScore(ctx context.Context, did, context string, epoch int64) (*Score, error)
	
	// StoreScore persists a computed score
	StoreScore(ctx context.Context, score *Score) error
}

// BudgetManager defines the interface for vouch budget management
type BudgetManager interface {
	// GetBudget returns the current vouch budget for a DID
	GetBudget(ctx context.Context, did, context string, epoch int64) (*VouchBudget, error)
	
	// SpendBudget attempts to spend from a DID's vouch budget
	SpendBudget(ctx context.Context, did, context string, epoch int64, amount float64) error
	
	// RefillBudget refills a DID's budget based on their current score
	RefillBudget(ctx context.Context, did, context string, epoch int64, score float64) error
	
	// GetSpentBudget returns total budget spent by a DID in an epoch
	GetSpentBudget(ctx context.Context, did, context string, epoch int64) (float64, error)
	
	// ValidateBudget ensures budget constraints are met
	ValidateBudget(ctx context.Context, did, context string, epoch int64) error
}

// GraphAnalyzer defines the interface for graph-based anti-collusion analysis
type GraphAnalyzer interface {
	// DetectCollusion analyzes the vouch graph for collusive behavior
	DetectCollusion(ctx context.Context, context string, epoch int64) ([]*CollusionCluster, error)
	
	// ComputeDiversity computes diversity score for a DID's vouch network
	ComputeDiversity(ctx context.Context, did, context string, epoch int64) (float64, error)
	
	// GetCommunityOverlap returns community overlap metrics for a DID
	GetCommunityOverlap(ctx context.Context, did, context string, epoch int64) (float64, error)
	
	// GetDenseSubgraphs identifies dense subgraphs that may indicate collusion
	GetDenseSubgraphs(ctx context.Context, context string, epoch int64, threshold float64) ([]*DenseSubgraph, error)
}

// DecayFunction defines the interface for decay computations
type DecayFunction interface {
	// ApplyDecay applies exponential decay to a value based on time elapsed
	ApplyDecay(value float64, elapsedEpochs int64, halfLife int64) float64
	
	// ApplyInactivityDecay applies decay for inactivity
	ApplyInactivityDecay(value float64, inactiveEpochs int64, decayRate float64) float64
	
	// ComputeTimeBonus computes time-based bonus with bounded growth
	ComputeTimeBonus(firstActivity, lastActivity int64, currentEpoch int64, maxGrowth float64) float64
}

// ScoreValidator defines the interface for score validation
type ScoreValidator interface {
	// ValidateInputData validates all input data for score computation
	ValidateInputData(ctx context.Context, did, context string, epoch int64) error
	
	// ValidateScoreRange ensures score is within expected bounds
	ValidateScoreRange(score float64) error
	
	// ValidateComponents ensures component values are consistent
	ValidateComponents(components *ScoreComponents, factors *ScoreFactors) error
	
	// ValidateIntegrity performs integrity checks on score data
	ValidateIntegrity(ctx context.Context, score *Score) error
}

// CollusionCluster represents a detected collusive group
type CollusionCluster struct {
	Context     string   `json:"context"`
	Epoch       int64    `json:"epoch"`
	Members     []string `json:"members"`      // DIDs in the cluster
	Density     float64  `json:"density"`      // Graph density
	VouchVolume float64  `json:"vouch_volume"` // Total vouch volume
	Confidence  float64  `json:"confidence"`   // Detection confidence
}

// DenseSubgraph represents a dense subgraph in the vouch network
type DenseSubgraph struct {
	Context   string   `json:"context"`
	Epoch     int64    `json:"epoch"`
	Nodes     []string `json:"nodes"`     // DIDs in subgraph
	Edges     int      `json:"edges"`     // Number of vouch edges
	Density   float64  `json:"density"`   // Subgraph density
	Suspicion float64  `json:"suspicion"` // Suspicion score
}

// ComputeRequest represents a score computation request
type ComputeRequest struct {
	DID     string `json:"did"`
	Context string `json:"context"`
	Epoch   int64  `json:"epoch"`
	
	// Optional parameters
	ForceRecompute bool `json:"force_recompute,omitempty"`
	IncludeProof   bool `json:"include_proof,omitempty"`
	IncludeFactors bool `json:"include_factors,omitempty"`
}

// ComputeResponse represents a score computation response
type ComputeResponse struct {
	Score      *Score          `json:"score"`
	Components *ScoreComponents `json:"components,omitempty"`
	Proof      *ScoreProof     `json:"proof,omitempty"`
	CachedFrom *int64          `json:"cached_from,omitempty"` // Epoch if from cache
}

// BatchComputeRequest represents a batch score computation request
type BatchComputeRequest struct {
	Requests []*ComputeRequest `json:"requests"`
}

// BatchComputeResponse represents a batch score computation response
type BatchComputeResponse struct {
	Responses []*ComputeResponse `json:"responses"`
	Errors    []string          `json:"errors,omitempty"`
}