package score

import (
	"context"
	"testing"
	"time"

	"github.com/ParichayaHQ/credence/internal/crypto"
)

func TestDeterministicEngine_ComputeScore(t *testing.T) {
	// Setup
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	budgetManager := NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := NewNetworkGraphAnalyzer(dataProvider)
	decayFunc := NewExponentialDecayFunction()
	validator := NewTestValidator()
	
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)
	
	engine := NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Test score computation
	score, err := engine.ComputeScore(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("ComputeScore failed: %v", err)
	}
	
	// Validate result
	if score == nil {
		t.Fatal("Score is nil")
	}
	
	if score.DID != did {
		t.Errorf("Expected DID %s, got %s", did, score.DID)
	}
	
	if score.Context != context {
		t.Errorf("Expected context %s, got %s", context, score.Context)
	}
	
	if score.Epoch != epoch {
		t.Errorf("Expected epoch %d, got %d", epoch, score.Epoch)
	}
	
	if score.Value < 0 {
		t.Errorf("Score value cannot be negative: %f", score.Value)
	}
	
	// Test that subsequent calls return cached result
	score2, err := engine.ComputeScore(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("Second ComputeScore failed: %v", err)
	}
	
	if score.Value != score2.Value {
		t.Errorf("Cached score differs: %f vs %f", score.Value, score2.Value)
	}
}

func TestDeterministicEngine_RecomputeScore(t *testing.T) {
	// Setup
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	budgetManager := NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := NewNetworkGraphAnalyzer(dataProvider)
	decayFunc := NewExponentialDecayFunction()
	validator := NewTestValidator()
	
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)
	
	engine := NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Compute initial score
	score1, err := engine.ComputeScore(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("Initial ComputeScore failed: %v", err)
	}
	
	// Modify test data
	testProvider := dataProvider.(*TestDataProvider)
	testProvider.AddVouch(&VouchData{
		FromDID:   "did:key:voucher",
		ToDID:     did,
		Context:   context,
		Strength:  25.0,
		Timestamp: time.Now(),
		Epoch:     epoch,
	})
	
	// Recompute score
	score2, err := engine.RecomputeScore(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("RecomputeScore failed: %v", err)
	}
	
	// Score should be different (higher due to additional vouch)
	if score1.Value >= score2.Value {
		t.Errorf("Recomputed score should be higher: %f vs %f", score1.Value, score2.Value)
	}
}

func TestDeterministicEngine_GetFactors(t *testing.T) {
	// Setup
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	budgetManager := NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := NewNetworkGraphAnalyzer(dataProvider)
	decayFunc := NewExponentialDecayFunction()
	validator := NewTestValidator()
	
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)
	
	engine := NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Get factors
	components, err := engine.GetFactors(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("GetFactors failed: %v", err)
	}
	
	if components == nil {
		t.Fatal("Components is nil")
	}
	
	// All factors should be non-negative
	if components.K < 0 {
		t.Errorf("K factor cannot be negative: %f", components.K)
	}
	if components.A < 0 {
		t.Errorf("A factor cannot be negative: %f", components.A)
	}
	if components.V < 0 {
		t.Errorf("V factor cannot be negative: %f", components.V)
	}
	if components.R < 0 {
		t.Errorf("R factor cannot be negative: %f", components.R)
	}
	if components.T < 0 {
		t.Errorf("T factor cannot be negative: %f", components.T)
	}
}

func TestDeterministicEngine_GetProofAndVerify(t *testing.T) {
	// Setup
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	budgetManager := NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := NewNetworkGraphAnalyzer(dataProvider)
	decayFunc := NewExponentialDecayFunction()
	validator := NewTestValidator()
	
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)
	
	engine := NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)
	
	ctx := context.Background()
	did := "did:key:test123"
	context := "test_context"
	epoch := int64(100)
	
	// Compute score
	score, err := engine.ComputeScore(ctx, did, context, epoch)
	if err != nil {
		t.Fatalf("ComputeScore failed: %v", err)
	}
	
	// Generate proof
	proof, err := engine.GetProof(ctx, score)
	if err != nil {
		t.Fatalf("GetProof failed: %v", err)
	}
	
	if proof == nil {
		t.Fatal("Proof is nil")
	}
	
	if proof.Score != score {
		t.Error("Proof score doesn't match original score")
	}
	
	if len(proof.Signature) == 0 {
		t.Error("Proof signature is empty")
	}
	
	if len(proof.PublicKey) == 0 {
		t.Error("Proof public key is empty")
	}
	
	// Verify proof
	err = engine.VerifyProof(ctx, proof)
	if err != nil {
		t.Fatalf("VerifyProof failed: %v", err)
	}
	
	// Test with invalid proof
	invalidProof := *proof
	invalidProof.Signature = []byte("invalid")
	err = engine.VerifyProof(ctx, &invalidProof)
	if err == nil {
		t.Error("VerifyProof should fail with invalid signature")
	}
}

