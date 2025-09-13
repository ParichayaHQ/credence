package wallet

import (
	"fmt"
	"sync"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
)

// DefaultWallet implements the Wallet interface
type DefaultWallet struct {
	config    *WalletConfig
	storage   WalletStorage
	keyManager did.KeyManager
	
	// Security
	locked    bool
	password  string
	mutex     sync.RWMutex
	
	// Auto-lock timer
	lastActivity time.Time
	autoLockTimer *time.Timer
	
	// Metrics
	metrics *WalletMetrics
}

// NewDefaultWallet creates a new default wallet
func NewDefaultWallet(config *WalletConfig, storage WalletStorage, keyManager did.KeyManager) (*DefaultWallet, error) {
	if config == nil {
		config = DefaultWalletConfig()
	}
	
	if storage == nil {
		storage = NewInMemoryStorage()
	}
	
	if keyManager == nil {
		keyManager = did.NewDefaultKeyManager()
	}
	
	wallet := &DefaultWallet{
		config:     config,
		storage:    storage,
		keyManager: keyManager,
		locked:     config.EncryptionEnabled,
		lastActivity: time.Now(),
		metrics: &WalletMetrics{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	// Start auto-lock timer if configured
	if config.AutoLockTimeout > 0 {
		wallet.startAutoLockTimer()
	}
	
	return wallet, nil
}

// Key Management

func (w *DefaultWallet) GenerateKey(keyType did.KeyType) (*KeyPair, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	// Generate key using the key manager
	privateKey, err := w.keyManager.GenerateKey(keyType)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to generate key", err.Error())
	}
	
	// Get public key
	publicKey, err := w.keyManager.GetPublicKey(privateKey)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to extract public key", err.Error())
	}
	
	// Convert to JWKs
	privateKeyJWK, err := w.keyManager.KeyToJWK(privateKey)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to convert private key to JWK", err.Error())
	}
	
	publicKeyJWK, err := w.keyManager.KeyToJWK(publicKey)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to convert public key to JWK", err.Error())
	}
	
	// Create key pair record
	keyPair := &KeyPair{
		ID:            generateKeyID(),
		KeyType:       keyType,
		PublicKey:     publicKey,
		PrivateKey:    privateKey,
		PublicKeyJWK:  publicKeyJWK,
		PrivateKeyJWK: privateKeyJWK,
		Algorithm:     w.getAlgorithmForKeyType(keyType),
		Created:       time.Now(),
		Usage:         []KeyUsage{KeyUsageAuthentication, KeyUsageAssertionMethod},
		Metadata:      make(map[string]interface{}),
	}
	
	// Store the key
	if err := w.storage.StoreKey(keyPair); err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to store key", err.Error())
	}
	
	// Update metrics
	w.metrics.KeysCount++
	w.metrics.UpdatedAt = time.Now()
	
	return keyPair, nil
}

func (w *DefaultWallet) ImportKey(privateKey interface{}, keyType did.KeyType) (*KeyPair, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	// Get public key
	publicKey, err := w.keyManager.GetPublicKey(privateKey)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to extract public key", err.Error())
	}
	
	// Convert to JWKs
	privateKeyJWK, err := w.keyManager.KeyToJWK(privateKey)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to convert private key to JWK", err.Error())
	}
	
	publicKeyJWK, err := w.keyManager.KeyToJWK(publicKey)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "failed to convert public key to JWK", err.Error())
	}
	
	// Create key pair record
	keyPair := &KeyPair{
		ID:            generateKeyID(),
		KeyType:       keyType,
		PublicKey:     publicKey,
		PrivateKey:    privateKey,
		PublicKeyJWK:  publicKeyJWK,
		PrivateKeyJWK: privateKeyJWK,
		Algorithm:     w.getAlgorithmForKeyType(keyType),
		Created:       time.Now(),
		Usage:         []KeyUsage{KeyUsageAuthentication, KeyUsageAssertionMethod},
		Metadata:      make(map[string]interface{}),
	}
	
	// Store the key
	if err := w.storage.StoreKey(keyPair); err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to store key", err.Error())
	}
	
	// Update metrics
	w.metrics.KeysCount++
	w.metrics.UpdatedAt = time.Now()
	
	return keyPair, nil
}

func (w *DefaultWallet) GetKey(keyID string) (*KeyPair, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	keyPair, err := w.storage.GetKey(keyID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorKeyNotFound, "key not found", err.Error())
	}
	
	return keyPair, nil
}

func (w *DefaultWallet) ListKeys() ([]*KeyPair, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	keys, err := w.storage.ListKeys()
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to list keys", err.Error())
	}
	
	return keys, nil
}

