package log

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/ParichayaHQ/credence/internal/crypto"
)

// MemoryTransparencyLog implements TransparencyLog using in-memory storage
// This is suitable for development, testing, and small-scale deployments
type MemoryTransparencyLog struct {
	config *Config
	
	// Storage
	leaves   []Leaf            // All leaves in order
	leafMap  map[string]int64  // Hash -> leaf index
	treeSize int64
	
	// Signing
	signer crypto.Signer
	
	// Synchronization
	mu sync.RWMutex
	
	// Statistics
	stats LogStats
	
	// State
	closed bool
}

// NewMemoryTransparencyLog creates a new in-memory transparency log
func NewMemoryTransparencyLog(config *Config) (*MemoryTransparencyLog, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	log := &MemoryTransparencyLog{
		config:  config,
		leaves:  make([]Leaf, 0),
		leafMap: make(map[string]int64),
		stats: LogStats{
			FirstAppend: time.Now(), // Will be overwritten on first append
		},
	}
	
	// Initialize signing key
	if err := log.initSigner(); err != nil {
		return nil, fmt.Errorf("failed to initialize signer: %w", err)
	}
	
	return log, nil
}

// initSigner initializes the cryptographic signer
func (m *MemoryTransparencyLog) initSigner() error {
	switch m.config.SigningKey.Algorithm {
	case "ed25519":
		// Generate a new key if none provided
		if m.config.SigningKey.PrivateKey == "" {
			keyPair, err := crypto.NewEd25519KeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
			}
			signer := crypto.NewEd25519Signer(keyPair)
			m.signer = signer
		} else {
			// TODO: Load from PEM or file
			return fmt.Errorf("loading Ed25519 keys from config not yet implemented")
		}
	default:
		return fmt.Errorf("unsupported signing algorithm: %s", m.config.SigningKey.Algorithm)
	}
	
	return nil
}

// AppendLeaves implements TransparencyLog.AppendLeaves
func (m *MemoryTransparencyLog) AppendLeaves(ctx context.Context, leaves []Leaf) (*AppendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.closed {
		return nil, fmt.Errorf("log is closed")
	}
	
	if len(leaves) == 0 {
		return nil, fmt.Errorf("no leaves to append")
	}
	
	// Validate and assign sequence numbers
	leafIndexes := make([]int64, len(leaves))
	now := time.Now()
	
	for i, leaf := range leaves {
		// Assign sequence number
		leafIndex := m.treeSize + int64(i)
		leafIndexes[i] = leafIndex
		
		// Set fields
		leaf.LeafIndex = leafIndex
		leaf.Timestamp = now
		
		// Validate hash if provided
		if len(leaf.LeafHash) == 0 {
			// Compute hash from leaf value
			hash := sha256.Sum256(leaf.LeafValue)
			leaf.LeafHash = hash[:]
		}
		
		// Store leaf
		m.leaves = append(m.leaves, leaf)
		m.leafMap[string(leaf.LeafHash)] = leafIndex
	}
	
	// Update tree size
	oldSize := m.treeSize
	m.treeSize += int64(len(leaves))
	
	// Compute new tree root
	rootHash := m.computeRootHash()
	
	// Create signed tree head
	sth, err := m.createSignedTreeHead(m.treeSize, rootHash, now)
	if err != nil {
		// Rollback changes
		m.leaves = m.leaves[:oldSize]
		m.treeSize = oldSize
		// Remove from leaf map
		for _, leaf := range leaves {
			delete(m.leafMap, string(leaf.LeafHash))
		}
		return nil, fmt.Errorf("failed to sign tree head: %w", err)
	}
	
	// Update statistics
	m.stats.TreeSize = m.treeSize
	m.stats.TotalLeaves = m.treeSize
	m.stats.LastAppend = now
	if oldSize == 0 {
		m.stats.FirstAppend = now
	}
	
	// Calculate append rate
	duration := now.Sub(m.stats.FirstAppend).Seconds()
	if duration > 0 {
		m.stats.AppendRate = float64(m.treeSize) / duration
	}
	
	return &AppendResult{
		TreeSize:       m.treeSize,
		RootHash:       rootHash,
		LeafIndexes:    leafIndexes,
		SignedTreeHead: sth,
	}, nil
}

