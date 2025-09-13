package cid

import (
	"crypto/sha256"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// CIDGenerator provides content identifier generation functionality
type CIDGenerator struct{}

// NewCIDGenerator creates a new CID generator
func NewCIDGenerator() *CIDGenerator {
	return &CIDGenerator{}
}

// GenerateFromBytes generates a CID from raw bytes using SHA-256
func (g *CIDGenerator) GenerateFromBytes(data []byte) (cid.Cid, error) {
	if len(data) == 0 {
		return cid.Undef, fmt.Errorf("cannot generate CID from empty data")
	}

	// Create SHA-256 hash
	hash := sha256.Sum256(data)

	// Create multihash from the SHA-256 hash
	mh, err := multihash.Encode(hash[:], multihash.SHA2_256)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create multihash: %w", err)
	}

	// Create CID v1 with dag-json codec
	c := cid.NewCidV1(cid.DagJSON, mh)
	
	return c, nil
}

// GenerateFromJSON generates a CID from canonical JSON bytes
func (g *CIDGenerator) GenerateFromJSON(jsonData []byte) (cid.Cid, error) {
	return g.GenerateFromBytes(jsonData)
}

// GenerateFromString generates a CID from a string
func (g *CIDGenerator) GenerateFromString(data string) (cid.Cid, error) {
	return g.GenerateFromBytes([]byte(data))
}

// ValidateCID validates that a CID string is valid
func (g *CIDGenerator) ValidateCID(cidStr string) error {
	_, err := cid.Parse(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}
	return nil
}

// ParseCID parses a CID string into a CID object
func (g *CIDGenerator) ParseCID(cidStr string) (cid.Cid, error) {
	c, err := cid.Parse(cidStr)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to parse CID: %w", err)
	}
	return c, nil
}

// ExtractHash extracts the raw hash bytes from a CID
func (g *CIDGenerator) ExtractHash(c cid.Cid) ([]byte, error) {
	mh := c.Hash()
	decoded, err := multihash.Decode(mh)
	if err != nil {
		return nil, fmt.Errorf("failed to decode multihash: %w", err)
	}
	return decoded.Digest, nil
}

// IsSHA256CID checks if a CID uses SHA-256 hashing
func (g *CIDGenerator) IsSHA256CID(c cid.Cid) bool {
	mh := c.Hash()
	decoded, err := multihash.Decode(mh)
	if err != nil {
		return false
	}
	return decoded.Code == multihash.SHA2_256
}

// CompareCIDs compares two CIDs for equality
func (g *CIDGenerator) CompareCIDs(c1, c2 cid.Cid) bool {
	return c1.Equals(c2)
}

// ContentAddressedStorage interface for storing content by CID
type ContentAddressedStorage interface {
	// Put stores data and returns its CID
	Put(data []byte) (cid.Cid, error)
	
	// Get retrieves data by CID
	Get(c cid.Cid) ([]byte, error)
	
	// Has checks if content exists for a CID
	Has(c cid.Cid) (bool, error)
	
	// Delete removes content by CID
	Delete(c cid.Cid) error
	
	// Size returns the size of content for a CID
	Size(c cid.Cid) (int64, error)
}

// ContentRouter interface for content routing and discovery
type ContentRouter interface {
	// Provide announces that this node can provide content for a CID
	Provide(c cid.Cid) error
	
	// FindProviders finds peers that can provide content for a CID
	FindProviders(c cid.Cid) ([]string, error)
	
	// GetClosestPeers finds the closest peers to a CID
	GetClosestPeers(c cid.Cid) ([]string, error)
}

// Helper functions for common CID operations

// GenerateCIDFromCanonicalJSON is a convenience function for generating CIDs from canonical JSON
func GenerateCIDFromCanonicalJSON(jsonData []byte) (cid.Cid, error) {
	generator := NewCIDGenerator()
	return generator.GenerateFromJSON(jsonData)
}

// ParseCIDString is a convenience function for parsing CID strings
func ParseCIDString(cidStr string) (cid.Cid, error) {
	generator := NewCIDGenerator()
	return generator.ParseCID(cidStr)
}

// ValidateCIDString is a convenience function for validating CID strings
func ValidateCIDString(cidStr string) error {
	generator := NewCIDGenerator()
	return generator.ValidateCID(cidStr)
}

// CIDToString converts a CID to its string representation
func CIDToString(c cid.Cid) string {
	return c.String()
}

// CIDToBytes converts a CID to its byte representation
func CIDToBytes(c cid.Cid) []byte {
	return c.Bytes()
}

// BytesToCID converts bytes back to a CID
func BytesToCID(data []byte) (cid.Cid, error) {
	return cid.Cast(data)
}