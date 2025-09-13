package consensus

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"testing"
	"time"
)

func TestDefaultVRFProvider_GenerateProof(t *testing.T) {
	provider := NewDefaultVRFProvider()
	
	// Generate Ed25519 key pair
	privateKey, publicKey, err := GenerateVRFKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	seed := []byte("test-vrf-seed-for-committee-selection")
	
	// Generate VRF proof
	proof, output, err := provider.GenerateProof(privateKey, seed)
	if err != nil {
		t.Errorf("Expected successful VRF proof generation, got error: %v", err)
	}
	
	if len(proof) == 0 {
		t.Error("Expected non-empty VRF proof")
	}
	
	if len(output) == 0 {
		t.Error("Expected non-empty VRF output")
	}
	
	// Verify the proof
	err = provider.VerifyProof(publicKey, seed, proof, output)
	if err != nil {
		t.Errorf("Expected valid VRF proof to verify, got error: %v", err)
	}
	
	// Test with invalid private key size
	invalidPrivateKey := []byte("too-short")
	_, _, err = provider.GenerateProof(invalidPrivateKey, seed)
	if err == nil {
		t.Error("Expected error for invalid private key size, got nil")
	}
	
	// Test with empty seed
	_, _, err = provider.GenerateProof(privateKey, []byte{})
	if err == nil {
		t.Error("Expected error for empty seed, got nil")
	}
}

func TestDefaultVRFProvider_VerifyProof(t *testing.T) {
	provider := NewDefaultVRFProvider()
	
	// Generate key pair
	privateKey, publicKey, err := GenerateVRFKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	seed := []byte("verification-test-seed")
	
	// Generate proof
	proof, output, err := provider.GenerateProof(privateKey, seed)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}
	
	// Test valid verification
	err = provider.VerifyProof(publicKey, seed, proof, output)
	if err != nil {
		t.Errorf("Expected valid proof to verify, got error: %v", err)
	}
	
	// Test with wrong seed
	wrongSeed := []byte("wrong-seed")
	err = provider.VerifyProof(publicKey, wrongSeed, proof, output)
	if err == nil {
		t.Error("Expected error for wrong seed, got nil")
	}
	
	// Test with wrong output
	wrongOutput := []byte("wrong-output-bytes")
	err = provider.VerifyProof(publicKey, seed, proof, wrongOutput)
	if err == nil {
		t.Error("Expected error for wrong output, got nil")
	}
	
	// Test with invalid public key size
	invalidPublicKey := []byte("invalid-size")
	err = provider.VerifyProof(invalidPublicKey, seed, proof, output)
	if err == nil {
		t.Error("Expected error for invalid public key size, got nil")
	}
	
	// Test with empty seed
	err = provider.VerifyProof(publicKey, []byte{}, proof, output)
	if err == nil {
		t.Error("Expected error for empty seed, got nil")
	}
	
	// Test with invalid proof size
	invalidProof := []byte("invalid-proof-size")
	err = provider.VerifyProof(publicKey, seed, invalidProof, output)
	if err == nil {
		t.Error("Expected error for invalid proof size, got nil")
	}
}

func TestDefaultVRFProvider_GetCurrentSeed(t *testing.T) {
	provider := NewDefaultVRFProvider()
	ctx := context.Background()
	
	seed, err := provider.GetCurrentSeed(ctx)
	if err != nil {
		t.Errorf("Expected successful seed generation, got error: %v", err)
	}
	
	if len(seed) == 0 {
		t.Error("Expected non-empty seed")
	}
	
	// Test determinism - same epoch should give same seed
	seed2, err := provider.GetCurrentSeed(ctx)
	if err != nil {
		t.Errorf("Expected successful seed generation, got error: %v", err)
	}
	
	if !bytesEqual(seed, seed2) {
		t.Error("Expected deterministic seed for same epoch")
	}
}

func TestCommitteeSelector_SelectCommittee(t *testing.T) {
	provider := NewDefaultVRFProvider()
	config := &CheckpointorConfig{
		CommitteeSize:      3,
		SignatureThreshold: 2,
		RotationPeriod:     24 * time.Hour,
	}
	selector := NewCommitteeSelector(provider, config)
	
	// Create test candidates
	candidates := make([]CommitteeCandidate, 5)
	for i := 0; i < 5; i++ {
		_, publicKey, err := GenerateVRFKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair for candidate %d: %v", i, err)
		}
		
		candidates[i] = CommitteeCandidate{
			DID:       fmt.Sprintf("did:key:candidate%d", i),
			PublicKey: publicKey,
			VRFProof:  []byte{},
			Stake:     int64((i + 1) * 100),
		}
		
		t.Logf("Candidate %d DID: %s", i, candidates[i].DID)
	}
	
	ctx := context.Background()
	seed := []byte("committee-selection-test-seed")
	
	// Test successful committee selection
	committee, err := selector.SelectCommittee(ctx, seed, candidates)
	if err != nil {
		t.Errorf("Expected successful committee selection, got error: %v", err)
	}
	
	if committee == nil {
		t.Fatal("Expected non-nil committee")
	}
	
	if len(committee.Members) != config.CommitteeSize {
		t.Errorf("Expected committee size %d, got %d", config.CommitteeSize, len(committee.Members))
	}
	
	if committee.Threshold != config.SignatureThreshold {
		t.Errorf("Expected threshold %d, got %d", config.SignatureThreshold, committee.Threshold)
	}
	
	if len(committee.VRFSeed) == 0 {
		t.Error("Expected non-empty VRF seed in committee")
	}
	
	// Test with insufficient candidates
	insufficientCandidates := candidates[:2] // Only 2 candidates for committee size 3
	_, err = selector.SelectCommittee(ctx, seed, insufficientCandidates)
	if err == nil {
		t.Error("Expected error for insufficient candidates, got nil")
	}
	
	// Test determinism - same seed should give same committee
	committee2, err := selector.SelectCommittee(ctx, seed, candidates)
	if err != nil {
		t.Errorf("Expected successful second committee selection, got error: %v", err)
	}
	
	if len(committee.Members) != len(committee2.Members) {
		t.Error("Expected deterministic committee selection")
	}
	
	for i, member := range committee.Members {
		if member != committee2.Members[i] {
			t.Error("Expected same committee members for same seed")
			break
		}
	}
	
	// Test with different seed gives different committee (probably)
	differentSeed := []byte("different-committee-selection-seed")
	committee3, err := selector.SelectCommittee(ctx, differentSeed, candidates)
	if err != nil {
		t.Errorf("Expected successful third committee selection, got error: %v", err)
	}
	
	// It's possible (but unlikely) that different seeds give same committee
	// Just check that we get a valid committee
	if len(committee3.Members) != config.CommitteeSize {
		t.Errorf("Expected committee size %d for different seed, got %d", config.CommitteeSize, len(committee3.Members))
	}
}

