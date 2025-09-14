---
layout: default
title: "System Architecture"
description: "High-level architecture and design principles"
collection: designs
---

# System Design: Decentralized Identity & Trust Score Network (Go)

## 0) Goals & Non-Goals

**Goals**

* Decentralized identity (DID + VC) ownership
* Verifiable, privacy-preserving **trust score** usable across apps
* Gasless operations: vouches, reports, revocations, scoring
* Sybil resistance via proof-of-personhood, budgets, and adjudication
* Transparency without full global consensus (CT-style append-only log)

**Non-Goals**

* Global financial settlement or token issuance
* Full censorship resistance against nation-state actors (v1 aims at practical decentralization)
* Perfect anonymity; we target **selective disclosure** and **threshold proofs**

---

## 1) High-Level Architecture

```
+-------------------------+            +---------------------------+
|         Issuers         |            |     Relying Applications  |
|  (KYC, .edu, employer)  |            | ("score ≥ θ?" checks)     |
+------------+------------+            +--------------+------------+
             |                                        |
             v                                        v
+------------+--------------------  P2P Fabric (libp2p)  -----------------------+
| Topics: events/*  revocations/*  rules/*  checkpoints/*   blobs/* (CID fetch) |
+----+----------------+-------------------+------------------+------------------+
     |                |                   |                  |
     v                v                   v                  v
+----+----+     +-----+-----+       +-----+-----+       +----+-----+
| Wallet |     | Full Node |       | Log Node  |       | Scorer   |
| Light  |     |  Storage  |       | (Trillian)|       |  Node     |
| Node   |     |  + Proofs |       |  + Proofs |       | (Determin.)|
+----+----+     +-----+-----+       +-----+-----+       +----+-----+
     |                |                   |                  |
     +----------------+---------+---------+------------------+
                               |
                               v
                      +--------+--------+
                      | Checkpoint Comm |
                      |  (BLS threshold)|
                      +-----------------+
```

---

## 2) Trust & Data Primitives

### 2.1 Identities & Credentials

* **DID Method**: `did:key` (v1) for simplicity; add `did:web`/`did:ion` later.
* **VC Format**: VC-JWT (JWS) with **SD-JWT** (selective disclosure).
* **Revocation**: StatusList 2021 (bitmaps) hosted via HTTP, mirrored on p2p.

### 2.2 Events (content-addressed, signed)

* `Vouch`: `(from DID, to DID, ctx, epoch, weight_hint?, nonce)`
* `Report`: `(reporter, target, ctx, reason_code, evidenceCID?, nonce)`
* `Appeal`: `(appellant, case_id, new_evidenceCID?)`
* `RevocationAnnounce`: `(issuer, statuslistURI, epoch, bitmapCID)`
* **CID** = SHA-256 multihash of canonical JSON; payloads gossiped and pinned.

### 2.3 Transparency & Checkpoints

* **Transparency Log**: Trillian (append-only Merkle tree of event CIDs + headers).
* **Checkpoint**: `{root, tree_id, epoch, signers[], threshold_sig}` produced every Δt (e.g., 10–30 min).

---

## 3) Services & Responsibilities (Go)

| Service          | Responsibilities                                                             | Key libs                   |
| ---------------- | ---------------------------------------------------------------------------- | -------------------------- |
| `walletd`        | DID/VC wallet; event signing; threshold proofs; proof retrieval              | `vc-go`, `jwx`, Ed25519    |
| `p2p-gateway`    | libp2p host; gossipsub; CID fetch; rate limits                               | `go-libp2p`, `go-cid`      |
| `fullnode`       | Blob store (RocksDB/FS); proof server; mirrors StatusLists/rules/checkpoints | `gorocksdb`, HTTP          |
| `lognode`        | Trillian personality: append event refs; inclusion/consistency proofs        | `google/trillian`          |
| `checkpointor`   | Committee sampling (VRF/drand), BLS aggregation, checkpoint gossip           | `herumi bls-eth-go-binary` |
| `scorer`         | Deterministic score compute; diversity/decay; proof bundling                 | custom (pure Go)           |
| `adjudicator`    | Reviewer sampling, vote collection, verdicts → log                           | VRF client, BLS verify     |
| `rules-registry` | Signed ruleset JSON hosting; versioning; time-lock                           | `jwx`                      |
| `issuer`         | VC issuing; StatusList publishing; DID rotation                              | `vc-go`                    |

