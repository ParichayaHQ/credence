package statuslist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// InMemoryStatusListProvider provides in-memory storage for status lists
type InMemoryStatusListProvider struct {
	mutex sync.RWMutex
	lists map[string]*StatusList2021
}

// NewInMemoryStatusListProvider creates a new in-memory status list provider
func NewInMemoryStatusListProvider() *InMemoryStatusListProvider {
	return &InMemoryStatusListProvider{
		lists: make(map[string]*StatusList2021),
	}
}

// FetchStatusList fetches a status list credential from storage or URL
func (p *InMemoryStatusListProvider) FetchStatusList(url string) (*StatusList2021, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	if list, exists := p.lists[url]; exists {
		return p.cloneStatusList(list), nil
	}
	
	return nil, NewStatusListError(ErrorListNotFound, "status list not found: "+url)
}

// StoreStatusList stores a status list credential
func (p *InMemoryStatusListProvider) StoreStatusList(list *StatusList2021) error {
	if list == nil {
		return NewStatusListError(ErrorInvalidStatusList, "status list cannot be nil")
	}
	
	if list.ID == "" {
		return NewStatusListError(ErrorInvalidStatusList, "status list ID cannot be empty")
	}
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.lists[list.ID] = p.cloneStatusList(list)
	return nil
}

// ListStatusLists lists all available status list IDs
func (p *InMemoryStatusListProvider) ListStatusLists() ([]string, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	ids := make([]string, 0, len(p.lists))
	for id := range p.lists {
		ids = append(ids, id)
	}
	
	return ids, nil
}

// DeleteStatusList removes a status list from storage
func (p *InMemoryStatusListProvider) DeleteStatusList(listID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if _, exists := p.lists[listID]; !exists {
		return NewStatusListError(ErrorListNotFound, "status list not found: "+listID)
	}
	
	delete(p.lists, listID)
	return nil
}

// Clear removes all status lists from storage
func (p *InMemoryStatusListProvider) Clear() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.lists = make(map[string]*StatusList2021)
	return nil
}

// Size returns the number of stored status lists
func (p *InMemoryStatusListProvider) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return len(p.lists)
}

func (p *InMemoryStatusListProvider) cloneStatusList(list *StatusList2021) *StatusList2021 {
	return &StatusList2021{
		Context:      append([]string{}, list.Context...),
		ID:           list.ID,
		Type:         append([]string{}, list.Type...),
		Issuer:       list.Issuer,
		IssuanceDate: list.IssuanceDate,
		CredentialSubject: StatusList{
			ID:           list.CredentialSubject.ID,
			Type:         list.CredentialSubject.Type,
			StatusPurpose: list.CredentialSubject.StatusPurpose,
			EncodedList:  list.CredentialSubject.EncodedList,
		},
		Proof: list.Proof, // Shallow copy is sufficient for proof
	}
}

// HTTPStatusListProvider fetches status lists from HTTP URLs
type HTTPStatusListProvider struct {
	client  *http.Client
	timeout time.Duration
	
	// Local cache for fetched status lists
	cache map[string]*cachedStatusList
	mutex sync.RWMutex
}

type cachedStatusList struct {
	list      *StatusList2021
	fetchTime time.Time
	ttl       time.Duration
}

// NewHTTPStatusListProvider creates a new HTTP-based status list provider
func NewHTTPStatusListProvider(timeout time.Duration) *HTTPStatusListProvider {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &HTTPStatusListProvider{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		cache:   make(map[string]*cachedStatusList),
	}
}

// FetchStatusList fetches a status list credential from an HTTP URL
func (p *HTTPStatusListProvider) FetchStatusList(url string) (*StatusList2021, error) {
	// Check cache first
	p.mutex.RLock()
	if cached, exists := p.cache[url]; exists {
		if time.Since(cached.fetchTime) < cached.ttl {
			p.mutex.RUnlock()
			return p.cloneStatusList(cached.list), nil
		}
	}
	p.mutex.RUnlock()
	
	// Fetch from URL
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorNetworkError, "failed to fetch status list", err.Error())
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, NewStatusListError(ErrorNetworkError, fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status))
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorNetworkError, "failed to read response body", err.Error())
	}
	
	// Parse JSON
	var statusList StatusList2021
	if err := json.Unmarshal(body, &statusList); err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorEncodingError, "failed to parse status list JSON", err.Error())
	}
	
	// Validate the status list
	if err := p.validateStatusList(&statusList); err != nil {
		return nil, err
	}
	
	// Cache the result (with 1 hour TTL by default)
	p.mutex.Lock()
	p.cache[url] = &cachedStatusList{
		list:      p.cloneStatusList(&statusList),
		fetchTime: time.Now(),
		ttl:       time.Hour,
	}
	p.mutex.Unlock()
	
	return &statusList, nil
}

// StoreStatusList is not supported by HTTP provider (read-only)
func (p *HTTPStatusListProvider) StoreStatusList(list *StatusList2021) error {
	return NewStatusListError(ErrorPermissionDenied, "HTTP provider is read-only")
}

