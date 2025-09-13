package wallet

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefaultWallet_GenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		keyType  did.KeyType
		locked   bool
		wantErr  bool
		errCode  string
	}{
		{
			name:    "generate Ed25519 key",
			keyType: did.KeyTypeEd25519,
			locked:  false,
			wantErr: false,
		},
		{
			name:    "generate secp256k1 key",
			keyType: did.KeyTypeSecp256k1,
			locked:  false,
			wantErr: false,
		},
		{
			name:    "wallet locked",
			keyType: did.KeyTypeEd25519,
			locked:  true,
			wantErr: true,
			errCode: ErrorWalletLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletWithMocks(t)
			
			if tt.locked {
				wallet.Lock("password")
			}
			
			keyPair, err := wallet.GenerateKey(tt.keyType)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				assert.Nil(t, keyPair)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, keyPair)
				assert.Equal(t, tt.keyType, keyPair.KeyType)
				assert.NotEmpty(t, keyPair.ID)
				assert.NotZero(t, keyPair.Created)
			}
		})
	}
}

func TestDefaultWallet_CreateDID(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		method     string
		setupKey   bool
		locked     bool
		wantErr    bool
		errCode    string
	}{
		{
			name:     "create did:key",
			keyID:    "key-1",
			method:   "key",
			setupKey: true,
			locked:   false,
			wantErr:  false,
		},
		{
			name:     "key not found",
			keyID:    "non-existent",
			method:   "key",
			setupKey: false,
			locked:   false,
			wantErr:  true,
			errCode:  ErrorKeyNotFound,
		},
		{
			name:     "wallet locked",
			keyID:    "key-1",
			method:   "key",
			setupKey: true,
			locked:   true,
			wantErr:  true,
			errCode:  ErrorWalletLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletWithMocks(t)
			
			if tt.setupKey {
				keyPair := &KeyPair{
					ID:      tt.keyID,
					KeyType: did.KeyTypeEd25519,
					Created: time.Now(),
				}
				wallet.(*DefaultWallet).storage.StoreKey(keyPair)
			}
			
			if tt.locked {
				wallet.Lock("password")
			}
			
			didRecord, err := wallet.CreateDID(tt.keyID, tt.method)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				assert.Nil(t, didRecord)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, didRecord)
				assert.Equal(t, tt.keyID, didRecord.KeyID)
				assert.Equal(t, tt.method, didRecord.Method)
				assert.NotEmpty(t, didRecord.DID)
				assert.NotZero(t, didRecord.Created)
			}
		})
	}
}

func TestDefaultWallet_StoreCredential(t *testing.T) {
	tests := []struct {
		name       string
		credential *vc.VerifiableCredential
		locked     bool
		wantErr    bool
		errCode    string
	}{
		{
			name: "store valid credential",
			credential: &vc.VerifiableCredential{
				Context:      []string{"https://www.w3.org/2018/credentials/v1"},
				ID:           "cred-1",
				Type:         []string{"VerifiableCredential"},
				Issuer:       "did:key:issuer",
				IssuanceDate: time.Now().Format(time.RFC3339),
				CredentialSubject: map[string]interface{}{
					"id": "did:key:subject",
				},
			},
			locked:  false,
			wantErr: false,
		},
		{
			name:       "nil credential",
			credential: nil,
			locked:     false,
			wantErr:    true,
			errCode:    ErrorInvalidCredential,
		},
		{
			name: "wallet locked",
			credential: &vc.VerifiableCredential{
				Context:      []string{"https://www.w3.org/2018/credentials/v1"},
				Type:         []string{"VerifiableCredential"},
				Issuer:       "did:key:issuer",
				IssuanceDate: time.Now().Format(time.RFC3339),
			},
			locked:  true,
			wantErr: true,
			errCode: ErrorWalletLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletWithMocks(t)
			
			if tt.locked {
				wallet.Lock("password")
			}
			
			credRecord, err := wallet.StoreCredential(tt.credential)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				assert.Nil(t, credRecord)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, credRecord)
				assert.NotEmpty(t, credRecord.ID)
				assert.Equal(t, tt.credential, credRecord.Credential)
				assert.Equal(t, CredentialStatusValid, credRecord.Status)
				assert.NotZero(t, credRecord.Created)
			}
		})
	}
}