**Deployment topologies**

* **Light Nodes**: wallet + p2p client (desktop/mobile)
* **Community Full Nodes**: proof/asset serving (hundreds–thousands)
* **Log Nodes**: 3–7 replicated Trillian clusters (HA)
* **Checkpoint Committee**: rotating set of N=50..150 operators; threshold t ≈ ⌊2N/3⌋

---

## 4) Data Schemas (Canonical JSON)

### 4.1 Canonical Event (signed body)

```json
{
  "type": "vouch|report|appeal|revocation_announce",
  "from": "did:key:z6MksA...",
  "to": "did:key:z6MkqB...",          // omit for revocation_announce
  "ctx": "general|commerce|hiring",
  "epoch": "2025-09",
  "payloadCID": "bafy... (optional)",
  "nonce": "base64-12-bytes",
  "issuedAt": "2025-09-12T19:12:45Z",
  "sig": "ed25519(...)"               // detached over canonical body
}
```

**Canonicalization**

* JSON with sorted keys; strings UTF-8; no insignificant whitespace.
* Hash = SHA-256 over canonical bytes; CID = multihash(sha2-256, digest).

### 4.2 Score Record

```json
{
  "did": "did:key:z6MkqB...",
  "ctx": "commerce",
  "ruleset": {"id":"v1.3","hash":"sha256:...","validFrom":"2025-09-01"},
  "checkpoint": {"root":"...","epoch":1024,"sig":"bls...","signers":["..."]},
  "score": 73.40,
  "factors": {
    "K": "comm(PoP,KYC)", "A": "comm(.edu,employer)", "V": "comm(web_of_trust)",
    "R": "comm(adjudications)", "T": "comm(time)"
  },
  "proofs": {
    "inclusion": [{"cid":"...","path":["..."]}],
    "consistency": [{"old":"...","new":"...","path":["..."]}],
    "statuslists": [{"issuer":"did:web:kyc1","epoch":1024,"bitmapCID":"bafy..."}]
  }
}
```

### 4.3 Ruleset (signed JSON)

```json
{
  "id": "v1.3",
  "weights": {"alpha":0.4,"beta":0.2,"gamma":0.25,"delta":0.1,"tau":0.05},
  "caps": {"K":1.0,"A":0.8,"V":0.9,"R":0.9,"T":0.2},
  "vouch": {
    "budget_base": 2,
    "budget_lambda": 1.2,
    "agg": "sqrt",
    "per_epoch": "monthly",
    "bond": {"type":"reputation","decay_on_abuse":0.05}
  },
  "decay": {"half_life_days":{"V":120,"R":180,"T":90}},
  "diversity": {"community_overlap_penalty":0.15,"min_clusters":3},
  "adjudication": {"pool_size":9,"quorum":6,"appeal_window_days":14},
  "signature": "jws(...)",
  "validFrom": "2025-09-01T00:00:00Z",
  "timeLockDays": 7
}
```

---

## 5) Protocols & Flows

### 5.1 Vouch Flow

```
Wallet        P2P        FullNode         LogNode          Checkpointor      Scorer
  |            |            |                |                   |              |
1 | create evt |            |                |                   |              |
2 | gossip --->|---> relay  |                |                   |              |
3 |            |            | store blob     |                   |              |
4 |            |            | appendRef ---->| leaf+seq          |              |
5 |            |            |<-- receipt ----|                   |              |
6 |            |            |                |--- roots ----->   | aggregate   |
7 |            |            |                |<-- checkpoint --- | broadcast   |
8 |            |            |                |                   |-- recompute |
9 |<---- score+proofs (from Scorer or FullNode on demand) ------ |   publish   |
```

**Notes**

* Wallet enforces **vouch budget** per epoch locally (per ruleset).
* Full nodes serve blob + inclusion proof; log node provides consistency proofs across checkpoints.

### 5.2 Report & Adjudication

```
Wallet -> gossip Report
Adjudicator: sample reviewers (VRF); collect signed votes; tally
Verdict -> log (Append); included in next checkpoint
Scorer applies: increase R(target); if malicious report: decrease reporter (slash bond/rep)
```

### 5.3 “Score ≥ θ?” Verification