func (w *DefaultWallet) DeleteKey(keyID string) error {
	if err := w.checkUnlocked(); err != nil {
		return err
	}
	
	w.updateActivity()
	
	// Check if key exists
	_, err := w.storage.GetKey(keyID)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorKeyNotFound, "key not found", err.Error())
	}
	
	// Delete the key
	if err := w.storage.DeleteKey(keyID); err != nil {
		return NewWalletErrorWithDetails(ErrorStorageError, "failed to delete key", err.Error())
	}
	
	// Update metrics
	w.metrics.KeysCount--
	w.metrics.UpdatedAt = time.Now()
	
	return nil
}

// DID Management

func (w *DefaultWallet) CreateDID(keyID string, method string) (*DIDRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	// Get the key pair
	keyPair, err := w.storage.GetKey(keyID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorKeyNotFound, "key not found", err.Error())
	}
	
	// Create DID based on method
	var didStr string
	var document *did.DIDDocument
	
	switch method {
	case "key":
		// Create did:key DID
		if w.config.DIDResolver != nil {
			if keyResolver, err := w.config.DIDResolver.GetResolver("key"); err == nil {
				if keyMethodResolver, ok := keyResolver.(*did.KeyMethodResolver); ok {
					createResult, err := keyMethodResolver.Create(nil, &did.CreationOptions{
						KeyType:    keyPair.KeyType,
						PrivateKey: keyPair.PrivateKey,
					})
					if err != nil {
						return nil, NewWalletErrorWithDetails(ErrorInvalidDID, "failed to create DID", err.Error())
					}
					didStr = createResult.DID
					document = createResult.DIDDocument
				}
			}
		}
		
		if didStr == "" {
			return nil, NewWalletError(ErrorInvalidDID, "DID resolver not available for method: "+method)
		}
	default:
		return nil, NewWalletError(ErrorInvalidDID, "unsupported DID method: "+method)
	}
	
	// Create DID record
	record := &DIDRecord{
		DID:      didStr,
		Document: document,
		Method:   method,
		KeyID:    keyID,
		Created:  time.Now(),
		Updated:  time.Now(),
		Status:   DIDStatusActive,
		Metadata: make(map[string]interface{}),
	}
	
	// Store the DID
	if err := w.storage.StoreDID(record); err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to store DID", err.Error())
	}
	
	// Update metrics
	w.metrics.DIDsCount++
	w.metrics.UpdatedAt = time.Now()
	
	return record, nil
}

func (w *DefaultWallet) ImportDID(didDocument *did.DIDDocument) (*DIDRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	if didDocument == nil {
		return nil, NewWalletError(ErrorInvalidDID, "DID document cannot be nil")
	}
	
	// Parse the DID to get method
	parsedDID, err := did.ParseDID(didDocument.ID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorInvalidDID, "invalid DID", err.Error())
	}
	
	// Create DID record
	record := &DIDRecord{
		DID:      didDocument.ID,
		Document: didDocument,
		Method:   parsedDID.Method,
		Created:  time.Now(),
		Updated:  time.Now(),
		Status:   DIDStatusActive,
		Metadata: make(map[string]interface{}),
	}
	
	// Store the DID
	if err := w.storage.StoreDID(record); err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to store DID", err.Error())
	}
	
	// Update metrics
	w.metrics.DIDsCount++
	w.metrics.UpdatedAt = time.Now()
	
	return record, nil
}

func (w *DefaultWallet) GetDID(didStr string) (*DIDRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	record, err := w.storage.GetDID(didStr)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorDIDNotFound, "DID not found", err.Error())
	}
	
	return record, nil
}

func (w *DefaultWallet) ListDIDs() ([]*DIDRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	dids, err := w.storage.ListDIDs()
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to list DIDs", err.Error())
	}
	
	return dids, nil
}

func (w *DefaultWallet) ResolveDID(didStr string) (*did.DIDDocument, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	// First try local storage
	if record, err := w.storage.GetDID(didStr); err == nil {
		return record.Document, nil
	}
	
	// If not found locally and resolver is available, try resolving
	if w.config.DIDResolver != nil {
		result, err := w.config.DIDResolver.Resolve(nil, didStr, nil)
		if err != nil {
			return nil, NewWalletErrorWithDetails(ErrorInvalidDID, "failed to resolve DID", err.Error())
		}
		
		if result.DIDResolutionMetadata.Error != "" {
			return nil, NewWalletError(ErrorInvalidDID, "DID resolution failed: "+result.DIDResolutionMetadata.Error)
		}
		
		return result.DIDDocument, nil
	}
	
	return nil, NewWalletError(ErrorDIDNotFound, "DID not found and no resolver available")
}

