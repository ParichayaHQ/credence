# Credence

A decentralized identity and trust score network built in Go implementing W3C standards for decentralized identifiers (DIDs), verifiable credentials (VCs), and P2P networking.

## Overview

Credence provides privacy-preserving trust scores that are verifiable and usable across applications without requiring global financial settlement. The system uses decentralized identifiers (DIDs), verifiable credentials (VCs), and a P2P network to enable gasless operations while maintaining Sybil resistance.

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

### Building

```bash
go build ./...
```

## Architecture

See [docs/designs/architecture.md](docs/designs/architecture.md) for detailed system design.

## License

MIT License - see [LICENSE](LICENSE) for details.