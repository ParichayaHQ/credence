package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Ed25519KeyPair represents an Ed25519 key pair
type Ed25519KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// Ed25519Signer provides Ed25519 signing functionality
type Ed25519Signer struct {
	keyPair *Ed25519KeyPair
}

// NewEd25519KeyPair generates a new Ed25519 key pair
func NewEd25519KeyPair() (*Ed25519KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	return &Ed25519KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// NewEd25519KeyPairFromSeed creates a key pair from a 32-byte seed
func NewEd25519KeyPairFromSeed(seed []byte) (*Ed25519KeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid seed size: expected %d bytes, got %d", ed25519.SeedSize, len(seed))
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &Ed25519KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// NewEd25519KeyPairFromPrivateKey creates a key pair from a private key
func NewEd25519KeyPairFromPrivateKey(privateKey ed25519.PrivateKey) (*Ed25519KeyPair, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &Ed25519KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// PublicKeyBase64 returns the public key as base64
func (kp *Ed25519KeyPair) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKey)
}

// PrivateKeyBase64 returns the private key as base64
func (kp *Ed25519KeyPair) PrivateKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PrivateKey)
}

// NewEd25519Signer creates a new Ed25519 signer
func NewEd25519Signer(keyPair *Ed25519KeyPair) *Ed25519Signer {
	return &Ed25519Signer{
		keyPair: keyPair,
	}
}

// NewEd25519SignerFromSeed creates a signer from a seed
func NewEd25519SignerFromSeed(seed []byte) (*Ed25519Signer, error) {
	keyPair, err := NewEd25519KeyPairFromSeed(seed)
	if err != nil {
		return nil, err
	}
	return NewEd25519Signer(keyPair), nil
}

// Sign signs data using Ed25519
func (s *Ed25519Signer) Sign(data []byte) ([]byte, error) {
	if s.keyPair == nil || s.keyPair.PrivateKey == nil {
		return nil, ErrNoPrivateKey
	}

	signature := ed25519.Sign(s.keyPair.PrivateKey, data)
	return signature, nil
}

// SignBase64 signs data and returns base64-encoded signature
func (s *Ed25519Signer) SignBase64(data []byte) (string, error) {
	signature, err := s.Sign(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// PublicKey returns the public key
func (s *Ed25519Signer) PublicKey() ed25519.PublicKey {
	if s.keyPair == nil {
		return nil
	}
	return s.keyPair.PublicKey
}

// PublicKeyBase64 returns the public key as base64
func (s *Ed25519Signer) PublicKeyBase64() string {
	if s.keyPair == nil {
		return ""
	}
	return s.keyPair.PublicKeyBase64()
}

// Ed25519Verifier provides Ed25519 signature verification
type Ed25519Verifier struct{}

// NewEd25519Verifier creates a new verifier
func NewEd25519Verifier() *Ed25519Verifier {
	return &Ed25519Verifier{}
}

// Verify verifies an Ed25519 signature
func (v *Ed25519Verifier) Verify(publicKey ed25519.PublicKey, data, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}
	if len(signature) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(publicKey, data, signature)
}

// VerifyBase64 verifies a base64-encoded signature
func (v *Ed25519Verifier) VerifyBase64(publicKeyB64, signatureB64 string, data []byte) (bool, error) {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return false, fmt.Errorf("invalid public key base64: %w", err)
	}

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, fmt.Errorf("invalid signature base64: %w", err)
	}

	return v.Verify(publicKey, data, signature), nil
}

// GenerateSecureRandom generates cryptographically secure random bytes
func GenerateSecureRandom(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateNonce generates a secure nonce for events
func GenerateNonce() (string, error) {
	nonce, err := GenerateSecureRandom(12) // 12 bytes = 96 bits
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(nonce), nil
}