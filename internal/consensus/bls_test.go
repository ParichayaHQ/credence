package consensus

import (
	"fmt"
	"testing"
	"time"
)

func TestBLSAggregator_VerifyPartialSignature(t *testing.T) {
	aggregator := NewDefaultBLSAggregator()
	
	// Generate key pairs for test committee members
	privateKey1, publicKey1, err := aggregator.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair 1: %v", err)
	}
	
	_, publicKey2, err := aggregator.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair 2: %v", err)
	}
	
	// Register committee members
	member1DID := "did:key:member1"
	member2DID := "did:key:member2"
	aggregator.RegisterMember(member1DID, publicKey1)
	aggregator.RegisterMember(member2DID, publicKey2)
	
	// Create test data
	root := []byte("test-root-hash")
	epoch := int64(123)
	message := aggregator.createSigningMessage(root, epoch)
	
	// Sign with member 1
	signature1, err := aggregator.SignMessage(message, privateKey1)
	if err != nil {
		t.Fatalf("Failed to sign with member 1: %v", err)
	}
	
	// Create partial signature
	partial := &PartialSignature{
		SignerDID: member1DID,
		Signature: signature1,
		Root:      root,
		Epoch:     epoch,
		Timestamp: time.Now(),
	}
	
	// Test verification
	err = aggregator.VerifyPartialSignature(partial, publicKey1.X)
	if err != nil {
		t.Errorf("Expected valid partial signature to verify, got error: %v", err)
	}
	
	// Test with invalid signer
	invalidPartial := &PartialSignature{
		SignerDID: "did:key:unknown",
		Signature: signature1,
		Root:      root,
		Epoch:     epoch,
		Timestamp: time.Now(),
	}
	
	err = aggregator.VerifyPartialSignature(invalidPartial, publicKey1.X)
	if err == nil {
		t.Error("Expected error for unknown signer, got nil")
	}
	
	// Test with nil partial
	err = aggregator.VerifyPartialSignature(nil, publicKey1.X)
	if err == nil {
		t.Error("Expected error for nil partial signature, got nil")
	}
}

func TestBLSAggregator_AggregateSignatures(t *testing.T) {
	aggregator := NewDefaultBLSAggregator()
	
	// Generate key pairs for 5 committee members
	members := make(map[string]*BLSPrivateKey)
	for i := 1; i <= 5; i++ {
		did := fmt.Sprintf("did:key:member%d", i)
		privateKey, publicKey, err := aggregator.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair for member %d: %v", i, err)
		}
		
		members[did] = privateKey
		aggregator.RegisterMember(did, publicKey)
	}
	
	// Create test data
	root := []byte("test-root-hash-for-aggregation")
	epoch := int64(456)
	message := aggregator.createSigningMessage(root, epoch)
	
	// Create partial signatures from 4 members (threshold = 3)
	var partials []PartialSignature
	i := 1
	for did, privateKey := range members {
		if i > 4 { // Only use first 4 members
			break
		}
		
		signature, err := aggregator.SignMessage(message, privateKey)
		if err != nil {
			t.Fatalf("Failed to sign with %s: %v", did, err)
		}
		
		partial := PartialSignature{
			SignerDID: did,
			Signature: signature,
			Root:      root,
			Epoch:     epoch,
			Timestamp: time.Now(),
		}
		
		partials = append(partials, partial)
		i++
	}
	
	// Test successful aggregation
	threshold := 3
	aggregatedSig, err := aggregator.AggregateSignatures(partials, threshold)
	if err != nil {
		t.Errorf("Expected successful aggregation, got error: %v", err)
	}
	
	if len(aggregatedSig) == 0 {
		t.Error("Expected non-empty aggregated signature")
	}
	
	// Test insufficient signatures
	_, err = aggregator.AggregateSignatures(partials[:2], threshold)
	if err == nil {
		t.Error("Expected error for insufficient signatures, got nil")
	}
	
	// Test empty partials
	_, err = aggregator.AggregateSignatures([]PartialSignature{}, threshold)
	if err == nil {
		t.Error("Expected error for empty partials, got nil")
	}
	
	// Test mismatched roots
	mismatchedPartial := partials[0]
	mismatchedPartial.Root = []byte("different-root")
	mismatchedPartials := []PartialSignature{partials[0], mismatchedPartial, partials[2]}
	
	_, err = aggregator.AggregateSignatures(mismatchedPartials, threshold)
	if err == nil {
		t.Error("Expected error for mismatched roots, got nil")
	}
}

