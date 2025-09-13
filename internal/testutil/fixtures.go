package testutil

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/ParichayaHQ/credence/internal/crypto"
	"github.com/ParichayaHQ/credence/internal/didvc"
	"github.com/ParichayaHQ/credence/internal/events"
	"github.com/ParichayaHQ/credence/pkg/types"
)

// TestKeyPair represents a test key pair with DID
type TestKeyPair struct {
	KeyPair *crypto.Ed25519KeyPair
	DID     *didvc.DID
	Signer  *crypto.Ed25519Signer
}

// NewTestKeyPair creates a new test key pair with associated DID
func NewTestKeyPair() (*TestKeyPair, error) {
	keyPair, err := crypto.NewEd25519KeyPair()
	if err != nil {
		return nil, err
	}

	did, err := didvc.CreateDIDKey(keyPair.PublicKey)
	if err != nil {
		return nil, err
	}

	signer := crypto.NewEd25519Signer(keyPair)

	return &TestKeyPair{
		KeyPair: keyPair,
		DID:     did,
		Signer:  signer,
	}, nil
}

// NewTestKeyPairFromSeed creates a test key pair from a deterministic seed
func NewTestKeyPairFromSeed(seed []byte) (*TestKeyPair, error) {
	keyPair, err := crypto.NewEd25519KeyPairFromSeed(seed)
	if err != nil {
		return nil, err
	}

	did, err := didvc.CreateDIDKey(keyPair.PublicKey)
	if err != nil {
		return nil, err
	}

	signer := crypto.NewEd25519Signer(keyPair)

	return &TestKeyPair{
		KeyPair: keyPair,
		DID:     did,
		Signer:  signer,
	}, nil
}

// CreateTestVouchEvent creates a test vouch event
func CreateTestVouchEvent(from, to *TestKeyPair, context string) (*events.Event, error) {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, err
	}

	signable := &events.SignableEvent{
		Type:     events.EventTypeVouch,
		From:     from.DID.String(),
		To:       to.DID.String(),
		Context:  context,
		Epoch:    "2025-09",
		Nonce:    nonce,
		IssuedAt: time.Now(),
	}

	// Canonicalize and sign
	canonical, err := events.CanonicalizeEvent(signable)
	if err != nil {
		return nil, err
	}

	signature, err := from.Signer.SignBase64(canonical)
	if err != nil {
		return nil, err
	}

	return &events.Event{
		Type:      signable.Type,
		From:      signable.From,
		To:        signable.To,
		Context:   signable.Context,
		Epoch:     signable.Epoch,
		Nonce:     signable.Nonce,
		IssuedAt:  signable.IssuedAt,
		Signature: signature,
	}, nil
}

// CreateTestReportEvent creates a test report event
func CreateTestReportEvent(from, to *TestKeyPair, context, reasonCode string) (*events.Event, error) {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, err
	}

	signable := &events.SignableEvent{
		Type:     events.EventTypeReport,
		From:     from.DID.String(),
		To:       to.DID.String(),
		Context:  context,
		Epoch:    "2025-09",
		Nonce:    nonce,
		IssuedAt: time.Now(),
	}

	canonical, err := events.CanonicalizeEvent(signable)
	if err != nil {
		return nil, err
	}

	signature, err := from.Signer.SignBase64(canonical)
	if err != nil {
		return nil, err
	}

	event := &events.Event{
		Type:      signable.Type,
		From:      signable.From,
		To:        signable.To,
		Context:   signable.Context,
		Epoch:     signable.Epoch,
		Nonce:     signable.Nonce,
		IssuedAt:  signable.IssuedAt,
		Signature: signature,
	}

	// Note: In a real implementation, we'd embed reasonCode in the payload
	// For this test utility, we'll add it as a comment or use PayloadCID

	return event, nil
}

// CreateTestRuleset creates a test ruleset with sensible defaults
func CreateTestRuleset() *types.Ruleset {
	return &types.Ruleset{
		ID: "test-v1.0",
		Weights: types.RulesetWeights{
			Alpha: 0.40,
			Beta:  0.20,
			Gamma: 0.25,
			Delta: 0.10,
			Tau:   0.05,
		},
		Caps: types.RulesetCaps{
			K: 1.0,
			A: 0.8,
			V: 0.9,
			R: 0.9,
			T: 0.2,
		},
		Vouch: types.VouchRules{
			BudgetBase:   2,
			BudgetLambda: 1.2,
			Aggregation:  "sqrt",
			PerEpoch:     "monthly",
			Bond: types.BondRules{
				Type:         "reputation",
				DecayOnAbuse: 0.05,
			},
		},
		Decay: types.DecayRules{
			HalfLifeDays: map[string]int{
				"V": 120,
				"R": 180,
				"T": 90,
			},
		},
		Diversity: types.DiversityRules{
			CommunityOverlapPenalty: 0.15,
			MinClusters:            3,
		},
		Adjudication: types.AdjudicationRules{
			PoolSize:         9,
			Quorum:          6,
			AppealWindowDays: 14,
		},
		ValidFrom:    time.Now(),
		TimeLockDays: 7,
	}
}

