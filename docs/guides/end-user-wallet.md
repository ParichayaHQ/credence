# End User Wallet Guide

> **⚠️ IMPLEMENTATION STATUS**  
> The Credence Desktop Wallet application exists and is functional. However, the **download links, installers (.dmg, .msi, .AppImage), auto-updater, and release infrastructure** described here are not yet implemented. The core wallet functionality works in development mode.

## Overview

For most users, all you need is the **Credence Desktop Wallet** - a simple download with no technical setup required.

## Installation

### Download & Install
```bash
# macOS
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/Credence-mac.dmg -o Credence.dmg
open Credence.dmg

# Windows
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/Credence-win.msi -o Credence.msi
start Credence.msi

# Linux
curl -L https://github.com/ParichayaHQ/credence/releases/latest/download/Credence-linux.AppImage -o Credence.AppImage
chmod +x Credence.AppImage
./Credence.AppImage
```

## What You Get

### Built-in Services
The wallet includes everything needed for basic operations:
- **Local wallet service** (`walletd`) - Key management and credential storage
- **P2P discovery** - Automatically finds network nodes
- **Network client** - Queries remote services for data

### No Server Setup Required
- No Docker containers to run
- No command line tools needed  
- No technical configuration
- Automatic updates via built-in updater

## First Launch

### Onboarding Wizard
1. **Welcome** - Overview of Credence features
2. **Security Setup** - Configure app lock and backup preferences
3. **Key Generation** - Create your first cryptographic key pair
4. **DID Creation** - Generate your decentralized identity
5. **Complete** - Ready to use!

### Network Connection
The wallet automatically:
- Discovers available network nodes via P2P gossip
- Connects to multiple nodes for redundancy
- Handles network failures gracefully
- Updates to new nodes as they join/leave

## Core Features

### Identity Management
- **DID Creation** - Generate `did:key` identities
- **Key Management** - Secure local key storage
- **Backup & Recovery** - Export/import wallet data

### Credential Operations
- **Receive Credentials** - Accept verifiable credentials from issuers
- **Create Presentations** - Bundle credentials for verification
- **Revocation Checking** - Verify credential status automatically

### Trust Network
- **Vouch for Others** - Build trust relationships
- **Check Trust Scores** - View computed trust ratings
- **Network Analytics** - Explore trust relationships

### Privacy & Security
- **Encrypted Storage** - All data encrypted at rest
- **App Lock** - PIN/password protection
- **Selective Disclosure** - Choose what data to share
- **Zero Knowledge** - Verify without revealing details

## Network Participation

### As a User Only
- **No infrastructure required** - Just use your wallet
- **Query existing nodes** - Network provides all services
- **Pay small fees** - Micropayments for queries/transactions

### Becoming a Node Operator (Optional)
If you want to contribute to network infrastructure:
- See [Light Node Guide](./light-node.md) for basic participation
- See [Full Node Guide](./full-node.md) for complete services
- Earn rewards for providing network services

## Support

### Help & Documentation
- **In-app help** - Built-in guides and tutorials
- **Community forum** - [credence.network/community](https://credence.network/community)
- **Developer docs** - [docs.credence.network](https://docs.credence.network)

### Troubleshooting
- **Connection issues** - Wallet will retry with different nodes
- **Sync problems** - Clear cache via Settings > Advanced
- **Lost keys** - Restore from backup phrase
- **App crashes** - Check logs in Settings > Diagnostics

## Privacy & Data

### Local Data Only
- **Keys never leave device** - All cryptographic operations local
- **Encrypted storage** - SQLite database with AES encryption
- **No tracking** - No analytics or telemetry sent

### Network Queries
- **Pseudonymous** - Queries use temporary identifiers
- **Distributed** - Data fetched from multiple nodes
- **Cached locally** - Reduce network requests

## Updates

### Automatic Updates
- **Background downloads** - New versions downloaded automatically
- **User consent** - You choose when to install updates
- **Rollback support** - Revert to previous version if needed
- **Security patches** - Critical updates applied immediately

The Credence wallet is designed to be **simple, secure, and private** - just download and start building your decentralized identity!