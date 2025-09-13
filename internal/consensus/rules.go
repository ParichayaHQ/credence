package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// RuleSet represents a set of consensus rules
type RuleSet struct {
	// Rule set metadata
	ID           string    `json:"id"`
	Version      string    `json:"version"`
	ValidFrom    time.Time `json:"valid_from"`
	TimeLockDays int       `json:"timelock_days"`
	CreatedAt    time.Time `json:"created_at"`
	
	// Scoring parameters
	ScoringWeights ScoringWeights `json:"scoring_weights"`
	
	// Budget and limits
	VouchBudgets    map[string]int `json:"vouch_budgets"`     // context -> budget per epoch
	ScoreCaps       ScoreCaps      `json:"score_caps"`        // various scoring caps
	DecayParameters DecayParams    `json:"decay_parameters"`  // time-based decay
	
	// Committee parameters
	CommitteeSize      int           `json:"committee_size"`
	SignatureThreshold int           `json:"signature_threshold"`
	RotationPeriod     time.Duration `json:"rotation_period"`
	
	// Network parameters
	CheckpointInterval time.Duration `json:"checkpoint_interval"`
	SignatureTimeout   time.Duration `json:"signature_timeout"`
	
	// Adjudication parameters
	DisputeWindow    time.Duration `json:"dispute_window"`
	SlashingPenalty  float64       `json:"slashing_penalty"`
	ReputationDecay  float64       `json:"reputation_decay"`
	
	// Digital signature
	Hash      []byte `json:"hash"`      // SHA256 of canonical JSON
	Signature []byte `json:"signature"` // Ed25519 signature
	SignerDID string `json:"signer_did"`
}

// ScoringWeights defines weights for trust score calculation
type ScoringWeights struct {
	KYCWeight        float64 `json:"kyc_weight"`         // α - KYC credential weight
	ActivityWeight   float64 `json:"activity_weight"`    // β - Activity weight
	VouchWeight      float64 `json:"vouch_weight"`       // γ - Vouch weight
	ReportWeight     float64 `json:"report_weight"`      // δ - Report penalty weight
	TenureWeight     float64 `json:"tenure_weight"`      // τ - Tenure bonus weight
}

// ScoreCaps defines various caps for scoring
type ScoreCaps struct {
	MaxIndividualScore float64 `json:"max_individual_score"`  // Maximum score for any individual
	MaxVouchValue      float64 `json:"max_vouch_value"`       // Cap per vouch
	DiversityBonus     float64 `json:"diversity_bonus"`       // Bonus for diverse vouchers
	MinScoreThreshold  float64 `json:"min_score_threshold"`   // Minimum viable score
}

// DecayParams defines time-based decay parameters
type DecayParams struct {
	VouchDecayHalfLife  time.Duration `json:"vouch_decay_halflife"`  // Vouch value decay
	ScoreDecayRate      float64       `json:"score_decay_rate"`      // Overall score decay
	ActivityWindow      time.Duration `json:"activity_window"`       // Recent activity window
	TenureBonusGracePeriod time.Duration `json:"tenure_bonus_grace"`   // Grace period for new members
}

// ProposedRuleChange represents a proposed change to rules
type ProposedRuleChange struct {
	ProposalID    string    `json:"proposal_id"`
	ProposerDID   string    `json:"proposer_did"`
	ProposedAt    time.Time `json:"proposed_at"`
	ActivationDate time.Time `json:"activation_date"`
	
	// The new ruleset
	NewRuleSet *RuleSet `json:"new_ruleset"`
	
	// Justification
	Title       string `json:"title"`
	Description string `json:"description"`
	Rationale   string `json:"rationale"`
	
	// Committee approvals
	Approvals []CommitteeApproval `json:"approvals"`
	Status    ProposalStatus      `json:"status"`
	
	// Signatures
	Hash      []byte `json:"hash"`
	Signature []byte `json:"signature"`
}

