package did

import (
	"context"
	"fmt"
	"sync"
)

// MultiDIDResolver implements a multi-method DID resolver
type MultiDIDResolver struct {
	methodResolvers map[string]MethodResolver
	defaultOptions  *DIDResolutionOptions
	mu              sync.RWMutex
}

// NewMultiDIDResolver creates a new multi-method DID resolver
func NewMultiDIDResolver() *MultiDIDResolver {
	resolver := &MultiDIDResolver{
		methodResolvers: make(map[string]MethodResolver),
		defaultOptions: &DIDResolutionOptions{
			Accept: "application/did+ld+json",
		},
	}
	
	// Register default did:key resolver
	keyResolver := NewKeyMethodResolver(NewDefaultKeyManager())
	resolver.RegisterMethod("key", keyResolver)
	
	return resolver
}

// Resolve resolves a DID to a DID document
func (r *MultiDIDResolver) Resolve(ctx context.Context, did string, options *DIDResolutionOptions) (*DIDResolutionResult, error) {
	parsed, err := ParseDID(did)
	if err != nil {
		return &DIDResolutionResult{
			DIDResolutionMetadata: DIDResolutionMetadata{
				Error:        ErrorInvalidDID,
				ErrorMessage: "invalid DID syntax: " + err.Error(),
			},
		}, nil
	}
	
	return r.ResolveWithMethod(ctx, did, parsed.Method, options)
}

// ResolveWithMethod resolves a DID using a specific method resolver
func (r *MultiDIDResolver) ResolveWithMethod(ctx context.Context, did string, method string, options *DIDResolutionOptions) (*DIDResolutionResult, error) {
	r.mu.RLock()
	methodResolver, exists := r.methodResolvers[method]
	r.mu.RUnlock()
	
	if !exists {
		return &DIDResolutionResult{
			DIDResolutionMetadata: DIDResolutionMetadata{
				Error:        ErrorMethodNotSupported,
				ErrorMessage: fmt.Sprintf("DID method '%s' is not supported", method),
			},
		}, nil
	}
	
	// Merge options with defaults
	resolveOptions := r.mergeOptions(options)
	
	return methodResolver.Resolve(ctx, did, resolveOptions)
}

// SupportsMethod checks if the resolver supports a specific DID method
func (r *MultiDIDResolver) SupportsMethod(method string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.methodResolvers[method]
	return exists
}

// SupportedMethods returns a list of supported DID methods
func (r *MultiDIDResolver) SupportedMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	methods := make([]string, 0, len(r.methodResolvers))
	for method := range r.methodResolvers {
		methods = append(methods, method)
	}
	
	return methods
}

