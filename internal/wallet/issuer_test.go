package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIssuerService_IssueCredential(t *testing.T) {
	tests := []struct {
		name        string
		request     *IssuanceRequest
		setupMocks  func(*MockDIDResolver, *MockCredentialIssuer, Wallet)
		wantErr     bool
		errCode     string
	}{
		{
			name: "valid issuance request",
			request: &IssuanceRequest{
				ID:       "cred-123",
				Context:  []string{"https://www.w3.org/2018/credentials/v1"},
				Type:     []string{"VerifiableCredential", "TestCredential"},
				Issuer:   "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				CredentialSubject: map[string]interface{}{
					"id":   "did:key:subject",
					"name": "Test User",
				},
				IssuanceDate:     time.Now(),
				SigningKeyID:     "key-1",
				Algorithm:        "EdDSA",
				EnableRevocation: false,
				StoreInWallet:    true,
			},
			setupMocks: func(resolver *MockDIDResolver, issuer *MockCredentialIssuer, wallet Wallet) {
				// Mock DID resolution
				resolver.On("Resolve", context.Background(), "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH", mock.AnythingOfType("*did.DIDResolutionOptions")).Return(&did.DIDResolutionResult{
					DIDDocument: &did.DIDDocument{
						ID: "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
					},
					DIDResolutionMetadata: did.DIDResolutionMetadata{},
				}, nil)

				// Add key to wallet
				keyPair := &KeyPair{
					ID:      "key-1",
					KeyType: did.KeyTypeEd25519,
				}
				wallet.(*DefaultWallet).storage.StoreKey(keyPair)

				// Mock credential issuance
				credential := &vc.VerifiableCredential{
					ID:           "cred-123",
					Context:      []string{"https://www.w3.org/2018/credentials/v1"},
					Type:         []string{"VerifiableCredential", "TestCredential"},
					Issuer:       "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
					IssuanceDate: time.Now().Format(time.RFC3339),
					CredentialSubject: map[string]interface{}{
						"id":   "did:key:subject",
						"name": "Test User",
					},
				}
				issuer.On("IssueCredential", 
					mock.AnythingOfType("*vc.CredentialTemplate"), 
					mock.AnythingOfType("*vc.IssuanceOptions")).Return(credential, nil)
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			request: nil,
			setupMocks: func(resolver *MockDIDResolver, issuer *MockCredentialIssuer, wallet Wallet) {
				// No setup needed
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "missing issuer",
			request: &IssuanceRequest{
				Type:             []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
				IssuanceDate:     time.Now(),
				SigningKeyID:     "key-1",
			},
			setupMocks: func(resolver *MockDIDResolver, issuer *MockCredentialIssuer, wallet Wallet) {
				// No setup needed
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "invalid issuer DID",
			request: &IssuanceRequest{
				Issuer:           "invalid-did",
				Type:             []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
				IssuanceDate:     time.Now(),
				SigningKeyID:     "key-1",
			},
			setupMocks: func(resolver *MockDIDResolver, issuer *MockCredentialIssuer, wallet Wallet) {
				// No setup needed
			},
			wantErr: true,
			errCode: ErrorInvalidDID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletForIssuer(t)
			resolver := &MockDIDResolver{}
			issuer := &MockCredentialIssuer{}
			
			tt.setupMocks(resolver, issuer, wallet)
			
			service := NewIssuerService(wallet, resolver, issuer)
			
			credential, err := service.IssueCredential(context.Background(), tt.request)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
				assert.Nil(t, credential)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, credential)
				assert.Equal(t, tt.request.Issuer, credential.Issuer)
				assert.Equal(t, tt.request.Type, credential.Type)
			}
		})
	}
}

func TestIssuerService_RevokeCredential(t *testing.T) {
	tests := []struct {
		name           string
		credentialID   string
		reason         string
		setupWallet    func(Wallet)
		wantErr        bool
		errCode        string
	}{
		{
			name:         "valid revocation",
			credentialID: "cred-1",
			reason:       "Test revocation",
			setupWallet: func(wallet Wallet) {
				// Add credential with status list
				cred := &vc.VerifiableCredential{
					ID:           "cred-1",
					Context:      []string{"https://www.w3.org/2018/credentials/v1"},
					Type:         []string{"VerifiableCredential"},
					IssuanceDate: time.Now().Format(time.RFC3339),
					CredentialSubject: map[string]interface{}{
						"id": "did:key:subject",
					},
					CredentialStatus: &vc.CredentialStatus{
						ID:   "status-list#1",
						Type: "StatusList2021Entry",
					},
				}
				wallet.StoreCredential(cred)
			},
			wantErr: false,
		},
		{
			name:         "empty credential ID",
			credentialID: "",
			reason:       "Test revocation",
			setupWallet:  func(wallet Wallet) {},
			wantErr:      true,
			errCode:      ErrorInvalidCredential,
		},
		{
			name:         "credential not found",
			credentialID: "non-existent",
			reason:       "Test revocation",
			setupWallet:  func(wallet Wallet) {},
			wantErr:      true,
			errCode:      ErrorCredentialNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletForIssuer(t)
			resolver := &MockDIDResolver{}
			issuer := &MockCredentialIssuer{}
			
			// For the valid revocation test, store a credential and use its actual ID
			credentialID := tt.credentialID
			if tt.name == "valid revocation" {
				// Add credential with status list
				cred := &vc.VerifiableCredential{
					ID:           "cred-1",
					Context:      []string{"https://www.w3.org/2018/credentials/v1"},
					Type:         []string{"VerifiableCredential"},
					IssuanceDate: time.Now().Format(time.RFC3339),
					CredentialSubject: map[string]interface{}{
						"id": "did:key:subject",
					},
					CredentialStatus: &vc.CredentialStatus{
						ID:   "status-list#1",
						Type: "StatusList2021Entry",
					},
				}
				credRecord, _ := wallet.StoreCredential(cred)
				credentialID = credRecord.ID
			} else {
				tt.setupWallet(wallet)
			}
			
			service := NewIssuerService(wallet, resolver, issuer)
			
			err := service.RevokeCredential(context.Background(), credentialID, tt.reason)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
			} else {
				require.NoError(t, err)
				
				// Verify credential status was updated
				credRecord, err := wallet.GetCredential(credentialID)
				require.NoError(t, err)
				assert.Equal(t, CredentialStatusRevoked, credRecord.Status)
				assert.Equal(t, tt.reason, credRecord.Metadata["revocationReason"])
			}
		})
	}
}

