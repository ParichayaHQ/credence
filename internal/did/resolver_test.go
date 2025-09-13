package did

import (
	"context"
	"testing"
)

func TestNewMultiDIDResolver(t *testing.T) {
	resolver := NewMultiDIDResolver()

	if resolver == nil {
		t.Error("expected resolver to be created")
	}

	// Should support did:key by default
	if !resolver.SupportsMethod("key") {
		t.Error("expected resolver to support did:key method")
	}

	methods := resolver.SupportedMethods()
	if len(methods) != 1 || methods[0] != "key" {
		t.Errorf("expected ['key'], got %v", methods)
	}
}

func TestMultiDIDResolver_RegisterMethod(t *testing.T) {
	resolver := NewMultiDIDResolver()
	mockResolver := &mockMethodResolver{}

	err := resolver.RegisterMethod("test", mockResolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resolver.SupportsMethod("test") {
		t.Error("expected resolver to support registered method")
	}

	// Test invalid method name
	err = resolver.RegisterMethod("", mockResolver)
	if err == nil {
		t.Error("expected error for empty method name")
	}

	// Test invalid method name format
	err = resolver.RegisterMethod("INVALID", mockResolver)
	if err == nil {
		t.Error("expected error for uppercase method name")
	}

	// Test nil resolver
	err = resolver.RegisterMethod("null", nil)
	if err == nil {
		t.Error("expected error for nil resolver")
	}
}

func TestMultiDIDResolver_UnregisterMethod(t *testing.T) {
	resolver := NewMultiDIDResolver()
	mockResolver := &mockMethodResolver{}

	// Register first
	err := resolver.RegisterMethod("test", mockResolver)
	if err != nil {
		t.Fatalf("failed to register method: %v", err)
	}

	// Unregister
	err = resolver.UnregisterMethod("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolver.SupportsMethod("test") {
		t.Error("expected method to be unregistered")
	}

	// Test unregistering non-existent method
	err = resolver.UnregisterMethod("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent method")
	}

	// Test empty method name
	err = resolver.UnregisterMethod("")
	if err == nil {
		t.Error("expected error for empty method name")
	}
}