func TestDefaultWallet_CreatePresentation(t *testing.T) {
	tests := []struct {
		name          string
		credentialIDs []string
		options       *PresentationOptions
		setupCreds    func(Wallet) []string // Returns credential IDs
		locked        bool
		wantErr       bool
		errCode       string
	}{
		{
			name:          "create valid presentation",
			credentialIDs: []string{}, // Will be populated by setupCreds
			options: &PresentationOptions{
				Holder:    "did:key:holder",
				KeyID:     "key-1",
				Algorithm: "EdDSA",
				Challenge: "test-challenge",
			},
			setupCreds: func(wallet Wallet) []string {
				// Add credential
				cred := &vc.VerifiableCredential{
					Context:      []string{"https://www.w3.org/2018/credentials/v1"},
					ID:           "cred-1",
					Type:         []string{"VerifiableCredential"},
					Issuer:       "did:key:issuer",
					IssuanceDate: time.Now().Format(time.RFC3339),
					CredentialSubject: map[string]interface{}{
						"id": "did:key:subject",
					},
				}
				credRecord, _ := wallet.StoreCredential(cred)
				
				// Add key
				keyPair := &KeyPair{
					ID:      "key-1",
					KeyType: did.KeyTypeEd25519,
				}
				wallet.(*DefaultWallet).storage.StoreKey(keyPair)
				
				return []string{credRecord.ID}
			},
			locked:  false,
			wantErr: false,
		},
		{
			name:          "empty credential IDs",
			credentialIDs: []string{},
			options: &PresentationOptions{
				Holder: "did:key:holder",
				KeyID:  "key-1",
			},
			setupCreds: func(wallet Wallet) []string {
				return []string{}
			},
			locked:  false,
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name:          "wallet locked",
			credentialIDs: []string{"cred-1"},
			options: &PresentationOptions{
				Holder: "did:key:holder",
				KeyID:  "key-1",
			},
			setupCreds: func(wallet Wallet) []string {
				return []string{}
			},
			locked:  true,
			wantErr: true,
			errCode: ErrorWalletLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletWithMocks(t)
			
			credentialIDs := tt.setupCreds(wallet)
			if len(credentialIDs) > 0 {
				// Use the actual credential IDs returned by setupCreds
				tt.credentialIDs = credentialIDs
			}
			
			if tt.locked {
				wallet.Lock("password")
			}
			
			presentation, err := wallet.CreatePresentation(tt.credentialIDs, tt.options)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				assert.Nil(t, presentation)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, presentation)
				assert.Equal(t, tt.options.Holder, presentation.Holder)
				assert.Len(t, presentation.VerifiableCredential, len(tt.credentialIDs))
			}
		})
	}
}

func TestDefaultWallet_LockUnlock(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "valid password",
			password: "test-password",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
			errCode:  ErrorInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletWithMocks(t)
			
			// Test lock
			err := wallet.Lock(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				return
			}
			
			require.NoError(t, err)
			assert.True(t, wallet.IsLocked())
			
			// Test unlock with correct password
			err = wallet.Unlock(tt.password)
			require.NoError(t, err)
			assert.False(t, wallet.IsLocked())
			
			// Test unlock with wrong password
			wallet.Lock(tt.password)
			err = wallet.Unlock("wrong-password")
			require.Error(t, err)
			var walletErr *WalletError
			require.ErrorAs(t, err, &walletErr)
			assert.Equal(t, ErrorInvalidPassword, walletErr.Code)
			assert.True(t, wallet.IsLocked())
		})
	}
}

func TestDefaultWallet_Export(t *testing.T) {
	tests := []struct {
		name     string
		password string
		addData  func(Wallet)
		wantErr  bool
		errCode  string
	}{
		{
			name:     "export with data",
			password: "export-password",
			addData: func(wallet Wallet) {
				// Add key
				keyPair, _ := wallet.GenerateKey(did.KeyTypeEd25519)
				
				// Add DID
				wallet.CreateDID(keyPair.ID, "key")
				
				// Add credential
				cred := &vc.VerifiableCredential{
					Context:      []string{"https://www.w3.org/2018/credentials/v1"},
					Type:         []string{"VerifiableCredential"},
					Issuer:       "did:key:issuer",
					IssuanceDate: time.Now().Format(time.RFC3339),
					CredentialSubject: map[string]interface{}{
						"id": "did:key:subject",
					},
				}
				wallet.StoreCredential(cred)
			},
			wantErr: false,
		},
		{
			name:     "empty password",
			password: "",
			addData:  func(wallet Wallet) {},
			wantErr:  true,
			errCode:  ErrorInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletWithMocks(t)
			
			tt.addData(wallet)
			
			data, err := wallet.Export(tt.password)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, data)
				assert.NotEmpty(t, data)
			}
		})
	}
}