// ListStatusLists is not supported by HTTP provider
func (p *HTTPStatusListProvider) ListStatusLists() ([]string, error) {
	return nil, NewStatusListError(ErrorPermissionDenied, "HTTP provider cannot list all status lists")
}

// ClearCache clears the internal cache
func (p *HTTPStatusListProvider) ClearCache() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.cache = make(map[string]*cachedStatusList)
}

// SetCacheTTL sets the cache TTL for a specific URL
func (p *HTTPStatusListProvider) SetCacheTTL(url string, ttl time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if cached, exists := p.cache[url]; exists {
		cached.ttl = ttl
	}
}

func (p *HTTPStatusListProvider) validateStatusList(list *StatusList2021) error {
	if list.ID == "" {
		return NewStatusListError(ErrorInvalidStatusList, "status list ID cannot be empty")
	}
	
	if len(list.Context) == 0 {
		return NewStatusListError(ErrorInvalidStatusList, "@context is required")
	}
	
	if len(list.Type) == 0 {
		return NewStatusListError(ErrorInvalidStatusList, "type is required")
	}
	
	// Check for required types
	hasVC := false
	hasSL := false
	for _, t := range list.Type {
		if t == "VerifiableCredential" {
			hasVC = true
		}
		if t == "StatusList2021Credential" {
			hasSL = true
		}
	}
	
	if !hasVC {
		return NewStatusListError(ErrorInvalidStatusList, "type must include VerifiableCredential")
	}
	
	if !hasSL {
		return NewStatusListError(ErrorInvalidStatusList, "type must include StatusList2021Credential")
	}
	
	if list.Issuer == "" {
		return NewStatusListError(ErrorInvalidStatusList, "issuer is required")
	}
	
	if list.CredentialSubject.Type != "StatusList2021" {
		return NewStatusListError(ErrorInvalidStatusList, "credentialSubject type must be StatusList2021")
	}
	
	if list.CredentialSubject.StatusPurpose != string(StatusPurposeRevocation) &&
		list.CredentialSubject.StatusPurpose != string(StatusPurposeSuspension) {
		return NewStatusListError(ErrorInvalidStatusList, "invalid status purpose")
	}
	
	if list.CredentialSubject.EncodedList == "" {
		return NewStatusListError(ErrorInvalidStatusList, "encodedList is required")
	}
	
	// Try to decode the encoded list to verify it's valid
	_, err := FromCompressedBase64(list.CredentialSubject.EncodedList)
	if err != nil {
		return NewStatusListErrorWithDetails(ErrorInvalidStatusList, "invalid encodedList", err.Error())
	}
	
	return nil
}

func (p *HTTPStatusListProvider) cloneStatusList(list *StatusList2021) *StatusList2021 {
	return &StatusList2021{
		Context:      append([]string{}, list.Context...),
		ID:           list.ID,
		Type:         append([]string{}, list.Type...),
		Issuer:       list.Issuer,
		IssuanceDate: list.IssuanceDate,
		CredentialSubject: StatusList{
			ID:           list.CredentialSubject.ID,
			Type:         list.CredentialSubject.Type,
			StatusPurpose: list.CredentialSubject.StatusPurpose,
			EncodedList:  list.CredentialSubject.EncodedList,
		},
		Proof: list.Proof,
	}
}

// CompositeStatusListProvider combines multiple providers
type CompositeStatusListProvider struct {
	providers []StatusListProvider
	writer    StatusListProvider // The provider used for write operations
}

// NewCompositeStatusListProvider creates a provider that tries multiple providers in order
func NewCompositeStatusListProvider(writer StatusListProvider, readers ...StatusListProvider) *CompositeStatusListProvider {
	providers := []StatusListProvider{writer}
	providers = append(providers, readers...)
	
	return &CompositeStatusListProvider{
		providers: providers,
		writer:    writer,
	}
}

// FetchStatusList tries each provider in order until one succeeds
func (p *CompositeStatusListProvider) FetchStatusList(url string) (*StatusList2021, error) {
	var lastErr error
	
	for _, provider := range p.providers {
		list, err := provider.FetchStatusList(url)
		if err == nil {
			return list, nil
		}
		lastErr = err
	}
	
	if lastErr != nil {
		return nil, lastErr
	}
	
	return nil, NewStatusListError(ErrorListNotFound, "status list not found in any provider")
}

// StoreStatusList uses the designated writer provider
func (p *CompositeStatusListProvider) StoreStatusList(list *StatusList2021) error {
	if p.writer == nil {
		return NewStatusListError(ErrorPermissionDenied, "no writer provider configured")
	}
	
	return p.writer.StoreStatusList(list)
}

// ListStatusLists uses the designated writer provider
func (p *CompositeStatusListProvider) ListStatusLists() ([]string, error) {
	if p.writer == nil {
		return nil, NewStatusListError(ErrorPermissionDenied, "no writer provider configured")
	}
	
	return p.writer.ListStatusLists()
}