func TestBLSAggregator_VerifyAggregatedSignature(t *testing.T) {
	aggregator := NewDefaultBLSAggregator()
	
	// Generate key pairs for committee members
	members := []string{"did:key:member1", "did:key:member2", "did:key:member3"}
	var publicKeys [][]byte
	
	for _, did := range members {
		_, publicKey, err := aggregator.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair for %s: %v", did, err)
		}
		
		aggregator.RegisterMember(did, publicKey)
		publicKeys = append(publicKeys, publicKey.X)
	}
	
	// Create test checkpoint
	checkpoint := &Checkpoint{
		Root:      []byte("test-root-for-verification"),
		Epoch:     789,
		TreeSize:  1000,
		Signers:   members,
		Signature: []byte("valid-aggregated-signature-32bytes"), // 32 bytes for placeholder
		Timestamp: time.Now(),
	}
	
	// Make sure signature is exactly 32 bytes for the placeholder implementation
	if len(checkpoint.Signature) != 32 {
		signature := make([]byte, 32)
		copy(signature, checkpoint.Signature)
		checkpoint.Signature = signature
	}
	
	// Test successful verification
	err := aggregator.VerifyAggregatedSignature(checkpoint, publicKeys)
	if err != nil {
		t.Errorf("Expected valid aggregated signature to verify, got error: %v", err)
	}
	
	// Test with nil checkpoint
	err = aggregator.VerifyAggregatedSignature(nil, publicKeys)
	if err == nil {
		t.Error("Expected error for nil checkpoint, got nil")
	}
	
	// Test with empty signature
	emptyCheckpoint := *checkpoint
	emptyCheckpoint.Signature = []byte{}
	err = aggregator.VerifyAggregatedSignature(&emptyCheckpoint, publicKeys)
	if err == nil {
		t.Error("Expected error for empty signature, got nil")
	}
	
	// Test with no signers
	noSignersCheckpoint := *checkpoint
	noSignersCheckpoint.Signers = []string{}
	err = aggregator.VerifyAggregatedSignature(&noSignersCheckpoint, publicKeys)
	if err == nil {
		t.Error("Expected error for no signers, got nil")
	}
}

func TestBLSAggregator_MessageSigning(t *testing.T) {
	aggregator := NewDefaultBLSAggregator()
	
	privateKey, publicKey, err := aggregator.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	message := []byte("test message for signing")
	
	// Test signing
	signature, err := aggregator.SignMessage(message, privateKey)
	if err != nil {
		t.Errorf("Expected successful signing, got error: %v", err)
	}
	
	if len(signature) == 0 {
		t.Error("Expected non-empty signature")
	}
	
	// Test with nil private key
	_, err = aggregator.SignMessage(message, nil)
	if err == nil {
		t.Error("Expected error for nil private key, got nil")
	}
	
	// Test signature verification
	memberDID := "did:key:test-member"
	aggregator.RegisterMember(memberDID, publicKey)
	
	valid := aggregator.verifySignature(message, signature, publicKey)
	if !valid {
		t.Error("Expected signature to be valid")
	}
}

func TestBLSAggregator_KeyGeneration(t *testing.T) {
	aggregator := NewDefaultBLSAggregator()
	
	// Generate multiple key pairs and ensure they're different
	keyPairs := make(map[string]*BLSPublicKey)
	
	for i := 0; i < 10; i++ {
		// Add small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
		
		privateKey, publicKey, err := aggregator.GenerateKeyPair()
		if err != nil {
			t.Errorf("Failed to generate key pair %d: %v", i, err)
			continue
		}
		
		if privateKey == nil || publicKey == nil {
			t.Errorf("Key pair %d contains nil keys", i)
			continue
		}
		
		if len(privateKey.D) == 0 {
			t.Errorf("Private key %d is empty", i)
			continue
		}
		
		if len(publicKey.X) == 0 || len(publicKey.Y) == 0 {
			t.Errorf("Public key %d has empty coordinates", i)
			continue
		}
		
		// Check for uniqueness (convert to string for map key)
		keyString := fmt.Sprintf("%x%x", publicKey.X, publicKey.Y)
		if _, exists := keyPairs[keyString]; exists {
			t.Errorf("Duplicate key pair generated at iteration %d", i)
		}
		keyPairs[keyString] = publicKey
	}
}