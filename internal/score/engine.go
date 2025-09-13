package score

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ParichayaHQ/credence/internal/crypto"
)

// DeterministicEngine implements the trust scoring algorithm
type DeterministicEngine struct {
	config        *ScoreConfig
	dataProvider  DataProvider
	budgetManager BudgetManager
	graphAnalyzer GraphAnalyzer
	decayFunc     DecayFunction
	validator     ScoreValidator
	signer        crypto.Signer
}

// NewDeterministicEngine creates a new deterministic scoring engine
func NewDeterministicEngine(
	config *ScoreConfig,
	dataProvider DataProvider,
	budgetManager BudgetManager,
	graphAnalyzer GraphAnalyzer,
	decayFunc DecayFunction,
	validator ScoreValidator,
	signer crypto.Signer,
) *DeterministicEngine {
	if config == nil {
		config = DefaultScoreConfig()
	}
	
	return &DeterministicEngine{
		config:        config,
		dataProvider:  dataProvider,
		budgetManager: budgetManager,
		graphAnalyzer: graphAnalyzer,
		decayFunc:     decayFunc,
		validator:     validator,
		signer:        signer,
	}
}

// ComputeScore implements Engine.ComputeScore
func (e *DeterministicEngine) ComputeScore(ctx context.Context, did, context string, epoch int64) (*Score, error) {
	// Check for cached score first
	if cachedScore, err := e.dataProvider.GetScore(ctx, did, context, epoch); err == nil && cachedScore != nil {
		return cachedScore, nil
	}
	
	// Validate input data
	if e.validator != nil {
		if err := e.validator.ValidateInputData(ctx, did, context, epoch); err != nil {
			return nil, fmt.Errorf("input validation failed: %w", err)
		}
	}
	
	// Compute each factor
	components, err := e.computeComponents(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to compute components: %w", err)
	}
	
	// Apply the scoring formula: S_i^c = α*K_i^c + β*A_i^c + γ*V_i^c - δ*R_i^c + τ*T_i^c
	factors := e.config.Factors
	totalScore := factors.Alpha*components.K +
		factors.Beta*components.A +
		factors.Gamma*components.V -
		factors.Delta*components.R +
		factors.Tau*components.T
	
	// Ensure score is non-negative
	if totalScore < 0 {
		totalScore = 0
	}
	
	// Create score object
	score := &Score{
		DID:        did,
		Context:    context,
		Value:      totalScore,
		Epoch:      epoch,
		Timestamp:  time.Now(),
		Components: *components,
		ComputedBy: "deterministic-engine",
		Version:    "1.0.0",
	}
	
	// Validate the computed score
	if e.validator != nil {
		if err := e.validator.ValidateScoreRange(totalScore); err != nil {
			return nil, fmt.Errorf("score validation failed: %w", err)
		}
		
		if err := e.validator.ValidateComponents(components, &factors); err != nil {
			return nil, fmt.Errorf("component validation failed: %w", err)
		}
	}
	
	// Store the computed score
	if err := e.dataProvider.StoreScore(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to store score: %w", err)
	}
	
	return score, nil
}

// computeComponents computes all score components
func (e *DeterministicEngine) computeComponents(ctx context.Context, did, context string, epoch int64) (*ScoreComponents, error) {
	var components ScoreComponents
	var err error
	
	// Compute K factor (PoP/KYC)
	components.K, err = e.computeKFactor(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("K factor computation failed: %w", err)
	}
	
	// Compute A factor (Attestations)
	components.A, err = e.computeAFactor(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("A factor computation failed: %w", err)
	}
	
	// Compute V factor (Vouches)
	components.V, err = e.computeVFactor(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("V factor computation failed: %w", err)
	}
	
	// Compute R factor (Reports)
	components.R, err = e.computeRFactor(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("R factor computation failed: %w", err)
	}
	
	// Compute T factor (Time)
	components.T, err = e.computeTFactor(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("T factor computation failed: %w", err)
	}
	
	return &components, nil
}

