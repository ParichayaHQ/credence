package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd25519KeyPair(t *testing.T) {
	t.Run("GenerateNewKeyPair", func(t *testing.T) {
		keyPair, err := NewEd25519KeyPair()
		require.NoError(t, err)
		assert.NotNil(t, keyPair)
		assert.Len(t, keyPair.PublicKey, ed25519.PublicKeySize)
		assert.Len(t, keyPair.PrivateKey, ed25519.PrivateKeySize)
	})

	t.Run("KeyPairFromSeed", func(t *testing.T) {
		seed := make([]byte, ed25519.SeedSize)
		for i := range seed {
			seed[i] = byte(i)
		}

		keyPair, err := NewEd25519KeyPairFromSeed(seed)
		require.NoError(t, err)
		assert.NotNil(t, keyPair)

		// Same seed should produce same key pair
		keyPair2, err := NewEd25519KeyPairFromSeed(seed)
		require.NoError(t, err)
		assert.Equal(t, keyPair.PublicKey, keyPair2.PublicKey)
		assert.Equal(t, keyPair.PrivateKey, keyPair2.PrivateKey)
	})

	t.Run("InvalidSeedSize", func(t *testing.T) {
		invalidSeed := []byte("too short")
		_, err := NewEd25519KeyPairFromSeed(invalidSeed)
		assert.Error(t, err)
	})

	t.Run("Base64Encoding", func(t *testing.T) {
		keyPair, err := NewEd25519KeyPair()
		require.NoError(t, err)

		pubB64 := keyPair.PublicKeyBase64()
		privB64 := keyPair.PrivateKeyBase64()

		assert.NotEmpty(t, pubB64)
		assert.NotEmpty(t, privB64)

		// Should be valid base64
		_, err = base64.StdEncoding.DecodeString(pubB64)
		assert.NoError(t, err)

		_, err = base64.StdEncoding.DecodeString(privB64)
		assert.NoError(t, err)
	})
}

func TestEd25519Signer(t *testing.T) {
	t.Run("SignAndVerify", func(t *testing.T) {
		keyPair, err := NewEd25519KeyPair()
		require.NoError(t, err)

		signer := NewEd25519Signer(keyPair)
		verifier := NewEd25519Verifier()

		testData := []byte("Hello, World!")

		signature, err := signer.Sign(testData)
		require.NoError(t, err)
		assert.Len(t, signature, ed25519.SignatureSize)

		// Verify signature
		isValid := verifier.Verify(keyPair.PublicKey, testData, signature)
		assert.True(t, isValid)

		// Wrong data should fail verification
		wrongData := []byte("Wrong data")
		isValid = verifier.Verify(keyPair.PublicKey, wrongData, signature)
		assert.False(t, isValid)
	})

	t.Run("SignBase64", func(t *testing.T) {
		keyPair, err := NewEd25519KeyPair()
		require.NoError(t, err)

		signer := NewEd25519Signer(keyPair)
		verifier := NewEd25519Verifier()

		testData := []byte("Test data for base64 signing")

		sigB64, err := signer.SignBase64(testData)
		require.NoError(t, err)
		assert.NotEmpty(t, sigB64)

		// Verify base64 signature
		pubB64 := signer.PublicKeyBase64()
		isValid, err := verifier.VerifyBase64(pubB64, sigB64, testData)
		require.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("SignerFromSeed", func(t *testing.T) {
		seed := make([]byte, ed25519.SeedSize)
		for i := range seed {
			seed[i] = byte(42) // Fixed seed for deterministic testing
		}

		signer, err := NewEd25519SignerFromSeed(seed)
		require.NoError(t, err)

		testData := []byte("Deterministic test")
		signature1, err := signer.Sign(testData)
		require.NoError(t, err)

		// Create another signer with same seed
		signer2, err := NewEd25519SignerFromSeed(seed)
		require.NoError(t, err)

		signature2, err := signer2.Sign(testData)
		require.NoError(t, err)

		// Signatures should be identical (deterministic)
		assert.Equal(t, signature1, signature2)
	})
}

func TestEd25519Verifier(t *testing.T) {
	t.Run("VerifyBase64Invalid", func(t *testing.T) {
		verifier := NewEd25519Verifier()

		// Invalid base64 public key
		_, err := verifier.VerifyBase64("invalid-base64", "dGVzdA==", []byte("data"))
		assert.Error(t, err)

		// Invalid base64 signature
		keyPair, err := NewEd25519KeyPair()
		require.NoError(t, err)
		pubB64 := keyPair.PublicKeyBase64()

		_, err = verifier.VerifyBase64(pubB64, "invalid-base64", []byte("data"))
		assert.Error(t, err)
	})

	t.Run("VerifyWrongKeySize", func(t *testing.T) {
		verifier := NewEd25519Verifier()

		wrongSizeKey := []byte("wrong size")
		signature := make([]byte, ed25519.SignatureSize)
		testData := []byte("test")

		isValid := verifier.Verify(wrongSizeKey, testData, signature)
		assert.False(t, isValid)
	})

	t.Run("VerifyWrongSignatureSize", func(t *testing.T) {
		verifier := NewEd25519Verifier()

		keyPair, err := NewEd25519KeyPair()
		require.NoError(t, err)

		wrongSizeSignature := []byte("wrong size")
		testData := []byte("test")

		isValid := verifier.Verify(keyPair.PublicKey, testData, wrongSizeSignature)
		assert.False(t, isValid)
	})
}

func TestSecureRandom(t *testing.T) {
	t.Run("GenerateSecureRandom", func(t *testing.T) {
		size := 32
		random1, err := GenerateSecureRandom(size)
		require.NoError(t, err)
		assert.Len(t, random1, size)

		random2, err := GenerateSecureRandom(size)
		require.NoError(t, err)
		assert.Len(t, random2, size)

		// Should be different (extremely high probability)
		assert.NotEqual(t, random1, random2)
	})

	t.Run("GenerateNonce", func(t *testing.T) {
		nonce1, err := GenerateNonce()
		require.NoError(t, err)
		assert.NotEmpty(t, nonce1)

		nonce2, err := GenerateNonce()
		require.NoError(t, err)
		assert.NotEmpty(t, nonce2)

		// Should be different
		assert.NotEqual(t, nonce1, nonce2)

		// Should be valid base64
		decoded, err := base64.StdEncoding.DecodeString(nonce1)
		require.NoError(t, err)
		assert.Len(t, decoded, 12) // 12 bytes as specified
	})
}