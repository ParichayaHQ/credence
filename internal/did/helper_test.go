package did

import (
	"context"
	"testing"
	"time"
)

func TestNewDocumentHelper(t *testing.T) {
	helper := NewDocumentHelper()
	if helper == nil {
		t.Error("expected helper to be created")
	}
}

func TestDocumentHelper_GetVerificationMethod(t *testing.T) {
	helper := NewDocumentHelper()

	document := &DIDDocument{
		ID: "did:key:z123",
		VerificationMethod: []VerificationMethod{
			{
				ID:         "did:key:z123#key1",
				Type:       "Ed25519VerificationKey2020",
				Controller: "did:key:z123",
			},
			{
				ID:         "did:key:z123#key2",
				Type:       "Ed25519VerificationKey2020",
				Controller: "did:key:z123",
			},
		},
	}

	// Test getting existing method
	method, err := helper.GetVerificationMethod(document, "did:key:z123#key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if method.ID != "did:key:z123#key1" {
		t.Errorf("expected method ID did:key:z123#key1, got %s", method.ID)
	}

	// Test getting method by fragment
	method, err = helper.GetVerificationMethod(document, "#key2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if method.ID != "did:key:z123#key2" {
		t.Errorf("expected method ID did:key:z123#key2, got %s", method.ID)
	}

	// Test non-existent method
	_, err = helper.GetVerificationMethod(document, "#nonexistent")
	if err == nil {
		t.Error("expected error for non-existent method")
	}

	// Test nil document
	_, err = helper.GetVerificationMethod(nil, "#key1")
	if err == nil {
		t.Error("expected error for nil document")
	}
}