// computeKFactor computes the PoP/KYC factor
func (e *DeterministicEngine) computeKFactor(ctx context.Context, did, context string, epoch int64) (float64, error) {
	kycData, err := e.dataProvider.GetKYCData(ctx, did, context, epoch)
	if err != nil {
		return 0, fmt.Errorf("failed to get KYC data: %w", err)
	}
	
	var totalWeight float64
	now := time.Now()
	
	for _, kyc := range kycData {
		// Skip expired credentials
		if kyc.ExpiresAt != nil && kyc.ExpiresAt.Before(now) {
			continue
		}
		
		// Apply decay based on age
		ageInEpochs := epoch - kyc.Epoch
		decayedWeight := e.decayFunc.ApplyDecay(kyc.Weight, ageInEpochs, e.config.VouchHalfLife)
		
		totalWeight += decayedWeight
	}
	
	return totalWeight, nil
}

// computeAFactor computes the attestations factor
func (e *DeterministicEngine) computeAFactor(ctx context.Context, did, context string, epoch int64) (float64, error) {
	attestations, err := e.dataProvider.GetAttestations(ctx, did, context, epoch)
	if err != nil {
		return 0, fmt.Errorf("failed to get attestations: %w", err)
	}
	
	var totalWeight float64
	
	for _, att := range attestations {
		// Weight attestation by issuer reputation and inherent weight
		weightedValue := att.Weight * att.IssuerReputation
		
		// Apply decay based on age
		ageInEpochs := epoch - att.Epoch
		decayedWeight := e.decayFunc.ApplyDecay(weightedValue, ageInEpochs, e.config.VouchHalfLife)
		
		totalWeight += decayedWeight
	}
	
	return totalWeight, nil
}

// computeVFactor computes the vouches factor with concave aggregation
func (e *DeterministicEngine) computeVFactor(ctx context.Context, did, context string, epoch int64) (float64, error) {
	vouches, err := e.dataProvider.GetVouches(ctx, did, context, epoch)
	if err != nil {
		return 0, fmt.Errorf("failed to get vouches: %w", err)
	}
	
	var weightedSum float64
	
	for _, vouch := range vouches {
		// Get the voucher's score (with cap)
		voucherScore, err := e.getVoucherScore(ctx, vouch.FromDID, context, epoch)
		if err != nil {
			// If we can't get the voucher's score, use a default low value
			voucherScore = 1.0
		}
		
		// Apply vouch cap
		cappedScore := math.Min(voucherScore, e.config.VouchCap)
		
		// Apply vouch strength and decay
		ageInEpochs := epoch - vouch.Epoch
		decayedStrength := e.decayFunc.ApplyDecay(vouch.Strength, ageInEpochs, e.config.VouchHalfLife)
		
		weightedSum += cappedScore * decayedStrength
	}
	
	// Apply concave aggregation (square root)
	concaveValue := math.Sqrt(weightedSum)
	
	// Apply diversity penalty if graph analyzer is available
	if e.graphAnalyzer != nil {
		diversity, err := e.graphAnalyzer.ComputeDiversity(ctx, did, context, epoch)
		if err == nil {
			diversityMultiplier := 1.0 - (e.config.DiversityPenalty * (1.0 - diversity))
			concaveValue *= diversityMultiplier
		}
	}
	
	return concaveValue, nil
}

// computeRFactor computes the reports factor
func (e *DeterministicEngine) computeRFactor(ctx context.Context, did, context string, epoch int64) (float64, error) {
	reports, err := e.dataProvider.GetReports(ctx, did, context, epoch)
	if err != nil {
		return 0, fmt.Errorf("failed to get reports: %w", err)
	}
	
	var totalPenalty float64
	
	for _, report := range reports {
		// Only consider adjudicated and upheld reports
		if !report.Adjudicated || !report.Upheld {
			continue
		}
		
		// Apply severity weighting
		severityWeight := report.Severity
		
		// Apply decay based on age
		ageInEpochs := epoch - report.Epoch
		decayedWeight := e.decayFunc.ApplyDecay(severityWeight, ageInEpochs, e.config.ReportHalfLife)
		
		totalPenalty += decayedWeight
	}
	
	return totalPenalty, nil
}

