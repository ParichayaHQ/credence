---
layout: guide
title: "Developer Integration Guide"
description: "Integrate Credence into your applications with SDKs and APIs"
collection: guides
status: draft
---

# Developer Integration Guide

> **⚠️ IMPLEMENTATION STATUS**  
> The core services provide APIs that can be integrated. However, the **packaged SDKs (@credence/sdk), npm packages, development tools, example applications, and some API endpoints** described here are not yet implemented. Integration is possible but requires direct API calls.

## Overview

This guide covers how to integrate Credence into your applications, whether you're building wallets, verifiers, or custom services that interact with the decentralized trust network.

## Development Environment Setup

### Prerequisites
```bash
# Install required tools
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Node.js 18+
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
```

### Clone & Setup
```bash
# Clone the repository
git clone https://github.com/ParichayaHQ/credence.git
cd credence/

# Start local development network
docker-compose -f docker-compose.dev.yml up -d

# Install client libraries
npm install @credence/sdk
# or
go get github.com/ParichayaHQ/credence/go-sdk
# or  
pip install credence-sdk
```

### Local Network
The development setup includes:
- **3 P2P nodes** for testing network gossip
- **1 full node** with all services
- **Mock issuer** for testing credential flows
- **Test data** pre-populated for development

```bash
# Check local network status
curl http://localhost:8080/status

# Expected response:
{
  "network": "dev",
  "peers": 3,
  "services": ["p2p-gateway", "fullnode", "scorer", "issuer"],
  "status": "healthy"
}
```

## SDK Integration

### JavaScript/TypeScript SDK

#### Installation
```bash
npm install @credence/sdk
```

#### Basic Usage
```typescript
import { CredenceClient, WalletManager } from '@credence/sdk';

// Initialize client
const client = new CredenceClient({
  network: 'testnet', // or 'mainnet'
  endpoints: ['http://localhost:8080'] // local dev
});

// Create wallet manager
const wallet = new WalletManager({
  storage: 'local', // or 'memory' for testing
  keyType: 'Ed25519'
});

// Generate identity
const identity = await wallet.createIdentity({
  type: 'did:key',
  alias: 'my-main-identity'
});

console.log('Created DID:', identity.did);
```

#### Credential Operations
```typescript
// Receive a credential
async function receiveCredential(credentialOffer: string) {
  const credential = await wallet.receiveCredential({
    offer: credentialOffer,
    identity: identity.did
  });
  
  await wallet.storeCredential(credential);
  return credential;
}

// Create presentation
async function createPresentation(requirements: any[]) {
  const presentation = await wallet.createPresentation({
    credentials: wallet.findCredentials(requirements),
    challenge: requirements.challenge,
    domain: requirements.domain
  });
  
  return presentation;
}

// Verify presentation
async function verifyPresentation(presentation: any) {
  const result = await client.verify({
    presentation,
    trustPolicy: {
      minimumTrustScore: 0.7,
      requiredCredentialTypes: ['UniversityDiploma']
    }
  });
  
  return result.verified;
}
```

#### Trust Score Operations
```typescript
// Get trust score
async function getTrustScore(did: string) {
  const score = await client.getTrustScore(did);
  return {
    score: score.value,
    confidence: score.confidence,
    lastUpdated: score.timestamp,
    factors: score.breakdown
  };
}

// Submit vouch
async function vouchForIdentity(targetDid: string, strength: number) {
  const vouch = await wallet.createVouch({
    target: targetDid,
    strength, // 0.0 to 1.0
    reason: 'Professional collaboration',
    metadata: {
      relationship: 'colleague',
      duration: '2-years'
    }
  });
  
  await client.publishEvent(vouch);
  return vouch.id;
}

// Report bad actor
async function reportIdentity(targetDid: string, reason: string) {
  const report = await wallet.createReport({
    target: targetDid,
    severity: 'high',
    category: 'fraud',
    description: reason,
    evidence: [] // Optional evidence attachments
  });
  
  await client.publishEvent(report);
  return report.id;
}
```

### Go SDK

#### Installation
```bash
go get github.com/ParichayaHQ/credence/go-sdk
```

#### Basic Usage
```go
package main

import (
    "context"
    "log"
    
    "github.com/ParichayaHQ/credence/go-sdk/client"
    "github.com/ParichayaHQ/credence/go-sdk/wallet"
)

func main() {
    // Initialize client
    client, err := client.New(&client.Config{
        Network: "testnet",
        Endpoints: []string{"http://localhost:8080"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Create wallet
    wallet, err := wallet.New(&wallet.Config{
        StoragePath: "./wallet-data",
        KeyType: "Ed25519",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate identity
    identity, err := wallet.CreateIdentity(context.Background(), &wallet.CreateIdentityRequest{
        Type: "did:key",
        Alias: "main-identity",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created DID: %s", identity.DID)
}
```