func BenchmarkDeterministicEngine_ComputeScore(b *testing.B) {
	// Setup
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	budgetManager := NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := NewNetworkGraphAnalyzer(dataProvider)
	decayFunc := NewExponentialDecayFunction()
	validator := NewTestValidator()
	
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)
	
	engine := NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)
	
	ctx := context.Background()
	did := "did:key:benchmark"
	context := "benchmark_context"
	epoch := int64(1000)
	
	// Add more test data for realistic benchmark
	testProvider := dataProvider.(*TestDataProvider)
	for i := 0; i < 100; i++ {
		testProvider.AddVouch(&VouchData{
			FromDID:   "did:key:voucher" + string(rune(i)),
			ToDID:     did,
			Context:   context,
			Strength:  10.0,
			Timestamp: time.Now(),
			Epoch:     epoch - int64(i),
		})
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Use different epochs to avoid caching
		testEpoch := epoch + int64(i)
		_, err := engine.ComputeScore(ctx, did, context, testEpoch)
		if err != nil {
			b.Fatalf("ComputeScore failed: %v", err)
		}
	}
}

func BenchmarkDeterministicEngine_GetFactors(b *testing.B) {
	// Setup
	config := DefaultScoreConfig()
	dataProvider := NewTestDataProvider()
	budgetManager := NewMemoryBudgetManager(config, dataProvider)
	graphAnalyzer := NewNetworkGraphAnalyzer(dataProvider)
	decayFunc := NewExponentialDecayFunction()
	validator := NewTestValidator()
	
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}
	signer := crypto.NewEd25519Signer(keyPair)
	
	engine := NewDeterministicEngine(
		config,
		dataProvider,
		budgetManager,
		graphAnalyzer,
		decayFunc,
		validator,
		signer,
	)
	
	ctx := context.Background()
	did := "did:key:benchmark"
	context := "benchmark_context"
	epoch := int64(1000)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := engine.GetFactors(ctx, did, context, epoch)
		if err != nil {
			b.Fatalf("GetFactors failed: %v", err)
		}
	}
}

// TestDataProvider provides a test implementation of DataProvider
type TestDataProvider struct {
	vouches      []*VouchData
	attestations []*AttestationData
	reports      []*ReportData
	kycData      []*KYCData
	timeData     map[string]*TimeData
	scores       map[string]*Score
}

func NewTestDataProvider() DataProvider {
	provider := &TestDataProvider{
		vouches:      make([]*VouchData, 0),
		attestations: make([]*AttestationData, 0),
		reports:      make([]*ReportData, 0),
		kycData:      make([]*KYCData, 0),
		timeData:     make(map[string]*TimeData),
		scores:       make(map[string]*Score),
	}
	
	// Add some default test data
	provider.AddVouch(&VouchData{
		FromDID:   "did:key:voucher1",
		ToDID:     "did:key:test123",
		Context:   "test_context",
		Strength:  15.0,
		Timestamp: time.Now().AddDate(0, 0, -10),
		Epoch:     90,
	})
	
	provider.AddAttestation(&AttestationData{
		DID:              "did:key:test123",
		Context:          "test_context",
		Type:             "employment",
		IssuerDID:        "did:key:employer",
		IssuerReputation: 0.9,
		Weight:           20.0,
		Timestamp:        time.Now().AddDate(0, -3, 0),
		Epoch:            10,
	})
	
	provider.AddKYC(&KYCData{
		DID:       "did:key:test123",
		Context:   "test_context",
		Type:      "kyc_level_1",
		Level:     1,
		IssuerDID: "did:key:kyc",
		Weight:    30.0,
		Timestamp: time.Now().AddDate(-1, 0, 0),
		Epoch:     1,
	})
	
	provider.SetTimeData("did:key:test123", "test_context", &TimeData{
		DID:           "did:key:test123",
		Context:       "test_context",
		FirstActivity: time.Now().AddDate(-1, 0, 0),
		LastActivity:  time.Now().AddDate(0, 0, -1),
		ActivityCount: 50,
		Epoch:         100,
	})
	
	return provider
}

