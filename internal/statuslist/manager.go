package statuslist

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
)

// DefaultStatusListManager implements the StatusListManager interface
type DefaultStatusListManager struct {
	config          *StatusListConfig
	provider        StatusListProvider
	cache           StatusListCache
	keyManager      did.KeyManager
	credentialSigner vc.CredentialIssuer
	
	// Index allocation tracking
	indexMutex sync.RWMutex
	nextIndex  map[string]int // listID -> next available index
}

// NewDefaultStatusListManager creates a new status list manager
func NewDefaultStatusListManager(
	config *StatusListConfig,
	provider StatusListProvider,
	cache StatusListCache,
	keyManager did.KeyManager,
	credentialSigner vc.CredentialIssuer,
) *DefaultStatusListManager {
	if config == nil {
		config = DefaultStatusListConfig()
	}
	
	return &DefaultStatusListManager{
		config:           config,
		provider:         provider,
		cache:            cache,
		keyManager:       keyManager,
		credentialSigner: credentialSigner,
		nextIndex:        make(map[string]int),
	}
}

// CreateStatusList creates a new status list credential
func (m *DefaultStatusListManager) CreateStatusList(issuer string, purpose StatusPurpose, size int) (*StatusList2021, error) {
	if issuer == "" {
		return nil, NewStatusListError(ErrorInvalidStatusList, "issuer cannot be empty")
	}
	
	if purpose != StatusPurposeRevocation && purpose != StatusPurposeSuspension {
		return nil, NewStatusListError(ErrorInvalidPurpose, "purpose must be revocation or suspension")
	}
	
	if size <= 0 {
		size = m.config.DefaultSize
	}
	
	if size > m.config.MaxSize {
		return nil, NewStatusListError(ErrorInvalidStatusList, "size exceeds maximum allowed")
	}
	
	// Create empty bitstring
	bitString := NewBitString(size)
	encodedList, err := bitString.ToCompressedBase64(m.config.CompressionLevel)
	if err != nil {
		return nil, err
	}
	
	// Generate unique ID for the status list
	listID := fmt.Sprintf("%s/status-lists/%s/%d", issuer, purpose, time.Now().Unix())
	
	// Create the status list credential
	statusList := &StatusList2021{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://w3id.org/vc/status-list/2021/v1",
		},
		ID:           listID,
		Type:         []string{"VerifiableCredential", "StatusList2021Credential"},
		Issuer:       issuer,
		IssuanceDate: time.Now().UTC().Format(time.RFC3339),
		CredentialSubject: StatusList{
			ID:           listID + "#list",
			Type:         "StatusList2021",
			StatusPurpose: string(purpose),
			EncodedList:  encodedList,
		},
	}
	
	// Store in provider
	if err := m.provider.StoreStatusList(statusList); err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorNetworkError, "failed to store status list", err.Error())
	}
	
	// Cache the status list
	if m.cache != nil {
		m.cache.Set(listID, statusList, m.config.CacheTimeout)
	}
	
	// Initialize index tracking
	m.indexMutex.Lock()
	m.nextIndex[listID] = 0
	m.indexMutex.Unlock()
	
	return statusList, nil
}

// GetStatusList retrieves a status list credential by ID
func (m *DefaultStatusListManager) GetStatusList(listID string) (*StatusList2021, error) {
	if listID == "" {
		return nil, NewStatusListError(ErrorInvalidStatusList, "list ID cannot be empty")
	}
	
	// Check cache first
	if m.cache != nil {
		if cached, err := m.cache.Get(listID); err == nil && cached != nil {
			return cached, nil
		}
	}
	
	// Fetch from provider
	statusList, err := m.provider.FetchStatusList(listID)
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorListNotFound, "failed to fetch status list", err.Error())
	}
	
	// Cache the result
	if m.cache != nil {
		m.cache.Set(listID, statusList, m.config.CacheTimeout)
	}
	
	return statusList, nil
}