func TestMultiDIDResolver_Resolve(t *testing.T) {
	resolver := NewMultiDIDResolver()
	ctx := context.Background()

	// Create a did:key DID first
	keyResolver := resolver.methodResolvers["key"]
	createResult, err := keyResolver.Create(ctx, nil)
	if err != nil {
		t.Fatalf("failed to create DID: %v", err)
	}

	// Now resolve it
	result, err := resolver.Resolve(ctx, createResult.DID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DIDResolutionMetadata.Error != "" {
		t.Errorf("unexpected resolution error: %s", result.DIDResolutionMetadata.Error)
	}

	if result.DIDDocument == nil {
		t.Error("expected DID document")
	}

	// Test invalid DID
	result, err = resolver.Resolve(ctx, "invalid-did", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DIDResolutionMetadata.Error == "" {
		t.Error("expected resolution error for invalid DID")
	}

	// Test unsupported method
	result, err = resolver.Resolve(ctx, "did:unsupported:123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DIDResolutionMetadata.Error != ErrorMethodNotSupported {
		t.Errorf("expected ErrorMethodNotSupported, got %s", result.DIDResolutionMetadata.Error)
	}
}

func TestMultiDIDResolver_ResolveWithMethod(t *testing.T) {
	resolver := NewMultiDIDResolver()
	ctx := context.Background()

	// Create a did:key DID first
	keyResolver := resolver.methodResolvers["key"]
	createResult, err := keyResolver.Create(ctx, nil)
	if err != nil {
		t.Fatalf("failed to create DID: %v", err)
	}

	// Resolve with specific method
	result, err := resolver.ResolveWithMethod(ctx, createResult.DID, "key", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DIDResolutionMetadata.Error != "" {
		t.Errorf("unexpected resolution error: %s", result.DIDResolutionMetadata.Error)
	}

	if result.DIDDocument == nil {
		t.Error("expected DID document")
	}

	// Test unsupported method
	result, err = resolver.ResolveWithMethod(ctx, createResult.DID, "unsupported", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DIDResolutionMetadata.Error != ErrorMethodNotSupported {
		t.Errorf("expected ErrorMethodNotSupported, got %s", result.DIDResolutionMetadata.Error)
	}
}

func TestMultiDIDResolver_GetResolver(t *testing.T) {
	resolver := NewMultiDIDResolver()

	// Get existing resolver
	keyResolver, err := resolver.GetResolver("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if keyResolver == nil {
		t.Error("expected resolver")
	}

	// Get non-existent resolver
	_, err = resolver.GetResolver("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent method")
	}
}

func TestMultiDIDResolver_DefaultOptions(t *testing.T) {
	resolver := NewMultiDIDResolver()

	// Test default options
	options := resolver.GetDefaultOptions()
	if options.Accept != "application/did+ld+json" {
		t.Errorf("expected default Accept header, got %s", options.Accept)
	}

	// Test setting new default options
	newOptions := &DIDResolutionOptions{
		Accept:    "application/json",
		VersionId: "1",
	}

	resolver.SetDefaultOptions(newOptions)

	retrieved := resolver.GetDefaultOptions()
	if retrieved.Accept != "application/json" {
		t.Errorf("expected updated Accept header, got %s", retrieved.Accept)
	}

	if retrieved.VersionId != "1" {
		t.Errorf("expected updated VersionId, got %s", retrieved.VersionId)
	}
}

func TestCachedResolver(t *testing.T) {
	baseResolver := NewMultiDIDResolver()
	cache := &mockCacheManager{}
	cachedResolver := NewCachedResolver(baseResolver, cache, 300)

	if cachedResolver == nil {
		t.Error("expected cached resolver to be created")
	}

	// Should delegate to base resolver
	if !cachedResolver.SupportsMethod("key") {
		t.Error("expected cached resolver to support key method")
	}

	methods := cachedResolver.SupportedMethods()
	if len(methods) != 1 || methods[0] != "key" {
		t.Errorf("expected ['key'], got %v", methods)
	}
}

// Mock method resolver for testing
type mockMethodResolver struct{}

func (m *mockMethodResolver) Method() string {
	return "test"
}

func (m *mockMethodResolver) Resolve(ctx context.Context, did string, options *DIDResolutionOptions) (*DIDResolutionResult, error) {
	return &DIDResolutionResult{
		DIDDocument: &DIDDocument{
			ID: did,
		},
	}, nil
}

func (m *mockMethodResolver) Create(ctx context.Context, options *CreationOptions) (*CreationResult, error) {
	return &CreationResult{
		DID: "did:test:123",
	}, nil
}

func (m *mockMethodResolver) Update(ctx context.Context, did string, document *DIDDocument, options *UpdateOptions) (*UpdateResult, error) {
	return nil, NewDIDError(ErrorMethodNotSupported, "not supported")
}

func (m *mockMethodResolver) Deactivate(ctx context.Context, did string, options *DeactivationOptions) (*DeactivationResult, error) {
	return nil, NewDIDError(ErrorMethodNotSupported, "not supported")
}

// Mock cache manager for testing
type mockCacheManager struct {
	cache map[string]*DIDDocument
}

func (m *mockCacheManager) Get(ctx context.Context, did string) (*DIDDocument, error) {
	if m.cache == nil {
		return nil, NewDIDError(ErrorNotFound, "not found")
	}
	
	if doc, exists := m.cache[did]; exists {
		return doc, nil
	}
	
	return nil, NewDIDError(ErrorNotFound, "not found")
}

func (m *mockCacheManager) Set(ctx context.Context, did string, document *DIDDocument, ttl int64) error {
	if m.cache == nil {
		m.cache = make(map[string]*DIDDocument)
	}
	
	m.cache[did] = document
	return nil
}

func (m *mockCacheManager) Invalidate(ctx context.Context, did string) error {
	if m.cache != nil {
		delete(m.cache, did)
	}
	return nil
}

func (m *mockCacheManager) Clear(ctx context.Context) error {
	m.cache = make(map[string]*DIDDocument)
	return nil
}