#### Credential Verification
```go
func verifyCredential(ctx context.Context, credential []byte) (*VerificationResult, error) {
    result, err := client.VerifyCredential(ctx, &client.VerifyCredentialRequest{
        Credential: credential,
        TrustPolicy: &client.TrustPolicy{
            MinimumTrustScore: 0.7,
            MaxCredentialAge: time.Hour * 24 * 30, // 30 days
            RequiredIssuers: []string{"did:web:university.edu"},
        },
    })
    if err != nil {
        return nil, err
    }
    
    return &VerificationResult{
        Valid: result.Valid,
        TrustScore: result.IssuerTrustScore,
        Reasons: result.ValidationReasons,
    }, nil
}
```

### Python SDK

#### Installation
```bash
pip install credence-sdk
```

#### Basic Usage
```python
from credence_sdk import CredenceClient, Wallet

# Initialize
client = CredenceClient(
    network="testnet",
    endpoints=["http://localhost:8080"]
)

wallet = Wallet(storage_path="./wallet_data")

# Create identity
identity = wallet.create_identity(
    type="did:key",
    alias="main-identity"
)

print(f"Created DID: {identity.did}")

# Get trust score
trust_score = client.get_trust_score(identity.did)
print(f"Trust score: {trust_score.value}")
```

## API Reference

### REST API Endpoints

#### Identity Operations
```bash
# Get DID document
GET /api/v1/identities/{did}

# Create new identity  
POST /api/v1/identities
{
  "type": "did:key",
  "keyType": "Ed25519"
}

# Update DID document
PUT /api/v1/identities/{did}
{
  "document": { /* DID document */ }
}
```

#### Credential Operations
```bash
# Issue credential
POST /api/v1/credentials/issue
{
  "issuer": "did:web:university.edu",
  "subject": "did:key:z6Mk...",
  "type": "UniversityDiploma",
  "claims": { /* credential data */ }
}

# Verify credential
POST /api/v1/credentials/verify
{
  "credential": "eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSJ9...",
  "trustPolicy": {
    "minimumTrustScore": 0.7
  }
}

# Check revocation status
GET /api/v1/credentials/{id}/status
```

#### Trust Score Operations
```bash
# Get trust score
GET /api/v1/trust/{did}

# Submit vouch
POST /api/v1/trust/vouch
{
  "target": "did:key:z6Mk...",
  "strength": 0.8,
  "reason": "Professional collaboration"
}

# Submit report
POST /api/v1/trust/report
{
  "target": "did:key:z6Mk...",
  "category": "fraud",
  "severity": "high",
  "description": "Fake credentials detected"
}
```

### gRPC API

#### Protocol Definitions
```protobuf
// identity.proto
service IdentityService {
  rpc CreateIdentity(CreateIdentityRequest) returns (Identity);
  rpc GetIdentity(GetIdentityRequest) returns (Identity);
  rpc UpdateIdentity(UpdateIdentityRequest) returns (Identity);
}

message Identity {
  string did = 1;
  string document = 2;
  int64 created_at = 3;
  int64 updated_at = 4;
}
```

#### Client Usage
```go
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewIdentityServiceClient(conn)

identity, err := client.CreateIdentity(context.Background(), &pb.CreateIdentityRequest{
    Type: "did:key",
    KeyType: "Ed25519",
})
```

## Integration Patterns

### Web Application
```html
<!DOCTYPE html>
<html>
<head>
    <script src="https://unpkg.com/@credence/sdk-web@latest/dist/credence.min.js"></script>
</head>
<body>
    <script>
        const credence = new Credence({
            network: 'testnet'
        });
        
        // Request credential verification
        async function verifyUser() {
            const presentation = await credence.requestPresentation({
                requirements: [{
                    type: 'UniversityDiploma',
                    issuer: 'did:web:university.edu'
                }]
            });
            
            const result = await credence.verify(presentation);
            if (result.verified) {
                // Grant access
                showUserDashboard();
            }
        }
    </script>
</body>
</html>
```

### Mobile Application (React Native)
```javascript
import { CredenceSDK } from '@credence/react-native-sdk';

const App = () => {
  const [credence, setCredence] = useState(null);
  
  useEffect(() => {
    const initCredence = async () => {
      const sdk = await CredenceSDK.initialize({
        network: 'testnet',
        storage: 'secure' // Uses device secure storage
      });
      setCredence(sdk);
    };
    initCredence();
  }, []);
  
  const handleCredentialRequest = async () => {
    try {
      const credential = await credence.receiveCredential({
        qrCode: scannedQRCode
      });
      
      Alert.alert('Success', 'Credential received!');
    } catch (error) {
      Alert.alert('Error', error.message);
    }
  };
  
  return (
    <View>
      <Button title="Scan Credential" onPress={handleCredentialRequest} />
    </View>
  );
};
```

