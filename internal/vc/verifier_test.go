package vc

import (
	"testing"

	"github.com/ParichayaHQ/credence/internal/did"
)

func TestNewDefaultCredentialVerifier(t *testing.T) {
	keyManager := did.NewDefaultKeyManager()
	resolver := did.NewMultiDIDResolver()
	
	verifier := NewDefaultCredentialVerifier(keyManager, resolver)
	
	if verifier == nil {
		t.Error("expected verifier to be created")
	}
	
	if verifier.keyManager != keyManager {
		t.Error("key manager not set correctly")
	}
	
	if verifier.didResolver != resolver {
		t.Error("DID resolver not set correctly")
	}
}

func TestDefaultCredentialVerifier_ValidateCredentialStructure(t *testing.T) {
	verifier := &DefaultCredentialVerifier{}
	
	// Test valid credential
	validCredential := &VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://example.org/examples/v1",
		},
		Type: []string{
			"VerifiableCredential",
			"ExampleCredential",
		},
		Issuer:            "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		IssuanceDate:      "2023-01-01T00:00:00Z",
		CredentialSubject: map[string]interface{}{"id": "did:example:123"},
	}
	
	err := verifier.validateCredentialStructure(validCredential)
	if err != nil {
		t.Fatalf("expected valid credential to pass validation: %v", err)
	}
	
	// Test missing @context
	invalidCredential := *validCredential
	invalidCredential.Context = nil
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for missing @context")
	}
	
	// Test invalid first @context
	invalidCredential = *validCredential
	invalidCredential.Context = []string{"https://example.org/contexts/v1"}
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for invalid first @context")
	}
	
	// Test missing type
	invalidCredential = *validCredential
	invalidCredential.Type = nil
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for missing type")
	}
	
	// Test missing VerifiableCredential type
	invalidCredential = *validCredential
	invalidCredential.Type = []string{"ExampleCredential"}
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for missing VerifiableCredential type")
	}
	
	// Test missing issuer
	invalidCredential = *validCredential
	invalidCredential.Issuer = nil
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for missing issuer")
	}
	
	// Test missing issuanceDate
	invalidCredential = *validCredential
	invalidCredential.IssuanceDate = ""
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for missing issuanceDate")
	}
	
	// Test missing credentialSubject
	invalidCredential = *validCredential
	invalidCredential.CredentialSubject = nil
	err = verifier.validateCredentialStructure(&invalidCredential)
	if err == nil {
		t.Error("expected error for missing credentialSubject")
	}
}

func TestDefaultCredentialVerifier_ValidatePresentationStructure(t *testing.T) {
	verifier := &DefaultCredentialVerifier{}
	
	// Test valid presentation
	validPresentation := &VerifiablePresentation{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
		},
		Type: []string{
			"VerifiablePresentation",
		},
		Holder: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
	}
	
	err := verifier.validatePresentationStructure(validPresentation)
	if err != nil {
		t.Fatalf("expected valid presentation to pass validation: %v", err)
	}
	
	// Test missing @context
	invalidPresentation := *validPresentation
	invalidPresentation.Context = nil
	err = verifier.validatePresentationStructure(&invalidPresentation)
	if err == nil {
		t.Error("expected error for missing @context")
	}
	
	// Test invalid first @context
	invalidPresentation = *validPresentation
	invalidPresentation.Context = []string{"https://example.org/contexts/v1"}
	err = verifier.validatePresentationStructure(&invalidPresentation)
	if err == nil {
		t.Error("expected error for invalid first @context")
	}
	
	// Test missing type
	invalidPresentation = *validPresentation
	invalidPresentation.Type = nil
	err = verifier.validatePresentationStructure(&invalidPresentation)
	if err == nil {
		t.Error("expected error for missing type")
	}
	
	// Test missing VerifiablePresentation type
	invalidPresentation = *validPresentation
	invalidPresentation.Type = []string{"ExamplePresentation"}
	err = verifier.validatePresentationStructure(&invalidPresentation)
	if err == nil {
		t.Error("expected error for missing VerifiablePresentation type")
	}
}