// UpdateStatus updates the status of a credential in a status list
func (m *DefaultStatusListManager) UpdateStatus(listID string, index int, status bool) error {
	if listID == "" {
		return NewStatusListError(ErrorInvalidStatusList, "list ID cannot be empty")
	}
	
	if index < 0 {
		return NewStatusListError(ErrorInvalidIndex, "index cannot be negative")
	}
	
	// Get the current status list
	statusList, err := m.GetStatusList(listID)
	if err != nil {
		return err
	}
	
	// Decode the current bitstring
	bitString, err := FromCompressedBase64(statusList.CredentialSubject.EncodedList)
	if err != nil {
		return err
	}
	
	// Update the bit
	if err := bitString.Set(index, status); err != nil {
		return err
	}
	
	// Re-encode the bitstring
	encodedList, err := bitString.ToCompressedBase64(m.config.CompressionLevel)
	if err != nil {
		return err
	}
	
	// Update the status list
	statusList.CredentialSubject.EncodedList = encodedList
	statusList.IssuanceDate = time.Now().UTC().Format(time.RFC3339)
	
	// Store the updated status list
	if err := m.provider.StoreStatusList(statusList); err != nil {
		return NewStatusListErrorWithDetails(ErrorNetworkError, "failed to store updated status list", err.Error())
	}
	
	// Invalidate cache
	if m.cache != nil {
		m.cache.Invalidate(listID)
	}
	
	return nil
}

// CheckStatus checks the status of a credential
func (m *DefaultStatusListManager) CheckStatus(entry *StatusListEntry) (*StatusResult, error) {
	if entry == nil {
		return nil, NewStatusListError(ErrorInvalidEntry, "status entry cannot be nil")
	}
	
	if entry.StatusListCredential == "" {
		return nil, NewStatusListError(ErrorInvalidEntry, "status list credential URL cannot be empty")
	}
	
	if entry.StatusListIndex == "" {
		return nil, NewStatusListError(ErrorInvalidEntry, "status list index cannot be empty")
	}
	
	// Parse the index
	index, err := strconv.Atoi(entry.StatusListIndex)
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorInvalidIndex, "invalid status list index", err.Error())
	}
	
	// Get the status list
	statusList, err := m.GetStatusList(entry.StatusListCredential)
	if err != nil {
		return nil, err
	}
	
	// Verify the purpose matches
	if statusList.CredentialSubject.StatusPurpose != entry.StatusPurpose {
		return nil, NewStatusListError(ErrorInvalidEntry, "status purpose mismatch")
	}
	
	// Decode the bitstring
	bitString, err := FromCompressedBase64(statusList.CredentialSubject.EncodedList)
	if err != nil {
		return nil, err
	}
	
	// Check the status
	status, err := bitString.Get(index)
	if err != nil {
		return nil, err
	}
	
	// Parse issuance date for last updated
	var lastUpdated *time.Time
	if statusList.IssuanceDate != "" {
		if t, err := time.Parse(time.RFC3339, statusList.IssuanceDate); err == nil {
			lastUpdated = &t
		}
	}
	
	return &StatusResult{
		Valid:       !status, // If status is true, credential is revoked/suspended (not valid)
		Status:      status,
		Purpose:     StatusPurpose(entry.StatusPurpose),
		Index:       index,
		ListID:      entry.StatusListCredential,
		LastUpdated: lastUpdated,
	}, nil
}

// AllocateIndex allocates the next available index in a status list
func (m *DefaultStatusListManager) AllocateIndex(listID string) (int, error) {
	if listID == "" {
		return -1, NewStatusListError(ErrorInvalidStatusList, "list ID cannot be empty")
	}
	
	m.indexMutex.Lock()
	defer m.indexMutex.Unlock()
	
	// Get the current status list to determine actual size
	statusList, err := m.GetStatusList(listID)
	if err != nil {
		return -1, err
	}
	
	// Decode the bitstring to check actual capacity
	bitString, err := FromCompressedBase64(statusList.CredentialSubject.EncodedList)
	if err != nil {
		return -1, err
	}
	
	// Find the next available index
	nextIndex, exists := m.nextIndex[listID]
	if !exists {
		nextIndex = 0
		m.nextIndex[listID] = nextIndex
	}
	
	// Find an unset bit starting from nextIndex
	for i := nextIndex; i < bitString.Length(); i++ {
		if bit, _ := bitString.Get(i); !bit {
			m.nextIndex[listID] = i + 1
			return i, nil
		}
	}
	
	// If auto-expand is enabled and we reached the end, expand the list
	if m.config.AutoExpand {
		newSize := bitString.Length() + m.config.ExpandIncrement
		if newSize > m.config.MaxSize {
			return -1, NewStatusListError(ErrorListFull, "status list cannot be expanded beyond maximum size")
		}
		
		// Expand the bitstring
		if err := bitString.Expand(newSize); err != nil {
			return -1, err
		}
		
		// Re-encode and update the status list
		encodedList, err := bitString.ToCompressedBase64(m.config.CompressionLevel)
		if err != nil {
			return -1, err
		}
		
		statusList.CredentialSubject.EncodedList = encodedList
		statusList.IssuanceDate = time.Now().UTC().Format(time.RFC3339)
		
		// Store the expanded status list
		if err := m.provider.StoreStatusList(statusList); err != nil {
			return -1, NewStatusListErrorWithDetails(ErrorNetworkError, "failed to store expanded status list", err.Error())
		}
		
		// Invalidate cache
		if m.cache != nil {
			m.cache.Invalidate(listID)
		}
		
		// Return the first new index
		allocatedIndex := bitString.Length() - m.config.ExpandIncrement
		m.nextIndex[listID] = allocatedIndex + 1
		return allocatedIndex, nil
	}
	
	return -1, NewStatusListError(ErrorListFull, "no available indexes in status list")
}

