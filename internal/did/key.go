package did

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/ParichayaHQ/credence/internal/crypto"
)

// KeyMethodResolver implements the did:key method resolver
type KeyMethodResolver struct {
	keyManager KeyManager
}

// NewKeyMethodResolver creates a new did:key method resolver
func NewKeyMethodResolver(keyManager KeyManager) *KeyMethodResolver {
	if keyManager == nil {
		keyManager = NewDefaultKeyManager()
	}
	
	return &KeyMethodResolver{
		keyManager: keyManager,
	}
}

// Method returns the DID method this resolver handles
func (r *KeyMethodResolver) Method() string {
	return "key"
}

// Resolve resolves a did:key DID to a DID document
func (r *KeyMethodResolver) Resolve(ctx context.Context, did string, options *DIDResolutionOptions) (*DIDResolutionResult, error) {
	parsed, err := ParseDID(did)
	if err != nil {
		return r.createErrorResult(ErrorInvalidDID, "invalid DID syntax", err), nil
	}
	
	if parsed.Method != "key" {
		return r.createErrorResult(ErrorMethodNotSupported, "method not supported: "+parsed.Method, nil), nil
	}
	
	// Extract the public key from the method-specific identifier
	publicKey, keyType, err := r.decodePublicKey(parsed.Identifier)
	if err != nil {
		return r.createErrorResult(ErrorInvalidKey, "failed to decode public key", err), nil
	}
	
	// Create the DID document
	document, err := r.createDIDDocument(did, publicKey, keyType)
	if err != nil {
		return r.createErrorResult(ErrorInvalidDocument, "failed to create DID document", err), nil
	}
	
	return &DIDResolutionResult{
		DIDDocument: document,
		DIDResolutionMetadata: DIDResolutionMetadata{
			ContentType: "application/did+ld+json",
			ResolutionTime: time.Now().UTC().Format(time.RFC3339),
			ResolutionMethod: "did:key",
		},
		DIDDocumentMetadata: DIDDocumentMetadata{
			Created: &[]time.Time{time.Now().UTC()}[0],
		},
	}, nil
}

// Create creates a new did:key DID
func (r *KeyMethodResolver) Create(ctx context.Context, options *CreationOptions) (*CreationResult, error) {
	if options == nil {
		options = &CreationOptions{
			KeyType: KeyTypeEd25519,
			KeyPurposes: []KeyPurpose{PurposeAuthentication, PurposeAssertionMethod},
		}
	}
	
	// Generate a new key pair
	var privateKey interface{}
	var err error
	
	if options.PrivateKey != nil {
		privateKey = options.PrivateKey
	} else if options.Seed != nil {
		// Create key from seed
		if len(options.Seed) != ed25519.SeedSize {
			return nil, NewDIDError(ErrorInvalidKey, "invalid seed size")
		}
		key := ed25519.NewKeyFromSeed(options.Seed)
		privateKey = key
	} else {
		// Generate new key
		privateKey, err = r.keyManager.GenerateKey(options.KeyType)
		if err != nil {
			return nil, NewDIDErrorWithCause(ErrorInternalError, "failed to generate key", err)
		}
	}
	
	// Get the public key
	publicKey, err := r.keyManager.GetPublicKey(privateKey)
	if err != nil {
		return nil, NewDIDErrorWithCause(ErrorInternalError, "failed to get public key", err)
	}
	
	// Create the DID from the public key
	did, err := r.createDIDFromPublicKey(publicKey, options.KeyType)
	if err != nil {
		return nil, NewDIDErrorWithCause(ErrorInternalError, "failed to create DID from public key", err)
	}
	
	// Create the DID document
	document, err := r.createDIDDocument(did, publicKey, options.KeyType)
	if err != nil {
		return nil, NewDIDErrorWithCause(ErrorInternalError, "failed to create DID document", err)
	}
	
	// Add any additional services
	if options.Services != nil {
		document.Service = options.Services
	}
	
	// Set additional properties
	if options.AlsoKnownAs != nil {
		document.AlsoKnownAs = options.AlsoKnownAs
	}
	
	if options.Controllers != nil {
		document.Controller = options.Controllers
	}
	
	// Convert private key to JWK if requested
	var privateKeyJWK *JWK
	if privateKey != nil {
		privateKeyJWK, _ = r.keyManager.KeyToJWK(privateKey)
	}
	
	return &CreationResult{
		DID:           did,
		DIDDocument:   document,
		PrivateKey:    privateKey,
		PrivateKeyJWK: privateKeyJWK,
	}, nil
}

