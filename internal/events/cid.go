package events

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
)

// GenerateCID generates a simple CID-like identifier from data
// This is a simplified version - in production you'd use the IPFS CID library
func GenerateCID(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("cannot generate CID for empty data")
	}
	
	// Create SHA-256 hash
	hash := sha256.Sum256(data)
	
	// Encode as base32 (similar to IPFS CIDv1)
	// Add a simple prefix to indicate this is our format
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash[:])
	
	// Add prefix to make it look like a CID
	return "bafy" + encoded, nil
}

// GenerateCIDFromJSON generates a CID from JSON data
func GenerateCIDFromJSON(jsonData []byte) (string, error) {
	return GenerateCID(jsonData)
}