1. App challenges wallet with `(ctx, θ, nonce)`
2. Wallet returns:

* **ZK threshold proof** that `S ≥ θ` (+ checkpoint ref), **or**
* Score Record + proofs (for deterministic verification)

3. App verifies ruleset JWS, checkpoint BLS sig, Merkle proofs, ZK proof (if provided)

---

## 6) Scoring Algorithm (Deterministic)

For user *i* in context *c*:

$$
S_i^c = \alpha K_i^c + \beta A_i^c + \gamma \cdot \sqrt{\sum_{j \in \mathcal{V}_i^c} \min(S_j^c, cap) \cdot q_{ij}} - \delta R_i^c + \tau T_i^c
$$

**Terms**

* `K_i^c` (0..1): PoP/KYC credentials; weighted by **issuer reputation** `w(issuer)`
* `A_i^c` (0..1): other VCs (employment, .edu); clipped by caps and revocation
* `V_i^c`: incoming vouches from set `V_i`, aggregated **concavely** (sqrt); `q_ij` adds:

  * **Diversity weight**: down-weight vouches from highly overlapping communities (Jaccard/overlap on neighbor sets)
  * **Recency**: exponential decay with half-life H\_V
* `R_i^c`: adjudicated reports (severity-weighted), decayed with half-life H\_R
* `T_i^c`: time component: bounded growth up to `caps.T`, decays on inactivity

**Vouch budget (per epoch)**

* `b_i = b0 + λ * log(1 + S_i^c)` (non-transferable; non-rollover)
* Each vouch posts an **implicit reputational bond**: if the vouchee is found abusive in window W, a fraction ε of the voucher’s `V` contribution decays.

**Issuer reputation**

* `w(issuer) ∈ [0.2, 1.0]`, initialized conservatively
* Updated by historical accuracy: false-pos/neg penalties

**Parameterization**

* All constants in signed ruleset; changes time-locked; clients pin versions.

---

## 7) P2P & Networking

### 7.1 libp2p Settings

* Transports: TCP + QUIC
* DHT: Kademlia (provider records for `blob/<CID>`)
* PubSub: gossipsub v1.1 with:

  * `mesh_n=8`, `mesh_n_low=5`, `mesh_n_high=12`
  * Per-peer rate limits, topic quotas
  * Flood-publish disabled except for committee checkpoints
* Topics:

  * `events/vouch`, `events/report`, `events/appeal`
  * `revocations/*`
  * `rules/active`
  * `checkpoints/epoch`

### 7.2 Message Size & Batching

* Event payloads ≤ 8–16 KB
* Large evidence stored as blobs (CID), fetched over P2P/HTTP
* Log appends in micro-batches (e.g., 100–1,000 events/append txn)

### 7.3 Peer Reputation & Anti-abuse

* Greylist bursty peers at topic level
* Simple PoW (hashcash) on **reports** to deter spam (configurable difficulty)
* Optional IP-diversity gating on reviewer sampling

---

## 8) Storage, Indexing & Proof Serving

### 8.1 Full Node Store

* **RocksDB**:

  * `events:{cid} -> bytes`
  * `idx:by_to:{did}:{epoch} -> cid[]`
  * `idx:by_from:{did}:{epoch} -> cid[]`
  * `statuslist:{issuer}:{epoch} -> bitmapCID`
  * `checkpoint:{epoch} -> json`
* **File store**: large evidence blobs, mapped by CID

### 8.2 Log Node (Trillian)

* MySQL/Postgres backend for nodes & leaves
* APIs:

  * `QueueLeaves(Leaf)` → seqNo
  * `GetInclusionProofByHash(hash, treeSize)` → audit path
  * `GetConsistencyProof(fromSize, toSize)` → proof path

### 8.3 Proof Bundling

* **Inclusion proofs** for all new events since previous checkpoint
* **Consistency proof**: last known → current checkpoint
* **StatusList proofs**: timestamped mirror receipts (sig of mirror + bitmap CID)

---

## 9) Committee & Adjudication

### 9.1 Checkpoint Committee

* Membership: rotating set sampled monthly; operational window: 1–4 weeks
* **Selection**: VRF seeded by drand; eligibility threshold on `S ≥ θ_op`
* **Protocol**:

  1. Each member fetches latest Trillian root
  2. Members sign `(tree_id, root, epoch)` with BLS
  3. Aggregator collects ≥ t partials → threshold signature
  4. Checkpoint gossiped + mirrored (HTTP)
