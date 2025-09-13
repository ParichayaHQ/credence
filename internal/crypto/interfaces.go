package crypto

import "crypto/ed25519"

// Signer interface for signing operations
type Signer interface {
	// Sign signs the given data and returns the signature
	Sign(data []byte) ([]byte, error)
	
	// SignBase64 signs data and returns base64-encoded signature
	SignBase64(data []byte) (string, error)
	
	// PublicKey returns the public key associated with this signer
	PublicKey() ed25519.PublicKey
	
	// PublicKeyBase64 returns the public key as base64
	PublicKeyBase64() string
}

// Verifier interface for signature verification
type Verifier interface {
	// Verify verifies a signature against data using the given public key
	Verify(publicKey ed25519.PublicKey, data, signature []byte) bool
	
	// VerifyBase64 verifies a base64-encoded signature
	VerifyBase64(publicKeyB64, signatureB64 string, data []byte) (bool, error)
}

// KeyManager interface for key management operations
type KeyManager interface {
	// GenerateKeyPair generates a new key pair
	GenerateKeyPair() (*Ed25519KeyPair, error)
	
	// ImportKeyPair imports a key pair from seed
	ImportKeyPair(seed []byte) (*Ed25519KeyPair, error)
	
	// ExportSeed exports the seed for a key pair
	ExportSeed(keyPair *Ed25519KeyPair) ([]byte, error)
}

// ThresholdSigner interface for BLS threshold signatures
type ThresholdSigner interface {
	// SignPartial creates a partial signature for threshold signing
	SignPartial(data []byte, signerIndex int) ([]byte, error)
	
	// AggregateSignatures combines partial signatures into a threshold signature
	AggregateSignatures(partialSigs map[int][]byte, threshold int) ([]byte, error)
	
	// VerifyThreshold verifies a threshold signature
	VerifyThreshold(publicKeys [][]byte, data, signature []byte, threshold int) bool
}

// VRFProver interface for VRF (Verifiable Random Function) operations
type VRFProver interface {
	// Prove generates a VRF proof for the given input
	Prove(input []byte) (output []byte, proof []byte, err error)
	
	// ProveBase64 generates VRF proof with base64 encoding
	ProveBase64(input []byte) (output string, proof string, err error)
}

// VRFVerifier interface for VRF proof verification
type VRFVerifier interface {
	// Verify verifies a VRF proof
	Verify(publicKey, input, output, proof []byte) bool
	
	// VerifyBase64 verifies base64-encoded VRF proof
	VerifyBase64(publicKeyB64, inputB64, outputB64, proofB64 string) (bool, error)
}

// RandomnessProvider interface for secure random number generation
type RandomnessProvider interface {
	// GenerateRandom generates cryptographically secure random bytes
	GenerateRandom(size int) ([]byte, error)
	
	// GenerateNonce generates a secure nonce
	GenerateNonce() (string, error)
	
	// GenerateSeed generates a seed for key derivation
	GenerateSeed() ([]byte, error)
}