func (t *TestDataProvider) AddVouch(vouch *VouchData) {
	t.vouches = append(t.vouches, vouch)
}

func (t *TestDataProvider) AddAttestation(att *AttestationData) {
	t.attestations = append(t.attestations, att)
}

func (t *TestDataProvider) AddReport(report *ReportData) {
	t.reports = append(t.reports, report)
}

func (t *TestDataProvider) AddKYC(kyc *KYCData) {
	t.kycData = append(t.kycData, kyc)
}

func (t *TestDataProvider) SetTimeData(did, context string, timeData *TimeData) {
	key := did + ":" + context
	t.timeData[key] = timeData
}

func (t *TestDataProvider) GetVouches(ctx context.Context, did, context string, maxEpoch int64) ([]*VouchData, error) {
	result := make([]*VouchData, 0)
	for _, vouch := range t.vouches {
		if vouch.ToDID == did && vouch.Context == context && vouch.Epoch <= maxEpoch {
			result = append(result, vouch)
		}
	}
	return result, nil
}

func (t *TestDataProvider) GetAttestations(ctx context.Context, did, context string, maxEpoch int64) ([]*AttestationData, error) {
	result := make([]*AttestationData, 0)
	for _, att := range t.attestations {
		if att.DID == did && att.Context == context && att.Epoch <= maxEpoch {
			result = append(result, att)
		}
	}
	return result, nil
}

func (t *TestDataProvider) GetReports(ctx context.Context, did, context string, maxEpoch int64) ([]*ReportData, error) {
	result := make([]*ReportData, 0)
	for _, report := range t.reports {
		if report.ReportedDID == did && report.Context == context && report.Epoch <= maxEpoch {
			result = append(result, report)
		}
	}
	return result, nil
}

func (t *TestDataProvider) GetKYCData(ctx context.Context, did, context string, maxEpoch int64) ([]*KYCData, error) {
	result := make([]*KYCData, 0)
	for _, kyc := range t.kycData {
		if kyc.DID == did && kyc.Context == context && kyc.Epoch <= maxEpoch {
			result = append(result, kyc)
		}
	}
	return result, nil
}

func (t *TestDataProvider) GetTimeData(ctx context.Context, did, context string, maxEpoch int64) (*TimeData, error) {
	key := did + ":" + context
	if timeData, exists := t.timeData[key]; exists {
		return timeData, nil
	}
	return nil, nil
}

func (t *TestDataProvider) GetScore(ctx context.Context, did, context string, epoch int64) (*Score, error) {
	key := did + ":" + context + ":" + string(rune(epoch))
	if score, exists := t.scores[key]; exists {
		return score, nil
	}
	return nil, nil
}

func (t *TestDataProvider) StoreScore(ctx context.Context, score *Score) error {
	key := score.DID + ":" + score.Context + ":" + string(rune(score.Epoch))
	t.scores[key] = score
	return nil
}

// TestValidator provides a test implementation of ScoreValidator
type TestValidator struct{}

func NewTestValidator() ScoreValidator {
	return &TestValidator{}
}

func (v *TestValidator) ValidateInputData(ctx context.Context, did, context string, epoch int64) error {
	if did == "" {
		return NewValidationError("DID cannot be empty")
	}
	return nil
}

func (v *TestValidator) ValidateScoreRange(score float64) error {
	if score < 0 {
		return NewValidationError("score cannot be negative")
	}
	if score > 10000 {
		return NewValidationError("score unreasonably high")
	}
	return nil
}

func (v *TestValidator) ValidateComponents(components *ScoreComponents, factors *ScoreFactors) error {
	if components.K < 0 || components.A < 0 || components.V < 0 || components.R < 0 || components.T < 0 {
		return NewValidationError("component values cannot be negative")
	}
	return nil
}

func (v *TestValidator) ValidateIntegrity(ctx context.Context, score *Score) error {
	if score.DID == "" {
		return NewValidationError("score DID cannot be empty")
	}
	if score.Value < 0 {
		return NewValidationError("score value cannot be negative")
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (e *ValidationError) Error() string {
	return e.message
}