func TestIssuerService_CreateCredentialTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template *CredentialTemplate
		wantErr  bool
		errCode  string
	}{
		{
			name: "valid template",
			template: &CredentialTemplate{
				ID:          "template-1",
				Name:        "Test Template",
				Description: "A test template",
				Context:     []string{"https://www.w3.org/2018/credentials/v1"},
				Type:        []string{"VerifiableCredential", "TestCredential"},
				RequiredFields: []string{"name", "email"},
				OptionalFields: []string{"phone"},
				DefaultValues: map[string]interface{}{
					"type": "TestCredential",
				},
			},
			wantErr: false,
		},
		{
			name:     "nil template",
			template: nil,
			wantErr:  true,
			errCode:  ErrorInvalidCredential,
		},
		{
			name: "empty template ID",
			template: &CredentialTemplate{
				Name:           "Test Template",
				Type:           []string{"VerifiableCredential"},
				RequiredFields: []string{"name"},
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "empty required fields",
			template: &CredentialTemplate{
				ID:   "template-1",
				Name: "Test Template",
				Type: []string{"VerifiableCredential"},
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletForIssuer(t)
			resolver := &MockDIDResolver{}
			issuer := &MockCredentialIssuer{}
			
			service := NewIssuerService(wallet, resolver, issuer)
			
			err := service.CreateCredentialTemplate(tt.template)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
			} else {
				require.NoError(t, err)
				
				// Verify template was stored
				key := "template_" + tt.template.ID
				stored, err := wallet.(*DefaultWallet).storage.GetMetadata(key)
				require.NoError(t, err)
				assert.NotNil(t, stored)
			}
		})
	}
}

func TestIssuerService_ValidateIssuanceRequest(t *testing.T) {
	tests := []struct {
		name    string
		request *IssuanceRequest
		wantErr bool
		errCode string
	}{
		{
			name: "valid request",
			request: &IssuanceRequest{
				Issuer:            "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				SigningKeyID:      "key-1",
				Type:              []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
				IssuanceDate:      time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty issuer",
			request: &IssuanceRequest{
				SigningKeyID:      "key-1",
				Type:              []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "empty signing key ID",
			request: &IssuanceRequest{
				Issuer:            "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				Type:              []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "empty type",
			request: &IssuanceRequest{
				Issuer:            "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				SigningKeyID:      "key-1",
				CredentialSubject: map[string]interface{}{"id": "subject"},
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "nil credential subject",
			request: &IssuanceRequest{
				Issuer:       "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				SigningKeyID: "key-1",
				Type:         []string{"VerifiableCredential"},
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "invalid issuer DID",
			request: &IssuanceRequest{
				Issuer:            "invalid-did",
				SigningKeyID:      "key-1",
				Type:              []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
			},
			wantErr: true,
			errCode: ErrorInvalidDID,
		},
		{
			name: "expiration before issuance",
			request: &IssuanceRequest{
				Issuer:            "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				SigningKeyID:      "key-1",
				Type:              []string{"VerifiableCredential"},
				CredentialSubject: map[string]interface{}{"id": "subject"},
				IssuanceDate:      time.Now(),
				ExpirationDate:    &[]time.Time{time.Now().Add(-time.Hour)}[0],
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWalletForIssuer(t)
			resolver := &MockDIDResolver{}
			issuer := &MockCredentialIssuer{}
			
			service := NewIssuerService(wallet, resolver, issuer)
			
			err := service.validateIssuanceRequest(tt.request)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var walletErr *WalletError
					require.ErrorAs(t, err, &walletErr)
					assert.Equal(t, tt.errCode, walletErr.Code)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper function to setup test wallet for issuer tests
func setupTestWalletForIssuer(t *testing.T) Wallet {
	config := DefaultWalletConfig()
	config.StorageType = "memory"
	config.EncryptionEnabled = false // Start unlocked for tests
	
	storage := NewInMemoryStorage()
	keyManager := &MockKeyManager{}
	
	wallet, err := NewDefaultWallet(config, storage, keyManager)
	require.NoError(t, err)
	
	return wallet
}

// MockCredentialIssuer is a mock implementation of vc.CredentialIssuer
type MockCredentialIssuer struct {
	mock.Mock
}

func (m *MockCredentialIssuer) IssueCredential(template *vc.CredentialTemplate, options *vc.IssuanceOptions) (*vc.VerifiableCredential, error) {
	args := m.Called(template, options)
	return args.Get(0).(*vc.VerifiableCredential), args.Error(1)
}

func (m *MockCredentialIssuer) IssueJWTCredential(template *vc.CredentialTemplate, options *vc.IssuanceOptions) (string, error) {
	args := m.Called(template, options)
	return args.String(0), args.Error(1)
}

func (m *MockCredentialIssuer) IssueSDJWT(template *vc.CredentialTemplate, options *vc.IssuanceOptions) (string, error) {
	args := m.Called(template, options)
	return args.String(0), args.Error(1)
}