// GenerateEntry generates a status list entry for a credential
func (m *DefaultStatusListManager) GenerateEntry(listID string, index int, purpose StatusPurpose) (*StatusListEntry, error) {
	if listID == "" {
		return nil, NewStatusListError(ErrorInvalidStatusList, "list ID cannot be empty")
	}
	
	if index < 0 {
		return nil, NewStatusListError(ErrorInvalidIndex, "index cannot be negative")
	}
	
	if purpose != StatusPurposeRevocation && purpose != StatusPurposeSuspension {
		return nil, NewStatusListError(ErrorInvalidPurpose, "invalid status purpose")
	}
	
	// Generate entry ID
	entryID := fmt.Sprintf("%s#%d", listID, index)
	
	return &StatusListEntry{
		ID:                   entryID,
		Type:                 "StatusList2021Entry",
		StatusPurpose:        string(purpose),
		StatusListIndex:      strconv.Itoa(index),
		StatusListCredential: listID,
	}, nil
}

// RevokeCredential revokes a credential by setting its status bit
func (m *DefaultStatusListManager) RevokeCredential(listID string, index int) error {
	return m.UpdateStatus(listID, index, true)
}

// SuspendCredential suspends a credential by setting its status bit
func (m *DefaultStatusListManager) SuspendCredential(listID string, index int) error {
	return m.UpdateStatus(listID, index, true)
}

// RestoreCredential restores a credential by clearing its status bit
func (m *DefaultStatusListManager) RestoreCredential(listID string, index int) error {
	return m.UpdateStatus(listID, index, false)
}

// GetMetrics returns metrics for status list operations
func (m *DefaultStatusListManager) GetMetrics() (*StatusListMetrics, error) {
	lists, err := m.provider.ListStatusLists()
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorNetworkError, "failed to list status lists", err.Error())
	}
	
	metrics := &StatusListMetrics{
		TotalLists:  len(lists),
		LastUpdated: time.Now().UTC(),
	}
	
	var totalSize int
	var totalCompressed int64
	var totalEntries int
	
	for _, listID := range lists {
		statusList, err := m.GetStatusList(listID)
		if err != nil {
			continue // Skip lists we can't access
		}
		
		bitString, err := FromCompressedBase64(statusList.CredentialSubject.EncodedList)
		if err != nil {
			continue // Skip lists we can't decode
		}
		
		totalSize += bitString.Length()
		totalCompressed += int64(len(statusList.CredentialSubject.EncodedList))
		
		// Count set bits based on purpose
		setBits := bitString.CountSet()
		if statusList.CredentialSubject.StatusPurpose == string(StatusPurposeRevocation) {
			metrics.RevokedCount += setBits
		} else if statusList.CredentialSubject.StatusPurpose == string(StatusPurposeSuspension) {
			metrics.SuspendedCount += setBits
		}
		
		totalEntries += setBits
	}
	
	if len(lists) > 0 {
		metrics.AverageListSize = float64(totalSize) / float64(len(lists))
		
		// Calculate compression ratio (original bits / compressed bytes)
		if totalCompressed > 0 {
			metrics.CompressionRatio = float64(totalSize) / float64(totalCompressed*8)
		}
	}
	
	metrics.TotalEntries = totalEntries
	
	return metrics, nil
}