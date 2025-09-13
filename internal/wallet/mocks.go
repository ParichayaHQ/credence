package wallet

import (
	"context"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
	"github.com/stretchr/testify/mock"
)

// MockDIDResolver is a mock implementation of did.MultiResolver
type MockDIDResolver struct {
	mock.Mock
}

func (m *MockDIDResolver) Resolve(ctx context.Context, didStr string, options *did.DIDResolutionOptions) (*did.DIDResolutionResult, error) {
	args := m.Called(ctx, didStr, options)
	return args.Get(0).(*did.DIDResolutionResult), args.Error(1)
}

func (m *MockDIDResolver) ResolveWithMethod(ctx context.Context, didStr string, method string, options *did.DIDResolutionOptions) (*did.DIDResolutionResult, error) {
	args := m.Called(ctx, didStr, method, options)
	return args.Get(0).(*did.DIDResolutionResult), args.Error(1)
}

func (m *MockDIDResolver) SupportsMethod(method string) bool {
	args := m.Called(method)
	return args.Bool(0)
}

func (m *MockDIDResolver) SupportedMethods() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockDIDResolver) RegisterMethod(method string, resolver did.MethodResolver) error {
	args := m.Called(method, resolver)
	return args.Error(0)
}

func (m *MockDIDResolver) UnregisterMethod(method string) error {
	args := m.Called(method)
	return args.Error(0)
}

func (m *MockDIDResolver) GetResolver(method string) (did.MethodResolver, error) {
	args := m.Called(method)
	return args.Get(0).(did.MethodResolver), args.Error(1)
}

func (m *MockDIDResolver) ListMethods() []string {
	return m.SupportedMethods()
}

func (m *MockDIDResolver) SetDefaultOptions(options *did.DIDResolutionOptions) {
	m.Called(options)
}

func (m *MockDIDResolver) GetDefaultOptions() *did.DIDResolutionOptions {
	args := m.Called()
	return args.Get(0).(*did.DIDResolutionOptions)
}

// MockCredentialVerifier is a mock implementation of vc.CredentialVerifier
type MockCredentialVerifier struct {
	mock.Mock
}

func (m *MockCredentialVerifier) VerifyCredential(credential *vc.VerifiableCredential, options *vc.VerificationOptions) (*vc.VerificationResult, error) {
	args := m.Called(credential, options)
	return args.Get(0).(*vc.VerificationResult), args.Error(1)
}

func (m *MockCredentialVerifier) VerifyPresentation(presentation *vc.VerifiablePresentation, options *vc.VerificationOptions) (*vc.VerificationResult, error) {
	args := m.Called(presentation, options)
	return args.Get(0).(*vc.VerificationResult), args.Error(1)
}

func (m *MockCredentialVerifier) VerifyJWTCredential(token string, options *vc.VerificationOptions) (*vc.VerificationResult, error) {
	args := m.Called(token, options)
	return args.Get(0).(*vc.VerificationResult), args.Error(1)
}

func (m *MockCredentialVerifier) VerifyJWTPresentation(token string, options *vc.VerificationOptions) (*vc.VerificationResult, error) {
	args := m.Called(token, options)
	return args.Get(0).(*vc.VerificationResult), args.Error(1)
}

func (m *MockCredentialVerifier) VerifySDJWT(sdjwt string, options *vc.VerificationOptions) (*vc.VerificationResult, error) {
	args := m.Called(sdjwt, options)
	return args.Get(0).(*vc.VerificationResult), args.Error(1)
}

// MockKeyManager is a mock implementation of did.KeyManager
type MockKeyManager struct {
	mock.Mock
}

func (m *MockKeyManager) GenerateKey(keyType did.KeyType) (interface{}, error) {
	args := m.Called(keyType)
	return args.Get(0), args.Error(1)
}

func (m *MockKeyManager) GetPublicKey(privateKey interface{}) (interface{}, error) {
	args := m.Called(privateKey)
	return args.Get(0), args.Error(1)
}

func (m *MockKeyManager) Sign(privateKey interface{}, data []byte) ([]byte, error) {
	args := m.Called(privateKey, data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockKeyManager) Verify(publicKey interface{}, data []byte, signature []byte) bool {
	args := m.Called(publicKey, data, signature)
	return args.Bool(0)
}

func (m *MockKeyManager) KeyToPEM(key interface{}) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockKeyManager) PEMToKey(pemData []byte) (interface{}, error) {
	args := m.Called(pemData)
	return args.Get(0), args.Error(1)
}

func (m *MockKeyManager) KeyToJWK(key interface{}) (*did.JWK, error) {
	args := m.Called(key)
	return args.Get(0).(*did.JWK), args.Error(1)
}

func (m *MockKeyManager) JWKToKey(jwk *did.JWK) (interface{}, error) {
	args := m.Called(jwk)
	return args.Get(0), args.Error(1)
}