package wallet

import (
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
)

// Wallet represents a digital wallet for managing keys, DIDs, and credentials
type Wallet interface {
	// Key Management
	GenerateKey(keyType did.KeyType) (*KeyPair, error)
	ImportKey(privateKey interface{}, keyType did.KeyType) (*KeyPair, error)
	GetKey(keyID string) (*KeyPair, error)
	ListKeys() ([]*KeyPair, error)
	DeleteKey(keyID string) error

	// DID Management
	CreateDID(keyID string, method string) (*DIDRecord, error)
	ImportDID(didDocument *did.DIDDocument) (*DIDRecord, error)
	GetDID(didStr string) (*DIDRecord, error)
	ListDIDs() ([]*DIDRecord, error)
	ResolveDID(didStr string) (*did.DIDDocument, error)

	// Credential Management
	StoreCredential(credential *vc.VerifiableCredential) (*CredentialRecord, error)
	GetCredential(credentialID string) (*CredentialRecord, error)
	ListCredentials(filter *CredentialFilter) ([]*CredentialRecord, error)
	DeleteCredential(credentialID string) error

	// Presentation Management
	CreatePresentation(credentialIDs []string, options *PresentationOptions) (*vc.VerifiablePresentation, error)
	StorePresentation(presentation *vc.VerifiablePresentation) (*PresentationRecord, error)
	GetPresentation(presentationID string) (*PresentationRecord, error)
	ListPresentations() ([]*PresentationRecord, error)

	// Wallet Operations
	Sign(keyID string, data []byte) ([]byte, error)
	Verify(keyID string, data []byte, signature []byte) (bool, error)
	Export(password string) ([]byte, error)
	Import(data []byte, password string) error
	Lock(password string) error
	Unlock(password string) error
	IsLocked() bool
}

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	ID          string           `json:"id"`
	KeyType     did.KeyType      `json:"keyType"`
	PublicKey   interface{}      `json:"-"` // Not serialized for security
	PrivateKey  interface{}      `json:"-"` // Not serialized for security
	PublicKeyJWK *did.JWK        `json:"publicKeyJwk,omitempty"`
	PrivateKeyJWK *did.JWK       `json:"privateKeyJwk,omitempty"`
	Algorithm   string           `json:"algorithm"`
	Created     time.Time        `json:"created"`
	Usage       []KeyUsage       `json:"usage"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DIDRecord represents a DID managed by the wallet
type DIDRecord struct {
	DID         string             `json:"did"`
	Document    *did.DIDDocument   `json:"document"`
	Method      string             `json:"method"`
	KeyID       string             `json:"keyId"` // Reference to controlling key
	Created     time.Time          `json:"created"`
	Updated     time.Time          `json:"updated"`
	Status      DIDStatus          `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CredentialRecord represents a verifiable credential in the wallet
type CredentialRecord struct {
	ID           string                    `json:"id"`
	Credential   *vc.VerifiableCredential  `json:"credential"`
	CredentialJWT string                   `json:"credentialJwt,omitempty"`
	Issuer       string                    `json:"issuer"`
	Subject      string                    `json:"subject"`
	IssuanceDate time.Time                 `json:"issuanceDate"`
	ExpirationDate *time.Time              `json:"expirationDate,omitempty"`
	Status       CredentialStatus          `json:"status"`
	Type         []string                  `json:"type"`
	Tags         []string                  `json:"tags,omitempty"`
	Created      time.Time                 `json:"created"`
	Metadata     map[string]interface{}    `json:"metadata,omitempty"`
}

// PresentationRecord represents a verifiable presentation in the wallet
type PresentationRecord struct {
	ID           string                      `json:"id"`
	Presentation *vc.VerifiablePresentation  `json:"presentation"`
	PresentationJWT string                   `json:"presentationJwt,omitempty"`
	Holder       string                      `json:"holder"`
	Verifier     string                      `json:"verifier,omitempty"`
	Created      time.Time                   `json:"created"`
	Challenge    string                      `json:"challenge,omitempty"`
	Domain       string                      `json:"domain,omitempty"`
	Purpose      string                      `json:"purpose,omitempty"`
	Credentials  []string                    `json:"credentials"` // IDs of included credentials
	Metadata     map[string]interface{}      `json:"metadata,omitempty"`
}

// KeyUsage represents the intended usage of a key
type KeyUsage string

const (
	KeyUsageAuthentication    KeyUsage = "authentication"
	KeyUsageAssertionMethod   KeyUsage = "assertionMethod"
	KeyUsageKeyAgreement      KeyUsage = "keyAgreement"
	KeyUsageCapabilityInvocation KeyUsage = "capabilityInvocation"
	KeyUsageCapabilityDelegation KeyUsage = "capabilityDelegation"
	KeyUsageSigning           KeyUsage = "signing"
	KeyUsageEncryption        KeyUsage = "encryption"
)

// DIDStatus represents the status of a DID
type DIDStatus string

const (
	DIDStatusActive      DIDStatus = "active"
	DIDStatusDeactivated DIDStatus = "deactivated"
	DIDStatusRevoked     DIDStatus = "revoked"
	DIDStatusSuspended   DIDStatus = "suspended"
)

// CredentialStatus represents the status of a credential
type CredentialStatus string

const (
	CredentialStatusValid     CredentialStatus = "valid"
	CredentialStatusRevoked   CredentialStatus = "revoked"
	CredentialStatusSuspended CredentialStatus = "suspended"
	CredentialStatusExpired   CredentialStatus = "expired"
)

