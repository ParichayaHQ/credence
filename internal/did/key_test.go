package did

import (
	"context"
	"crypto/ed25519"
	"testing"
)

func TestKeyMethodResolver_Method(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	if resolver.Method() != "key" {
		t.Errorf("expected method 'key', got %s", resolver.Method())
	}
}

func TestKeyMethodResolver_Create(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	ctx := context.Background()

	// Test with default options
	result, err := resolver.Create(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DID == "" {
		t.Error("expected DID to be set")
	}

	if !IsValidDID(result.DID) {
		t.Errorf("created DID is not valid: %s", result.DID)
	}

	if result.DIDDocument == nil {
		t.Error("expected DID document to be set")
	}

	if result.PrivateKey == nil {
		t.Error("expected private key to be set")
	}

	if result.PrivateKeyJWK == nil {
		t.Error("expected private key JWK to be set")
	}

	// Verify the DID starts with did:key:z
	if len(result.DID) < 8 || result.DID[:8] != "did:key:" || result.DID[8] != 'z' {
		t.Errorf("expected did:key:z... format, got %s", result.DID)
	}
}

func TestKeyMethodResolver_CreateWithSeed(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	ctx := context.Background()

	// Create a known seed for reproducible tests
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}

	options := &CreationOptions{
		KeyType: KeyTypeEd25519,
		Seed:    seed,
	}

	result1, err := resolver.Create(ctx, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result2, err := resolver.Create(ctx, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create the same DID from the same seed
	if result1.DID != result2.DID {
		t.Errorf("expected same DID from same seed, got %s and %s", result1.DID, result2.DID)
	}
}

func TestKeyMethodResolver_Resolve(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	ctx := context.Background()

	// First create a DID
	createResult, err := resolver.Create(ctx, nil)
	if err != nil {
		t.Fatalf("failed to create DID: %v", err)
	}

	// Now resolve it
	resolveResult, err := resolver.Resolve(ctx, createResult.DID, nil)
	if err != nil {
		t.Fatalf("failed to resolve DID: %v", err)
	}

	if resolveResult.DIDDocument == nil {
		t.Error("expected DID document in resolve result")
	}

	if resolveResult.DIDResolutionMetadata.Error != "" {
		t.Errorf("unexpected resolution error: %s", resolveResult.DIDResolutionMetadata.Error)
	}

	// Verify document structure
	doc := resolveResult.DIDDocument
	if doc.ID != createResult.DID {
		t.Errorf("expected document ID %s, got %s", createResult.DID, doc.ID)
	}

	if len(doc.VerificationMethod) != 1 {
		t.Errorf("expected 1 verification method, got %d", len(doc.VerificationMethod))
	}

	if len(doc.Authentication) != 1 {
		t.Errorf("expected 1 authentication method, got %d", len(doc.Authentication))
	}
}

func TestKeyMethodResolver_ResolveInvalidDID(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	ctx := context.Background()

	tests := []struct {
		name string
		did  string
	}{
		{"empty DID", ""},
		{"invalid syntax", "not-a-did"},
		{"wrong method", "did:web:example.com"},
		{"invalid key", "did:key:invalid"},
		{"malformed multibase", "did:key:a123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(ctx, tt.did, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.DIDResolutionMetadata.Error == "" {
				t.Error("expected resolution error")
			}

			if result.DIDDocument != nil {
				t.Error("expected no DID document on error")
			}
		})
	}
}

func TestKeyMethodResolver_UpdateNotSupported(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	ctx := context.Background()

	result, err := resolver.Update(ctx, "did:key:z123", nil, nil)
	if err == nil {
		t.Error("expected error for unsupported operation")
	}

	if result != nil {
		t.Error("expected nil result for unsupported operation")
	}

	didErr, ok := err.(*DIDError)
	if !ok {
		t.Error("expected DIDError")
	} else if didErr.Code != ErrorMethodNotSupported {
		t.Errorf("expected ErrorMethodNotSupported, got %s", didErr.Code)
	}
}

