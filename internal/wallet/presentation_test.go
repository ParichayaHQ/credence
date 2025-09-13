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

func TestPresentationService_CreatePresentation(t *testing.T) {
	tests := []struct {
		name        string
		request     *PresentationRequest
		setupWallet func() Wallet
		wantErr     bool
		errCode     string
	}{
		{
			name: "valid presentation request",
			request: &PresentationRequest{
				CredentialIDs: []string{"cred-1"},
				Holder:        "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				KeyID:         "key-1",
				Algorithm:     "EdDSA",
				Challenge:     "test-challenge",
				Domain:        "example.com",
			},
			setupWallet: func() Wallet {
				wallet := setupTestWallet(t)
				
				// Add test credential
				cred := &vc.VerifiableCredential{
					Context:      []string{"https://www.w3.org/2018/credentials/v1"},
					Type:         []string{"VerifiableCredential", "TestCredential"},
					Issuer:       "did:key:issuer",
					IssuanceDate: time.Now().Format(time.RFC3339),
					CredentialSubject: map[string]interface{}{
						"id":   "did:key:subject",
						"name": "Test User",
					},
				}
				
				_, err := wallet.StoreCredential(cred)
				require.NoError(t, err)
				
				return wallet
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			request: nil,
			setupWallet: func() Wallet {
				return setupTestWallet(t)
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := tt.setupWallet()
			resolver := &MockDIDResolver{}
			verifier := &MockCredentialVerifier{}
			
			ps := NewPresentationService(wallet, verifier, resolver)
			
			presentation, err := ps.CreatePresentation(context.Background(), tt.request)
			
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
				assert.Equal(t, tt.request.Holder, presentation.Holder)
			}
		})
	}
}

func TestPresentationService_VerifyPresentation(t *testing.T) {
	tests := []struct {
		name           string
		presentation   *vc.VerifiablePresentation
		options        *VerificationOptions
		setupMocks     func(*MockDIDResolver, *MockCredentialVerifier)
		wantValid      bool
		wantErrorCount int
	}{
		{
			name: "valid presentation",
			presentation: &vc.VerifiablePresentation{
				Context: []string{"https://www.w3.org/2018/credentials/v1"},
				Type:    []string{"VerifiablePresentation"},
				Holder:  "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				VerifiableCredential: []interface{}{
					&vc.VerifiableCredential{
						Context:      []string{"https://www.w3.org/2018/credentials/v1"},
						Type:         []string{"VerifiableCredential"},
						Issuer:       "did:key:issuer",
						IssuanceDate: time.Now().Format(time.RFC3339),
						CredentialSubject: map[string]interface{}{
							"id": "did:key:subject",
						},
					},
				},
				Proof: map[string]interface{}{
					"type":               "Ed25519Signature2018",
					"challenge":          "test-challenge",
					"domain":             "example.com",
					"proofPurpose":       "authentication",
					"verificationMethod": "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH#key-1",
				},
			},
			options: &VerificationOptions{
				ExpectedChallenge: "test-challenge",
				ExpectedDomain:    "example.com",
				ExpectedPurpose:   "authentication",
			},
			setupMocks: func(resolver *MockDIDResolver, verifier *MockCredentialVerifier) {
				resolver.On("Resolve", context.Background(), "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH", mock.AnythingOfType("*did.DIDResolutionOptions")).Return(&did.DIDResolutionResult{
					DIDDocument: &did.DIDDocument{
						ID: "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
					},
					DIDResolutionMetadata: did.DIDResolutionMetadata{},
				}, nil)
				
				verifier.On("VerifyPresentation", 
					mock.AnythingOfType("*vc.VerifiablePresentation"), 
					mock.AnythingOfType("*vc.VerificationOptions")).Return(&vc.VerificationResult{
					Verified: true,
				}, nil)
				
				verifier.On("VerifyCredential", 
					mock.AnythingOfType("*vc.VerifiableCredential"), 
					mock.AnythingOfType("*vc.VerificationOptions")).Return(&vc.VerificationResult{
					Verified: true,
				}, nil)
			},
			wantValid:      true,
			wantErrorCount: 0,
		},
		{
			name:           "nil presentation",
			presentation:   nil,
			wantValid:      false,
			wantErrorCount: 0, // This should return an error, not a result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWallet(t)
			resolver := &MockDIDResolver{}
			verifier := &MockCredentialVerifier{}
			
			if tt.setupMocks != nil {
				tt.setupMocks(resolver, verifier)
			}
			
			ps := NewPresentationService(wallet, verifier, resolver)
			
			result, err := ps.VerifyPresentation(context.Background(), tt.presentation, tt.options)
			
			if tt.presentation == nil {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}
			
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantValid, result.Valid)
			assert.Len(t, result.Errors, tt.wantErrorCount)
			assert.NotZero(t, result.VerifiedAt)
		})
	}
}

func TestPresentationService_ValidatePresentationRequest(t *testing.T) {
	tests := []struct {
		name    string
		request *PresentationRequest
		wantErr bool
		errCode string
	}{
		{
			name: "valid request",
			request: &PresentationRequest{
				CredentialIDs: []string{"cred-1"},
				Holder:        "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				KeyID:         "key-1",
			},
			wantErr: false,
		},
		{
			name: "empty credential IDs",
			request: &PresentationRequest{
				CredentialIDs: []string{},
				Holder:        "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
				KeyID:         "key-1",
			},
			wantErr: true,
			errCode: ErrorInvalidCredential,
		},
		{
			name: "invalid holder DID",
			request: &PresentationRequest{
				CredentialIDs: []string{"cred-1"},
				Holder:        "invalid-did",
				KeyID:         "key-1",
			},
			wantErr: true,
			errCode: ErrorInvalidDID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := setupTestWallet(t)
			resolver := &MockDIDResolver{}
			verifier := &MockCredentialVerifier{}
			
			ps := NewPresentationService(wallet, verifier, resolver)
			
			err := ps.validatePresentationRequest(tt.request)
			
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

// Helper function to setup test wallet
func setupTestWallet(t *testing.T) Wallet {
	config := DefaultWalletConfig()
	config.StorageType = "memory"
	config.EncryptionEnabled = false // Start unlocked for tests
	
	storage := NewInMemoryStorage()
	keyManager := &MockKeyManager{}
	
	wallet, err := NewDefaultWallet(config, storage, keyManager)
	require.NoError(t, err)
	
	return wallet
}