// GetLeafByHash implements TransparencyLog.GetLeafByHash
func (m *MemoryTransparencyLog) GetLeafByHash(ctx context.Context, leafHash []byte) (*Leaf, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return nil, fmt.Errorf("log is closed")
	}
	
	leafIndex, exists := m.leafMap[string(leafHash)]
	if !exists {
		return nil, fmt.Errorf("leaf not found")
	}
	
	if leafIndex >= int64(len(m.leaves)) {
		return nil, fmt.Errorf("leaf index out of range")
	}
	
	leaf := m.leaves[leafIndex]
	return &leaf, nil
}

// GetLeavesByRange implements TransparencyLog.GetLeavesByRange
func (m *MemoryTransparencyLog) GetLeavesByRange(ctx context.Context, startSeq, endSeq int64) ([]Leaf, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return nil, fmt.Errorf("log is closed")
	}
	
	if startSeq < 0 || endSeq < startSeq || startSeq >= m.treeSize {
		return nil, fmt.Errorf("invalid range: start=%d, end=%d, treeSize=%d", startSeq, endSeq, m.treeSize)
	}
	
	// Adjust end to be within bounds
	if endSeq >= m.treeSize {
		endSeq = m.treeSize - 1
	}
	
	result := make([]Leaf, endSeq-startSeq+1)
	copy(result, m.leaves[startSeq:endSeq+1])
	
	return result, nil
}

// GetInclusionProof implements TransparencyLog.GetInclusionProof
func (m *MemoryTransparencyLog) GetInclusionProof(ctx context.Context, leafHash []byte, treeSize int64) (*InclusionProof, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return nil, fmt.Errorf("log is closed")
	}
	
	leafIndex, exists := m.leafMap[string(leafHash)]
	if !exists {
		return nil, fmt.Errorf("leaf not found")
	}
	
	if treeSize <= 0 || treeSize > m.treeSize {
		treeSize = m.treeSize
	}
	
	if leafIndex >= treeSize {
		return nil, fmt.Errorf("leaf not included in tree of size %d", treeSize)
	}
	
	// Compute audit path
	auditPath := m.computeAuditPath(leafIndex, treeSize)
	
	return &InclusionProof{
		LeafIndex: leafIndex,
		TreeSize:  treeSize,
		AuditPath: auditPath,
	}, nil
}

// GetConsistencyProof implements TransparencyLog.GetConsistencyProof
func (m *MemoryTransparencyLog) GetConsistencyProof(ctx context.Context, fromSize, toSize int64) (*ConsistencyProof, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return nil, fmt.Errorf("log is closed")
	}
	
	if fromSize < 0 || toSize < fromSize || toSize > m.treeSize {
		return nil, fmt.Errorf("invalid size range: from=%d, to=%d, current=%d", fromSize, toSize, m.treeSize)
	}
	
	if fromSize == toSize {
		return &ConsistencyProof{
			FirstTreeSize:  fromSize,
			SecondTreeSize: toSize,
			ProofPath:      [][]byte{},
		}, nil
	}
	
	// Compute consistency proof path
	proofPath := m.computeConsistencyPath(fromSize, toSize)
	
	return &ConsistencyProof{
		FirstTreeSize:  fromSize,
		SecondTreeSize: toSize,
		ProofPath:      proofPath,
	}, nil
}

// GetSignedTreeHead implements TransparencyLog.GetSignedTreeHead
func (m *MemoryTransparencyLog) GetSignedTreeHead(ctx context.Context) (*SignedTreeHead, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return nil, fmt.Errorf("log is closed")
	}
	
	rootHash := m.computeRootHash()
	return m.createSignedTreeHead(m.treeSize, rootHash, time.Now())
}

