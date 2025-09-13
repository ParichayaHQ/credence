package tests

import (
	"context"
	"testing"

	"github.com/ParichayaHQ/credence/internal/events"
	"github.com/ParichayaHQ/credence/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventLifecycle tests the complete event lifecycle
func TestEventLifecycle(t *testing.T) {
	ctx := context.Background()
	
	// Create test identities
	alice, err := testutil.NewTestKeyPair()
	require.NoError(t, err)
	
	bob, err := testutil.NewTestKeyPair()
	require.NoError(t, err)
	
	// Create services
	eventService := testutil.NewMockEventService()
	_ = testutil.NewMockStorageService() // Available for future use
	
	// Test 1: Create and submit a vouch event
	t.Run("CreateAndSubmitVouch", func(t *testing.T) {
		vouch, err := testutil.CreateTestVouchEvent(alice, bob, "commerce")
		require.NoError(t, err)
		
		// Validate event
		err = events.ValidateEvent(vouch)
		require.NoError(t, err)
		
		// Submit event
		receipt, err := eventService.SubmitEvent(ctx, vouch)
		require.NoError(t, err)
		assert.Equal(t, "processed", receipt.Status)
		
		// Verify event can be retrieved
		retrievedEvent, err := eventService.GetEvent(ctx, receipt.CID)
		require.NoError(t, err)
		assert.Equal(t, vouch.From, retrievedEvent.From)
		assert.Equal(t, vouch.To, retrievedEvent.To)
		assert.Equal(t, vouch.Type, retrievedEvent.Type)
	})
	
	// Test 2: Signature validation
	t.Run("SignatureValidation", func(t *testing.T) {
		vouch, err := testutil.CreateTestVouchEvent(alice, bob, "general")
		require.NoError(t, err)
		
		validator := testutil.NewTestEventValidator()
		isValid, err := validator.ValidateEventSignature(vouch)
		require.NoError(t, err)
		assert.True(t, isValid)
		
		// Test invalid signature
		vouch.Signature = "invalid-signature"
		isValid, err = validator.ValidateEventSignature(vouch)
		require.Error(t, err) // Should fail due to invalid base64
	})
	
	// Test 3: Event queries
	t.Run("EventQueries", func(t *testing.T) {
		// Create multiple events
		vouch1, err := testutil.CreateTestVouchEvent(alice, bob, "commerce")
		require.NoError(t, err)
		
		vouch2, err := testutil.CreateTestVouchEvent(bob, alice, "commerce")
		require.NoError(t, err)
		
		report, err := testutil.CreateTestReportEvent(alice, bob, "commerce", "spam")
		require.NoError(t, err)
		
		// Submit all events
		_, err = eventService.SubmitEvent(ctx, vouch1)
		require.NoError(t, err)
		
		_, err = eventService.SubmitEvent(ctx, vouch2)
		require.NoError(t, err)
		
		_, err = eventService.SubmitEvent(ctx, report)
		require.NoError(t, err)
		
		// Query events for Alice
		aliceEvents, err := eventService.GetEventsByDID(ctx, alice.DID.String(), "", 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(aliceEvents), 2) // At least 2 events involving Alice
		
		// Query only vouch events for Alice
		aliceVouches, err := eventService.GetEventsByDID(ctx, alice.DID.String(), events.EventTypeVouch, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(aliceVouches), 1) // At least 1 vouch involving Alice
		
		// Query with limit
		limitedEvents, err := eventService.GetEventsByDID(ctx, alice.DID.String(), "", 1)
		require.NoError(t, err)
		assert.Equal(t, 1, len(limitedEvents))
	})
}

// TestTrustScoreComputation tests trust score calculation
func TestTrustScoreComputation(t *testing.T) {
	ctx := context.Background()
	
	// Create test network
	network, err := testutil.CreateTestNetwork(5)
	require.NoError(t, err)
	
	scorer := testutil.NewMockScorerService()
	
	t.Run("BasicScoreComputation", func(t *testing.T) {
		alice := network[0]
		
		// Set a specific score for testing
		scorer.SetScore(alice.DID.String(), "commerce", 75.5)
		
		// Compute score
		scoreRecord, err := scorer.ComputeScore(ctx, alice.DID.String(), "commerce")
		require.NoError(t, err)
		assert.Equal(t, 75.5, scoreRecord.Score)
		assert.Equal(t, alice.DID.String(), scoreRecord.DID)
		assert.Equal(t, "commerce", scoreRecord.Context)
	})
	
	t.Run("ThresholdProofGeneration", func(t *testing.T) {
		alice := network[0]
		
		// Set score above threshold
		scorer.SetScore(alice.DID.String(), "hiring", 80.0)
		
		// Generate threshold proof for score >= 75
		proof, err := scorer.GenerateThresholdProof(ctx, alice.DID.String(), "hiring", 75.0, "test-nonce")
		require.NoError(t, err)
		assert.Equal(t, "hiring", proof.Context)
		assert.Equal(t, 75.0, proof.Threshold)
		assert.Equal(t, "test-nonce", proof.Nonce)
		
		// Verify the proof
		isValid, err := scorer.VerifyThresholdProof(ctx, proof)
		require.NoError(t, err)
		assert.True(t, isValid)
	})
}

// TestVouchNetwork tests vouch network creation and validation
func TestVouchNetwork(t *testing.T) {
	// Create a test network of 5 identities
	network, err := testutil.CreateTestNetwork(5)
	require.NoError(t, err)
	
	// Create a vouch chain: A vouches for B, B vouches for C, etc.
	vouches, err := testutil.CreateVouchChain(network, "general")
	require.NoError(t, err)
	assert.Equal(t, 4, len(vouches)) // 5 identities = 4 vouches
	
	// Validate all vouches
	validator := testutil.NewTestEventValidator()
	for i, vouch := range vouches {
		t.Run("ValidateVouch"+string(rune('A'+i)), func(t *testing.T) {
			err := events.ValidateEvent(vouch)
			require.NoError(t, err)
			
			isValidSig, err := validator.ValidateEventSignature(vouch)
			require.NoError(t, err)
			assert.True(t, isValidSig)
		})
	}
}

// TestEventCanonicalization tests canonical JSON generation
func TestEventCanonicalization(t *testing.T) {
	alice, err := testutil.NewTestKeyPair()
	require.NoError(t, err)
	
	bob, err := testutil.NewTestKeyPair()
	require.NoError(t, err)
	
	// Create two identical events with same data
	vouch1, err := testutil.CreateTestVouchEvent(alice, bob, "commerce")
	require.NoError(t, err)
	
	// Create signable event with same data
	signable := &events.SignableEvent{
		Type:     vouch1.Type,
		From:     vouch1.From,
		To:       vouch1.To,
		Context:  vouch1.Context,
		Epoch:    vouch1.Epoch,
		Nonce:    vouch1.Nonce,
		IssuedAt: vouch1.IssuedAt,
	}
	
	// Canonicalize both
	canonical1, err := events.CanonicalizeEventWithSignature(vouch1)
	require.NoError(t, err)
	
	canonical2, err := events.CanonicalizeEvent(signable)
	require.NoError(t, err)
	
	// The canonical representation of signable event should be deterministic
	canonical3, err := events.CanonicalizeEvent(signable)
	require.NoError(t, err)
	
	assert.Equal(t, canonical2, canonical3, "Canonicalization should be deterministic")
	
	// Validate canonical JSON format
	err = events.ValidateCanonicalJSON(canonical1)
	assert.NoError(t, err)
	
	err = events.ValidateCanonicalJSON(canonical2)
	assert.NoError(t, err)
}

// TestCheckpointSystem tests checkpoint creation and verification
func TestCheckpointSystem(t *testing.T) {
	ctx := context.Background()
	
	checkpointService := testutil.NewMockCheckpointService()
	
	t.Run("CreateAndRetrieveCheckpoint", func(t *testing.T) {
		// Create a checkpoint
		checkpoint, err := checkpointService.CreateCheckpoint(ctx, "test-root", "test-tree", 100)
		require.NoError(t, err)
		assert.Equal(t, "test-root", checkpoint.Root)
		assert.Equal(t, int64(100), checkpoint.Epoch)
		
		// Retrieve the checkpoint
		retrieved, err := checkpointService.GetCheckpoint(ctx, 100)
		require.NoError(t, err)
		assert.Equal(t, checkpoint.Root, retrieved.Root)
		assert.Equal(t, checkpoint.Epoch, retrieved.Epoch)
		
		// Get latest checkpoint
		latest, err := checkpointService.GetLatestCheckpoint(ctx)
		require.NoError(t, err)
		assert.Equal(t, checkpoint.Epoch, latest.Epoch)
	})
	
	t.Run("VerifyCheckpoint", func(t *testing.T) {
		checkpoint := testutil.CreateTestCheckpoint(200, "another-root")
		
		isValid, err := checkpointService.VerifyCheckpoint(ctx, checkpoint)
		require.NoError(t, err)
		assert.True(t, isValid)
	})
}

// BenchmarkEventCreation benchmarks event creation performance
func BenchmarkEventCreation(b *testing.B) {
	alice, err := testutil.NewTestKeyPair()
	require.NoError(b, err)
	
	bob, err := testutil.NewTestKeyPair()
	require.NoError(b, err)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := testutil.CreateTestVouchEvent(alice, bob, "commerce")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCanonicalization benchmarks JSON canonicalization performance
func BenchmarkCanonicalization(b *testing.B) {
	alice, err := testutil.NewTestKeyPair()
	require.NoError(b, err)
	
	bob, err := testutil.NewTestKeyPair()
	require.NoError(b, err)
	
	vouch, err := testutil.CreateTestVouchEvent(alice, bob, "commerce")
	require.NoError(b, err)
	
	signable := vouch.ToSignable()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := events.CanonicalizeEvent(signable)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestPropertyBasedCanonicalization uses property-based testing for canonicalization
func TestPropertyBasedCanonicalization(t *testing.T) {
	// Property: Canonicalization should be deterministic
	for i := 0; i < 100; i++ {
		alice, err := testutil.NewTestKeyPair()
		require.NoError(t, err)
		
		bob, err := testutil.NewTestKeyPair()
		require.NoError(t, err)
		
		vouch, err := testutil.CreateTestVouchEvent(alice, bob, "general")
		require.NoError(t, err)
		
		signable := vouch.ToSignable()
		
		// Canonicalize multiple times
		canonical1, err := events.CanonicalizeEvent(signable)
		require.NoError(t, err)
		
		canonical2, err := events.CanonicalizeEvent(signable)
		require.NoError(t, err)
		
		// Should always be identical
		assert.Equal(t, canonical1, canonical2, 
			"Canonicalization should be deterministic for iteration %d", i)
	}
}