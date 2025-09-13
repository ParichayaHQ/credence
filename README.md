# Credence

A decentralized identity and trust score network built in Go.

## Overview

Credence provides privacy-preserving trust scores that are verifiable and usable across applications without requiring global financial settlement. The system uses decentralized identifiers (DIDs), verifiable credentials (VCs), and a P2P network to enable gasless operations while maintaining Sybil resistance.

## Architecture

See [docs/designs/architecture.md](docs/designs/architecture.md) for detailed system design.

## Development

### Prerequisites

- Go 1.21+
- Docker (for local development stack)

### Building

```bash
go mod tidy
go build ./...
```

### Testing

```bash
go test ./...
```

### Services

This repository contains multiple services:

- **walletd**: DID/VC wallet and event signing
- **p2p-gateway**: libp2p networking layer
- **fullnode**: Blob storage and proof serving
- **lognode**: Transparency log (Trillian personality)
- **checkpointor**: Committee checkpoint aggregation
- **scorer**: Deterministic trust score computation

## Development Status

🚧 **In Development** - Core Infrastructure phase

## License

TBD