func TestVRFCommitteeManager(t *testing.T) {
	provider := NewDefaultVRFProvider()
	config := DefaultCheckpointorConfig()
	selector := NewCommitteeSelector(provider, config)
	store := NewMemoryCheckpointStore()
	
	manager := NewVRFCommitteeManager(selector, store)
	ctx := context.Background()
	
	// Initially no committee
	_, err := manager.GetCurrentCommittee(ctx)
	if err == nil {
		t.Error("Expected error for no current committee, got nil")
	}
	
	// Store a test committee
	testCommittee := &Committee{
		Members:   []string{"did:key:member1", "did:key:member2", "did:key:member3"},
		Threshold: 2,
		Epoch:     123,
		VRFSeed:   []byte("test-seed"),
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour),
	}
	
	err = store.StoreCommittee(ctx, testCommittee)
	if err != nil {
		t.Fatalf("Failed to store test committee: %v", err)
	}
	
	// Now we should be able to get the current committee
	committee, err := manager.GetCurrentCommittee(ctx)
	if err != nil {
		t.Errorf("Expected to get current committee, got error: %v", err)
	}
	
	if committee == nil {
		t.Fatal("Expected non-nil committee")
	}
	
	if len(committee.Members) != len(testCommittee.Members) {
		t.Errorf("Expected %d members, got %d", len(testCommittee.Members), len(committee.Members))
	}
	
	// Test IsMember
	isMember, err := manager.IsMember(ctx, "did:key:member1")
	if err != nil {
		t.Errorf("Expected successful member check, got error: %v", err)
	}
	
	if !isMember {
		t.Error("Expected member1 to be in committee")
	}
	
	// Test non-member
	isNotMember, err := manager.IsMember(ctx, "did:key:nonmember")
	if err != nil {
		t.Errorf("Expected successful non-member check, got error: %v", err)
	}
	
	if isNotMember {
		t.Error("Expected nonmember to not be in committee")
	}
}

func TestVRFDeterminism(t *testing.T) {
	provider := NewDefaultVRFProvider()
	
	// Generate key pair
	privateKey, publicKey, err := GenerateVRFKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	seed := []byte("determinism-test-seed")
	
	// Generate VRF multiple times with same inputs
	proof1, output1, err := provider.GenerateProof(privateKey, seed)
	if err != nil {
		t.Fatalf("Failed to generate first proof: %v", err)
	}
	
	proof2, output2, err := provider.GenerateProof(privateKey, seed)
	if err != nil {
		t.Fatalf("Failed to generate second proof: %v", err)
	}
	
	// Should be identical
	if !bytesEqual(proof1, proof2) {
		t.Error("Expected identical proofs for same inputs")
	}
	
	if !bytesEqual(output1, output2) {
		t.Error("Expected identical outputs for same inputs")
	}
	
	// Verify both proofs
	err = provider.VerifyProof(publicKey, seed, proof1, output1)
	if err != nil {
		t.Errorf("Expected first proof to verify: %v", err)
	}
	
	err = provider.VerifyProof(publicKey, seed, proof2, output2)
	if err != nil {
		t.Errorf("Expected second proof to verify: %v", err)
	}
}

func TestGenerateVRFKeyPair(t *testing.T) {
	// Generate multiple key pairs and ensure they're different
	keyPairs := make(map[string]bool)
	
	for i := 0; i < 10; i++ {
		privateKey, publicKey, err := GenerateVRFKeyPair()
		if err != nil {
			t.Errorf("Failed to generate key pair %d: %v", i, err)
			continue
		}
		
		if len(privateKey) != ed25519.PrivateKeySize {
			t.Errorf("Private key %d has wrong size: expected %d, got %d", i, ed25519.PrivateKeySize, len(privateKey))
		}
		
		if len(publicKey) != ed25519.PublicKeySize {
			t.Errorf("Public key %d has wrong size: expected %d, got %d", i, ed25519.PublicKeySize, len(publicKey))
		}
		
		// Check uniqueness
		keyString := fmt.Sprintf("%x", publicKey)
		if keyPairs[keyString] {
			t.Errorf("Duplicate key pair generated at iteration %d", i)
		}
		keyPairs[keyString] = true
	}
}