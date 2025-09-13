package didvc

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
)

// DIDMethod represents the supported DID methods
type DIDMethod string

const (
	DIDMethodKey DIDMethod = "key"
	DIDMethodWeb DIDMethod = "web"
	DIDMethodION DIDMethod = "ion"
)

// DID represents a decentralized identifier
type DID struct {
	Method     DIDMethod `json:"method"`
	Identifier string    `json:"identifier"`
	Fragment   string    `json:"fragment,omitempty"`
	Path       string    `json:"path,omitempty"`
	Query      string    `json:"query,omitempty"`
}

// DIDDocument represents a DID document according to W3C spec
type DIDDocument struct {
	Context              []string                 `json:"@context"`
	ID                   string                   `json:"id"`
	VerificationMethod   []VerificationMethod     `json:"verificationMethod"`
	Authentication       []interface{}            `json:"authentication,omitempty"`
	AssertionMethod      []interface{}            `json:"assertionMethod,omitempty"`
	KeyAgreement         []interface{}            `json:"keyAgreement,omitempty"`
	CapabilityInvocation []interface{}            `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []interface{}            `json:"capabilityDelegation,omitempty"`
	Service              []Service                `json:"service,omitempty"`
}

// VerificationMethod represents a verification method in a DID document
type VerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
	PublicKeyJwk       *JWK   `json:"publicKeyJwk,omitempty"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

// Service represents a service endpoint in a DID document
type Service struct {
	ID              string      `json:"id"`
	Type            string      `json:"type"`
	ServiceEndpoint interface{} `json:"serviceEndpoint"`
}

// DIDResolver interface for resolving DIDs to DID documents
type DIDResolver interface {
	// Resolve resolves a DID to its DID document
	Resolve(did string) (*DIDDocument, error)
	
	// SupportsMethod checks if the resolver supports a specific DID method
	SupportsMethod(method DIDMethod) bool
}

// DIDKeyResolver resolves did:key DIDs
type DIDKeyResolver struct{}

// NewDIDKeyResolver creates a new did:key resolver
func NewDIDKeyResolver() *DIDKeyResolver {
	return &DIDKeyResolver{}
}

// Resolve resolves a did:key DID to a DID document
func (r *DIDKeyResolver) Resolve(didStr string) (*DIDDocument, error) {
	did, err := ParseDID(didStr)
	if err != nil {
		return nil, err
	}

	if did.Method != DIDMethodKey {
		return nil, fmt.Errorf("unsupported DID method: %s", did.Method)
	}

	// Extract public key from the identifier (remove 'z' multibase prefix)
	if !strings.HasPrefix(did.Identifier, "z") {
		return nil, fmt.Errorf("invalid did:key identifier format")
	}

	// For simplicity, assume Ed25519 keys (in production, would need multicodec parsing)
	keyBytes, err := base64.RawURLEncoding.DecodeString(did.Identifier[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid Ed25519 public key size")
	}

	// Create verification method
	vmID := fmt.Sprintf("%s#%s", didStr, did.Identifier)
	verificationMethod := VerificationMethod{
		ID:                 vmID,
		Type:               "Ed25519VerificationKey2020",
		Controller:         didStr,
		PublicKeyMultibase: did.Identifier,
	}

	// Create DID document
	doc := &DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID:                 didStr,
		VerificationMethod: []VerificationMethod{verificationMethod},
		Authentication:     []interface{}{vmID},
		AssertionMethod:    []interface{}{vmID},
		CapabilityInvocation: []interface{}{vmID},
		CapabilityDelegation: []interface{}{vmID},
	}

	return doc, nil
}

// SupportsMethod checks if this resolver supports the given DID method
func (r *DIDKeyResolver) SupportsMethod(method DIDMethod) bool {
	return method == DIDMethodKey
}

// ParseDID parses a DID string into a DID struct
func ParseDID(didStr string) (*DID, error) {
	if didStr == "" {
		return nil, fmt.Errorf("empty DID string")
	}

	// Basic DID format: did:method:identifier
	parts := strings.Split(didStr, ":")
	if len(parts) < 3 || parts[0] != "did" {
		return nil, fmt.Errorf("invalid DID format: %s", didStr)
	}

	method := DIDMethod(parts[1])
	identifier := strings.Join(parts[2:], ":")

	// Handle fragments, paths, queries (simplified)
	var fragment, path, query string
	
	// Check for fragment (#)
	if idx := strings.Index(identifier, "#"); idx != -1 {
		fragment = identifier[idx+1:]
		identifier = identifier[:idx]
	}

	// Check for query (?)
	if idx := strings.Index(identifier, "?"); idx != -1 {
		query = identifier[idx+1:]
		identifier = identifier[:idx]
	}

	// Check for path (/)
	if idx := strings.Index(identifier, "/"); idx != -1 {
		path = identifier[idx:]
		identifier = identifier[:idx]
	}

	return &DID{
		Method:     method,
		Identifier: identifier,
		Fragment:   fragment,
		Path:       path,
		Query:      query,
	}, nil
}

// String returns the string representation of the DID
func (d *DID) String() string {
	didStr := fmt.Sprintf("did:%s:%s", d.Method, d.Identifier)
	
	if d.Path != "" {
		didStr += d.Path
	}
	
	if d.Query != "" {
		didStr += "?" + d.Query
	}
	
	if d.Fragment != "" {
		didStr += "#" + d.Fragment
	}
	
	return didStr
}

// IsValid checks if the DID is valid
func (d *DID) IsValid() bool {
	if d == nil {
		return false
	}
	
	return d.Method != "" && d.Identifier != ""
}

// CreateDIDKey creates a did:key DID from an Ed25519 public key
func CreateDIDKey(publicKey ed25519.PublicKey) (*DID, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid Ed25519 public key size")
	}

	// Encode public key with multibase (z prefix for base58btc)
	identifier := "z" + base64.RawURLEncoding.EncodeToString(publicKey)

	return &DID{
		Method:     DIDMethodKey,
		Identifier: identifier,
	}, nil
}

// ExtractPublicKeyFromDIDKey extracts Ed25519 public key from did:key DID
func ExtractPublicKeyFromDIDKey(did *DID) (ed25519.PublicKey, error) {
	if did.Method != DIDMethodKey {
		return nil, fmt.Errorf("not a did:key DID")
	}

	if !strings.HasPrefix(did.Identifier, "z") {
		return nil, fmt.Errorf("invalid did:key identifier format")
	}

	keyBytes, err := base64.RawURLEncoding.DecodeString(did.Identifier[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid Ed25519 public key size")
	}

	return keyBytes, nil
}