* **Slashing**: publish non-matching signatures → member reputation decay

### 9.2 Adjudication Pools

* **Pool size** per case: 9 (quorum 6); larger for severe reports
* **Sampling**: VRF lottery among reviewers with `S ≥ θ_rev` and diversity constraints (ASN/IP/country/graph partition balance)
* **Incentives**:

  * Agreement with final outcome ⇒ reviewer reputation ↑
  * Outlier behavior repeatedly ⇒ reputation ↓
* **Outputs**:

  * `AdjudicationResult{case_id, target, reporter, outcome, severity, evidenceRefs[]}` → appended to log

---

## 10) Privacy Model

* **Selective disclosure** via SD-JWT/BBS+: apps learn “KYC-passed” (claim) without SSN or issuer details unless consented
* **Threshold proofs**: ZK proving `S ≥ θ` without revealing factors
* **No PII on p2p/log**: only commitments and issuer/status references
* **Per-app consent**: wallet policy to disclose minimal claims

---

## 11) Security & Threat Model

**Threats & Mitigations**

* **Sybil farms**: require PoP/KYC before outbound vouch; vouch budgets; concave aggregation; diversity weighting
* **Collusion rings**: graph anomaly detection (sudden dense subgraphs), velocity caps; down-weight not block
* **Spam reports**: PoW + rate limits + adjudication slashing
* **Key compromise**: wallet social recovery; freeze & migrate DIDs with proof chain
* **Log equivocation**: Trillian consistency proofs + widely mirrored checkpoints
* **Committee capture**: rotation + VRF + diversity + time-locked rulesets

---

## 12) APIs (HTTP+JSON)

### 12.1 Wallet

* `POST /v1/events` → `{cid, receipt}`
  Body: Event (signed)
* `GET /v1/scores?did=&ctx=` → Score Record + proofs
* `POST /v1/threshold-proof` → `{proof, checkpoint}`
  Body: `{ctx, threshold, nonce}`

### 12.2 Full Node

* `GET /v1/blobs/{cid}` → raw bytes
* `GET /v1/proofs/inclusion?cid=&epoch=` → inclusion proof
* `GET /v1/proofs/consistency?from=&to=` → consistency proof
* `GET /v1/status/{issuer}/{epoch}` → bitmap & mirror receipt

### 12.3 Rules Registry

* `GET /v1/rules/active` → signed ruleset JSON
* `GET /v1/rules/{id}`

### 12.4 Log Node (edge proxy to Trillian)

* `POST /v1/log/append` → `{seqNo, leafHash}`
* `GET /v1/log/inclusion?hash=&size=`
* `GET /v1/log/consistency?from=&to=`

**Auth**: requests are either public (proofs) or signed (event submission). Rate limits by peer ID + DID.

---

## 13) Rate Limits & Quotas

* **Events**:

  * Vouch: ≤ `b_i` per epoch enforced by wallet; network also limits ≤ 5/day if abuse detected
  * Report: ≤ 3/day, PoW required (e.g., 2^20 target)
* **P2P**: topic‐level quotas; backpressure and greylisting
* **Proof serving**: per IP/DID burst limits; cache common proofs/CDNs

---

## 14) Observability & SLOs

**SLOs**

* 99% `AppendRef` latency < 800 ms (regional)
* 99% `InclusionProof` < 300 ms from nearest full node
* Checkpoint issuance every 10 min ± 2 min
* Score recompute lag: P95 < 5 min post-checkpoint

**Telemetry**

* Prometheus metrics per service (latency, error rates, topic throughput)
* Structured logs (Zap) with correlation IDs (CID/epoch)
* Tracing (OpenTelemetry) across wallet→full→log→scorer

---

## 15) Capacity Planning (v1)

Assumptions:

* 1M identities; 100k MAU
* 200k vouches/day; 10k reports/day
* Avg event 1 KB; inclusion proof \~ 600–800 B

**Throughput**

* Log appends: \~2.5 events/sec avg; bursts 50–100 eps → micro-batching OK
* Storage: 210 GB/year raw events + indexes; 3x replication → \~630 GB/year
* Full nodes: 200–500 global nodes suffice; each serving \~1–5 rps proofs peak