// CreateTestCheckpoint creates a test checkpoint
func CreateTestCheckpoint(epoch int64, root string) *types.Checkpoint {
	return &types.Checkpoint{
		Root:      root,
		TreeID:    "test-tree",
		Epoch:     epoch,
		Signers:   []string{"signer1", "signer2", "signer3"},
		Signature: "test-bls-signature",
		Timestamp: time.Now(),
	}
}

// CreateTestScoreRecord creates a test score record
func CreateTestScoreRecord(did, context string, score float64) *types.ScoreRecord {
	return &types.ScoreRecord{
		DID:     did,
		Context: context,
		Ruleset: types.RulesetRef{
			ID:        "test-v1.0",
			Hash:      "test-hash",
			ValidFrom: time.Now(),
		},
		Checkpoint: types.CheckpointRef{
			Root:      "test-root",
			Epoch:     1024,
			Signature: "test-sig",
			Signers:   []string{"signer1", "signer2"},
		},
		Score: score,
		Factors: types.ScoreFactors{
			K: "commitment-k",
			A: "commitment-a",
			V: "commitment-v",
			R: "commitment-r",
			T: "commitment-t",
		},
		Proofs: types.ScoreProofs{
			Inclusion: []types.InclusionProof{
				{CID: "test-cid", Path: []string{"proof", "path"}},
			},
			Consistency: []types.ConsistencyProof{
				{OldRoot: "old-root", NewRoot: "new-root", Path: []string{"consistency", "path"}},
			},
			StatusLists: []types.StatusListProof{
				{Issuer: "did:key:test", Epoch: 1024, BitmapCID: "bitmap-cid"},
			},
		},
		ComputedAt: time.Now(),
	}
}

// GenerateTestSeed generates a deterministic test seed
func GenerateTestSeed(identifier string) []byte {
	seed := make([]byte, ed25519.SeedSize)
	copy(seed, []byte(identifier))
	// Pad to required size
	for i := len(identifier); i < ed25519.SeedSize; i++ {
		seed[i] = byte(i % 256)
	}
	return seed
}

// GenerateRandomSeed generates a random seed for testing
func GenerateRandomSeed() ([]byte, error) {
	seed := make([]byte, ed25519.SeedSize)
	_, err := rand.Read(seed)
	return seed, err
}

// CreateTestNetwork creates a set of interconnected test key pairs
func CreateTestNetwork(size int) ([]*TestKeyPair, error) {
	network := make([]*TestKeyPair, size)
	
	for i := 0; i < size; i++ {
		seed := GenerateTestSeed(string(rune('A' + i)))
		keyPair, err := NewTestKeyPairFromSeed(seed)
		if err != nil {
			return nil, err
		}
		network[i] = keyPair
	}
	
	return network, nil
}

// CreateVouchChain creates a chain of vouches between test identities
func CreateVouchChain(network []*TestKeyPair, context string) ([]*events.Event, error) {
	if len(network) < 2 {
		return nil, nil
	}
	
	vouches := make([]*events.Event, 0, len(network)-1)
	
	for i := 0; i < len(network)-1; i++ {
		vouch, err := CreateTestVouchEvent(network[i], network[i+1], context)
		if err != nil {
			return nil, err
		}
		vouches = append(vouches, vouch)
	}
	
	return vouches, nil
}

// TestEventValidator provides validation utilities for tests
type TestEventValidator struct {
	verifier *crypto.Ed25519Verifier
}

// NewTestEventValidator creates a new test event validator
func NewTestEventValidator() *TestEventValidator {
	return &TestEventValidator{
		verifier: crypto.NewEd25519Verifier(),
	}
}

// ValidateEventSignature validates an event signature against its DID
func (v *TestEventValidator) ValidateEventSignature(event *events.Event) (bool, error) {
	// Parse DID and extract public key
	did, err := didvc.ParseDID(event.From)
	if err != nil {
		return false, err
	}
	
	publicKey, err := didvc.ExtractPublicKeyFromDIDKey(did)
	if err != nil {
		return false, err
	}
	
	// Canonicalize signable event
	signable := event.ToSignable()
	canonical, err := events.CanonicalizeEvent(signable)
	if err != nil {
		return false, err
	}
	
	// Verify signature
	signature, err := base64.StdEncoding.DecodeString(event.Signature)
	if err != nil {
		return false, err
	}
	
	return v.verifier.Verify(publicKey, canonical, signature), nil
}