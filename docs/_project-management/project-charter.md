---
layout: default
title: "Project Charter"
description: "Vision, goals, and roadmap for the Credence project"
collection: project-management
---

# Team-Based Parallel Development Strategy

## Core Team Structure (6-8 Teams)

### Team 1: **Core Infrastructure** (2-3 devs)
```
Responsibilities:
├── Repository structure & shared libraries
├── Event schemas & canonicalization  
├── Cryptographic primitives (Ed25519, BLS, VRF)
├── CID generation & content addressing
├── Common data structures & interfaces
└── Integration test framework
```
**Deliverable**: Foundation that all other teams depend on

### Team 2: **P2P Networking** (2-3 devs)
```
Responsibilities:
├── libp2p host implementation
├── gossipsub with exact mesh parameters
├── DHT for content discovery
├── Rate limiting & peer reputation
├── Topic management (events/*, revocations/*, etc.)
└── HTTP bridge for internal services
```
**Deliverable**: p2p-gateway service
**Dependencies**: Core Infrastructure schemas

### Team 3: **Storage & Persistence** (2-3 devs)
```
Responsibilities:
├── fullnode (RocksDB + blob storage)
├── Event indexing strategies
├── Mirror services (checkpoints, rules, StatusLists)
├── lognode (Trillian personality)
├── Database schemas & migrations
└── Proof serving infrastructure
```
**Deliverable**: fullnode + lognode services
**Dependencies**: Core Infrastructure, some P2P coordination

### Team 4: **Trust Scoring Engine** (2-3 devs)
```
Responsibilities:
├── Complete scoring algorithm implementation
├── Decay functions & diversity weighting
├── Vouch budget enforcement
├── Graph analytics for anti-collusion
├── Score caching & incremental updates
└── Proof bundling for verification
```
**Deliverable**: scorer service
**Dependencies**: Core Infrastructure, Storage for reading events

### Team 5: **Identity & Credentials** (2-3 devs)
```
Responsibilities:
├── DID method implementation (did:key, did:web)
├── VC-JWT & SD-JWT integration
├── StatusList 2021 revocation
├── Wallet key management
├── Credential presentation flows
└── issuer service for VC issuance
```
**Deliverable**: walletd service + issuer service
**Dependencies**: Core Infrastructure, some P2P for publishing

### Team 6: **Consensus & Governance** (2-3 devs)
```
Responsibilities:
├── Checkpoint committee implementation
├── BLS threshold signature aggregation
├── VRF-based committee selection
├── rules-registry with time-locks
├── Committee rotation logic
└── Governance proposal system
```
**Deliverable**: checkpointor + rules-registry services
**Dependencies**: Core Infrastructure, Storage for checkpoints

### Team 7: **Client Applications** (3-4 devs)
```
Responsibilities:
├── Web wallet interface
├── Mobile wallet (React Native/Flutter)
├── Desktop wallet application
├── Demo applications (forum, marketplace)
├── API client libraries
└── Developer documentation
```
**Deliverable**: User-facing applications
**Dependencies**: Most other services via APIs

### Team 8: **DevOps & Integration** (2-3 devs)
```
Responsibilities:
├── Docker containerization
├── Kubernetes deployment configs
├── CI/CD pipelines
├── Monitoring & observability setup
├── Integration testing infrastructure
└── Production deployment automation
```
**Deliverable**: Complete deployment pipeline
**Dependencies**: All services

## Development Stages & Integration Points

### Stage 1: Foundation
```
Team Priority Order:
1. Core Infrastructure
2. P2P Networking, Storage, Identity
3. Scoring, Consensus
4. Client Apps, DevOps
```

### Stage 2: Service Integration
```
Integration Milestones:
├── Core + P2P: Event publishing works
├── Storage + P2P: Event persistence & retrieval
├── Scoring + Storage: Score computation from events
├── Identity + All: Full vouch → score flow
├── Consensus + All: Checkpoint creation & verification
└── DevOps: All services running in containers
```

### Stage 3: Application Integration
```
End-to-End Flows:
├── Wallet creates & publishes vouch
├── Storage persists, Scoring updates
├── Consensus creates checkpoint
├── Wallet fetches score with proofs
├── Demo app verifies score ≥ threshold
└── Full monitoring & alerting
```

## Team Coordination Strategy

### 1. **Interface-First Development**
Each team starts by defining their service's:
- gRPC/HTTP API contracts
- Database schemas  
- Configuration interfaces
- Error types and handling

### 2. **Shared Repository Structure**
```
/services
  /core-infrastructure/    # Team 1
  /p2p-gateway/           # Team 2  
  /fullnode/              # Team 3a
  /lognode/               # Team 3b
  /scorer/                # Team 4
  /walletd/               # Team 5a
  /issuer/                # Team 5b
  /checkpointor/          # Team 6a
  /rules-registry/        # Team 6b
  /client-apps/           # Team 7
  /deploy/                # Team 8

/shared                   # Owned by Team 1
  /internal/events/       # Event schemas
  /internal/crypto/       # Crypto primitives
  /internal/api/          # API definitions
  /internal/config/       # Configuration
```

### 3. **Weekly Integration Cycles**
```
Monday: Team updates & dependency planning
Wednesday: Integration builds & conflict resolution  
Friday: End-to-end testing & weekly demo
```

### 4. **Dependency Management**
```yaml
# Clear dependency graph
core-infrastructure: []
p2p-gateway: [core-infrastructure]
storage: [core-infrastructure, p2p-gateway]
scoring: [core-infrastructure, storage]
identity: [core-infrastructure, p2p-gateway]
consensus: [core-infrastructure, storage]
clients: [ALL_SERVICES]
devops: [ALL_SERVICES]
```

## AI Agent Integration per Team

Each team can use AI agents for:
- **Boilerplate generation** within their service boundaries
- **Test suite creation** for their components
- **Documentation generation** for their APIs
- **Performance optimization** within their domain

But **human coordination** handles:
- Cross-service API design
- Integration testing
- Architectural decisions
- Conflict resolution

## Success Metrics & Gates

### Phase 1 Gate: "Services Start"
- ✅ All service interfaces defined
- ✅ Core infrastructure APIs stable  
- ✅ Each service can start & expose health checks

### Phase 2 Gate: "Basic Integration"
- ✅ Vouch event: create → publish → store → score
- ✅ All services communicate via defined APIs
- ✅ Basic monitoring & logging working

### Phase 3 Gate: "End-to-End Demo"
- ✅ Complete user flow: vouch → score → verify
- ✅ Demo applications working
- ✅ Production deployment ready

## Risk Mitigation

### **Integration Hell Prevention:**
- Daily integration builds from all teams
- Shared API testing contracts  
- Early & frequent cross-team communication

### **Scope Creep Control:**
- Teams stick to exact architecture specifications
- Changes require cross-team approval
- AI agents prevent feature drift within services

### **Dependency Bottlenecks:**
- Core Infrastructure team prioritizes blockers
- Mock services for parallel development
- Clear escalation paths for conflicts

This approach lets you scale development horizontally while maintaining architectural coherence. Each team can move fast within their domain while integration points keep everyone aligned.