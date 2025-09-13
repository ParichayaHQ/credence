# Claude Development Notes

This file contains development notes and context for AI assistants working on the Credence project.

## Project Overview

Credence is a decentralized identity and trust score network built in Go. The system provides privacy-preserving trust scores that are verifiable and usable across applications without requiring global financial settlement.

## Architecture Context

- **P2P Fabric**: Uses libp2p for networking with gossipsub messaging
- **Trust Scoring**: Deterministic algorithm combining K(YC/PoP) + A(credentials) + V(vouches) - R(reports) + T(time)
- **Transparency**: Events logged in Merkle tree with periodic checkpoints
- **Identity**: Uses DIDs (did:key) with Ed25519 signatures and SD-JWT for selective disclosure
- **Storage**: Content-addressed using IPFS CIDs with RocksDB persistence

## Development Commands

### Testing
```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -race -v ./...

# Run specific packages
go test -v ./internal/events ./internal/crypto

# Run integration tests
go test -v ./tests

# Run benchmarks
go test -bench=. ./...
```

### Building
```bash
# Build all packages
go build ./...

# Build specific service (when implemented)
go build ./cmd/walletd
```

### Code Quality
```bash
# Format code
go fmt ./...

# Lint (if golangci-lint is available)
golangci-lint run

# Tidy dependencies
go mod tidy
```

## Project Status

**📋 For current project status and TODO tracking, always refer to `TODO.md`**

This file contains the authoritative project roadmap, completion status, and task breakdown. When working on the project:

1. **Check TODO.md first** to understand current status and priorities
2. **Update TODO.md** as you complete tasks or identify new work
3. **Use the checklist format** to track progress on complex features
4. **Reference TODO.md** when providing status updates or planning next steps

The TODO.md file is the single source of truth for project progress and should be kept current during development work.

## Key Design Decisions

### Security-First Approach
- All cryptographic operations use constant-time algorithms where applicable
- Deterministic canonicalization for consensus-critical functions
- Comprehensive input validation and sanitization

### Interface-Driven Development
- Clean separation between internal implementation and public APIs
- Mock implementations available for all services
- Type-safe contracts enable parallel team development

### Testing Strategy
- Property-based testing for canonicalization (determinism)
- Integration tests covering full event lifecycle
- Benchmarks for performance-critical operations
- Race detection for concurrent safety

## Known Limitations & Future Work

1. **DID Format**: Currently using base64url encoding instead of proper base58btc multicodec for did:key identifiers. This should be updated for production compatibility.

2. **BLS Signatures**: Interfaces defined but implementation pending. Required for checkpoint committee threshold signatures.

3. **VRF Implementation**: Interfaces ready but VRF proofs not yet implemented. Needed for committee selection.

4. **Status List 2021**: Referenced in data structures but revocation checking not fully implemented.

## Team Coordination

### For Service Teams
- Use interfaces in `pkg/interfaces/` for service contracts
- Reference data structures in `pkg/types/`
- Leverage test utilities in `internal/testutil/` for mocking dependencies

### For Integration
- Event schemas in `internal/events/` are the canonical format
- CID generation in `internal/cid/` ensures content addressing consistency
- Crypto primitives in `internal/crypto/` provide secure operations

### Common Patterns
- All services should implement health check interfaces
- Use context.Context for cancellation and timeouts
- Follow structured error handling with typed errors
- Maintain deterministic behavior for consensus components

## Debugging Tips

### Event Validation Issues
- Check DID format matches regex in `internal/events/validation.go`
- Ensure all required fields are present and non-empty
- Verify epoch format is "YYYY-MM"
- Confirm nonce is valid base64 with at least 12 bytes when decoded

### Canonicalization Problems
- JSON keys must be sorted alphabetically
- Empty/nil values are omitted from canonical form
- Use `events.CanonicalizeJSON()` for consistent serialization

### Signature Verification
- Signatures are detached (not embedded in canonical JSON)
- Use `event.ToSignable()` to get the signable portion
- Verify against the canonicalized bytes, not the original JSON

## Performance Considerations

### Benchmarks (as of implementation)
- Event creation: ~100μs per event
- Canonicalization: ~50μs for typical events
- Ed25519 signing: ~100μs per signature
- Ed25519 verification: ~200μs per verification

### Optimization Opportunities
- CID generation could be cached for repeated content
- Event validation could use compiled regex patterns
- Batch signature verification for multiple events

## Development Guidelines for Claude Code

When working on this project with AI agents, follow these guidelines to maintain code quality and architectural consistency:

### Code Standards
1. **Security First**: Never compromise on cryptographic security or input validation
2. **Determinism**: Ensure all consensus-critical operations produce identical results across runs
3. **Error Handling**: Use typed errors from respective packages (e.g., `events.ErrInvalidSignature`)
4. **Testing**: Write tests for all new functionality with proper error cases
5. **Documentation**: Include clear comments for complex algorithms and cryptographic operations

### Architecture Principles
1. **Interface Segregation**: Keep interfaces focused and minimal
2. **Dependency Injection**: Use interfaces for external dependencies to enable testing
3. **Immutability**: Prefer immutable data structures, especially for events and signatures
4. **Context Propagation**: Always use `context.Context` for cancellation and timeouts
5. **Type Safety**: Leverage Go's type system to prevent runtime errors

### AI Agent Coordination
1. **Read First**: Always examine existing code patterns before implementing new features
2. **Test Integration**: Run tests after significant changes to ensure no regressions
3. **Interface Compliance**: Ensure new services implement required interfaces from `pkg/interfaces/`
4. **Mock Implementations**: Provide mock versions for testing in `internal/testutil/`
5. **Incremental Development**: Build features incrementally with tests at each step

### Code Organization
1. **Package Structure**: Follow the established directory layout
   - `internal/` for private packages
   - `pkg/` for public APIs
   - `cmd/` for service entry points
   - `tests/` for integration tests
2. **Import Order**: Standard library, third-party, local packages
3. **Naming**: Use clear, descriptive names following Go conventions
4. **File Structure**: Keep files focused on single responsibilities

### Performance Guidelines
1. **Crypto Operations**: Cache expensive operations where safe (e.g., public key derivation)
2. **Memory Management**: Avoid unnecessary allocations in hot paths
3. **Concurrency**: Use goroutines for I/O bound operations, avoid for CPU-bound crypto
4. **Benchmarking**: Add benchmarks for performance-critical functions

### Security Guidelines
1. **Input Validation**: Validate all external inputs using the validation framework
2. **Constant Time**: Use constant-time operations for cryptographic comparisons
3. **Key Management**: Never log or expose private key material
4. **Random Generation**: Use `crypto/rand` for all cryptographic randomness
5. **Error Leakage**: Avoid leaking sensitive information in error messages

### Testing Strategy
1. **Unit Tests**: Test individual functions with edge cases
2. **Integration Tests**: Test complete workflows end-to-end
3. **Property Tests**: Use property-based testing for deterministic operations
4. **Mock Usage**: Use provided mocks for external dependencies
5. **Race Detection**: Run tests with `-race` flag for concurrent code

### Common Patterns to Follow

#### Event Creation
```go
// Always validate before processing
if err := events.ValidateEvent(event); err != nil {
    return nil, fmt.Errorf("invalid event: %w", err)
}

// Canonicalize before signing/hashing
canonical, err := events.CanonicalizeEvent(signable)
if err != nil {
    return nil, fmt.Errorf("canonicalization failed: %w", err)
}
```

#### Service Implementation
```go
type MyService struct {
    // Inject dependencies via interfaces
    storage interfaces.StorageService
    crypto  interfaces.Signer
}

func (s *MyService) DoWork(ctx context.Context, input string) error {
    // Always check context first
    if ctx.Err() != nil {
        return ctx.Err()
    }
    
    // Validate inputs
    if input == "" {
        return ErrInvalidInput
    }
    
    // Implement logic...
    return nil
}
```

#### Error Handling
```go
// Use typed errors
var ErrMySpecificError = errors.New("my specific error")

// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// Check error types when needed
if errors.Is(err, events.ErrInvalidSignature) {
    // Handle signature error specifically
}
```

### Forbidden Patterns
1. **Direct Crypto Libraries**: Use wrappers in `internal/crypto/` instead of calling crypto libraries directly
2. **Hardcoded Values**: Use constants or configuration for all magic numbers/strings
3. **Global State**: Avoid global variables, use dependency injection
4. **Panic in Services**: Services should return errors, not panic
5. **Unsafe Operations**: No `unsafe` package usage without explicit approval

## Dependencies

### Core Dependencies
- `github.com/ipfs/go-cid` - Content addressing
- `golang.org/x/crypto/ed25519` - Cryptographic operations
- `github.com/multiformats/go-multihash` - Multihash support
- `github.com/go-playground/validator/v10` - Validation framework

### Test Dependencies
- `github.com/stretchr/testify` - Testing assertions and utilities

All dependencies are regularly maintained and security-audited packages from the Go ecosystem.