### Backend Verification Service
```javascript
const express = require('express');
const { CredenceClient } = require('@credence/sdk');

const app = express();
const credence = new CredenceClient({
  network: 'mainnet',
  endpoints: process.env.CREDENCE_ENDPOINTS.split(',')
});

app.post('/verify', async (req, res) => {
  try {
    const { presentation, requirements } = req.body;
    
    const result = await credence.verifyPresentation({
      presentation,
      trustPolicy: {
        minimumTrustScore: 0.8,
        requiredCredentialTypes: requirements.types,
        maxCredentialAge: '30d'
      }
    });
    
    if (result.verified) {
      // Extract verified claims
      const claims = result.verifiedClaims;
      res.json({
        verified: true,
        claims: claims,
        trustScore: result.issuerTrustScore
      });
    } else {
      res.status(400).json({
        verified: false,
        reasons: result.failureReasons
      });
    }
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.listen(3000);
```

## Testing & Development

### Local Testing
```bash
# Start test network
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
npm test

# Test specific credential flow
./scripts/test-credential-flow.sh --issuer university --holder student --verifier employer
```

### Mock Services
```javascript
// Use mock services for testing
const credence = new CredenceClient({
  network: 'mock',
  endpoints: ['http://localhost:8080'],
  options: {
    mockTrustScores: {
      'did:key:alice': 0.9,
      'did:key:bob': 0.7,
      'did:key:mallory': 0.1
    }
  }
});
```

### Test Data Generation
```bash
# Generate test identities
./scripts/generate-test-data.sh --identities 100 --vouches 500

# Create test credentials
./scripts/create-test-credentials.sh --type diploma --count 50 --issuer university.edu
```

## Performance & Optimization

### Caching Strategies
```javascript
// Cache trust scores locally
const trustScoreCache = new Map();

async function getTrustScore(did) {
  if (trustScoreCache.has(did)) {
    const cached = trustScoreCache.get(did);
    if (Date.now() - cached.timestamp < 300000) { // 5 minutes
      return cached.score;
    }
  }
  
  const score = await credence.getTrustScore(did);
  trustScoreCache.set(did, {
    score,
    timestamp: Date.now()
  });
  
  return score;
}
```

### Batch Operations
```javascript
// Batch verify multiple credentials
const results = await credence.batchVerify({
  credentials: credentialList,
  concurrency: 5, // Verify 5 at a time
  trustPolicy: commonTrustPolicy
});
```

### Connection Pooling
```go
// Use connection pooling for high-throughput applications
client, err := client.New(&client.Config{
    Network: "mainnet",
    Endpoints: endpoints,
    PoolSize: 20,
    MaxRetries: 3,
    Timeout: time.Second * 30,
})
```

## Security Best Practices

### Key Management
```javascript
// Use hardware security modules in production
const wallet = new WalletManager({
  keyStorage: 'hsm',
  hsmConfig: {
    provider: 'aws-cloudhsm',
    region: 'us-east-1'
  }
});
```

### Input Validation
```javascript
// Always validate inputs
function validateDID(did) {
  const didRegex = /^did:[a-z0-9]+:[a-zA-Z0-9._%-]*[a-zA-Z0-9._%-]+$/;
  if (!didRegex.test(did)) {
    throw new Error('Invalid DID format');
  }
  return did;
}
```

### Rate Limiting
```javascript
// Implement rate limiting for API calls
const rateLimiter = new RateLimiter({
  points: 100, // Number of requests
  duration: 60, // Per 60 seconds
});

await rateLimiter.consume(userKey);
```

## Troubleshooting

### Common Issues

**Connection timeouts:**
```javascript
// Increase timeout and add retry logic
const client = new CredenceClient({
  endpoints: [...],
  timeout: 30000,
  retries: 3,
  retryDelay: 1000
});
```

**Invalid signatures:**
```javascript
// Verify key material is correct
const isValid = await wallet.verifyKeyPair(publicKey, privateKey);
if (!isValid) {
  throw new Error('Key pair validation failed');
}
```

**Trust score calculation delays:**
```javascript
// Use cached scores for better UX
const score = await getTrustScoreWithFallback(did);

async function getTrustScoreWithFallback(did) {
  try {
    return await credence.getTrustScore(did, { timeout: 5000 });
  } catch (error) {
    // Return cached score or default
    return getCachedTrustScore(did) || { value: 0.5, confidence: 'low' };
  }
}
```

### Debug Mode
```javascript
// Enable debug logging
const credence = new CredenceClient({
  debug: true,
  logLevel: 'debug'
});

// Monitor network requests
credence.on('request', (req) => {
  console.log('Request:', req);
});

credence.on('response', (res) => {
  console.log('Response:', res);
});
```

## Resources

### Documentation
- **API Reference**: [api.credence.network](https://api.credence.network)
- **SDK Documentation**: [docs.credence.network/sdk](https://docs.credence.network/sdk)
- **Examples Repository**: [github.com/ParichayaHQ/credence-examples](https://github.com/ParichayaHQ/credence-examples)

### Community
- **Discord**: [discord.gg/credence](https://discord.gg/credence)
- **Forum**: [forum.credence.network](https://forum.credence.network)
- **Stack Overflow**: Tag `credence-network`

### Support
- **GitHub Issues**: [github.com/ParichayaHQ/credence/issues](https://github.com/ParichayaHQ/credence/issues)
- **Email**: developers@credence.network

Happy building with Credence! 🚀