package did

import (
	"encoding/json"
	"time"
)

// DID represents a Decentralized Identifier
type DID struct {
	Method     string `json:"method"`     // e.g., "key", "web", "ion"
	Identifier string `json:"identifier"` // method-specific identifier
	Fragment   string `json:"fragment,omitempty"`   // optional fragment
	Path       string `json:"path,omitempty"`       // optional path
	Query      string `json:"query,omitempty"`      // optional query
}

// String returns the DID as a string (did:method:identifier)
func (d *DID) String() string {
	result := "did:" + d.Method + ":" + d.Identifier
	
	if d.Path != "" {
		result += "/" + d.Path
	}
	
	if d.Query != "" {
		result += "?" + d.Query
	}
	
	if d.Fragment != "" {
		result += "#" + d.Fragment
	}
	
	return result
}

// DIDDocument represents a DID Document as per W3C DID specification
type DIDDocument struct {
	Context            []string                 `json:"@context"`
	ID                 string                   `json:"id"`
	AlsoKnownAs        []string                 `json:"alsoKnownAs,omitempty"`
	Controller         []string                 `json:"controller,omitempty"`
	VerificationMethod []VerificationMethod     `json:"verificationMethod,omitempty"`
	Authentication     []interface{}            `json:"authentication,omitempty"`     // can be strings or VerificationMethod objects
	AssertionMethod    []interface{}            `json:"assertionMethod,omitempty"`    // can be strings or VerificationMethod objects
	KeyAgreement       []interface{}            `json:"keyAgreement,omitempty"`       // can be strings or VerificationMethod objects
	CapabilityInvocation []interface{}          `json:"capabilityInvocation,omitempty"` // can be strings or VerificationMethod objects
	CapabilityDelegation []interface{}          `json:"capabilityDelegation,omitempty"` // can be strings or VerificationMethod objects
	Service            []Service                `json:"service,omitempty"`
	Created            *time.Time               `json:"created,omitempty"`
	Updated            *time.Time               `json:"updated,omitempty"`
	Deactivated        *bool                    `json:"deactivated,omitempty"`
}