// CommitteeApproval represents approval from a committee member
type CommitteeApproval struct {
	MemberDID   string    `json:"member_did"`
	ApprovedAt  time.Time `json:"approved_at"`
	Signature   []byte    `json:"signature"`
	Comments    string    `json:"comments,omitempty"`
}

// ProposalStatus represents the status of a rule change proposal
type ProposalStatus string

const (
	ProposalStatusPending   ProposalStatus = "pending"
	ProposalStatusApproved  ProposalStatus = "approved"
	ProposalStatusRejected  ProposalStatus = "rejected"
	ProposalStatusExecuted  ProposalStatus = "executed"
	ProposalStatusExpired   ProposalStatus = "expired"
)

// RulesRegistry manages rule sets and governance
type RulesRegistry interface {
	// Get the currently active ruleset
	GetActiveRuleSet(ctx context.Context) (*RuleSet, error)
	
	// Get a specific ruleset by ID
	GetRuleSet(ctx context.Context, id string) (*RuleSet, error)
	
	// List all rulesets
	ListRuleSets(ctx context.Context) ([]*RuleSet, error)
	
	// Propose a new ruleset
	ProposeRuleChange(ctx context.Context, proposal *ProposedRuleChange) error
	
	// Approve a proposed rule change (committee member)
	ApproveProposal(ctx context.Context, proposalID string, approval *CommitteeApproval) error
	
	// Execute an approved proposal (after timelock)
	ExecuteProposal(ctx context.Context, proposalID string) error
	
	// Get proposal by ID
	GetProposal(ctx context.Context, proposalID string) (*ProposedRuleChange, error)
	
	// List all proposals
	ListProposals(ctx context.Context, status ProposalStatus) ([]*ProposedRuleChange, error)
}

// DefaultRulesRegistry implements RulesRegistry
type DefaultRulesRegistry struct {
	store         RulesStore
	committee     CommitteeManager
	timeProvider  TimeProvider
}

// RulesStore defines storage interface for rules
type RulesStore interface {
	StoreRuleSet(ctx context.Context, ruleSet *RuleSet) error
	GetRuleSet(ctx context.Context, id string) (*RuleSet, error)
	ListRuleSets(ctx context.Context) ([]*RuleSet, error)
	GetActiveRuleSet(ctx context.Context) (*RuleSet, error)
	SetActiveRuleSet(ctx context.Context, id string) error
	
	StoreProposal(ctx context.Context, proposal *ProposedRuleChange) error
	GetProposal(ctx context.Context, id string) (*ProposedRuleChange, error)
	ListProposals(ctx context.Context, status ProposalStatus) ([]*ProposedRuleChange, error)
	UpdateProposal(ctx context.Context, proposal *ProposedRuleChange) error
}

// TimeProvider allows mocking time for tests
type TimeProvider interface {
	Now() time.Time
}

// DefaultTimeProvider implements TimeProvider
type DefaultTimeProvider struct{}

func (t *DefaultTimeProvider) Now() time.Time {
	return time.Now()
}

// NewDefaultRulesRegistry creates a new rules registry
func NewDefaultRulesRegistry(store RulesStore, committee CommitteeManager) *DefaultRulesRegistry {
	return &DefaultRulesRegistry{
		store:         store,
		committee:     committee,
		timeProvider:  &DefaultTimeProvider{},
	}
}

// GetActiveRuleSet returns the currently active ruleset
func (r *DefaultRulesRegistry) GetActiveRuleSet(ctx context.Context) (*RuleSet, error) {
	return r.store.GetActiveRuleSet(ctx)
}

// GetRuleSet returns a specific ruleset by ID
func (r *DefaultRulesRegistry) GetRuleSet(ctx context.Context, id string) (*RuleSet, error) {
	return r.store.GetRuleSet(ctx, id)
}