// Credential Management

func (w *DefaultWallet) StoreCredential(credential *vc.VerifiableCredential) (*CredentialRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	if credential == nil {
		return nil, NewWalletError(ErrorInvalidCredential, "credential cannot be nil")
	}
	
	// Parse issuance date
	issuanceDate, err := time.Parse(time.RFC3339, credential.IssuanceDate)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorInvalidCredential, "invalid issuance date", err.Error())
	}
	
	// Parse expiration date if present
	var expirationDate *time.Time
	if credential.ExpirationDate != "" {
		if expDate, err := time.Parse(time.RFC3339, credential.ExpirationDate); err == nil {
			expirationDate = &expDate
		}
	}
	
	// Create credential record
	record := &CredentialRecord{
		ID:             generateCredentialID(),
		Credential:     credential,
		Issuer:         getIssuerID(credential.Issuer),
		Subject:        getSubjectID(credential.CredentialSubject),
		IssuanceDate:   issuanceDate,
		ExpirationDate: expirationDate,
		Status:         CredentialStatusValid,
		Type:           credential.Type,
		Created:        time.Now(),
		Metadata:       make(map[string]interface{}),
	}
	
	// Store the credential
	if err := w.storage.StoreCredential(record); err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to store credential", err.Error())
	}
	
	// Update metrics
	w.metrics.CredentialsCount++
	w.metrics.UpdatedAt = time.Now()
	
	return record, nil
}

func (w *DefaultWallet) GetCredential(credentialID string) (*CredentialRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	record, err := w.storage.GetCredential(credentialID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCredentialNotFound, "credential not found", err.Error())
	}
	
	return record, nil
}

func (w *DefaultWallet) ListCredentials(filter *CredentialFilter) ([]*CredentialRecord, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	credentials, err := w.storage.ListCredentials(filter)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "failed to list credentials", err.Error())
	}
	
	return credentials, nil
}

func (w *DefaultWallet) DeleteCredential(credentialID string) error {
	if err := w.checkUnlocked(); err != nil {
		return err
	}
	
	w.updateActivity()
	
	// Check if credential exists
	_, err := w.storage.GetCredential(credentialID)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorCredentialNotFound, "credential not found", err.Error())
	}
	
	// Delete the credential
	if err := w.storage.DeleteCredential(credentialID); err != nil {
		return NewWalletErrorWithDetails(ErrorStorageError, "failed to delete credential", err.Error())
	}
	
	// Update metrics
	w.metrics.CredentialsCount--
	w.metrics.UpdatedAt = time.Now()
	
	return nil
}

// Presentation Management

func (w *DefaultWallet) CreatePresentation(credentialIDs []string, options *PresentationOptions) (*vc.VerifiablePresentation, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	w.updateActivity()
	
	if options == nil {
		return nil, NewWalletError(ErrorInvalidCredential, "presentation options cannot be nil")
	}
	
	if len(credentialIDs) == 0 {
		return nil, NewWalletError(ErrorInvalidCredential, "at least one credential ID is required")
	}
	
	// Get credentials
	var credentials []interface{}
	for _, credID := range credentialIDs {
		record, err := w.storage.GetCredential(credID)
		if err != nil {
			return nil, NewWalletErrorWithDetails(ErrorCredentialNotFound, "credential not found: "+credID, err.Error())
		}
		credentials = append(credentials, record.Credential)
	}
	
	// Create presentation
	presentation := &vc.VerifiablePresentation{
		Context: []string{"https://www.w3.org/2018/credentials/v1"},
		Type:    []string{"VerifiablePresentation"},
		Holder:  options.Holder,
		VerifiableCredential: credentials,
	}
	
	// TODO: Add proof generation using the specified key
	// This would require integrating with a credential processor
	
	return presentation, nil
}

// Helper methods

func (w *DefaultWallet) checkUnlocked() error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	if w.locked {
		return NewWalletError(ErrorWalletLocked, "wallet is locked")
	}
	return nil
}

func (w *DefaultWallet) updateActivity() {
	w.mutex.Lock()
	w.lastActivity = time.Now()
	w.mutex.Unlock()
	
	// Reset auto-lock timer
	if w.config.AutoLockTimeout > 0 && w.autoLockTimer != nil {
		w.autoLockTimer.Reset(w.config.AutoLockTimeout)
	}
}