// GetTreeSize implements TransparencyLog.GetTreeSize
func (m *MemoryTransparencyLog) GetTreeSize(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return 0, fmt.Errorf("log is closed")
	}
	
	return m.treeSize, nil
}

// Close implements TransparencyLog.Close
func (m *MemoryTransparencyLog) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.closed = true
	return nil
}

// GetStats returns log statistics
func (m *MemoryTransparencyLog) GetStats() LogStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := m.stats
	stats.StorageBytes = int64(len(m.leaves)) * 256 // Rough estimate
	
	return stats
}

// computeRootHash computes the Merkle tree root hash
func (m *MemoryTransparencyLog) computeRootHash() []byte {
	if m.treeSize == 0 {
		return nil
	}
	
	// Simple implementation: hash all leaf hashes together
	// In a real implementation, this would be a proper Merkle tree
	hasher := sha256.New()
	
	for i := int64(0); i < m.treeSize; i++ {
		hasher.Write(m.leaves[i].LeafHash)
	}
	
	hash := hasher.Sum(nil)
	return hash
}

// computeAuditPath computes the audit path for inclusion proof
func (m *MemoryTransparencyLog) computeAuditPath(leafIndex, treeSize int64) [][]byte {
	// Simplified implementation - in practice this would compute
	// the actual Merkle tree audit path
	var path [][]byte
	
	// For demonstration, return sibling hashes at each level
	// This is NOT a correct Merkle audit path implementation
	currentIndex := leafIndex
	currentSize := treeSize
	
	for currentSize > 1 {
		// Find sibling
		var siblingIndex int64
		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
		} else {
			siblingIndex = currentIndex - 1
		}
		
		// Add sibling hash if it exists
		if siblingIndex < currentSize && siblingIndex < int64(len(m.leaves)) {
			path = append(path, m.leaves[siblingIndex].LeafHash)
		}
		
		// Move up one level
		currentIndex /= 2
		currentSize = (currentSize + 1) / 2
	}
	
	return path
}

// computeConsistencyPath computes the consistency proof path
func (m *MemoryTransparencyLog) computeConsistencyPath(fromSize, toSize int64) [][]byte {
	// Simplified implementation - in practice this would compute
	// the actual Merkle tree consistency path
	var path [][]byte
	
	// For demonstration, include intermediate root hashes
	// This is NOT a correct consistency proof implementation
	if fromSize < toSize && fromSize < int64(len(m.leaves)) {
		// Add a hash representing the state at fromSize
		hasher := sha256.New()
		for i := int64(0); i < fromSize; i++ {
			hasher.Write(m.leaves[i].LeafHash)
		}
		fromHash := hasher.Sum(nil)
		path = append(path, fromHash)
	}
	
	return path
}

// createSignedTreeHead creates and signs a tree head
func (m *MemoryTransparencyLog) createSignedTreeHead(treeSize int64, rootHash []byte, timestamp time.Time) (*SignedTreeHead, error) {
	sth := &SignedTreeHead{
		TreeSize:           treeSize,
		RootHash:           rootHash,
		Timestamp:          timestamp,
		TreeID:             m.config.TreeID,
		SignatureAlgorithm: m.config.SigningKey.Algorithm,
		KeyID:              m.config.SigningKey.KeyID,
	}
	
	// Create canonical representation for signing
	canonical := fmt.Sprintf("%d:%x:%d:%d", 
		sth.TreeID, 
		sth.RootHash, 
		sth.TreeSize, 
		sth.Timestamp.Unix())
	
	// Sign
	signature, err := m.signer.Sign([]byte(canonical))
	if err != nil {
		return nil, fmt.Errorf("failed to sign tree head: %w", err)
	}
	
	sth.Signature = signature
	return sth, nil
}