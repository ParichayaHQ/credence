package consensus

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"
)

// VRFOutput represents the output of a VRF function
type VRFOutput struct {
	Proof  []byte `json:"proof"`
	Output []byte `json:"output"`
}

// DefaultVRFProvider implements VRFProvider interface
// NOTE: This is a simplified VRF implementation for demonstration
// In production, use a proper VRF library like go-vrf or libsodium VRF
type DefaultVRFProvider struct {
	// For this implementation, we'll use Ed25519 keys for VRF
	// In production, use dedicated VRF keys
}

// NewDefaultVRFProvider creates a new VRF provider
func NewDefaultVRFProvider() *DefaultVRFProvider {
	return &DefaultVRFProvider{}
}

// GenerateProof generates a VRF proof and output
func (v *DefaultVRFProvider) GenerateProof(privateKey []byte, seed []byte) ([]byte, []byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}
	
	if len(seed) == 0 {
		return nil, nil, fmt.Errorf("seed cannot be empty")
	}
	
	// Create the message to sign (seed + VRF domain separator)
	message := v.createVRFMessage(seed)
	
	// Sign the message (this is our VRF proof)
	proof := ed25519.Sign(privateKey, message)
	
	// Derive deterministic output from proof
	output := v.deriveOutput(proof, seed)
	
	return proof, output, nil
}

// VerifyProof verifies a VRF proof and output
func (v *DefaultVRFProvider) VerifyProof(publicKey []byte, seed []byte, proof []byte, output []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(publicKey))
	}
	
	if len(seed) == 0 {
		return fmt.Errorf("seed cannot be empty")
	}
	
	if len(proof) != ed25519.SignatureSize {
		return fmt.Errorf("invalid proof size: expected %d, got %d", ed25519.SignatureSize, len(proof))
	}
	
	// Create the message that should have been signed
	message := v.createVRFMessage(seed)
	
	// Verify the signature (proof)
	if !ed25519.Verify(publicKey, message, proof) {
		return fmt.Errorf("invalid VRF proof")
	}
	
	// Verify the output is correctly derived from proof
	expectedOutput := v.deriveOutput(proof, seed)
	if !bytesEqual(output, expectedOutput) {
		return fmt.Errorf("invalid VRF output")
	}
	
	return nil
}

// GetCurrentSeed gets the current VRF seed (placeholder implementation)
func (v *DefaultVRFProvider) GetCurrentSeed(ctx context.Context) ([]byte, error) {
	// In production, this would fetch from drand or another beacon
	// For now, return a deterministic seed based on current time
	
	// Use current epoch (simplified)
	seed := make([]byte, 32)
	hasher := sha256.New()
	hasher.Write([]byte("CREDENCE_VRF_SEED_V1"))
	// Add some time-based entropy (rounded to hours for stability)
	// In production, use actual beacon rounds
	epoch := getCurrentEpoch()
	hasher.Write(int64ToBytes(epoch))
	copy(seed, hasher.Sum(nil))
	
	return seed, nil
}

// CommitteeSelector implements committee selection using VRF
type CommitteeSelector struct {
	vrfProvider VRFProvider
	config      *CheckpointorConfig
}

// NewCommitteeSelector creates a new committee selector
func NewCommitteeSelector(vrfProvider VRFProvider, config *CheckpointorConfig) *CommitteeSelector {
	return &CommitteeSelector{
		vrfProvider: vrfProvider,
		config:      config,
	}
}

// SelectCommittee selects a committee using VRF-based sortition
func (cs *CommitteeSelector) SelectCommittee(ctx context.Context, vrfSeed []byte, candidates []CommitteeCandidate) (*Committee, error) {
	if len(candidates) < cs.config.CommitteeSize {
		return nil, fmt.Errorf("insufficient candidates: need %d, got %d", cs.config.CommitteeSize, len(candidates))
	}
	
	// Calculate VRF outputs for all candidates
	candidateOutputs := make([]CandidateWithVRF, 0, len(candidates))
	
	for _, candidate := range candidates {
		// In production, candidates would provide their own VRF proofs
		// For testing, we'll simulate the VRF output from their public key
		
		// Simulate VRF output deterministically from public key and seed
		output := cs.simulateVRFOutput(candidate.PublicKey, vrfSeed)
		proof := []byte("simulated-proof") // Placeholder proof
		
		candidateWithVRF := CandidateWithVRF{
			Candidate: candidate,
			VRFOutput: output,
			VRFProof:  proof,
		}
		
		candidateOutputs = append(candidateOutputs, candidateWithVRF)
	}
	
	if len(candidateOutputs) < cs.config.CommitteeSize {
		return nil, fmt.Errorf("insufficient valid VRF candidates: need %d, got %d", cs.config.CommitteeSize, len(candidateOutputs))
	}
	
	// Sort candidates by VRF output (lowest first)
	sort.Slice(candidateOutputs, func(i, j int) bool {
		return compareBytes(candidateOutputs[i].VRFOutput, candidateOutputs[j].VRFOutput) < 0
	})
	
	// Select the top committee size candidates
	selectedCandidates := candidateOutputs[:cs.config.CommitteeSize]
	
	// Create committee
	members := make([]string, len(selectedCandidates))
	for i, candidate := range selectedCandidates {
		members[i] = candidate.Candidate.DID
	}
	
	committee := &Committee{
		Members:   members,
		Threshold: cs.config.SignatureThreshold,
		Epoch:     getCurrentEpoch(),
		VRFSeed:   vrfSeed,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(cs.config.RotationPeriod),
	}
	
	return committee, nil
}

