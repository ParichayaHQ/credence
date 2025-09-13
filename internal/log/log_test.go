package log

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryTransparencyLog(t *testing.T) {
	config := DefaultConfig()
	log, err := NewMemoryTransparencyLog(config)
	require.NoError(t, err)
	defer log.Close()
	
	ctx := context.Background()
	
	t.Run("InitialState", func(t *testing.T) {
		size, err := log.GetTreeSize(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
		
		sth, err := log.GetSignedTreeHead(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), sth.TreeSize)
		assert.Equal(t, config.TreeID, sth.TreeID)
	})
	
	t.Run("AppendSingleLeaf", func(t *testing.T) {
		testData := []byte("test leaf data")
		hash := sha256.Sum256(testData)
		
		leaf := Leaf{
			LeafValue: testData,
			LeafHash:  hash[:],
		}
		
		result, err := log.AppendLeaves(ctx, []Leaf{leaf})
		require.NoError(t, err)
		
		assert.Equal(t, int64(1), result.TreeSize)
		assert.Len(t, result.LeafIndexes, 1)
		assert.Equal(t, int64(0), result.LeafIndexes[0])
		assert.NotNil(t, result.SignedTreeHead)
		assert.NotEmpty(t, result.RootHash)
	})
	
	t.Run("AppendMultipleLeaves", func(t *testing.T) {
		leaves := make([]Leaf, 3)
		for i := 0; i < 3; i++ {
			testData := []byte("leaf " + string(rune('A'+i)))
			hash := sha256.Sum256(testData)
			leaves[i] = Leaf{
				LeafValue: testData,
				LeafHash:  hash[:],
			}
		}
		
		result, err := log.AppendLeaves(ctx, leaves)
		require.NoError(t, err)
		
		assert.Equal(t, int64(4), result.TreeSize) // 1 from previous test + 3 new
		assert.Len(t, result.LeafIndexes, 3)
		assert.Equal(t, []int64{1, 2, 3}, result.LeafIndexes)
	})
	
	t.Run("RetrieveLeafByHash", func(t *testing.T) {
		testData := []byte("retrieve me")
		hash := sha256.Sum256(testData)
		
		leaf := Leaf{
			LeafValue: testData,
			LeafHash:  hash[:],
		}
		
		// Append leaf
		result, err := log.AppendLeaves(ctx, []Leaf{leaf})
		require.NoError(t, err)
		
		// Retrieve by hash
		retrieved, err := log.GetLeafByHash(ctx, hash[:])
		require.NoError(t, err)
		
		assert.Equal(t, testData, retrieved.LeafValue)
		assert.Equal(t, hash[:], retrieved.LeafHash)
		assert.Equal(t, result.LeafIndexes[0], retrieved.LeafIndex)
	})
	
	t.Run("GetLeavesByRange", func(t *testing.T) {
		// Get current tree size
		_, err := log.GetTreeSize(ctx)
		require.NoError(t, err)
		
		// Add some more leaves
		leaves := make([]Leaf, 5)
		for i := 0; i < 5; i++ {
			testData := []byte("range test " + string(rune('1'+i)))
			hash := sha256.Sum256(testData)
			leaves[i] = Leaf{
				LeafValue: testData,
				LeafHash:  hash[:],
			}
		}
		
		result, err := log.AppendLeaves(ctx, leaves)
		require.NoError(t, err)
		
		// Get range of leaves
		startIdx := result.LeafIndexes[0]
		endIdx := result.LeafIndexes[2] // First 3 of the new leaves
		
		rangeLeaves, err := log.GetLeavesByRange(ctx, startIdx, endIdx)
		require.NoError(t, err)
		
		assert.Len(t, rangeLeaves, 3)
		assert.Equal(t, leaves[0].LeafValue, rangeLeaves[0].LeafValue)
		assert.Equal(t, leaves[2].LeafValue, rangeLeaves[2].LeafValue)
	})
	
	t.Run("GetInclusionProof", func(t *testing.T) {
		testData := []byte("proof test")
		hash := sha256.Sum256(testData)
		
		leaf := Leaf{
			LeafValue: testData,
			LeafHash:  hash[:],
		}
		
		// Append leaf
		result, err := log.AppendLeaves(ctx, []Leaf{leaf})
		require.NoError(t, err)
		
		// Get inclusion proof
		proof, err := log.GetInclusionProof(ctx, hash[:], result.TreeSize)
		require.NoError(t, err)
		
		assert.Equal(t, result.LeafIndexes[0], proof.LeafIndex)
		assert.Equal(t, result.TreeSize, proof.TreeSize)
		assert.NotNil(t, proof.AuditPath)
	})
	
	t.Run("GetConsistencyProof", func(t *testing.T) {
		// Get current tree size
		fromSize, err := log.GetTreeSize(ctx)
		require.NoError(t, err)
		
		// Add more leaves
		testData := []byte("consistency test")
		hash := sha256.Sum256(testData)
		leaf := Leaf{
			LeafValue: testData,
			LeafHash:  hash[:],
		}
		
		result, err := log.AppendLeaves(ctx, []Leaf{leaf})
		require.NoError(t, err)
		
		toSize := result.TreeSize
		
		// Get consistency proof
		proof, err := log.GetConsistencyProof(ctx, fromSize, toSize)
		require.NoError(t, err)
		
		assert.Equal(t, fromSize, proof.FirstTreeSize)
		assert.Equal(t, toSize, proof.SecondTreeSize)
		assert.NotNil(t, proof.ProofPath)
	})
	
	t.Run("SignedTreeHeadProgression", func(t *testing.T) {
		// Get initial STH
		sth1, err := log.GetSignedTreeHead(ctx)
		require.NoError(t, err)
		
		initialSize := sth1.TreeSize
		
		// Add a leaf
		testData := []byte("sth progression test")
		hash := sha256.Sum256(testData)
		leaf := Leaf{
			LeafValue: testData,
			LeafHash:  hash[:],
		}
		
		result, err := log.AppendLeaves(ctx, []Leaf{leaf})
		require.NoError(t, err)
		
		// Get new STH
		sth2, err := log.GetSignedTreeHead(ctx)
		require.NoError(t, err)
		
		// Verify progression
		assert.Equal(t, initialSize+1, sth2.TreeSize)
		assert.Equal(t, result.TreeSize, sth2.TreeSize)
		assert.NotEqual(t, sth1.RootHash, sth2.RootHash)
		assert.NotEqual(t, sth1.Signature, sth2.Signature)
		assert.True(t, sth2.Timestamp.After(sth1.Timestamp) || sth2.Timestamp.Equal(sth1.Timestamp))
	})
	
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test empty leaves
		_, err := log.AppendLeaves(ctx, []Leaf{})
		assert.Error(t, err)
		
		// Test non-existent leaf
		fakeHash := []byte("nonexistent")
		_, err = log.GetLeafByHash(ctx, fakeHash)
		assert.Error(t, err)
		
		// Test invalid range
		_, err = log.GetLeavesByRange(ctx, -1, 0)
		assert.Error(t, err)
		
		// Test invalid consistency proof range
		_, err = log.GetConsistencyProof(ctx, 100, 50)
		assert.Error(t, err)
	})
}