// Update is not supported for did:key method
func (r *KeyMethodResolver) Update(ctx context.Context, did string, document *DIDDocument, options *UpdateOptions) (*UpdateResult, error) {
	return nil, NewDIDError(ErrorMethodNotSupported, "did:key documents cannot be updated")
}

// Deactivate is not supported for did:key method
func (r *KeyMethodResolver) Deactivate(ctx context.Context, did string, options *DeactivationOptions) (*DeactivationResult, error) {
	return nil, NewDIDError(ErrorMethodNotSupported, "did:key documents cannot be deactivated")
}

// decodePublicKey decodes the public key from the method-specific identifier
func (r *KeyMethodResolver) decodePublicKey(identifier string) (interface{}, KeyType, error) {
	// For did:key, the identifier is a multibase-encoded public key
	// Format: z + base58btc-encoded multicodec public key
	
	if !strings.HasPrefix(identifier, "z") {
		return nil, "", NewDIDError(ErrorInvalidKey, "did:key identifier must start with 'z'")
	}
	
	// Remove the 'z' prefix
	encoded := identifier[1:]
	
	// Decode base58
	decoded, err := base58Decode(encoded)
	if err != nil {
		return nil, "", NewDIDErrorWithCause(ErrorInvalidKey, "failed to decode base58", err)
	}
	
	// Check multicodec prefix
	if len(decoded) < 2 {
		return nil, "", NewDIDError(ErrorInvalidKey, "decoded key too short")
	}
	
	// Ed25519 public keys have multicodec prefix 0xed 0x01
	if decoded[0] == 0xed && decoded[1] == 0x01 {
		if len(decoded) != 34 { // 2 bytes prefix + 32 bytes key
			return nil, "", NewDIDError(ErrorInvalidKey, "invalid Ed25519 key length")
		}
		
		publicKey := ed25519.PublicKey(decoded[2:])
		return publicKey, KeyTypeEd25519, nil
	}
	
	// Add support for other key types as needed
	// X25519: 0xec 0x01
	// secp256k1: 0xe7 0x01
	
	return nil, "", NewDIDError(ErrorInvalidKey, "unsupported key type")
}

// createDIDFromPublicKey creates a did:key DID from a public key
func (r *KeyMethodResolver) createDIDFromPublicKey(publicKey interface{}, keyType KeyType) (string, error) {
	var encoded string
	
	switch keyType {
	case KeyTypeEd25519:
		ed25519Key, ok := publicKey.(ed25519.PublicKey)
		if !ok {
			return "", NewDIDError(ErrorInvalidKey, "invalid Ed25519 public key")
		}
		
		// Add multicodec prefix (0xed 0x01 for Ed25519)
		prefixed := append([]byte{0xed, 0x01}, ed25519Key...)
		
		// Encode with base58 and add 'z' prefix
		encoded = "z" + base58Encode(prefixed)
		
	default:
		return "", NewDIDError(ErrorInvalidKey, "unsupported key type: "+string(keyType))
	}
	
	return "did:key:" + encoded, nil
}

// createDIDDocument creates a DID document for a did:key DID
func (r *KeyMethodResolver) createDIDDocument(did string, publicKey interface{}, keyType KeyType) (*DIDDocument, error) {
	now := time.Now().UTC()
	
	// Create the verification method
	verificationMethod, err := r.createVerificationMethod(did, publicKey, keyType)
	if err != nil {
		return nil, err
	}
	
	document := &DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID: did,
		VerificationMethod: []VerificationMethod{*verificationMethod},
		Authentication: []interface{}{verificationMethod.ID},
		AssertionMethod: []interface{}{verificationMethod.ID},
		CapabilityInvocation: []interface{}{verificationMethod.ID},
		CapabilityDelegation: []interface{}{verificationMethod.ID},
		Created: &now,
	}
	
	return document, nil
}

// createVerificationMethod creates a verification method for a public key
func (r *KeyMethodResolver) createVerificationMethod(did string, publicKey interface{}, keyType KeyType) (*VerificationMethod, error) {
	methodID := did + "#" + did[8:] // Remove "did:key:" prefix for fragment
	
	switch keyType {
	case KeyTypeEd25519:
		ed25519Key, ok := publicKey.(ed25519.PublicKey)
		if !ok {
			return nil, NewDIDError(ErrorInvalidKey, "invalid Ed25519 public key")
		}
		
		// Encode public key as multibase
		prefixed := append([]byte{0xed, 0x01}, ed25519Key...)
		multibaseKey := "z" + base58Encode(prefixed)
		
		return &VerificationMethod{
			ID:                 methodID,
			Type:               string(KeyTypeEd25519),
			Controller:         did,
			PublicKeyMultibase: &multibaseKey,
		}, nil
		
	default:
		return nil, NewDIDError(ErrorInvalidKey, "unsupported key type: "+string(keyType))
	}
}