// computeTFactor computes the time factor
func (e *DeterministicEngine) computeTFactor(ctx context.Context, did, context string, epoch int64) (float64, error) {
	timeData, err := e.dataProvider.GetTimeData(ctx, did, context, epoch)
	if err != nil {
		return 0, fmt.Errorf("failed to get time data: %w", err)
	}
	
	if timeData == nil {
		return 0, nil
	}
	
	// Compute age bonus with bounded growth
	firstEpoch := timeData.FirstActivity.Unix() / 86400 // Convert to epoch days
	lastEpoch := timeData.LastActivity.Unix() / 86400
	
	ageBonus := e.decayFunc.ComputeTimeBonus(firstEpoch, lastEpoch, epoch, e.config.TimeMaxGrowth)
	
	// Apply inactivity decay
	inactiveEpochs := epoch - lastEpoch
	if inactiveEpochs > 0 {
		ageBonus = e.decayFunc.ApplyInactivityDecay(ageBonus, inactiveEpochs, e.config.TimeInactivityDecay)
	}
	
	return ageBonus, nil
}

// getVoucherScore gets a score for a voucher, handling circular dependencies
func (e *DeterministicEngine) getVoucherScore(ctx context.Context, voucherDID, context string, epoch int64) (float64, error) {
	// Try to get cached score first
	if score, err := e.dataProvider.GetScore(ctx, voucherDID, context, epoch); err == nil && score != nil {
		return score.Value, nil
	}
	
	// For bootstrapping, use a simple heuristic based on voucher data
	vouches, err := e.dataProvider.GetVouches(ctx, voucherDID, context, epoch)
	if err != nil {
		return 1.0, nil // Default score
	}
	
	// Simple score approximation based on vouch count
	baseScore := 10.0 + float64(len(vouches))*2.0
	return math.Min(baseScore, e.config.VouchCap), nil
}

// ComputeScores implements Engine.ComputeScores
func (e *DeterministicEngine) ComputeScores(ctx context.Context, dids []string, context string, epoch int64) ([]*Score, error) {
	scores := make([]*Score, len(dids))
	
	for i, did := range dids {
		score, err := e.ComputeScore(ctx, did, context, epoch)
		if err != nil {
			return nil, fmt.Errorf("failed to compute score for %s: %w", did, err)
		}
		scores[i] = score
	}
	
	return scores, nil
}

// RecomputeScore implements Engine.RecomputeScore
func (e *DeterministicEngine) RecomputeScore(ctx context.Context, did, context string, epoch int64) (*Score, error) {
	// Force recomputation by not checking cache
	components, err := e.computeComponents(ctx, did, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to compute components: %w", err)
	}
	
	factors := e.config.Factors
	totalScore := factors.Alpha*components.K +
		factors.Beta*components.A +
		factors.Gamma*components.V -
		factors.Delta*components.R +
		factors.Tau*components.T
	
	if totalScore < 0 {
		totalScore = 0
	}
	
	score := &Score{
		DID:        did,
		Context:    context,
		Value:      totalScore,
		Epoch:      epoch,
		Timestamp:  time.Now(),
		Components: *components,
		ComputedBy: "deterministic-engine",
		Version:    "1.0.0",
	}
	
	// Store the recomputed score
	if err := e.dataProvider.StoreScore(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to store recomputed score: %w", err)
	}
	
	return score, nil
}

// GetFactors implements Engine.GetFactors
func (e *DeterministicEngine) GetFactors(ctx context.Context, did, context string, epoch int64) (*ScoreComponents, error) {
	return e.computeComponents(ctx, did, context, epoch)
}

// ValidateScore implements Engine.ValidateScore
func (e *DeterministicEngine) ValidateScore(ctx context.Context, score *Score) error {
	if e.validator != nil {
		return e.validator.ValidateIntegrity(ctx, score)
	}
	return nil
}

