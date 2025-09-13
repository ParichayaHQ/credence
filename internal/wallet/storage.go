package wallet

import (
	"encoding/json"
	"sync"
	"time"
)

// InMemoryStorage provides an in-memory implementation of WalletStorage
type InMemoryStorage struct {
	mutex sync.RWMutex
	
	keys         map[string]*KeyPair
	dids         map[string]*DIDRecord
	credentials  map[string]*CredentialRecord
	presentations map[string]*PresentationRecord
	metadata     map[string]interface{}
}

// NewInMemoryStorage creates a new in-memory wallet storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		keys:          make(map[string]*KeyPair),
		dids:          make(map[string]*DIDRecord),
		credentials:   make(map[string]*CredentialRecord),
		presentations: make(map[string]*PresentationRecord),
		metadata:      make(map[string]interface{}),
	}
}

// Key Storage

func (s *InMemoryStorage) StoreKey(keyPair *KeyPair) error {
	if keyPair == nil {
		return NewWalletError(ErrorStorageError, "key pair cannot be nil")
	}
	
	if keyPair.ID == "" {
		return NewWalletError(ErrorStorageError, "key ID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if key already exists
	if _, exists := s.keys[keyPair.ID]; exists {
		return NewWalletError(ErrorKeyAlreadyExists, "key already exists: "+keyPair.ID)
	}
	
	// Clone the key pair to avoid external modifications
	cloned := s.cloneKeyPair(keyPair)
	s.keys[keyPair.ID] = cloned
	
	return nil
}

func (s *InMemoryStorage) GetKey(keyID string) (*KeyPair, error) {
	if keyID == "" {
		return nil, NewWalletError(ErrorStorageError, "key ID cannot be empty")
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	keyPair, exists := s.keys[keyID]
	if !exists {
		return nil, NewWalletError(ErrorKeyNotFound, "key not found: "+keyID)
	}
	
	return s.cloneKeyPair(keyPair), nil
}

func (s *InMemoryStorage) ListKeys() ([]*KeyPair, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	keys := make([]*KeyPair, 0, len(s.keys))
	for _, keyPair := range s.keys {
		keys = append(keys, s.cloneKeyPair(keyPair))
	}
	
	return keys, nil
}

func (s *InMemoryStorage) DeleteKey(keyID string) error {
	if keyID == "" {
		return NewWalletError(ErrorStorageError, "key ID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.keys[keyID]; !exists {
		return NewWalletError(ErrorKeyNotFound, "key not found: "+keyID)
	}
	
	delete(s.keys, keyID)
	return nil
}

// DID Storage

func (s *InMemoryStorage) StoreDID(record *DIDRecord) error {
	if record == nil {
		return NewWalletError(ErrorStorageError, "DID record cannot be nil")
	}
	
	if record.DID == "" {
		return NewWalletError(ErrorStorageError, "DID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if DID already exists
	if _, exists := s.dids[record.DID]; exists {
		return NewWalletError(ErrorDIDAlreadyExists, "DID already exists: "+record.DID)
	}
	
	// Clone the record to avoid external modifications
	cloned := s.cloneDIDRecord(record)
	s.dids[record.DID] = cloned
	
	return nil
}

func (s *InMemoryStorage) GetDID(did string) (*DIDRecord, error) {
	if did == "" {
		return nil, NewWalletError(ErrorStorageError, "DID cannot be empty")
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	record, exists := s.dids[did]
	if !exists {
		return nil, NewWalletError(ErrorDIDNotFound, "DID not found: "+did)
	}
	
	return s.cloneDIDRecord(record), nil
}

func (s *InMemoryStorage) ListDIDs() ([]*DIDRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	dids := make([]*DIDRecord, 0, len(s.dids))
	for _, record := range s.dids {
		dids = append(dids, s.cloneDIDRecord(record))
	}
	
	return dids, nil
}

func (s *InMemoryStorage) DeleteDID(did string) error {
	if did == "" {
		return NewWalletError(ErrorStorageError, "DID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.dids[did]; !exists {
		return NewWalletError(ErrorDIDNotFound, "DID not found: "+did)
	}
	
	delete(s.dids, did)
	return nil
}

// Credential Storage

func (s *InMemoryStorage) StoreCredential(record *CredentialRecord) error {
	if record == nil {
		return NewWalletError(ErrorStorageError, "credential record cannot be nil")
	}
	
	if record.ID == "" {
		return NewWalletError(ErrorStorageError, "credential ID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Clone the record to avoid external modifications
	cloned := s.cloneCredentialRecord(record)
	s.credentials[record.ID] = cloned
	
	return nil
}

func (s *InMemoryStorage) GetCredential(credentialID string) (*CredentialRecord, error) {
	if credentialID == "" {
		return nil, NewWalletError(ErrorStorageError, "credential ID cannot be empty")
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	record, exists := s.credentials[credentialID]
	if !exists {
		return nil, NewWalletError(ErrorCredentialNotFound, "credential not found: "+credentialID)
	}
	
	return s.cloneCredentialRecord(record), nil
}

func (s *InMemoryStorage) ListCredentials(filter *CredentialFilter) ([]*CredentialRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var credentials []*CredentialRecord
	
	for _, record := range s.credentials {
		if s.matchesFilter(record, filter) {
			credentials = append(credentials, s.cloneCredentialRecord(record))
		}
	}
	
	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(credentials) {
			credentials = credentials[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(credentials) {
			credentials = credentials[:filter.Limit]
		}
	}
	
	return credentials, nil
}

func (s *InMemoryStorage) DeleteCredential(credentialID string) error {
	if credentialID == "" {
		return NewWalletError(ErrorStorageError, "credential ID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.credentials[credentialID]; !exists {
		return NewWalletError(ErrorCredentialNotFound, "credential not found: "+credentialID)
	}
	
	delete(s.credentials, credentialID)
	return nil
}

// Presentation Storage

func (s *InMemoryStorage) StorePresentation(record *PresentationRecord) error {
	if record == nil {
		return NewWalletError(ErrorStorageError, "presentation record cannot be nil")
	}
	
	if record.ID == "" {
		return NewWalletError(ErrorStorageError, "presentation ID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Clone the record to avoid external modifications
	cloned := s.clonePresentationRecord(record)
	s.presentations[record.ID] = cloned
	
	return nil
}

func (s *InMemoryStorage) GetPresentation(presentationID string) (*PresentationRecord, error) {
	if presentationID == "" {
		return nil, NewWalletError(ErrorStorageError, "presentation ID cannot be empty")
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	record, exists := s.presentations[presentationID]
	if !exists {
		return nil, NewWalletError(ErrorPresentationNotFound, "presentation not found: "+presentationID)
	}
	
	return s.clonePresentationRecord(record), nil
}

func (s *InMemoryStorage) ListPresentations() ([]*PresentationRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	presentations := make([]*PresentationRecord, 0, len(s.presentations))
	for _, record := range s.presentations {
		presentations = append(presentations, s.clonePresentationRecord(record))
	}
	
	return presentations, nil
}

func (s *InMemoryStorage) DeletePresentation(presentationID string) error {
	if presentationID == "" {
		return NewWalletError(ErrorStorageError, "presentation ID cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.presentations[presentationID]; !exists {
		return NewWalletError(ErrorPresentationNotFound, "presentation not found: "+presentationID)
	}
	
	delete(s.presentations, presentationID)
	return nil
}

// Metadata Storage

func (s *InMemoryStorage) SetMetadata(key string, value interface{}) error {
	if key == "" {
		return NewWalletError(ErrorStorageError, "metadata key cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.metadata[key] = value
	return nil
}

func (s *InMemoryStorage) GetMetadata(key string) (interface{}, error) {
	if key == "" {
		return nil, NewWalletError(ErrorStorageError, "metadata key cannot be empty")
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	value, exists := s.metadata[key]
	if !exists {
		return nil, NewWalletError(ErrorStorageError, "metadata not found: "+key)
	}
	
	return value, nil
}

func (s *InMemoryStorage) DeleteMetadata(key string) error {
	if key == "" {
		return NewWalletError(ErrorStorageError, "metadata key cannot be empty")
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.metadata, key)
	return nil
}

// Backup and Recovery

func (s *InMemoryStorage) Export() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	data := map[string]interface{}{
		"keys":          s.keys,
		"dids":          s.dids,
		"credentials":   s.credentials,
		"presentations": s.presentations,
		"metadata":      s.metadata,
		"exported_at":   time.Now().UTC(),
	}
	
	return json.Marshal(data)
}

func (s *InMemoryStorage) Import(data []byte) error {
	var imported map[string]interface{}
	if err := json.Unmarshal(data, &imported); err != nil {
		return NewWalletErrorWithDetails(ErrorSerializationError, "failed to unmarshal import data", err.Error())
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Import keys
	if keysData, ok := imported["keys"]; ok {
		keysBytes, _ := json.Marshal(keysData)
		var keys map[string]*KeyPair
		if err := json.Unmarshal(keysBytes, &keys); err == nil {
			s.keys = keys
		}
	}
	
	// Import DIDs
	if didsData, ok := imported["dids"]; ok {
		didsBytes, _ := json.Marshal(didsData)
		var dids map[string]*DIDRecord
		if err := json.Unmarshal(didsBytes, &dids); err == nil {
			s.dids = dids
		}
	}
	
	// Import credentials
	if credsData, ok := imported["credentials"]; ok {
		credsBytes, _ := json.Marshal(credsData)
		var credentials map[string]*CredentialRecord
		if err := json.Unmarshal(credsBytes, &credentials); err == nil {
			s.credentials = credentials
		}
	}
	
	// Import presentations
	if presData, ok := imported["presentations"]; ok {
		presBytes, _ := json.Marshal(presData)
		var presentations map[string]*PresentationRecord
		if err := json.Unmarshal(presBytes, &presentations); err == nil {
			s.presentations = presentations
		}
	}
	
	// Import metadata
	if metaData, ok := imported["metadata"]; ok {
		metaBytes, _ := json.Marshal(metaData)
		var metadata map[string]interface{}
		if err := json.Unmarshal(metaBytes, &metadata); err == nil {
			s.metadata = metadata
		}
	}
	
	return nil
}

func (s *InMemoryStorage) Clear() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.keys = make(map[string]*KeyPair)
	s.dids = make(map[string]*DIDRecord)
	s.credentials = make(map[string]*CredentialRecord)
	s.presentations = make(map[string]*PresentationRecord)
	s.metadata = make(map[string]interface{})
	
	return nil
}

// Helper methods for filtering credentials
func (s *InMemoryStorage) matchesFilter(record *CredentialRecord, filter *CredentialFilter) bool {
	if filter == nil {
		return true
	}
	
	// Filter by issuer
	if filter.Issuer != "" && record.Issuer != filter.Issuer {
		return false
	}
	
	// Filter by subject
	if filter.Subject != "" && record.Subject != filter.Subject {
		return false
	}
	
	// Filter by status
	if filter.Status != "" && record.Status != filter.Status {
		return false
	}
	
	// Filter by type
	if len(filter.Type) > 0 {
		hasMatchingType := false
		for _, filterType := range filter.Type {
			for _, recordType := range record.Type {
				if filterType == recordType {
					hasMatchingType = true
					break
				}
			}
			if hasMatchingType {
				break
			}
		}
		if !hasMatchingType {
			return false
		}
	}
	
	// Filter by tags
	if len(filter.Tags) > 0 {
		hasMatchingTag := false
		for _, filterTag := range filter.Tags {
			for _, recordTag := range record.Tags {
				if filterTag == recordTag {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if !hasMatchingTag {
			return false
		}
	}
	
	// Filter by issuance date range
	if filter.IssuedAfter != nil && record.IssuanceDate.Before(*filter.IssuedAfter) {
		return false
	}
	
	if filter.IssuedBefore != nil && record.IssuanceDate.After(*filter.IssuedBefore) {
		return false
	}
	
	// Filter by expiration date range
	if record.ExpirationDate != nil {
		if filter.ExpiresAfter != nil && record.ExpirationDate.Before(*filter.ExpiresAfter) {
			return false
		}
		
		if filter.ExpiresBefore != nil && record.ExpirationDate.After(*filter.ExpiresBefore) {
			return false
		}
	}
	
	return true
}

// Helper methods for deep cloning records
func (s *InMemoryStorage) cloneKeyPair(original *KeyPair) *KeyPair {
	if original == nil {
		return nil
	}
	
	cloned := &KeyPair{
		ID:          original.ID,
		KeyType:     original.KeyType,
		PublicKey:   original.PublicKey,   // Note: shallow copy of interface{}
		PrivateKey:  original.PrivateKey,  // Note: shallow copy of interface{}
		Algorithm:   original.Algorithm,
		Created:     original.Created,
		Usage:       make([]KeyUsage, len(original.Usage)),
		Metadata:    make(map[string]interface{}),
	}
	
	// Copy usage slice
	copy(cloned.Usage, original.Usage)
	
	// Deep copy metadata
	for k, v := range original.Metadata {
		cloned.Metadata[k] = v
	}
	
	// Clone JWKs if present
	if original.PublicKeyJWK != nil {
		cloned.PublicKeyJWK = &(*original.PublicKeyJWK)
	}
	if original.PrivateKeyJWK != nil {
		cloned.PrivateKeyJWK = &(*original.PrivateKeyJWK)
	}
	
	return cloned
}

func (s *InMemoryStorage) cloneDIDRecord(original *DIDRecord) *DIDRecord {
	if original == nil {
		return nil
	}
	
	cloned := &DIDRecord{
		DID:      original.DID,
		Document: original.Document, // Note: shallow copy
		Method:   original.Method,
		KeyID:    original.KeyID,
		Created:  original.Created,
		Updated:  original.Updated,
		Status:   original.Status,
		Metadata: make(map[string]interface{}),
	}
	
	// Deep copy metadata
	for k, v := range original.Metadata {
		cloned.Metadata[k] = v
	}
	
	return cloned
}

func (s *InMemoryStorage) cloneCredentialRecord(original *CredentialRecord) *CredentialRecord {
	if original == nil {
		return nil
	}
	
	cloned := &CredentialRecord{
		ID:             original.ID,
		Credential:     original.Credential, // Note: shallow copy
		CredentialJWT:  original.CredentialJWT,
		Issuer:         original.Issuer,
		Subject:        original.Subject,
		IssuanceDate:   original.IssuanceDate,
		Status:         original.Status,
		Type:           make([]string, len(original.Type)),
		Tags:           make([]string, len(original.Tags)),
		Created:        original.Created,
		Metadata:       make(map[string]interface{}),
	}
	
	// Copy slices
	copy(cloned.Type, original.Type)
	copy(cloned.Tags, original.Tags)
	
	// Copy expiration date if present
	if original.ExpirationDate != nil {
		expDate := *original.ExpirationDate
		cloned.ExpirationDate = &expDate
	}
	
	// Deep copy metadata
	for k, v := range original.Metadata {
		cloned.Metadata[k] = v
	}
	
	return cloned
}

func (s *InMemoryStorage) clonePresentationRecord(original *PresentationRecord) *PresentationRecord {
	if original == nil {
		return nil
	}
	
	cloned := &PresentationRecord{
		ID:              original.ID,
		Presentation:    original.Presentation, // Note: shallow copy
		PresentationJWT: original.PresentationJWT,
		Holder:          original.Holder,
		Verifier:        original.Verifier,
		Created:         original.Created,
		Challenge:       original.Challenge,
		Domain:          original.Domain,
		Purpose:         original.Purpose,
		Credentials:     make([]string, len(original.Credentials)),
		Metadata:        make(map[string]interface{}),
	}
	
	// Copy credentials slice
	copy(cloned.Credentials, original.Credentials)
	
	// Deep copy metadata
	for k, v := range original.Metadata {
		cloned.Metadata[k] = v
	}
	
	return cloned
}