// createErrorResult creates a DID resolution result with an error
func (r *KeyMethodResolver) createErrorResult(code, message string, cause error) *DIDResolutionResult {
	errorMessage := message
	if cause != nil {
		errorMessage += ": " + cause.Error()
	}
	
	return &DIDResolutionResult{
		DIDResolutionMetadata: DIDResolutionMetadata{
			Error:        code,
			ErrorMessage: errorMessage,
		},
		DIDDocumentMetadata: DIDDocumentMetadata{},
	}
}

// DefaultKeyManager provides a default implementation of KeyManager
type DefaultKeyManager struct{}

// NewDefaultKeyManager creates a new default key manager
func NewDefaultKeyManager() KeyManager {
	return &DefaultKeyManager{}
}

// GenerateKey generates a new key of the specified type
func (km *DefaultKeyManager) GenerateKey(keyType KeyType) (interface{}, error) {
	switch keyType {
	case KeyTypeEd25519:
		keyPair, err := crypto.NewEd25519KeyPair()
		if err != nil {
			return nil, err
		}
		return keyPair.PrivateKey, nil
		
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// GetPublicKey returns the public key for a private key
func (km *DefaultKeyManager) GetPublicKey(privateKey interface{}) (interface{}, error) {
	switch key := privateKey.(type) {
	case ed25519.PrivateKey:
		return key.Public().(ed25519.PublicKey), nil
		
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}
}

// Sign signs data with a private key
func (km *DefaultKeyManager) Sign(privateKey interface{}, data []byte) ([]byte, error) {
	switch key := privateKey.(type) {
	case ed25519.PrivateKey:
		return ed25519.Sign(key, data), nil
		
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}
}

// Verify verifies a signature with a public key
func (km *DefaultKeyManager) Verify(publicKey interface{}, data []byte, signature []byte) bool {
	switch key := publicKey.(type) {
	case ed25519.PublicKey:
		return ed25519.Verify(key, data, signature)
		
	default:
		return false
	}
}

// KeyToPEM converts a key to PEM format
func (km *DefaultKeyManager) KeyToPEM(key interface{}) ([]byte, error) {
	return nil, fmt.Errorf("PEM conversion not implemented")
}

// PEMToKey converts PEM data to a key
func (km *DefaultKeyManager) PEMToKey(pemData []byte) (interface{}, error) {
	return nil, fmt.Errorf("PEM conversion not implemented")
}

// KeyToJWK converts a key to JWK format
func (km *DefaultKeyManager) KeyToJWK(key interface{}) (*JWK, error) {
	switch k := key.(type) {
	case ed25519.PrivateKey:
		publicKey := k.Public().(ed25519.PublicKey)
		return &JWK{
			Kty: "OKP",
			Crv: "Ed25519",
			X:   base64.RawURLEncoding.EncodeToString(publicKey),
			D:   base64.RawURLEncoding.EncodeToString(k[:32]),
		}, nil
		
	case ed25519.PublicKey:
		return &JWK{
			Kty: "OKP",
			Crv: "Ed25519",
			X:   base64.RawURLEncoding.EncodeToString(k),
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported key type")
	}
}

// JWKToKey converts a JWK to a key
func (km *DefaultKeyManager) JWKToKey(jwk *JWK) (interface{}, error) {
	if jwk.Kty != "OKP" || jwk.Crv != "Ed25519" {
		return nil, fmt.Errorf("unsupported JWK type")
	}
	
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("invalid X value: %w", err)
	}
	
	if len(xBytes) != 32 {
		return nil, fmt.Errorf("invalid public key length")
	}
	
	if jwk.D != "" {
		// Private key
		dBytes, err := base64.RawURLEncoding.DecodeString(jwk.D)
		if err != nil {
			return nil, fmt.Errorf("invalid D value: %w", err)
		}
		
		if len(dBytes) != 32 {
			return nil, fmt.Errorf("invalid private key length")
		}
		
		return ed25519.NewKeyFromSeed(dBytes), nil
	}
	
	// Public key only
	return ed25519.PublicKey(xBytes), nil
}