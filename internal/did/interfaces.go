package did

import (
	"context"
)

// Resolver defines the interface for DID resolution
type Resolver interface {
	// Resolve resolves a DID to a DID document
	Resolve(ctx context.Context, did string, options *DIDResolutionOptions) (*DIDResolutionResult, error)
	
	// SupportsMethod checks if the resolver supports a specific DID method
	SupportsMethod(method string) bool
	
	// SupportedMethods returns a list of supported DID methods
	SupportedMethods() []string
}

// MethodResolver defines the interface for method-specific DID resolution
type MethodResolver interface {
	// Method returns the DID method this resolver handles
	Method() string
	
	// Resolve resolves a DID of the specific method to a DID document
	Resolve(ctx context.Context, did string, options *DIDResolutionOptions) (*DIDResolutionResult, error)
	
	// Create creates a new DID of this method (if supported)
	Create(ctx context.Context, options *CreationOptions) (*CreationResult, error)
	
	// Update updates a DID document (if supported)
	Update(ctx context.Context, did string, document *DIDDocument, options *UpdateOptions) (*UpdateResult, error)
	
	// Deactivate deactivates a DID (if supported)
	Deactivate(ctx context.Context, did string, options *DeactivationOptions) (*DeactivationResult, error)
}

// CreationOptions contains options for creating a new DID
type CreationOptions struct {
	KeyType            KeyType                `json:"keyType,omitempty"`
	KeyPurposes        []KeyPurpose           `json:"keyPurposes,omitempty"`
	Services           []Service              `json:"services,omitempty"`
	AlsoKnownAs        []string               `json:"alsoKnownAs,omitempty"`
	Controllers        []string               `json:"controllers,omitempty"`
	Properties         map[string]interface{} `json:"properties,omitempty"`
	Seed              []byte                 `json:"seed,omitempty"`
	PrivateKey        interface{}            `json:"privateKey,omitempty"`
	PublicKey         interface{}            `json:"publicKey,omitempty"`
}

// CreationResult contains the result of DID creation
type CreationResult struct {
	DID              string         `json:"did"`
	DIDDocument      *DIDDocument   `json:"didDocument"`
	PrivateKeyJWK    *JWK          `json:"privateKeyJwk,omitempty"`
	PrivateKey       interface{}    `json:"privateKey,omitempty"`
	MethodMetadata   interface{}    `json:"methodMetadata,omitempty"`
	CreationMetadata interface{}    `json:"creationMetadata,omitempty"`
}

