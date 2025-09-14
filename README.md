# Credence

A decentralized identity and trust score network built in Go implementing W3C standards for decentralized identifiers (DIDs), verifiable credentials (VCs), and P2P networking.

## Overview

Credence provides privacy-preserving trust scores that are verifiable and usable across applications without requiring global financial settlement. The system uses decentralized identifiers (DIDs), verifiable credentials (VCs), and a P2P network to enable gasless operations while maintaining Sybil resistance.

## Quick Start

### Prerequisites

- Go 1.21+
- Docker (for local development stack)

## Architecture

See [docs/designs/architecture.md](docs/designs/architecture.md) for detailed system design.

## Deployment & Integration Guides

Choose your path based on your role:

### 🔧 For Users
- **[End User Wallet](docs/guides/end-user-wallet.md)** - Download and use the Credence Desktop Wallet (5 min setup)

### 🏗️ For Network Operators  
- **[Light Node Guide](docs/guides/light-node.md)** - Run a basic network node (30 min setup, basic rewards)
- **[Full Node Guide](docs/guides/full-node.md)** - Deploy complete infrastructure (2-4 hour setup, high rewards)

### 🏢 For Organizations
- **[Service Provider Guide](docs/guides/service-provider.md)** - Issue credentials for your institution (1-2 hour setup)

### 👩‍💻 For Developers
- **[Developer Integration Guide](docs/guides/developer.md)** - Integrate Credence into your applications (30 min setup)

### 📖 Overview
- **[All Deployment Guides](docs/guides/README.md)** - Complete guide navigation with quick-start matrix

> **Note**: Core services are implemented but some deployment infrastructure (Docker containers, installers, SDKs) is still in development. See individual guides for implementation status.

## License

MIT License - see [LICENSE](LICENSE) for details.