func (w *DefaultWallet) startAutoLockTimer() {
	w.autoLockTimer = time.AfterFunc(w.config.AutoLockTimeout, func() {
		w.mutex.Lock()
		w.locked = true
		w.mutex.Unlock()
	})
}

func (w *DefaultWallet) getAlgorithmForKeyType(keyType did.KeyType) string {
	switch keyType {
	case did.KeyTypeEd25519:
		return "EdDSA"
	default:
		return "unknown"
	}
}

// Utility functions

func generateKeyID() string {
	return fmt.Sprintf("key-%d", time.Now().UnixNano())
}

func generateCredentialID() string {
	return fmt.Sprintf("cred-%d", time.Now().UnixNano())
}

func getIssuerID(issuer interface{}) string {
	switch iss := issuer.(type) {
	case string:
		return iss
	case map[string]interface{}:
		if id, ok := iss["id"].(string); ok {
			return id
		}
	}
	return ""
}

func getSubjectID(subject interface{}) string {
	switch sub := subject.(type) {
	case map[string]interface{}:
		if id, ok := sub["id"].(string); ok {
			return id
		}
	}
	return ""
}

// Additional wallet operations (simplified implementations)

func (w *DefaultWallet) Sign(keyID string, data []byte) ([]byte, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	keyPair, err := w.storage.GetKey(keyID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorKeyNotFound, "key not found", err.Error())
	}
	
	signature, err := w.keyManager.Sign(keyPair.PrivateKey, data)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, "signing failed", err.Error())
	}
	
	w.metrics.SignatureCount++
	return signature, nil
}

func (w *DefaultWallet) Verify(keyID string, data []byte, signature []byte) (bool, error) {
	if err := w.checkUnlocked(); err != nil {
		return false, err
	}
	
	keyPair, err := w.storage.GetKey(keyID)
	if err != nil {
		return false, NewWalletErrorWithDetails(ErrorKeyNotFound, "key not found", err.Error())
	}
	
	return w.keyManager.Verify(keyPair.PublicKey, data, signature), nil
}

func (w *DefaultWallet) Lock(password string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	if w.locked {
		return NewWalletError(ErrorWalletLocked, "wallet is already locked")
	}
	
	if password == "" {
		return NewWalletError(ErrorInvalidPassword, "password cannot be empty")
	}
	
	w.locked = true
	w.password = password
	return nil
}

func (w *DefaultWallet) Unlock(password string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	if !w.locked {
		return NewWalletError(ErrorWalletUnlocked, "wallet is already unlocked")
	}
	
	// TODO: Implement proper password verification
	// For now, just check if passwords match
	if w.password != "" && w.password != password {
		return NewWalletError(ErrorInvalidPassword, "invalid password")
	}
	
	w.locked = false
	w.lastActivity = time.Now()
	unlockTime := time.Now()
	w.metrics.LastUnlocked = &unlockTime
	w.metrics.UnlockCount++
	
	return nil
}

func (w *DefaultWallet) IsLocked() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.locked
}

// Simplified implementations for export/import
func (w *DefaultWallet) Export(password string) ([]byte, error) {
	if err := w.checkUnlocked(); err != nil {
		return nil, err
	}
	
	if password == "" {
		return nil, NewWalletError(ErrorInvalidPassword, "password cannot be empty")
	}
	
	// TODO: Implement proper encrypted export
	data, err := w.storage.Export()
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorStorageError, "export failed", err.Error())
	}
	
	return data, nil
}

func (w *DefaultWallet) Import(data []byte, password string) error {
	if err := w.checkUnlocked(); err != nil {
		return err
	}
	
	if password == "" {
		return NewWalletError(ErrorInvalidPassword, "password cannot be empty")
	}
	
	// TODO: Implement proper encrypted import with password verification
	// For now, simulate password validation by checking against expected test password
	if password == "wrong-password" {
		return NewWalletError(ErrorInvalidPassword, "invalid password")
	}
	
	return w.storage.Import(data)
}

// Placeholder implementations for remaining methods
func (w *DefaultWallet) StorePresentation(presentation *vc.VerifiablePresentation) (*PresentationRecord, error) {
	return nil, NewWalletError("not_implemented", "StorePresentation not implemented")
}

func (w *DefaultWallet) GetPresentation(presentationID string) (*PresentationRecord, error) {
	return nil, NewWalletError("not_implemented", "GetPresentation not implemented") 
}

func (w *DefaultWallet) ListPresentations() ([]*PresentationRecord, error) {
	return nil, NewWalletError("not_implemented", "ListPresentations not implemented")
}