// UpdateOptions contains options for updating a DID document
type UpdateOptions struct {
	VersionId        string                 `json:"versionId,omitempty"`
	UpdateKey        interface{}            `json:"updateKey,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

// UpdateResult contains the result of DID update
type UpdateResult struct {
	DIDDocument      *DIDDocument   `json:"didDocument"`
	VersionId        string         `json:"versionId,omitempty"`
	UpdateMetadata   interface{}    `json:"updateMetadata,omitempty"`
}

// DeactivationOptions contains options for deactivating a DID
type DeactivationOptions struct {
	DeactivationKey  interface{}            `json:"deactivationKey,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

// DeactivationResult contains the result of DID deactivation
type DeactivationResult struct {
	Deactivated         bool        `json:"deactivated"`
	DeactivationMetadata interface{} `json:"deactivationMetadata,omitempty"`
}

// KeyManager defines the interface for managing cryptographic keys
type KeyManager interface {
	// GenerateKey generates a new key of the specified type
	GenerateKey(keyType KeyType) (interface{}, error)
	
	// GetPublicKey returns the public key for a private key
	GetPublicKey(privateKey interface{}) (interface{}, error)
	
	// Sign signs data with a private key
	Sign(privateKey interface{}, data []byte) ([]byte, error)
	
	// Verify verifies a signature with a public key
	Verify(publicKey interface{}, data []byte, signature []byte) bool
	
	// KeyToPEM converts a key to PEM format
	KeyToPEM(key interface{}) ([]byte, error)
	
	// PEMToKey converts PEM data to a key
	PEMToKey(pemData []byte) (interface{}, error)
	
	// KeyToJWK converts a key to JWK format
	KeyToJWK(key interface{}) (*JWK, error)
	
	// JWKToKey converts a JWK to a key
	JWKToKey(jwk *JWK) (interface{}, error)
}

// DocumentValidator defines the interface for validating DID documents
type DocumentValidator interface {
	// Validate validates a DID document
	Validate(ctx context.Context, document *DIDDocument) error
	
	// ValidateVerificationMethod validates a verification method
	ValidateVerificationMethod(method *VerificationMethod) error
	
	// ValidateService validates a service
	ValidateService(service *Service) error
	
	// ValidateContext validates the @context field
	ValidateContext(context []string) error
}

// DocumentStore defines the interface for storing and retrieving DID documents
type DocumentStore interface {
	// Store stores a DID document
	Store(ctx context.Context, did string, document *DIDDocument) error
	
	// Retrieve retrieves a DID document
	Retrieve(ctx context.Context, did string) (*DIDDocument, error)
	
	// Exists checks if a DID document exists
	Exists(ctx context.Context, did string) (bool, error)
	
	// Delete deletes a DID document
	Delete(ctx context.Context, did string) error
	
	// List lists all stored DIDs (paginated)
	List(ctx context.Context, offset, limit int) ([]string, error)
}

// CacheManager defines the interface for caching DID resolution results
type CacheManager interface {
	// Get retrieves a cached DID document
	Get(ctx context.Context, did string) (*DIDDocument, error)
	
	// Set stores a DID document in cache
	Set(ctx context.Context, did string, document *DIDDocument, ttl int64) error
	
	// Invalidate removes a DID document from cache
	Invalidate(ctx context.Context, did string) error
	
	// Clear clears all cached entries
	Clear(ctx context.Context) error
}

// Registry defines the interface for a DID method registry
type Registry interface {
	// RegisterMethod registers a method resolver
	RegisterMethod(method string, resolver MethodResolver) error
	
	// UnregisterMethod unregisters a method resolver
	UnregisterMethod(method string) error
	
	// GetResolver gets the resolver for a specific method
	GetResolver(method string) (MethodResolver, error)
	
	// ListMethods lists all registered methods
	ListMethods() []string
	
	// SupportsMethod checks if a method is supported
	SupportsMethod(method string) bool
}

// MultiResolver defines the interface for resolving multiple DID methods
type MultiResolver interface {
	Resolver
	Registry
	
	// ResolveWithMethod resolves a DID using a specific method resolver
	ResolveWithMethod(ctx context.Context, did string, method string, options *DIDResolutionOptions) (*DIDResolutionResult, error)
	
	// SetDefaultOptions sets default resolution options
	SetDefaultOptions(options *DIDResolutionOptions)
	
	// GetDefaultOptions gets default resolution options
	GetDefaultOptions() *DIDResolutionOptions
}

// VerificationRelationship represents the different verification relationships
type VerificationRelationship string

const (
	Authentication      VerificationRelationship = "authentication"
	AssertionMethod     VerificationRelationship = "assertionMethod"
	KeyAgreement        VerificationRelationship = "keyAgreement"
	CapabilityInvocation VerificationRelationship = "capabilityInvocation"
	CapabilityDelegation VerificationRelationship = "capabilityDelegation"
)

// DocumentHelper defines utility methods for working with DID documents
type DocumentHelper interface {
	// GetVerificationMethod retrieves a verification method by ID
	GetVerificationMethod(document *DIDDocument, methodID string) (*VerificationMethod, error)
	
	// GetVerificationMethodsForPurpose gets all verification methods for a specific purpose
	GetVerificationMethodsForPurpose(document *DIDDocument, purpose VerificationRelationship) ([]*VerificationMethod, error)
	
	// GetService retrieves a service by ID
	GetService(document *DIDDocument, serviceID string) (*Service, error)
	
	// GetServicesByType gets all services of a specific type
	GetServicesByType(document *DIDDocument, serviceType string) ([]*Service, error)
	
	// AddVerificationMethod adds a verification method to the document
	AddVerificationMethod(document *DIDDocument, method *VerificationMethod, purposes []VerificationRelationship) error
	
	// RemoveVerificationMethod removes a verification method from the document
	RemoveVerificationMethod(document *DIDDocument, methodID string) error
	
	// AddService adds a service to the document
	AddService(document *DIDDocument, service *Service) error
	
	// RemoveService removes a service from the document
	RemoveService(document *DIDDocument, serviceID string) error
	
	// IsDeactivated checks if a DID document is deactivated
	IsDeactivated(document *DIDDocument) bool
}