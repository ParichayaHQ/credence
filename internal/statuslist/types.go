package statuslist

import (
	"time"
)

// StatusList2021 represents a W3C StatusList 2021 credential
type StatusList2021 struct {
	Context           []string    `json:"@context"`
	ID                string      `json:"id"`
	Type              []string    `json:"type"`
	Issuer            string      `json:"issuer"`
	IssuanceDate      string      `json:"issuanceDate"`
	CredentialSubject StatusList  `json:"credentialSubject"`
	Proof             interface{} `json:"proof,omitempty"`
}

// StatusList represents the credential subject of a StatusList credential
type StatusList struct {
	ID           string `json:"id"`
	Type         string `json:"type"` // "StatusList2021"
	StatusPurpose string `json:"statusPurpose"` // "revocation" or "suspension"
	EncodedList  string `json:"encodedList"` // GZIP + Base64 encoded bitstring
}

// StatusListEntry represents a credential's status entry reference
type StatusListEntry struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"` // "StatusList2021Entry"
	StatusPurpose        string `json:"statusPurpose"` // "revocation" or "suspension"
	StatusListIndex      string `json:"statusListIndex"` // Index position as string
	StatusListCredential string `json:"statusListCredential"` // URL to StatusList credential
}

// BitString represents an expandable bit array for status tracking
type BitString struct {
	bits   []byte
	length int
}

// StatusListManager manages StatusList 2021 credentials and operations
type StatusListManager interface {
	// CreateStatusList creates a new status list credential
	CreateStatusList(issuer string, purpose StatusPurpose, size int) (*StatusList2021, error)
	
	// GetStatusList retrieves a status list credential by ID
	GetStatusList(listID string) (*StatusList2021, error)
	
	// UpdateStatus updates the status of a credential in a status list
	UpdateStatus(listID string, index int, status bool) error
	
	// CheckStatus checks the status of a credential
	CheckStatus(entry *StatusListEntry) (*StatusResult, error)
	
	// AllocateIndex allocates the next available index in a status list
	AllocateIndex(listID string) (int, error)
	
	// GenerateEntry generates a status list entry for a credential
	GenerateEntry(listID string, index int, purpose StatusPurpose) (*StatusListEntry, error)
}

// StatusListProvider provides access to status list credentials
type StatusListProvider interface {
	// FetchStatusList fetches a status list credential from a URL
	FetchStatusList(url string) (*StatusList2021, error)
	
	// StoreStatusList stores a status list credential
	StoreStatusList(list *StatusList2021) error
	
	// ListStatusLists lists all available status list IDs
	ListStatusLists() ([]string, error)
}

// StatusPurpose represents the purpose of a status list
type StatusPurpose string

const (
	// StatusPurposeRevocation indicates the list tracks revoked credentials
	StatusPurposeRevocation StatusPurpose = "revocation"
	
	// StatusPurposeSuspension indicates the list tracks suspended credentials
	StatusPurposeSuspension StatusPurpose = "suspension"
)

// StatusResult represents the result of a status check
type StatusResult struct {
	// Valid indicates if the credential is valid (not revoked/suspended)
	Valid bool `json:"valid"`
	
	// Status indicates the actual status (true = revoked/suspended, false = valid)
	Status bool `json:"status"`
	
	// Purpose indicates what the status represents
	Purpose StatusPurpose `json:"purpose"`
	
	// Index is the position in the status list
	Index int `json:"index"`
	
	// ListID is the ID of the status list credential
	ListID string `json:"listId"`
	
	// LastUpdated is when the status was last updated
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
}

// StatusListConfig contains configuration for status list management
type StatusListConfig struct {
	// DefaultSize is the default size for new status lists
	DefaultSize int `json:"defaultSize"`
	
	// MaxSize is the maximum size for status lists
	MaxSize int `json:"maxSize"`
	
	// CompressionLevel for GZIP compression (1-9)
	CompressionLevel int `json:"compressionLevel"`
	
	// CacheTimeout for status list caching
	CacheTimeout time.Duration `json:"cacheTimeout"`
	
	// AutoExpand enables automatic expansion of status lists
	AutoExpand bool `json:"autoExpand"`
	
	// ExpandIncrement is the size increment when expanding
	ExpandIncrement int `json:"expandIncrement"`
}

// DefaultStatusListConfig returns default configuration
func DefaultStatusListConfig() *StatusListConfig {
	return &StatusListConfig{
		DefaultSize:      131072, // 128KB when compressed typically becomes ~16KB
		MaxSize:          1048576, // 1MB
		CompressionLevel: 6,
		CacheTimeout:     time.Hour,
		AutoExpand:       true,
		ExpandIncrement:  65536, // 64KB
	}
}

// StatusListError represents an error in status list operations
type StatusListError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *StatusListError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Error codes for status list operations
const (
	ErrorInvalidStatusList = "invalid_status_list"
	ErrorInvalidIndex      = "invalid_index"
	ErrorListNotFound      = "list_not_found"
	ErrorListFull          = "list_full"
	ErrorInvalidEntry      = "invalid_entry"
	ErrorNetworkError      = "network_error"
	ErrorCompressionError  = "compression_error"
	ErrorEncodingError     = "encoding_error"
	ErrorPermissionDenied  = "permission_denied"
	ErrorInvalidPurpose    = "invalid_purpose"
)

// NewStatusListError creates a new status list error
func NewStatusListError(code, message string) *StatusListError {
	return &StatusListError{
		Code:    code,
		Message: message,
	}
}

// NewStatusListErrorWithDetails creates a new status list error with details
func NewStatusListErrorWithDetails(code, message, details string) *StatusListError {
	return &StatusListError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// StatusListMetrics contains metrics for status list operations
type StatusListMetrics struct {
	// TotalLists is the total number of status lists
	TotalLists int `json:"totalLists"`
	
	// TotalEntries is the total number of allocated entries across all lists
	TotalEntries int `json:"totalEntries"`
	
	// RevokedCount is the number of revoked credentials
	RevokedCount int `json:"revokedCount"`
	
	// SuspendedCount is the number of suspended credentials
	SuspendedCount int `json:"suspendedCount"`
	
	// AverageListSize is the average size of status lists
	AverageListSize float64 `json:"averageListSize"`
	
	// CompressionRatio is the average compression ratio
	CompressionRatio float64 `json:"compressionRatio"`
	
	// LastUpdated is when metrics were last calculated
	LastUpdated time.Time `json:"lastUpdated"`
}

// StatusListCache provides caching for status lists
type StatusListCache interface {
	// Get retrieves a cached status list
	Get(listID string) (*StatusList2021, error)
	
	// Set stores a status list in cache
	Set(listID string, list *StatusList2021, ttl time.Duration) error
	
	// Invalidate removes a status list from cache
	Invalidate(listID string) error
	
	// Clear removes all cached status lists
	Clear() error
	
	// Stats returns cache statistics
	Stats() *CacheStats
}

// CacheStats contains cache statistics
type CacheStats struct {
	// Hits is the number of cache hits
	Hits int64 `json:"hits"`
	
	// Misses is the number of cache misses
	Misses int64 `json:"misses"`
	
	// Size is the current cache size
	Size int `json:"size"`
	
	// MaxSize is the maximum cache size
	MaxSize int `json:"maxSize"`
	
	// HitRatio is the cache hit ratio
	HitRatio float64 `json:"hitRatio"`
}