func TestEventReference(t *testing.T) {
	t.Run("SerializeEventReference", func(t *testing.T) {
		eventRef := EventReference{
			CID:         "QmTestCID123",
			ContentHash: []byte("test-hash"),
			Type:        "vouch",
			From:        "did:key:test1",
			To:          "did:key:test2",
			Epoch:       "2025-09",
			Timestamp:   time.Now(),
			ContentSize: 256,
		}
		
		// Should be able to JSON marshal/unmarshal
		data, err := eventRef.MarshalJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		
		var decoded EventReference
		err = decoded.UnmarshalJSON(data)
		require.NoError(t, err)
		
		assert.Equal(t, eventRef.CID, decoded.CID)
		assert.Equal(t, eventRef.Type, decoded.Type)
		assert.Equal(t, eventRef.From, decoded.From)
		assert.Equal(t, eventRef.To, decoded.To)
	})
}

func TestLogConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		
		assert.Equal(t, int64(1), config.TreeID)
		assert.Equal(t, "ed25519", config.SigningKey.Algorithm)
		assert.Equal(t, "memory", config.Storage.Backend)
		assert.Equal(t, "sha256", config.HashAlgorithm)
		assert.True(t, config.Batching.EnableBatching)
		assert.True(t, config.Batching.MaxBatchSize > 0)
		assert.True(t, config.Batching.MaxBatchDelay > 0)
	})
}