func TestDefaultWallet_Import(t *testing.T) {
	wallet1 := setupTestWalletWithMocks(t)
	
	// Add some data to first wallet
	keyPair, err := wallet1.GenerateKey(did.KeyTypeEd25519)
	require.NoError(t, err)
	
	didRecord, err := wallet1.CreateDID(keyPair.ID, "key")
	require.NoError(t, err)
	
	cred := &vc.VerifiableCredential{
		Context:      []string{"https://www.w3.org/2018/credentials/v1"},
		Type:         []string{"VerifiableCredential"},
		Issuer:       "did:key:issuer",
		IssuanceDate: time.Now().Format(time.RFC3339),
		CredentialSubject: map[string]interface{}{
			"id": "did:key:subject",
		},
	}
	credRecord, err := wallet1.StoreCredential(cred)
	require.NoError(t, err)
	
	// Export data
	password := "test-password"
	exportData, err := wallet1.Export(password)
	require.NoError(t, err)
	
	tests := []struct {
		name     string
		data     []byte
		password string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "valid import",
			data:     exportData,
			password: password,
			wantErr:  false,
		},
		{
			name:     "wrong password",
			data:     exportData,
			password: "wrong-password",
			wantErr:  true,
			errCode:  ErrorInvalidPassword,
		},
		{
			name:     "invalid data",
			data:     []byte("invalid"),
			password: password,
			wantErr:  true,
			errCode:  ErrorSerializationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet2 := setupTestWalletWithMocks(t)
			
			err := wallet2.Import(tt.data, tt.password)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
			} else {
				require.NoError(t, err)
				
				// Verify data was imported
				importedKey, err := wallet2.GetKey(keyPair.ID)
				require.NoError(t, err)
				assert.Equal(t, keyPair.ID, importedKey.ID)
				
				importedDID, err := wallet2.GetDID(didRecord.DID)
				require.NoError(t, err)
				assert.Equal(t, didRecord.DID, importedDID.DID)
				
				importedCred, err := wallet2.GetCredential(credRecord.ID)
				require.NoError(t, err)
				assert.Equal(t, credRecord.ID, importedCred.ID)
			}
		})
	}
}

func TestDefaultWallet_AutoLock(t *testing.T) {
	config := DefaultWalletConfig()
	config.AutoLockTimeout = 100 * time.Millisecond // Short timeout for testing
	
	storage := NewInMemoryStorage()
	keyManager := &MockKeyManager{}
	
	wallet, err := NewDefaultWallet(config, storage, keyManager)
	require.NoError(t, err)
	
	// Wallet should start unlocked
	assert.False(t, wallet.IsLocked())
	
	// Wait for auto-lock
	time.Sleep(150 * time.Millisecond)
	
	// Wallet should now be locked
	assert.True(t, wallet.IsLocked())
	
	// Operations should fail
	_, err = wallet.GenerateKey(did.KeyTypeEd25519)
	require.Error(t, err)
	
	var walletErr *WalletError
	require.ErrorAs(t, err, &walletErr)
	assert.Equal(t, ErrorWalletLocked, walletErr.Code)
}

func TestDefaultWallet_ActivityTracking(t *testing.T) {
	wallet := setupTestWalletWithMocks(t)
	
	// Initial metrics
	initialMetrics := wallet.(*DefaultWallet).metrics
	assert.Equal(t, int64(0), initialMetrics.UnlockCount)
	assert.Equal(t, int64(0), initialMetrics.SignatureCount)
	
	// Generate key (should increment activity)
	_, err := wallet.GenerateKey(did.KeyTypeEd25519)
	require.NoError(t, err)
	
	// Check metrics updated
	updatedMetrics := wallet.(*DefaultWallet).metrics
	assert.Equal(t, int64(0), updatedMetrics.UnlockCount) // Still 0, no lock/unlock
	assert.Equal(t, 1, updatedMetrics.KeysCount)
	
	// Lock and unlock
	err = wallet.Lock("password")
	require.NoError(t, err)
	
	err = wallet.Unlock("password")
	require.NoError(t, err)
	
	// Check unlock count incremented
	finalMetrics := wallet.(*DefaultWallet).metrics
	assert.Equal(t, int64(1), finalMetrics.UnlockCount)
	assert.NotNil(t, finalMetrics.LastUnlocked)
}