// GetProof implements Engine.GetProof
func (e *DeterministicEngine) GetProof(ctx context.Context, score *Score) (*ScoreProof, error) {
	if e.signer == nil {
		return nil, fmt.Errorf("no signer configured")
	}
	
	// Generate deterministic input hash
	inputHash, err := e.computeInputHash(ctx, score.DID, score.Context, score.Epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to compute input hash: %w", err)
	}
	
	// Create canonical representation for signing
	canonical := fmt.Sprintf("%s|%s|%.6f|%d|%s", 
		score.DID, score.Context, score.Value, score.Epoch, inputHash)
	
	// Sign the canonical representation
	signature, err := e.signer.Sign([]byte(canonical))
	if err != nil {
		return nil, fmt.Errorf("failed to sign score: %w", err)
	}
	
	// Get public key
	publicKey := []byte(e.signer.PublicKey())
	
	return &ScoreProof{
		Score:     score,
		InputHash: inputHash,
		Signature: signature,
		PublicKey: publicKey,
		Algorithm: "ed25519", // Assuming Ed25519
		Timestamp: time.Now(),
	}, nil
}

// VerifyProof implements Engine.VerifyProof
func (e *DeterministicEngine) VerifyProof(ctx context.Context, proof *ScoreProof) error {
	// Recreate canonical representation
	canonical := fmt.Sprintf("%s|%s|%.6f|%d|%s", 
		proof.Score.DID, proof.Score.Context, proof.Score.Value, 
		proof.Score.Epoch, proof.InputHash)
	
	// Verify signature
	verifier := crypto.NewEd25519Verifier()
	if !verifier.Verify(proof.PublicKey, []byte(canonical), proof.Signature) {
		return fmt.Errorf("signature verification failed")
	}
	
	return nil
}

// computeInputHash creates a deterministic hash of all input data
func (e *DeterministicEngine) computeInputHash(ctx context.Context, did, context string, epoch int64) (string, error) {
	hasher := sha256.New()
	
	// Hash vouches data
	vouches, err := e.dataProvider.GetVouches(ctx, did, context, epoch)
	if err == nil {
		// Sort for deterministic ordering
		sort.Slice(vouches, func(i, j int) bool {
			return vouches[i].FromDID < vouches[j].FromDID
		})
		for _, v := range vouches {
			hasher.Write([]byte(fmt.Sprintf("v:%s:%s:%.6f:%d", v.FromDID, v.ToDID, v.Strength, v.Epoch)))
		}
	}
	
	// Hash attestations data
	attestations, err := e.dataProvider.GetAttestations(ctx, did, context, epoch)
	if err == nil {
		sort.Slice(attestations, func(i, j int) bool {
			return attestations[i].IssuerDID < attestations[j].IssuerDID
		})
		for _, a := range attestations {
			hasher.Write([]byte(fmt.Sprintf("a:%s:%s:%.6f:%d", a.IssuerDID, a.Type, a.Weight, a.Epoch)))
		}
	}
	
	// Hash reports data
	reports, err := e.dataProvider.GetReports(ctx, did, context, epoch)
	if err == nil {
		sort.Slice(reports, func(i, j int) bool {
			return reports[i].ReporterDID < reports[j].ReporterDID
		})
		for _, r := range reports {
			hasher.Write([]byte(fmt.Sprintf("r:%s:%.6f:%t:%t:%d", r.ReporterDID, r.Severity, r.Adjudicated, r.Upheld, r.Epoch)))
		}
	}
	
	// Hash KYC data
	kycData, err := e.dataProvider.GetKYCData(ctx, did, context, epoch)
	if err == nil {
		sort.Slice(kycData, func(i, j int) bool {
			return kycData[i].IssuerDID < kycData[j].IssuerDID
		})
		for _, k := range kycData {
			expiry := ""
			if k.ExpiresAt != nil {
				expiry = k.ExpiresAt.Format(time.RFC3339)
			}
			hasher.Write([]byte(fmt.Sprintf("k:%s:%s:%d:%.6f:%s:%d", k.IssuerDID, k.Type, k.Level, k.Weight, expiry, k.Epoch)))
		}
	}
	
	// Hash time data
	timeData, err := e.dataProvider.GetTimeData(ctx, did, context, epoch)
	if err == nil && timeData != nil {
		hasher.Write([]byte(fmt.Sprintf("t:%s:%s:%d:%d", 
			timeData.FirstActivity.Format(time.RFC3339),
			timeData.LastActivity.Format(time.RFC3339),
			timeData.ActivityCount,
			timeData.Epoch)))
	}
	
	return hex.EncodeToString(hasher.Sum(nil)), nil
}