func TestDocumentHelper_GetVerificationMethodsForPurpose(t *testing.T) {
	helper := NewDocumentHelper()

	document := &DIDDocument{
		ID: "did:key:z123",
		VerificationMethod: []VerificationMethod{
			{
				ID:         "did:key:z123#key1",
				Type:       "Ed25519VerificationKey2020",
				Controller: "did:key:z123",
			},
			{
				ID:         "did:key:z123#key2",
				Type:       "Ed25519VerificationKey2020",
				Controller: "did:key:z123",
			},
		},
		Authentication: []interface{}{
			"did:key:z123#key1",
			"#key2", // relative reference
		},
		AssertionMethod: []interface{}{
			"did:key:z123#key1",
		},
	}

	// Test authentication methods
	methods, err := helper.GetVerificationMethodsForPurpose(document, Authentication)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(methods) != 2 {
		t.Errorf("expected 2 authentication methods, got %d", len(methods))
	}

	// Test assertion methods
	methods, err = helper.GetVerificationMethodsForPurpose(document, AssertionMethod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(methods) != 1 {
		t.Errorf("expected 1 assertion method, got %d", len(methods))
	}

	if methods[0].ID != "did:key:z123#key1" {
		t.Errorf("expected key1, got %s", methods[0].ID)
	}

	// Test empty purpose
	methods, err = helper.GetVerificationMethodsForPurpose(document, KeyAgreement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(methods) != 0 {
		t.Errorf("expected 0 key agreement methods, got %d", len(methods))
	}

	// Test nil document
	_, err = helper.GetVerificationMethodsForPurpose(nil, Authentication)
	if err == nil {
		t.Error("expected error for nil document")
	}

	// Test invalid purpose
	_, err = helper.GetVerificationMethodsForPurpose(document, VerificationRelationship("invalid"))
	if err == nil {
		t.Error("expected error for invalid purpose")
	}
}

func TestDocumentHelper_GetService(t *testing.T) {
	helper := NewDocumentHelper()

	document := &DIDDocument{
		ID: "did:key:z123",
		Service: []Service{
			{
				ID:   "did:key:z123#agent",
				Type: "DIDCommMessaging",
				ServiceEndpoint: map[string]interface{}{
					"uri": "https://example.com/endpoint",
				},
			},
		},
	}

	// Test getting existing service
	service, err := helper.GetService(document, "did:key:z123#agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if service.ID != "did:key:z123#agent" {
		t.Errorf("expected service ID did:key:z123#agent, got %s", service.ID)
	}

	// Test getting service by fragment
	service, err = helper.GetService(document, "#agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if service.ID != "did:key:z123#agent" {
		t.Errorf("expected service ID did:key:z123#agent, got %s", service.ID)
	}

	// Test non-existent service
	_, err = helper.GetService(document, "#nonexistent")
	if err == nil {
		t.Error("expected error for non-existent service")
	}

	// Test nil document
	_, err = helper.GetService(nil, "#agent")
	if err == nil {
		t.Error("expected error for nil document")
	}
}

func TestDocumentHelper_GetServicesByType(t *testing.T) {
	helper := NewDocumentHelper()

	document := &DIDDocument{
		ID: "did:key:z123",
		Service: []Service{
			{
				ID:   "did:key:z123#agent",
				Type: "DIDCommMessaging",
				ServiceEndpoint: map[string]interface{}{
					"uri": "https://example.com/endpoint",
				},
			},
			{
				ID:   "did:key:z123#resolver",
				Type: "LinkedDomains",
				ServiceEndpoint: map[string]interface{}{
					"origins": []string{"https://example.com"},
				},
			},
			{
				ID:   "did:key:z123#agent2",
				Type: "DIDCommMessaging",
				ServiceEndpoint: map[string]interface{}{
					"uri": "https://example2.com/endpoint",
				},
			},
		},
	}

	// Test getting services by type
	services, err := helper.GetServicesByType(document, "DIDCommMessaging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("expected 2 DIDCommMessaging services, got %d", len(services))
	}

	// Test getting services by different type
	services, err = helper.GetServicesByType(document, "LinkedDomains")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("expected 1 LinkedDomains service, got %d", len(services))
	}

	if services[0].ID != "did:key:z123#resolver" {
		t.Errorf("expected resolver service, got %s", services[0].ID)
	}

	// Test non-existent type
	services, err = helper.GetServicesByType(document, "NonExistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(services) != 0 {
		t.Errorf("expected 0 services, got %d", len(services))
	}

	// Test nil document
	_, err = helper.GetServicesByType(nil, "DIDCommMessaging")
	if err == nil {
		t.Error("expected error for nil document")
	}
}

func TestDocumentHelper_AddVerificationMethod(t *testing.T) {
	helper := NewDocumentHelper()

	document := &DIDDocument{
		ID:                 "did:key:z123",
		VerificationMethod: []VerificationMethod{},
	}

	multibaseKey := "z123"
	method := &VerificationMethod{
		ID:                 "did:key:z123#key1",
		Type:               "Ed25519VerificationKey2020",
		Controller:         "did:key:z123",
		PublicKeyMultibase: &multibaseKey,
	}

	err := helper.AddVerificationMethod(document, method, []VerificationRelationship{Authentication, AssertionMethod})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(document.VerificationMethod) != 1 {
		t.Errorf("expected 1 verification method, got %d", len(document.VerificationMethod))
	}

	if len(document.Authentication) != 1 {
		t.Errorf("expected 1 authentication method, got %d", len(document.Authentication))
	}

	if len(document.AssertionMethod) != 1 {
		t.Errorf("expected 1 assertion method, got %d", len(document.AssertionMethod))
	}

	// Test adding duplicate method
	err = helper.AddVerificationMethod(document, method, []VerificationRelationship{Authentication})
	if err == nil {
		t.Error("expected error for duplicate method")
	}

	// Test nil document
	err = helper.AddVerificationMethod(nil, method, []VerificationRelationship{Authentication})
	if err == nil {
		t.Error("expected error for nil document")
	}

	// Test nil method
	err = helper.AddVerificationMethod(document, nil, []VerificationRelationship{Authentication})
	if err == nil {
		t.Error("expected error for nil method")
	}
}

func TestDocumentHelper_RemoveVerificationMethod(t *testing.T) {
	helper := NewDocumentHelper()

	multibaseKey := "z123"
	document := &DIDDocument{
		ID: "did:key:z123",
		VerificationMethod: []VerificationMethod{
			{
				ID:                 "did:key:z123#key1",
				Type:               "Ed25519VerificationKey2020",
				Controller:         "did:key:z123",
				PublicKeyMultibase: &multibaseKey,
			},
		},
		Authentication:       []interface{}{"did:key:z123#key1"},
		AssertionMethod:      []interface{}{"did:key:z123#key1"},
		CapabilityInvocation: []interface{}{"did:key:z123#key1"},
	}

	err := helper.RemoveVerificationMethod(document, "did:key:z123#key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(document.VerificationMethod) != 0 {
		t.Errorf("expected 0 verification methods, got %d", len(document.VerificationMethod))
	}

	if len(document.Authentication) != 0 {
		t.Errorf("expected 0 authentication methods, got %d", len(document.Authentication))
	}

	if len(document.AssertionMethod) != 0 {
		t.Errorf("expected 0 assertion methods, got %d", len(document.AssertionMethod))
	}

	// Test removing non-existent method
	err = helper.RemoveVerificationMethod(document, "did:key:z123#nonexistent")
	if err == nil {
		t.Error("expected error for non-existent method")
	}

	// Test nil document
	err = helper.RemoveVerificationMethod(nil, "did:key:z123#key1")
	if err == nil {
		t.Error("expected error for nil document")
	}
}

func TestDocumentHelper_IsDeactivated(t *testing.T) {
	helper := NewDocumentHelper()

	// Test non-deactivated document
	document := &DIDDocument{
		ID: "did:key:z123",
	}

	if helper.IsDeactivated(document) {
		t.Error("expected document not to be deactivated")
	}

	// Test deactivated document
	deactivated := true
	document.Deactivated = &deactivated

	if !helper.IsDeactivated(document) {
		t.Error("expected document to be deactivated")
	}

	// Test explicitly not deactivated
	deactivated = false
	document.Deactivated = &deactivated

	if helper.IsDeactivated(document) {
		t.Error("expected document not to be deactivated")
	}

	// Test nil document
	if helper.IsDeactivated(nil) {
		t.Error("expected nil document not to be deactivated")
	}
}

func TestDocumentHelper_ValidateVerificationMethod(t *testing.T) {
	helper := NewDocumentHelper().(*DefaultDocumentHelper)

	multibaseKey := "z123"
	validMethod := &VerificationMethod{
		ID:                 "did:key:z123#key1",
		Type:               "Ed25519VerificationKey2020",
		Controller:         "did:key:z123",
		PublicKeyMultibase: &multibaseKey,
	}

	err := helper.ValidateVerificationMethod(validMethod)
	if err != nil {
		t.Fatalf("unexpected error for valid method: %v", err)
	}

	// Test missing ID
	invalidMethod := *validMethod
	invalidMethod.ID = ""
	err = helper.ValidateVerificationMethod(&invalidMethod)
	if err == nil {
		t.Error("expected error for missing ID")
	}

	// Test missing type
	invalidMethod = *validMethod
	invalidMethod.Type = ""
	err = helper.ValidateVerificationMethod(&invalidMethod)
	if err == nil {
		t.Error("expected error for missing type")
	}

	// Test missing controller
	invalidMethod = *validMethod
	invalidMethod.Controller = ""
	err = helper.ValidateVerificationMethod(&invalidMethod)
	if err == nil {
		t.Error("expected error for missing controller")
	}

	// Test missing key representation
	invalidMethod = *validMethod
	invalidMethod.PublicKeyMultibase = nil
	err = helper.ValidateVerificationMethod(&invalidMethod)
	if err == nil {
		t.Error("expected error for missing key representation")
	}
}

func TestDocumentHelper_ValidateContext(t *testing.T) {
	helper := NewDocumentHelper().(*DefaultDocumentHelper)

	// Test valid context
	validContext := []string{
		"https://www.w3.org/ns/did/v1",
		"https://w3id.org/security/suites/ed25519-2020/v1",
	}

	err := helper.ValidateContext(validContext)
	if err != nil {
		t.Fatalf("unexpected error for valid context: %v", err)
	}

	// Test empty context
	err = helper.ValidateContext([]string{})
	if err == nil {
		t.Error("expected error for empty context")
	}

	// Test invalid first context
	invalidContext := []string{
		"https://example.com/context",
		"https://w3id.org/security/suites/ed25519-2020/v1",
	}

	err = helper.ValidateContext(invalidContext)
	if err == nil {
		t.Error("expected error for invalid first context")
	}
}

func TestNewDocumentValidator(t *testing.T) {
	validator := NewDocumentValidator()
	if validator == nil {
		t.Error("expected validator to be created")
	}
}

func TestDocumentValidator_Validate(t *testing.T) {
	validator := NewDocumentValidator()
	ctx := context.Background()

	now := time.Now().UTC()
	multibaseKey := "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"

	validDocument := &DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		VerificationMethod: []VerificationMethod{
			{
				ID:                 "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				Type:               "Ed25519VerificationKey2020",
				Controller:         "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				PublicKeyMultibase: &multibaseKey,
			},
		},
		Authentication: []interface{}{
			"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
		Created: &now,
	}

	err := validator.Validate(ctx, validDocument)
	if err != nil {
		t.Fatalf("unexpected error for valid document: %v", err)
	}

	// Test nil document
	err = validator.Validate(ctx, nil)
	if err == nil {
		t.Error("expected error for nil document")
	}

	// Test missing ID
	invalidDocument := *validDocument
	invalidDocument.ID = ""
	err = validator.Validate(ctx, &invalidDocument)
	if err == nil {
		t.Error("expected error for missing ID")
	}

	// Test invalid DID
	invalidDocument = *validDocument
	invalidDocument.ID = "not-a-did"
	err = validator.Validate(ctx, &invalidDocument)
	if err == nil {
		t.Error("expected error for invalid DID")
	}
}