// CredentialFilter provides filtering options for credentials
type CredentialFilter struct {
	Issuer         string                 `json:"issuer,omitempty"`
	Subject        string                 `json:"subject,omitempty"`
	Type           []string               `json:"type,omitempty"`
	Status         CredentialStatus       `json:"status,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	IssuedAfter    *time.Time             `json:"issuedAfter,omitempty"`
	IssuedBefore   *time.Time             `json:"issuedBefore,omitempty"`
	ExpiresAfter   *time.Time             `json:"expiresAfter,omitempty"`
	ExpiresBefore  *time.Time             `json:"expiresBefore,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Limit          int                    `json:"limit,omitempty"`
	Offset         int                    `json:"offset,omitempty"`
}

// PresentationOptions provides options for creating presentations
type PresentationOptions struct {
	Holder      string                 `json:"holder"`
	Verifier    string                 `json:"verifier,omitempty"`
	Challenge   string                 `json:"challenge,omitempty"`
	Domain      string                 `json:"domain,omitempty"`
	Purpose     string                 `json:"purpose,omitempty"`
	KeyID       string                 `json:"keyId"`
	Algorithm   string                 `json:"algorithm"`
	SelectiveDisclosure map[string][]string `json:"selectiveDisclosure,omitempty"` // credentialID -> fields
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WalletStorage provides storage interface for wallet data
type WalletStorage interface {
	// Key Storage
	StoreKey(keyPair *KeyPair) error
	GetKey(keyID string) (*KeyPair, error)
	ListKeys() ([]*KeyPair, error)
	DeleteKey(keyID string) error

	// DID Storage
	StoreDID(record *DIDRecord) error
	GetDID(did string) (*DIDRecord, error)
	ListDIDs() ([]*DIDRecord, error)
	DeleteDID(did string) error

	// Credential Storage
	StoreCredential(record *CredentialRecord) error
	GetCredential(credentialID string) (*CredentialRecord, error)
	ListCredentials(filter *CredentialFilter) ([]*CredentialRecord, error)
	DeleteCredential(credentialID string) error

	// Presentation Storage
	StorePresentation(record *PresentationRecord) error
	GetPresentation(presentationID string) (*PresentationRecord, error)
	ListPresentations() ([]*PresentationRecord, error)
	DeletePresentation(presentationID string) error

	// Wallet Metadata
	SetMetadata(key string, value interface{}) error
	GetMetadata(key string) (interface{}, error)
	DeleteMetadata(key string) error

	// Backup and Recovery
	Export() ([]byte, error)
	Import(data []byte) error
	Clear() error
}

// WalletConfig contains configuration for the wallet
type WalletConfig struct {
	// Storage configuration
	StorageType string `json:"storageType"` // "memory", "file", "encrypted"
	StoragePath string `json:"storagePath,omitempty"`
	
	// Security configuration
	EncryptionEnabled bool   `json:"encryptionEnabled"`
	KeyDerivationRounds int  `json:"keyDerivationRounds"`
	
	// Key management
	DefaultKeyType    did.KeyType `json:"defaultKeyType"`
	DefaultAlgorithm  string      `json:"defaultAlgorithm"`
	
	// Auto-lock configuration
	AutoLockTimeout   time.Duration `json:"autoLockTimeout"`
	
	// DID resolution
	DIDResolver       did.MultiResolver `json:"-"`
	
	// Credential verification
	CredentialVerifier vc.CredentialVerifier `json:"-"`
}

// DefaultWalletConfig returns default wallet configuration
func DefaultWalletConfig() *WalletConfig {
	return &WalletConfig{
		StorageType:         "memory",
		EncryptionEnabled:   true,
		KeyDerivationRounds: 100000,
		DefaultKeyType:      did.KeyTypeEd25519,
		DefaultAlgorithm:    "EdDSA",
		AutoLockTimeout:     time.Hour,
	}
}

// WalletError represents wallet-specific errors
type WalletError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *WalletError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Error codes
const (
	ErrorWalletLocked       = "wallet_locked"
	ErrorWalletUnlocked     = "wallet_unlocked"
	ErrorInvalidPassword    = "invalid_password"
	ErrorKeyNotFound        = "key_not_found"
	ErrorDIDNotFound        = "did_not_found"
	ErrorCredentialNotFound = "credential_not_found"
	ErrorPresentationNotFound = "presentation_not_found"
	ErrorKeyAlreadyExists   = "key_already_exists"
	ErrorDIDAlreadyExists   = "did_already_exists"
	ErrorInvalidKeyType     = "invalid_key_type"
	ErrorInvalidDID         = "invalid_did"
	ErrorInvalidCredential  = "invalid_credential"
	ErrorStorageError       = "storage_error"
	ErrorCryptoError        = "crypto_error"
	ErrorSerializationError = "serialization_error"
)

// NewWalletError creates a new wallet error
func NewWalletError(code, message string) *WalletError {
	return &WalletError{
		Code:    code,
		Message: message,
	}
}

// NewWalletErrorWithDetails creates a new wallet error with details
func NewWalletErrorWithDetails(code, message, details string) *WalletError {
	return &WalletError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WalletMetrics contains wallet usage metrics
type WalletMetrics struct {
	KeysCount         int       `json:"keysCount"`
	DIDsCount         int       `json:"didsCount"`
	CredentialsCount  int       `json:"credentialsCount"`
	PresentationsCount int      `json:"presentationsCount"`
	LastUnlocked      *time.Time `json:"lastUnlocked,omitempty"`
	UnlockCount       int64     `json:"unlockCount"`
	SignatureCount    int64     `json:"signatureCount"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}