// ListRuleSets returns all rulesets
func (r *DefaultRulesRegistry) ListRuleSets(ctx context.Context) ([]*RuleSet, error) {
	return r.store.ListRuleSets(ctx)
}

// ProposeRuleChange proposes a new rule change
func (r *DefaultRulesRegistry) ProposeRuleChange(ctx context.Context, proposal *ProposedRuleChange) error {
	// Validate proposal
	if proposal.NewRuleSet == nil {
		return fmt.Errorf("new ruleset cannot be nil")
	}
	
	if proposal.ProposerDID == "" {
		return fmt.Errorf("proposer DID cannot be empty")
	}
	
	// Check if proposer is a committee member
	isMember, err := r.committee.IsMember(ctx, proposal.ProposerDID)
	if err != nil {
		return fmt.Errorf("failed to check committee membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("proposer %s is not a committee member", proposal.ProposerDID)
	}
	
	// Set proposal metadata
	now := r.timeProvider.Now()
	proposal.ProposedAt = now
	proposal.Status = ProposalStatusPending
	proposal.Approvals = []CommitteeApproval{}
	
	// Calculate activation date (must respect timelock)
	timeLockDuration := time.Duration(proposal.NewRuleSet.TimeLockDays) * 24 * time.Hour
	proposal.ActivationDate = now.Add(timeLockDuration)
	
	// Calculate hash of the proposal
	proposal.Hash = r.calculateProposalHash(proposal)
	
	// Store the proposal
	return r.store.StoreProposal(ctx, proposal)
}

// ApproveProposal adds committee approval to a proposal
func (r *DefaultRulesRegistry) ApproveProposal(ctx context.Context, proposalID string, approval *CommitteeApproval) error {
	// Get the proposal
	proposal, err := r.store.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}
	
	if proposal.Status != ProposalStatusPending {
		return fmt.Errorf("proposal is not in pending status")
	}
	
	// Check if member is in committee
	isMember, err := r.committee.IsMember(ctx, approval.MemberDID)
	if err != nil {
		return fmt.Errorf("failed to check committee membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("approver %s is not a committee member", approval.MemberDID)
	}
	
	// Check if member already approved
	for _, existing := range proposal.Approvals {
		if existing.MemberDID == approval.MemberDID {
			return fmt.Errorf("member %s has already approved this proposal", approval.MemberDID)
		}
	}
	
	// Add approval
	approval.ApprovedAt = r.timeProvider.Now()
	proposal.Approvals = append(proposal.Approvals, *approval)
	
	// Check if we have enough approvals
	committee, err := r.committee.GetCurrentCommittee(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current committee: %w", err)
	}
	
	if len(proposal.Approvals) >= committee.Threshold {
		proposal.Status = ProposalStatusApproved
	}
	
	// Update the proposal
	return r.store.UpdateProposal(ctx, proposal)
}

// ExecuteProposal executes an approved proposal after timelock expires
func (r *DefaultRulesRegistry) ExecuteProposal(ctx context.Context, proposalID string) error {
	proposal, err := r.store.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}
	
	if proposal.Status != ProposalStatusApproved {
		return fmt.Errorf("proposal is not approved")
	}
	
	// Check if timelock has expired
	now := r.timeProvider.Now()
	if now.Before(proposal.ActivationDate) {
		return fmt.Errorf("timelock has not expired yet (activation: %v, now: %v)", 
			proposal.ActivationDate, now)
	}
	
	// Finalize the new ruleset
	newRuleSet := proposal.NewRuleSet
	newRuleSet.ValidFrom = now
	newRuleSet.Hash = r.calculateRuleSetHash(newRuleSet)
	
	// Store the new ruleset
	if err := r.store.StoreRuleSet(ctx, newRuleSet); err != nil {
		return fmt.Errorf("failed to store new ruleset: %w", err)
	}
	
	// Set it as active
	if err := r.store.SetActiveRuleSet(ctx, newRuleSet.ID); err != nil {
		return fmt.Errorf("failed to set active ruleset: %w", err)
	}
	
	// Mark proposal as executed
	proposal.Status = ProposalStatusExecuted
	if err := r.store.UpdateProposal(ctx, proposal); err != nil {
		return fmt.Errorf("failed to update proposal status: %w", err)
	}
	
	return nil
}