func TestDefaultCredentialVerifier_LooksLikeSDJWT(t *testing.T) {
	verifier := &DefaultCredentialVerifier{}
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSJ9.eyJpc3MiOiJkaWQ6a2V5Onp4IiwidmMiOnsiaWQiOiJ1cm46dXVpZDoxIn19.sig~WyJzYWx0IiwgImNsYWltIiwgInZhbHVlIl0~", true},
		{"eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSJ9.eyJpc3MiOiJkaWQ6a2V5Onp4IiwidmMiOnsiaWQiOiJ1cm46dXVpZDoxIn19.sig", false},
		{"regular~string~with~tildes", true},
		{"regular.jwt.token", false},
		{"", false},
	}
	
	for _, tt := range tests {
		result := verifier.looksLikeSDJWT(tt.input)
		if result != tt.expected {
			t.Errorf("looksLikeSDJWT(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultCredentialVerifier_LooksLikeJWT(t *testing.T) {
	verifier := &DefaultCredentialVerifier{}
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSJ9.eyJpc3MiOiJkaWQ6a2V5Onp4In0.sig", true},
		{"header.payload.signature.extra", false},
		{"header.payload", false},
		{"single-part", false},
		{"", false},
	}
	
	for _, tt := range tests {
		result := verifier.looksLikeJWT(tt.input)
		if result != tt.expected {
			t.Errorf("looksLikeJWT(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultCredentialVerifier_VerifyCredential(t *testing.T) {
	keyManager := did.NewDefaultKeyManager()
	resolver := did.NewMultiDIDResolver()
	verifier := NewDefaultCredentialVerifier(keyManager, resolver)
	
	// Test nil credential
	result, err := verifier.VerifyCredential(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.Verified {
		t.Error("expected verification to fail for nil credential")
	}
	
	// Test credential with valid structure but no proof
	credential := &VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
		},
		Type: []string{
			"VerifiableCredential",
		},
		Issuer:            "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		IssuanceDate:      "2023-01-01T00:00:00Z",
		CredentialSubject: map[string]interface{}{"id": "did:example:123"},
	}
	
	result, err = verifier.VerifyCredential(credential, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.Verified {
		t.Error("expected verification to fail for credential without proof")
	}
	
	// Test credential with proof
	credential.Proof = map[string]interface{}{
		"type":    "Ed25519Signature2020",
		"created": "2023-01-01T00:00:00Z",
	}
	
	result, err = verifier.VerifyCredential(credential, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// For now, our simplified implementation returns true for any proof
	if !result.Verified {
		t.Error("expected verification to succeed for credential with proof")
	}
}

func TestNewDefaultCredentialIssuer(t *testing.T) {
	keyManager := did.NewDefaultKeyManager()
	resolver := did.NewMultiDIDResolver()
	
	issuer := NewDefaultCredentialIssuer(keyManager, resolver)
	
	if issuer == nil {
		t.Error("expected issuer to be created")
	}
	
	if issuer.keyManager != keyManager {
		t.Error("key manager not set correctly")
	}
	
	if issuer.didResolver != resolver {
		t.Error("DID resolver not set correctly")
	}
}

func TestDefaultCredentialIssuer_IssueCredential(t *testing.T) {
	keyManager := did.NewDefaultKeyManager()
	resolver := did.NewMultiDIDResolver()
	issuer := NewDefaultCredentialIssuer(keyManager, resolver)
	
	// Test nil template
	credential, err := issuer.IssueCredential(nil, nil)
	if err == nil {
		t.Error("expected error for nil template")
	}
	
	if credential != nil {
		t.Error("expected nil credential for nil template")
	}
	
	// Test valid template
	template := &CredentialTemplate{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://example.org/examples/v1",
		},
		Type: []string{
			"VerifiableCredential",
			"ExampleCredential",
		},
		Issuer: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		CredentialSubject: map[string]interface{}{
			"id":   "did:example:123",
			"name": "John Doe",
		},
	}
	
	options := &IssuanceOptions{
		KeyID:     "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#key1",
		Algorithm: "EdDSA",
	}
	
	credential, err = issuer.IssueCredential(template, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if credential == nil {
		t.Error("expected credential to be issued")
	}
	
	// Verify the issued credential has required fields
	if len(credential.Context) == 0 {
		t.Error("expected credential to have @context")
	}
	
	if len(credential.Type) == 0 {
		t.Error("expected credential to have type")
	}
	
	if credential.Issuer == nil {
		t.Error("expected credential to have issuer")
	}
	
	if credential.IssuanceDate == "" {
		t.Error("expected credential to have issuanceDate")
	}
	
	if credential.CredentialSubject == nil {
		t.Error("expected credential to have credentialSubject")
	}
	
	if credential.Proof == nil {
		t.Error("expected credential to have proof")
	}
}

func TestGetIssuerID(t *testing.T) {
	tests := []struct {
		name     string
		issuer   interface{}
		expected string
	}{
		{
			name:     "string issuer",
			issuer:   "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
		{
			name: "Issuer struct",
			issuer: &Issuer{
				ID:   "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				Name: "Example Issuer",
			},
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
		{
			name: "map issuer",
			issuer: map[string]interface{}{
				"id":   "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				"name": "Example Issuer",
			},
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
		{
			name:     "nil issuer",
			issuer:   nil,
			expected: "",
		},
		{
			name:     "invalid issuer",
			issuer:   123,
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIssuerID(tt.issuer)
			if result != tt.expected {
				t.Errorf("getIssuerID() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetCredentialSubjectID(t *testing.T) {
	tests := []struct {
		name     string
		subject  interface{}
		expected string
	}{
		{
			name: "CredentialSubject struct",
			subject: &CredentialSubject{
				ID: "did:example:123",
			},
			expected: "did:example:123",
		},
		{
			name: "map subject",
			subject: map[string]interface{}{
				"id":   "did:example:123",
				"name": "John Doe",
			},
			expected: "did:example:123",
		},
		{
			name: "map without id",
			subject: map[string]interface{}{
				"name": "John Doe",
			},
			expected: "",
		},
		{
			name:     "nil subject",
			subject:  nil,
			expected: "",
		},
		{
			name:     "invalid subject",
			subject:  "string",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCredentialSubjectID(tt.subject)
			if result != tt.expected {
				t.Errorf("getCredentialSubjectID() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestVCError(t *testing.T) {
	// Test basic error
	err := NewVCError("test_code", "test message")
	if err.Code != "test_code" {
		t.Errorf("expected code 'test_code', got %q", err.Code)
	}
	
	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", err.Message)
	}
	
	if err.Error() != "test message" {
		t.Errorf("expected Error() 'test message', got %q", err.Error())
	}
	
	// Test error with details
	err = NewVCErrorWithDetails("test_code", "test message", "additional details")
	if err.Details != "additional details" {
		t.Errorf("expected details 'additional details', got %q", err.Details)
	}
	
	if err.Error() != "test message: additional details" {
		t.Errorf("expected Error() 'test message: additional details', got %q", err.Error())
	}
}