func TestKeyMethodResolver_DeactivateNotSupported(t *testing.T) {
	resolver := NewKeyMethodResolver(nil)
	ctx := context.Background()

	result, err := resolver.Deactivate(ctx, "did:key:z123", nil)
	if err == nil {
		t.Error("expected error for unsupported operation")
	}

	if result != nil {
		t.Error("expected nil result for unsupported operation")
	}

	didErr, ok := err.(*DIDError)
	if !ok {
		t.Error("expected DIDError")
	} else if didErr.Code != ErrorMethodNotSupported {
		t.Errorf("expected ErrorMethodNotSupported, got %s", didErr.Code)
	}
}

func TestDefaultKeyManager_GenerateKey(t *testing.T) {
	km := NewDefaultKeyManager()

	key, err := km.GenerateKey(KeyTypeEd25519)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key == nil {
		t.Error("expected key to be generated")
	}

	// Verify it's an Ed25519 private key
	_, ok := key.(ed25519.PrivateKey)
	if !ok {
		t.Error("expected Ed25519 private key")
	}
}

func TestDefaultKeyManager_GetPublicKey(t *testing.T) {
	km := NewDefaultKeyManager()

	privateKey, err := km.GenerateKey(KeyTypeEd25519)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	publicKey, err := km.GetPublicKey(privateKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if publicKey == nil {
		t.Error("expected public key")
	}

	// Verify it's an Ed25519 public key
	_, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		t.Error("expected Ed25519 public key")
	}
}

func TestDefaultKeyManager_SignAndVerify(t *testing.T) {
	km := NewDefaultKeyManager()

	privateKey, err := km.GenerateKey(KeyTypeEd25519)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	publicKey, err := km.GetPublicKey(privateKey)
	if err != nil {
		t.Fatalf("failed to get public key: %v", err)
	}

	testData := []byte("test message")

	signature, err := km.Sign(privateKey, testData)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	if len(signature) == 0 {
		t.Error("expected signature")
	}

	// Verify signature
	if !km.Verify(publicKey, testData, signature) {
		t.Error("signature verification failed")
	}

	// Verify with wrong data should fail
	if km.Verify(publicKey, []byte("wrong data"), signature) {
		t.Error("signature verification should have failed with wrong data")
	}
}

func TestDefaultKeyManager_KeyToJWK(t *testing.T) {
	km := NewDefaultKeyManager()

	privateKey, err := km.GenerateKey(KeyTypeEd25519)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	jwk, err := km.KeyToJWK(privateKey)
	if err != nil {
		t.Fatalf("failed to convert key to JWK: %v", err)
	}

	if jwk.Kty != "OKP" {
		t.Errorf("expected kty OKP, got %s", jwk.Kty)
	}

	if jwk.Crv != "Ed25519" {
		t.Errorf("expected crv Ed25519, got %s", jwk.Crv)
	}

	if jwk.X == "" {
		t.Error("expected X parameter")
	}

	if jwk.D == "" {
		t.Error("expected D parameter for private key")
	}
}

func TestDefaultKeyManager_JWKToKey(t *testing.T) {
	km := NewDefaultKeyManager()

	privateKey, err := km.GenerateKey(KeyTypeEd25519)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	jwk, err := km.KeyToJWK(privateKey)
	if err != nil {
		t.Fatalf("failed to convert key to JWK: %v", err)
	}

	reconstructedKey, err := km.JWKToKey(jwk)
	if err != nil {
		t.Fatalf("failed to convert JWK to key: %v", err)
	}

	// Sign with original key
	testData := []byte("test message")
	signature1, err := km.Sign(privateKey, testData)
	if err != nil {
		t.Fatalf("failed to sign with original key: %v", err)
	}

	// Sign with reconstructed key
	signature2, err := km.Sign(reconstructedKey, testData)
	if err != nil {
		t.Fatalf("failed to sign with reconstructed key: %v", err)
	}

	// Both signatures should be valid
	publicKey, _ := km.GetPublicKey(privateKey)
	if !km.Verify(publicKey, testData, signature1) {
		t.Error("original signature verification failed")
	}

	if !km.Verify(publicKey, testData, signature2) {
		t.Error("reconstructed signature verification failed")
	}
}