// CandidateWithVRF represents a candidate with VRF computation
type CandidateWithVRF struct {
	Candidate CommitteeCandidate
	VRFOutput []byte
	VRFProof  []byte
}

// simulateVRFOutput creates a deterministic VRF-like output for testing
func (cs *CommitteeSelector) simulateVRFOutput(publicKey ed25519.PublicKey, seed []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte("SIMULATE_VRF_OUTPUT_V1"))
	hasher.Write(publicKey)
	hasher.Write(seed)
	return hasher.Sum(nil)
}

// Helper methods

// createVRFMessage creates the message to be signed for VRF
func (v *DefaultVRFProvider) createVRFMessage(seed []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte("CREDENCE_VRF_V1"))
	hasher.Write(seed)
	return hasher.Sum(nil)
}

// deriveOutput derives deterministic output from VRF proof
func (v *DefaultVRFProvider) deriveOutput(proof []byte, seed []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte("CREDENCE_VRF_OUTPUT_V1"))
	hasher.Write(proof)
	hasher.Write(seed)
	return hasher.Sum(nil)
}

// getCurrentEpoch returns current epoch (simplified)
func getCurrentEpoch() int64 {
	// In production, use proper epoch calculation
	return time.Now().Unix() / 3600 // Hour-based epochs
}

// compareBytes compares two byte slices lexicographically
func compareBytes(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	
	return 0
}

// GenerateVRFKeyPair generates Ed25519 key pair for VRF (testing)
func GenerateVRFKeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	return privateKey, publicKey, nil
}

// VRFCommitteeManager combines VRF provider with committee management
type VRFCommitteeManager struct {
	selector *CommitteeSelector
	store    CheckpointStore
}

// NewVRFCommitteeManager creates a new VRF-based committee manager
func NewVRFCommitteeManager(selector *CommitteeSelector, store CheckpointStore) *VRFCommitteeManager {
	return &VRFCommitteeManager{
		selector: selector,
		store:    store,
	}
}

// GetCurrentCommittee implements CommitteeManager interface
func (m *VRFCommitteeManager) GetCurrentCommittee(ctx context.Context) (*Committee, error) {
	return m.store.GetCurrentCommittee(ctx)
}

// GetNextCommittee returns the next committee (during rotation)
func (m *VRFCommitteeManager) GetNextCommittee(ctx context.Context) (*Committee, error) {
	// For now, return nil as next committee is not pre-computed
	return nil, fmt.Errorf("next committee not implemented")
}

// SelectCommittee selects a new committee using VRF
func (m *VRFCommitteeManager) SelectCommittee(ctx context.Context, vrfSeed []byte, candidates []CommitteeCandidate) (*Committee, error) {
	return m.selector.SelectCommittee(ctx, vrfSeed, candidates)
}

// RotateCommittee rotates to a new committee
func (m *VRFCommitteeManager) RotateCommittee(ctx context.Context) error {
	// This would typically:
	// 1. Get current candidates from registry
	// 2. Generate new VRF seed
	// 3. Select new committee
	// 4. Store new committee
	
	// For now, return not implemented
	return fmt.Errorf("committee rotation not fully implemented")
}

// IsMember checks if a DID is in the current committee
func (m *VRFCommitteeManager) IsMember(ctx context.Context, memberDID string) (bool, error) {
	committee, err := m.GetCurrentCommittee(ctx)
	if err != nil {
		return false, err
	}
	
	for _, member := range committee.Members {
		if member == memberDID {
			return true, nil
		}
	}
	
	return false, nil
}