// RegisterMethod registers a method resolver
func (r *MultiDIDResolver) RegisterMethod(method string, resolver MethodResolver) error {
	if method == "" {
		return NewDIDError(ErrorInvalidDID, "method name cannot be empty")
	}
	
	if resolver == nil {
		return NewDIDError(ErrorInternalError, "resolver cannot be nil")
	}
	
	if !IsValidMethod(method) {
		return NewDIDError(ErrorInvalidDID, "invalid method name: "+method)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.methodResolvers[method] = resolver
	return nil
}

// UnregisterMethod unregisters a method resolver
func (r *MultiDIDResolver) UnregisterMethod(method string) error {
	if method == "" {
		return NewDIDError(ErrorInvalidDID, "method name cannot be empty")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.methodResolvers[method]; !exists {
		return NewDIDError(ErrorMethodNotSupported, "method not registered: "+method)
	}
	
	delete(r.methodResolvers, method)
	return nil
}

// GetResolver gets the resolver for a specific method
func (r *MultiDIDResolver) GetResolver(method string) (MethodResolver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	resolver, exists := r.methodResolvers[method]
	if !exists {
		return nil, NewDIDError(ErrorMethodNotSupported, "method not supported: "+method)
	}
	
	return resolver, nil
}

// ListMethods lists all registered methods
func (r *MultiDIDResolver) ListMethods() []string {
	return r.SupportedMethods()
}

// SetDefaultOptions sets default resolution options
func (r *MultiDIDResolver) SetDefaultOptions(options *DIDResolutionOptions) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.defaultOptions = options
}

// GetDefaultOptions gets default resolution options
func (r *MultiDIDResolver) GetDefaultOptions() *DIDResolutionOptions {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.defaultOptions == nil {
		return &DIDResolutionOptions{}
	}
	
	// Return a copy to prevent external modification
	return &DIDResolutionOptions{
		Accept:      r.defaultOptions.Accept,
		VersionId:   r.defaultOptions.VersionId,
		VersionTime: r.defaultOptions.VersionTime,
		Transform:   r.defaultOptions.Transform,
		Properties:  copyProperties(r.defaultOptions.Properties),
	}
}

// mergeOptions merges provided options with default options
func (r *MultiDIDResolver) mergeOptions(options *DIDResolutionOptions) *DIDResolutionOptions {
	if options == nil {
		return r.GetDefaultOptions()
	}
	
	defaults := r.GetDefaultOptions()
	merged := &DIDResolutionOptions{
		Accept:      options.Accept,
		VersionId:   options.VersionId,
		VersionTime: options.VersionTime,
		Transform:   options.Transform,
		Properties:  copyProperties(options.Properties),
	}
	
	// Use defaults for empty fields
	if merged.Accept == "" {
		merged.Accept = defaults.Accept
	}
	
	// Merge properties
	if defaults.Properties != nil {
		if merged.Properties == nil {
			merged.Properties = make(map[string]interface{})
		}
		
		for key, value := range defaults.Properties {
			if _, exists := merged.Properties[key]; !exists {
				merged.Properties[key] = value
			}
		}
	}
	
	return merged
}

// copyProperties creates a deep copy of properties map
func copyProperties(properties map[string]interface{}) map[string]interface{} {
	if properties == nil {
		return nil
	}
	
	copied := make(map[string]interface{})
	for key, value := range properties {
		copied[key] = value
	}
	
	return copied
}

// CachedResolver wraps a resolver with caching capabilities
type CachedResolver struct {
	resolver     MultiResolver
	cache        CacheManager
	defaultTTL   int64
}

// NewCachedResolver creates a new cached resolver
func NewCachedResolver(resolver MultiResolver, cache CacheManager, defaultTTL int64) *CachedResolver {
	return &CachedResolver{
		resolver:   resolver,
		cache:      cache,
		defaultTTL: defaultTTL,
	}
}

// Resolve resolves a DID with caching
func (r *CachedResolver) Resolve(ctx context.Context, did string, options *DIDResolutionOptions) (*DIDResolutionResult, error) {
	// Try cache first (unless version-specific resolution)
	if options == nil || (options.VersionId == "" && options.VersionTime == "") {
		if document, err := r.cache.Get(ctx, did); err == nil && document != nil {
			return &DIDResolutionResult{
				DIDDocument: document,
				DIDResolutionMetadata: DIDResolutionMetadata{
					ContentType: "application/did+ld+json",
				},
				DIDDocumentMetadata: DIDDocumentMetadata{},
			}, nil
		}
	}
	
	// Resolve using underlying resolver
	result, err := r.resolver.Resolve(ctx, did, options)
	if err != nil {
		return result, err
	}
	
	// Cache successful results (non-error, non-version-specific)
	if result.DIDResolutionMetadata.Error == "" && 
	   result.DIDDocument != nil &&
	   (options == nil || (options.VersionId == "" && options.VersionTime == "")) {
		r.cache.Set(ctx, did, result.DIDDocument, r.defaultTTL)
	}
	
	return result, nil
}

// SupportsMethod checks if the resolver supports a specific DID method
func (r *CachedResolver) SupportsMethod(method string) bool {
	return r.resolver.SupportsMethod(method)
}

// SupportedMethods returns a list of supported DID methods
func (r *CachedResolver) SupportedMethods() []string {
	return r.resolver.SupportedMethods()
}

// RegisterMethod registers a method resolver
func (r *CachedResolver) RegisterMethod(method string, resolver MethodResolver) error {
	return r.resolver.RegisterMethod(method, resolver)
}

// UnregisterMethod unregisters a method resolver
func (r *CachedResolver) UnregisterMethod(method string) error {
	return r.resolver.UnregisterMethod(method)
}

// GetResolver gets the resolver for a specific method
func (r *CachedResolver) GetResolver(method string) (MethodResolver, error) {
	return r.resolver.GetResolver(method)
}

// ListMethods lists all registered methods
func (r *CachedResolver) ListMethods() []string {
	return r.resolver.ListMethods()
}

// ResolveWithMethod resolves a DID using a specific method resolver
func (r *CachedResolver) ResolveWithMethod(ctx context.Context, did string, method string, options *DIDResolutionOptions) (*DIDResolutionResult, error) {
	return r.resolver.ResolveWithMethod(ctx, did, method, options)
}

// SetDefaultOptions sets default resolution options
func (r *CachedResolver) SetDefaultOptions(options *DIDResolutionOptions) {
	r.resolver.SetDefaultOptions(options)
}

// GetDefaultOptions gets default resolution options
func (r *CachedResolver) GetDefaultOptions() *DIDResolutionOptions {
	return r.resolver.GetDefaultOptions()
}

// InvalidateCache invalidates the cache for a specific DID
func (r *CachedResolver) InvalidateCache(ctx context.Context, did string) error {
	return r.cache.Invalidate(ctx, did)
}

// ClearCache clears all cached entries
func (r *CachedResolver) ClearCache(ctx context.Context) error {
	return r.cache.Clear(ctx)
}