# Credence

A decentralized identity and trust score network built in Go implementing W3C standards for decentralized identifiers (DIDs), verifiable credentials (VCs), and P2P networking.

## Overview

Credence provides privacy-preserving trust scores that are verifiable and usable across applications without requiring global financial settlement. The system uses decentralized identifiers (DIDs), verifiable credentials (VCs), and a P2P network to enable gasless operations while maintaining Sybil resistance.

## Features

### ✅ **Core Identity Infrastructure**
- **Decentralized Identifiers (DIDs)** with `did:key` method support
- **Verifiable Credentials (VCs)** with VC-JWT and SD-JWT formats  
- **StatusList 2021** bitmap-based credential revocation
- **Ed25519** cryptographic operations
- **Comprehensive wallet system** for key and credential management

### ✅ **P2P Networking**
- Trust scoring engine
- Peer discovery and management
- Message routing and validation
- Network consensus mechanisms

### 🔧 **In Progress**
- Full wallet encryption for secure export/import
- Secp256k1 key type support
- Production-grade error handling

## Quick Start

### Prerequisites

- Go 1.21+
- Docker (for local development stack)

### Installation

```bash
git clone https://github.com/ParichayaHQ/credence.git
cd credence
go mod tidy
```

### Testing Core Functionality

```bash
# Test core DID/VC operations (77 tests)
go test ./internal/did/... ./internal/vc/... ./internal/statuslist/...

# Test wallet functionality
go test ./internal/wallet/ -run "TestDefaultWallet_(GenerateKey|CreateDID|StoreCredential)"

# Test P2P networking
go test ./internal/p2p/...

# Run all tests
go test ./...
```

### Building

```bash
go build ./...
```

## Architecture

The system is built with a modular architecture:

```
├── internal/
│   ├── did/          # Decentralized Identifiers
│   ├── vc/           # Verifiable Credentials  
│   ├── statuslist/   # Credential revocation
│   ├── wallet/       # Key & credential management
│   └── p2p/          # Networking layer
├── docs/designs/     # Architecture documentation
└── services/         # Deployable services
```

See [docs/designs/architecture.md](docs/designs/architecture.md) for detailed system design.

## Services

This repository contains multiple services:

- **walletd**: DID/VC wallet and event signing
- **p2p-gateway**: libp2p networking layer
- **fullnode**: Blob storage and proof serving
- **lognode**: Transparency log (Trillian personality)
- **checkpointor**: Committee checkpoint aggregation
- **scorer**: Deterministic trust score computation

## Development Status

🚧 **Core Infrastructure Phase** - Foundation components implemented and tested

**Current Status:**
- ✅ Core DID/VC functionality (90%+ tests passing)
- ✅ Wallet operations (key management, credentials, presentations)
- ✅ Trust scoring system
- ✅ P2P communication primitives
- 🔧 Remaining: Full encryption, additional key types, production hardening

## Usage Examples

### Creating DIDs and Credentials

```go
// Generate a key pair
keyPair, err := wallet.GenerateKey(did.KeyTypeEd25519)

// Create a DID
didRecord, err := wallet.CreateDID(keyPair.ID, "key")

// Issue a credential
credential := &vc.VerifiableCredential{
    Context:      []string{"https://www.w3.org/2018/credentials/v1"},
    Type:         []string{"VerifiableCredential"},
    Issuer:       didRecord.DID,
    IssuanceDate: time.Now().Format(time.RFC3339),
    CredentialSubject: map[string]interface{}{
        "id": "did:key:subject",
    },
}
```

### Managing Trust Scores

```go
// Calculate trust score for a peer
score := trustEngine.CalculateScore(peerID, interactions)

// Update peer reputation
trustEngine.UpdateReputation(peerID, outcome)
```

## License

MIT License - see [LICENSE](LICENSE) for details.