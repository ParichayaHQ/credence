package vc

import (
	"time"
)

// VerifiableCredential represents a W3C Verifiable Credential
type VerifiableCredential struct {
	Context           []string               `json:"@context"`
	ID                string                 `json:"id,omitempty"`
	Type              []string               `json:"type"`
	Issuer            interface{}            `json:"issuer"`
	IssuanceDate      string                 `json:"issuanceDate"`
	ExpirationDate    string                 `json:"expirationDate,omitempty"`
	CredentialSubject interface{}            `json:"credentialSubject"`
	Proof             interface{}            `json:"proof,omitempty"`
	CredentialStatus  *CredentialStatus      `json:"credentialStatus,omitempty"`
	CredentialSchema  []CredentialSchema     `json:"credentialSchema,omitempty"`
	RefreshService    *RefreshService        `json:"refreshService,omitempty"`
	TermsOfUse        []TermsOfUse           `json:"termsOfUse,omitempty"`
	Evidence          []Evidence             `json:"evidence,omitempty"`
	
	// Additional fields for JWT representation
	JWT string `json:"-"` // The JWT representation when applicable
}

// VerifiablePresentation represents a W3C Verifiable Presentation
type VerifiablePresentation struct {
	Context              []string                 `json:"@context"`
	ID                   string                   `json:"id,omitempty"`
	Type                 []string                 `json:"type"`
	Holder               string                   `json:"holder,omitempty"`
	VerifiableCredential []interface{}            `json:"verifiableCredential,omitempty"`
	Proof                interface{}              `json:"proof,omitempty"`
	
	// Additional fields for JWT representation
	JWT string `json:"-"` // The JWT representation when applicable
}

// Issuer represents the issuer of a credential
type Issuer struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// CredentialSubject represents the subject of a credential
type CredentialSubject struct {
	ID string `json:"id,omitempty"`
	// Additional properties are stored as arbitrary JSON
	Claims map[string]interface{} `json:"-"`
}

// CredentialStatus represents the status information for a credential
type CredentialStatus struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// CredentialSchema represents schema information for a credential
type CredentialSchema struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// RefreshService represents refresh service information
type RefreshService struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// TermsOfUse represents terms of use for a credential
type TermsOfUse struct {
	Type string `json:"type"`
}

// Evidence represents evidence supporting a credential
type Evidence struct {
	Type string `json:"type"`
}

// JWTCredential represents the JWT format of a Verifiable Credential
type JWTCredential struct {
	// JWT Header
	Header JWTHeader `json:"-"`
	
	// JWT Claims
	Issuer         string                 `json:"iss"`
	Subject        string                 `json:"sub,omitempty"`
	Audience       interface{}            `json:"aud,omitempty"`
	ExpirationTime *int64                 `json:"exp,omitempty"`
	NotBefore      *int64                 `json:"nbf,omitempty"`
	IssuedAt       *int64                 `json:"iat,omitempty"`
	JWTID          string                 `json:"jti,omitempty"`
	
	// VC-specific claims
	VC *VerifiableCredential `json:"vc"`
	
	// Raw JWT string
	Token string `json:"-"`
}

// JWTPresentation represents the JWT format of a Verifiable Presentation
type JWTPresentation struct {
	// JWT Header
	Header JWTHeader `json:"-"`
	
	// JWT Claims
	Issuer         string                  `json:"iss"`
	Subject        string                  `json:"sub,omitempty"`
	Audience       interface{}             `json:"aud,omitempty"`
	ExpirationTime *int64                  `json:"exp,omitempty"`
	NotBefore      *int64                  `json:"nbf,omitempty"`
	IssuedAt       *int64                  `json:"iat,omitempty"`
	JWTID          string                  `json:"jti,omitempty"`
	Nonce          string                  `json:"nonce,omitempty"`
	
	// VP-specific claims
	VP *VerifiablePresentation `json:"vp"`
	
	// Raw JWT string
	Token string `json:"-"`
}

// JWTHeader represents a JWT header
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// SD-JWT specific types

// SDJWTCredential represents a Selective Disclosure JWT Credential
type SDJWTCredential struct {
	// The SD-JWT consists of a JWT and optional disclosures
	JWT         string      `json:"-"`
	Disclosures []Disclosure `json:"-"`
	KeyBinding  *KeyBinding `json:"-"`
	
	// Parsed claims
	Header JWTHeader              `json:"-"`
	Claims map[string]interface{} `json:"-"`
}

// Disclosure represents a selective disclosure
type Disclosure struct {
	Salt  string      `json:"-"`
	Claim string      `json:"-"`
	Value interface{} `json:"-"`
	
	// Base64-encoded disclosure
	Encoded string `json:"-"`
}

// KeyBinding represents a key binding JWT for SD-JWT
type KeyBinding struct {
	JWT string `json:"-"`
	
	// Parsed claims
	Header JWTHeader              `json:"-"`
	Claims map[string]interface{} `json:"-"`
}

// Verification interfaces

// CredentialVerifier verifies verifiable credentials
type CredentialVerifier interface {
	// VerifyCredential verifies a verifiable credential
	VerifyCredential(credential *VerifiableCredential, options *VerificationOptions) (*VerificationResult, error)
	
	// VerifyPresentation verifies a verifiable presentation
	VerifyPresentation(presentation *VerifiablePresentation, options *VerificationOptions) (*VerificationResult, error)
	
	// VerifyJWTCredential verifies a JWT-format credential
	VerifyJWTCredential(token string, options *VerificationOptions) (*VerificationResult, error)
	
	// VerifyJWTPresentation verifies a JWT-format presentation
	VerifyJWTPresentation(token string, options *VerificationOptions) (*VerificationResult, error)
	
	// VerifySDJWT verifies a Selective Disclosure JWT
	VerifySDJWT(sdjwt string, options *VerificationOptions) (*VerificationResult, error)
}