func TestDefaultWallet_ListOperations(t *testing.T) {
	wallet := setupTestWalletWithMocks(t)
	
	// Add test data
	keyPair1, err := wallet.GenerateKey(did.KeyTypeEd25519)
	require.NoError(t, err)
	
	keyPair2, err := wallet.GenerateKey(did.KeyTypeSecp256k1)
	require.NoError(t, err)
	
	_, err = wallet.CreateDID(keyPair1.ID, "key")
	require.NoError(t, err)
	
	_, err = wallet.CreateDID(keyPair2.ID, "key")
	require.NoError(t, err)
	
	cred1 := &vc.VerifiableCredential{
		Context:      []string{"https://www.w3.org/2018/credentials/v1"},
		Type:         []string{"VerifiableCredential", "TestCredential"},
		Issuer:       "did:key:issuer1",
		IssuanceDate: time.Now().Format(time.RFC3339),
		CredentialSubject: map[string]interface{}{
			"id": "did:key:subject1",
		},
	}
	credRecord1, err := wallet.StoreCredential(cred1)
	require.NoError(t, err)
	
	cred2 := &vc.VerifiableCredential{
		Context:      []string{"https://www.w3.org/2018/credentials/v1"},
		Type:         []string{"VerifiableCredential", "AnotherCredential"},
		Issuer:       "did:key:issuer2",
		IssuanceDate: time.Now().Format(time.RFC3339),
		CredentialSubject: map[string]interface{}{
			"id": "did:key:subject2",
		},
	}
	credRecord2, err := wallet.StoreCredential(cred2)
	require.NoError(t, err)
	
	// Test list keys
	keys, err := wallet.ListKeys()
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	
	// Test list DIDs
	dids, err := wallet.ListDIDs()
	require.NoError(t, err)
	assert.Len(t, dids, 2)
	
	// Test list all credentials
	allCreds, err := wallet.ListCredentials(nil)
	require.NoError(t, err)
	assert.Len(t, allCreds, 2)
	
	// Test filtered credentials by type
	filter := &CredentialFilter{
		Type: []string{"TestCredential"},
	}
	filteredCreds, err := wallet.ListCredentials(filter)
	require.NoError(t, err)
	assert.Len(t, filteredCreds, 1)
	assert.Equal(t, credRecord1.ID, filteredCreds[0].ID)
	
	// Test filtered credentials by issuer
	filter = &CredentialFilter{
		Issuer: "did:key:issuer2",
	}
	filteredCreds, err = wallet.ListCredentials(filter)
	require.NoError(t, err)
	assert.Len(t, filteredCreds, 1)
	assert.Equal(t, credRecord2.ID, filteredCreds[0].ID)
}

// Helper function to setup test wallet with mocks
func setupTestWalletWithMocks(t *testing.T) Wallet {
	config := DefaultWalletConfig()
	config.StorageType = "memory"
	config.AutoLockTimeout = 0 // Disable auto-lock for tests
	config.EncryptionEnabled = false // Start unlocked for tests
	
	storage := NewInMemoryStorage()
	keyManager := &MockKeyManager{}
	
	// Generate a test Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	
	// Setup mock key manager
	keyManager.On("GenerateKey", mock.AnythingOfType("did.KeyType")).Return(privateKey, nil)
	keyManager.On("GetPublicKey", privateKey).Return(publicKey, nil)
	keyManager.On("KeyToJWK", mock.Anything).Return(&did.JWK{Kty: "OKP"}, nil)
	keyManager.On("Sign", mock.Anything, mock.AnythingOfType("[]uint8")).Return([]byte("signature"), nil)
	
	// Setup DID resolver
	didResolver := &MockDIDResolver{}
	keyResolver := did.NewKeyMethodResolver(keyManager)
	didResolver.On("GetResolver", "key").Return(keyResolver, nil)
	config.DIDResolver = didResolver
	
	wallet, err := NewDefaultWallet(config, storage, keyManager)
	require.NoError(t, err)
	
	return wallet
}