// GetProposal returns a proposal by ID
func (r *DefaultRulesRegistry) GetProposal(ctx context.Context, proposalID string) (*ProposedRuleChange, error) {
	return r.store.GetProposal(ctx, proposalID)
}

// ListProposals returns proposals filtered by status
func (r *DefaultRulesRegistry) ListProposals(ctx context.Context, status ProposalStatus) ([]*ProposedRuleChange, error) {
	return r.store.ListProposals(ctx, status)
}

// Helper methods

// calculateRuleSetHash calculates SHA256 hash of a ruleset
func (r *DefaultRulesRegistry) calculateRuleSetHash(ruleSet *RuleSet) []byte {
	// Create a copy without hash and signature for canonical hashing
	canonical := *ruleSet
	canonical.Hash = nil
	canonical.Signature = nil
	
	// Marshal to canonical JSON
	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		return nil
	}
	
	hasher := sha256.New()
	hasher.Write(jsonBytes)
	return hasher.Sum(nil)
}

// calculateProposalHash calculates SHA256 hash of a proposal
func (r *DefaultRulesRegistry) calculateProposalHash(proposal *ProposedRuleChange) []byte {
	// Create a copy without hash and signature for canonical hashing
	canonical := *proposal
	canonical.Hash = nil
	canonical.Signature = nil
	
	// Marshal to canonical JSON
	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		return nil
	}
	
	hasher := sha256.New()
	hasher.Write(jsonBytes)
	return hasher.Sum(nil)
}

// DefaultRuleSet returns a sensible default ruleset
func DefaultRuleSet() *RuleSet {
	return &RuleSet{
		ID:           "default-v1",
		Version:      "1.0.0",
		ValidFrom:    time.Now(),
		TimeLockDays: 7, // One week timelock
		CreatedAt:    time.Now(),
		
		ScoringWeights: ScoringWeights{
			KYCWeight:      1.0,  // α = 1.0
			ActivityWeight: 0.5,  // β = 0.5
			VouchWeight:    0.8,  // γ = 0.8
			ReportWeight:   2.0,  // δ = 2.0 (penalty)
			TenureWeight:   0.3,  // τ = 0.3
		},
		
		VouchBudgets: map[string]int{
			"general":  5,  // 5 vouches per epoch in general context
			"commerce": 3,  // 3 vouches per epoch in commerce context
			"hiring":   2,  // 2 vouches per epoch in hiring context
		},
		
		ScoreCaps: ScoreCaps{
			MaxIndividualScore: 100.0,
			MaxVouchValue:      10.0,
			DiversityBonus:     1.2,
			MinScoreThreshold:  5.0,
		},
		
		DecayParameters: DecayParams{
			VouchDecayHalfLife:     30 * 24 * time.Hour, // 30 days
			ScoreDecayRate:         0.02,                // 2% decay per epoch
			ActivityWindow:         7 * 24 * time.Hour,  // 1 week activity window
			TenureBonusGracePeriod: 90 * 24 * time.Hour, // 90 days grace period
		},
		
		CommitteeSize:      7,
		SignatureThreshold: 5, // 5 of 7
		RotationPeriod:     30 * 24 * time.Hour, // 30 days
		
		CheckpointInterval: 10 * time.Minute,
		SignatureTimeout:   5 * time.Minute,
		
		DisputeWindow:    24 * time.Hour, // 24 hours to dispute
		SlashingPenalty:  0.1,            // 10% penalty
		ReputationDecay:  0.05,           // 5% reputation decay
	}
}