// VerificationMethod represents a verification method in a DID document
type VerificationMethod struct {
	ID                 string      `json:"id"`
	Type               string      `json:"type"`               // e.g., "Ed25519VerificationKey2020"
	Controller         string      `json:"controller"`
	PublicKeyMultibase *string     `json:"publicKeyMultibase,omitempty"`
	PublicKeyJwk       *JWK        `json:"publicKeyJwk,omitempty"`
	PublicKeyBase58    *string     `json:"publicKeyBase58,omitempty"`     // deprecated but still used
	PublicKeyBase64    *string     `json:"publicKeyBase64,omitempty"`     // deprecated but still used
	PublicKeyHex       *string     `json:"publicKeyHex,omitempty"`        // deprecated but still used
	BlockchainAccountId *string    `json:"blockchainAccountId,omitempty"` // for blockchain methods
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`           // Key Type
	Crv string `json:"crv,omitempty"` // Curve (for EC keys)
	X   string `json:"x,omitempty"`   // X coordinate (for EC keys)
	Y   string `json:"y,omitempty"`   // Y coordinate (for EC keys)
	D   string `json:"d,omitempty"`   // Private key component
	Use string `json:"use,omitempty"` // Public Key Use
	Alg string `json:"alg,omitempty"` // Algorithm
	Kid string `json:"kid,omitempty"` // Key ID
}

// Service represents a service endpoint in a DID document
type Service struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	ServiceEndpoint interface{}            `json:"serviceEndpoint"` // can be string, object, or array
	RoutingKeys     []string               `json:"routingKeys,omitempty"`
	Accept          []string               `json:"accept,omitempty"`
	Properties      map[string]interface{} `json:"properties,omitempty"`
}

// DIDResolutionResult represents the result of DID resolution
type DIDResolutionResult struct {
	DIDDocument      *DIDDocument                   `json:"didDocument,omitempty"`
	DIDResolutionMetadata DIDResolutionMetadata     `json:"didResolutionMetadata"`
	DIDDocumentMetadata   DIDDocumentMetadata       `json:"didDocumentMetadata"`
}

// DIDResolutionMetadata contains metadata about the resolution process
type DIDResolutionMetadata struct {
	ContentType     string `json:"contentType,omitempty"`
	Error           string `json:"error,omitempty"`
	ErrorMessage    string `json:"errorMessage,omitempty"`
	ResolutionTime  string `json:"resolutionTime,omitempty"`
	ResolutionMethod string `json:"resolutionMethod,omitempty"`
}

// DIDDocumentMetadata contains metadata about the DID document
type DIDDocumentMetadata struct {
	Created         *time.Time             `json:"created,omitempty"`
	Updated         *time.Time             `json:"updated,omitempty"`
	Deactivated     *bool                  `json:"deactivated,omitempty"`
	NextUpdate      *time.Time             `json:"nextUpdate,omitempty"`
	VersionId       string                 `json:"versionId,omitempty"`
	NextVersionId   string                 `json:"nextVersionId,omitempty"`
	EquivalentId    []string               `json:"equivalentId,omitempty"`
	CanonicalId     string                 `json:"canonicalId,omitempty"`
	Properties      map[string]interface{} `json:"properties,omitempty"`
}

// DIDResolutionOptions contains options for DID resolution
type DIDResolutionOptions struct {
	Accept       string                 `json:"accept,omitempty"`
	VersionId    string                 `json:"versionId,omitempty"`
	VersionTime  string                 `json:"versionTime,omitempty"`
	Transform    string                 `json:"transform,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// KeyPurpose represents the different purposes a key can serve
type KeyPurpose string

const (
	PurposeAuthentication      KeyPurpose = "authentication"
	PurposeAssertionMethod     KeyPurpose = "assertionMethod"
	PurposeKeyAgreement        KeyPurpose = "keyAgreement"
	PurposeCapabilityInvocation KeyPurpose = "capabilityInvocation"
	PurposeCapabilityDelegation KeyPurpose = "capabilityDelegation"
)

// KeyType represents supported key types
type KeyType string

const (
	KeyTypeEd25519           KeyType = "Ed25519VerificationKey2020"
	KeyTypeX25519            KeyType = "X25519KeyAgreementKey2020"
	KeyTypeSecp256k1         KeyType = "EcdsaSecp256k1VerificationKey2019"
	KeyTypeSecp256r1         KeyType = "JsonWebKey2020"
	KeyTypeBLS12381G1        KeyType = "Bls12381G1Key2020"
	KeyTypeBLS12381G2        KeyType = "Bls12381G2Key2020"
)

// DIDError represents errors that can occur during DID operations
type DIDError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e *DIDError) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Message + " (cause: " + e.Cause.Error() + ")"
	}
	return e.Code + ": " + e.Message
}

// Common DID error codes
const (
	ErrorInvalidDID            = "invalidDid"
	ErrorMethodNotSupported    = "methodNotSupported"
	ErrorNotFound              = "notFound"
	ErrorInvalidDocument       = "invalidDidDocument"
	ErrorInvalidKey            = "invalidKey"
	ErrorRepresentationNotSupported = "representationNotSupported"
	ErrorInternalError         = "internalError"
)

// NewDIDError creates a new DID error
func NewDIDError(code, message string) *DIDError {
	return &DIDError{
		Code:    code,
		Message: message,
	}
}

// NewDIDErrorWithCause creates a new DID error with a cause
func NewDIDErrorWithCause(code, message string, cause error) *DIDError {
	return &DIDError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// MarshalJSON custom marshaling for DIDDocument to handle interface{} fields properly
func (d *DIDDocument) MarshalJSON() ([]byte, error) {
	type Alias DIDDocument
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	})
}

// UnmarshalJSON custom unmarshaling for DIDDocument
func (d *DIDDocument) UnmarshalJSON(data []byte) error {
	type Alias DIDDocument
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	return json.Unmarshal(data, &aux)
}