---

## 16) Rollout Plan

**Phase 0 (Local Net)**

* docker-compose: 1 log cluster (Trillian+MySQL), 2 full nodes, 1 scorer, 1 checkpointor
* Demo web app with “score ≥ 50?” gate

**Phase 1 (Pilot)**

* 3 issuer partners (KYC, .edu, employer)
* 5 apps (forum, marketplace, collaboration tool, gaming guild, job board)
* 50 community full nodes; committee N=25

**Phase 2 (Open Beta)**

* Open light clients; committee N=75; checkpoints every 10 min
* Add adjudication pools; enable appeals

**Phase 3 (Prod)**

* ZK threshold proofs
* Governance for rulesets (time-locked proposals; require supermajority committee cosign)

---

## 17) Failure Modes & Recovery

* **Wallet lost keys**: social recovery guardians; migration VC links old→new DID; freeze old
* **Log node outage**: rely on replicas; consistency proofs resume after gap
* **Checkpoint delays**: scorers continue with last checkpoint; mark freshness bit; apps can accept `S ≥ θ` with max-staleness policy
* **Issuer StatusList missing**: treat as last known; after TTL expire, down-weight affected credentials

---

## 18) Graph Analytics (anti-collusion)

Periodic scorer job:

* **Features**: degree distribution, clustering coeff., edge velocity, community overlap
* **Detections**:

  * Dense bipartite cores appearing quickly
  * Reciprocal vouch cycles
  * High vouch out-degree with low in-degree over short window
* **Actions**: reduce `q_ij` weights, increase proof-of-work for implicated cluster’s reports, alert maintainers (no hard bans)

---

## 19) Developer-Facing Repo Layout

```
/cmd
  /walletd          // CLI + gRPC/HTTP
  /p2p-gateway      // libp2p host
  /fullnode         // blob + proofs
  /lognode          // trillian personality
  /checkpointor     // bls aggregation
  /scorer           // score calculator
  /adjudicator      // reviewer pools
  /rules-registry   // signed rulesets
/internal
  /didvc            // vc-go adapters, sd-jwt
  /events           // schemas, canonicalization
  /p2p              // pubsub, dht, topics
  /log              // trillian client
  /score            // algorithms, decay, diversity
  /proof            // bundling; (zk circuits later)
  /store            // rocksdb, blob fs
  /crypto           // ed25519, bls, vrf
/api
  /http             // OpenAPI
  /proto            // gRPC defs
/deploy
  /compose          // local stack
  /helm             // k8s charts
/tests
  /e2e              // simulated networks
```

---

## 20) Test Strategy

* **Unit**: canonicalization, CID, signature verification, decay math, diversity weights
* **Property tests**: idempotent scoring under event reordering; consistency proofs
* **Adversarial sims**: sybil clusters, spam waves, collusion rings, malicious committee member
* **Soak**: 7-day run with synthetic traffic; checkpoint churn and log node failures
* **Back-compat**: ruleset vN→vN+1 migrations; time-locks honored

---

## 21) Example Constants (v1.3 draft)

* Weights: `α=0.40, β=0.20, γ=0.25, δ=0.10, τ=0.05`
* Caps: `K≤1.0, A≤0.8, V≤0.9, R≤0.9, T≤0.2`
* Half-lives: `V=120d, R=180d, T=90d`
* Diversity penalty: 0.15 if >60% overlap in 2-hop neighbors
* Vouch budget: `b0=2/epoch, λ=1.2`; max vouch impact per sender per recipient per epoch: 0.05
* Reviewer thresholds: `θ_rev=65`, pool 9, quorum 6, appeal window 14d
* Checkpoint every 10 min; committee N=75; threshold t=50

---

## 22) What to Build First (MVP Cut)

1. **Event schema + canonicalization + CID**
2. **libp2p** pubsub + simple peer limits
3. **Full node** blob store + inclusion proof proxy to Trillian
4. **Trillian** single-region cluster + append API
5. **Scorer** with `K/A/V/R/T` (no diversity yet), simple decay
6. **Demo wallet** (CLI) to create vouch/report, query score+proof
7. **Demo app**: “score ≥ 50?” gate with deterministic verification

---