// CredentialIssuer issues verifiable credentials
type CredentialIssuer interface {
	// IssueCredential issues a verifiable credential
	IssueCredential(template *CredentialTemplate, options *IssuanceOptions) (*VerifiableCredential, error)
	
	// IssueJWTCredential issues a JWT-format credential
	IssueJWTCredential(template *CredentialTemplate, options *IssuanceOptions) (string, error)
	
	// IssueSDJWT issues a Selective Disclosure JWT
	IssueSDJWT(template *CredentialTemplate, options *IssuanceOptions) (string, error)
}

// PresentationCreator creates verifiable presentations
type PresentationCreator interface {
	// CreatePresentation creates a verifiable presentation
	CreatePresentation(credentials []interface{}, options *PresentationOptions) (*VerifiablePresentation, error)
	
	// CreateJWTPresentation creates a JWT-format presentation
	CreateJWTPresentation(credentials []interface{}, options *PresentationOptions) (string, error)
}

// Options and results

// VerificationOptions contains options for credential verification
type VerificationOptions struct {
	// Challenge for presentation verification
	Challenge string `json:"challenge,omitempty"`
	
	// Domain for presentation verification
	Domain string `json:"domain,omitempty"`
	
	// Expected audience
	Audience string `json:"audience,omitempty"`
	
	// Current time for time-based validation
	Now *time.Time `json:"now,omitempty"`
	
	// Whether to check credential status
	CheckStatus bool `json:"checkStatus,omitempty"`
	
	// Whether to validate credential schema
	ValidateSchema bool `json:"validateSchema,omitempty"`
	
	// Trust framework to use for policy-based verification
	TrustFramework string `json:"trustFramework,omitempty"`
	
	// Additional verification parameters
	Params map[string]interface{} `json:"params,omitempty"`
}

// VerificationResult contains the result of credential verification
type VerificationResult struct {
	// Whether verification was successful
	Verified bool `json:"verified"`
	
	// Error message if verification failed
	Error string `json:"error,omitempty"`
	
	// Verification details
	Details map[string]interface{} `json:"details,omitempty"`
	
	// Extracted credential/presentation
	Credential   *VerifiableCredential   `json:"credential,omitempty"`
	Presentation *VerifiablePresentation `json:"presentation,omitempty"`
	
	// JWT-specific results
	JWTCredential   *JWTCredential   `json:"jwtCredential,omitempty"`
	JWTPresentation *JWTPresentation `json:"jwtPresentation,omitempty"`
	
	// SD-JWT specific results
	SDJWTCredential *SDJWTCredential `json:"sdjwtCredential,omitempty"`
}

// CredentialTemplate contains the template for credential issuance
type CredentialTemplate struct {
	Context           []string               `json:"@context"`
	Type              []string               `json:"type"`
	Issuer            interface{}            `json:"issuer"`
	CredentialSubject interface{}            `json:"credentialSubject"`
	ExpirationDate    string                 `json:"expirationDate,omitempty"`
	CredentialStatus  *CredentialStatus      `json:"credentialStatus,omitempty"`
	CredentialSchema  []CredentialSchema     `json:"credentialSchema,omitempty"`
	
	// SD-JWT specific fields
	SelectivelyDisclosable []string `json:"selectivelyDisclosable,omitempty"`
}

// IssuanceOptions contains options for credential issuance
type IssuanceOptions struct {
	// Signing key identifier
	KeyID string `json:"keyId"`
	
	// Signing algorithm
	Algorithm string `json:"algorithm"`
	
	// Additional JWT claims
	AdditionalClaims map[string]interface{} `json:"additionalClaims,omitempty"`
	
	// For SD-JWT: salt generation function
	SaltGenerator func() string `json:"-"`
	
	// For SD-JWT: whether to require key binding
	RequireKeyBinding bool `json:"requireKeyBinding,omitempty"`
}

// PresentationOptions contains options for presentation creation
type PresentationOptions struct {
	// Holder DID
	Holder string `json:"holder"`
	
	// Challenge from verifier
	Challenge string `json:"challenge,omitempty"`
	
	// Domain of verifier
	Domain string `json:"domain,omitempty"`
	
	// Signing key identifier
	KeyID string `json:"keyId"`
	
	// Signing algorithm
	Algorithm string `json:"algorithm"`
	
	// Additional JWT claims
	AdditionalClaims map[string]interface{} `json:"additionalClaims,omitempty"`
	
	// For SD-JWT: selective disclosures to include
	Disclosures []string `json:"disclosures,omitempty"`
}

// Error types

// VCError represents a verifiable credential error
type VCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *VCError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Common error codes
const (
	ErrorInvalidCredential   = "invalid_credential"
	ErrorInvalidPresentation = "invalid_presentation"
	ErrorInvalidJWT          = "invalid_jwt"
	ErrorInvalidSignature    = "invalid_signature"
	ErrorExpiredCredential   = "expired_credential"
	ErrorRevokedCredential   = "revoked_credential"
	ErrorInvalidIssuer       = "invalid_issuer"
	ErrorInvalidHolder       = "invalid_holder"
	ErrorInvalidProof        = "invalid_proof"
	ErrorMissingProof        = "missing_proof"
	ErrorInvalidSchema       = "invalid_schema"
	ErrorInvalidContext      = "invalid_context"
)

// NewVCError creates a new VC error
func NewVCError(code, message string) *VCError {
	return &VCError{
		Code:    code,
		Message: message,
	}
}

// NewVCErrorWithDetails creates a new VC error with details
func NewVCErrorWithDetails(code, message, details string) *VCError {
	return &VCError{
		Code:    code,
		Message: message,
		Details: details,
	}
}