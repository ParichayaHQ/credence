---
layout: default
title: "Component Design"
description: "Detailed component specifications and interfaces"
collection: designs
---

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
+----+----+     +-----+-----+       +-----+-----+       +----+-------+
| Wallet  |     | Full Node |       | Log Node  |       | Scorer     |
| Light   |     |  Storage  |       | (Trillian)|       |  Node      |
| Node    |     |  + Proofs |       |  + Proofs |       | (Determin.)|
+----+----+     +-----+-----+       +-----+-----+       +----+-------+
     |                |                   |                  |
     +----------------+---------+---------+------------------+
                               |
                               v
                      +--------+--------+
                      | Checkpoint Comm |
                      |  (BLS threshold)|
                      +-----------------+
```

- [wallet (light node)](#wallet-light-node)
- [p2p-gateway](#p2p-gateway)
- [fullnode](#fullnode)
- [lognode (transparency log)](#lognode-transparency-log)
- [checkpointor](#checkpointor)
- [scorer](#scorer)

---

# wallet (light node)

## purpose

End-user client that holds keys & VCs, creates signed events (vouch/report), requests/assembles proofs, and returns threshold proofs or score bundles to apps.

## responsibilities

* DID key mgmt (`did:key` v1; later `did:web`/ION)
* Store/present VCs (VC-JWT + SD-JWT)
* Create/sign events; enforce **vouch budgets**
* Fetch inclusion/consistency proofs & checkpoints
* Produce **threshold proofs** (optional, later)

## external API (HTTP/CLI)

* `POST /v1/events` → submit signed `Event` to p2p gateway
  body: Event (detached JWS or inline `sig`)
* `GET /v1/scores?did=&ctx=` → return cached ScoreRecord+proofs (via scorer/fullnode)
* `POST /v1/threshold-proof` → `{proof, checkpoint}`
  body: `{ctx, threshold, nonce}`

## schemas (canonical JSON)

**Event**

```json
{
  "type":"vouch|report",
  "from":"did:key:z6Mk...",
  "to":"did:key:z6Mk...", 
  "ctx":"general|commerce|hiring",
  "epoch":"2025-09",
  "payloadCID":"bafy... (optional)",
  "nonce":"base64-12",
  "issuedAt":"2025-09-12T19:12:45Z",
  "sig":"ed25519..."
}
```

## state & storage

* Keystore (Ed25519) + mnemonic/guardians (social recovery later)
* VC store (encrypted)
* Local counters: `vouch_budget[ctx][epoch]`
* Cache: recent checkpoints, proofs, score bundles

## dependencies

* p2p-gateway (publish/fetch)
* fullnode/lognode (proofs via gateway)
* scorer (for score bundles)
* rules-registry (active ruleset)

## flows (happy paths)

1. **vouch**: build event → sign → `POST /events` → receive CID receipt
2. **prove score**: request `GET /scores` (or build zk) → return to app
3. **fetch proofs**: on demand, call gateway for inclusion/consistency proofs

## SLOs & limits

* sign+publish event: P95 < 200 ms (local)
* cached score fetch: P95 < 100 ms
* vouch budget: enforced client-side; non-rollover per epoch

## security

* Keys in OS secure enclave if available
* Outbound vouch locked until PoP/KYC VC present
* Opt-in telemetry only; no PII in network messages

## tests

* canonicalization determinism, CID stability
* signature round-trip, key rotation
* budget enforcement; replay protection (nonce)

---

# p2p-gateway

## purpose

Single entry for libp2p: gossip publish/subscribe, CID fetch, rate limiting, HTTP bridge for other services.

## responsibilities

* libp2p host (gossipsub + DHT)
* Topics: `events/*`, `revocations/*`, `rules/*`, `checkpoints/*`, `blobs/*`
* HTTP façade for internal services (wallet/fullnode/lognode/scorer)

## external API (HTTP)

* `POST /v1/publish?topic=` → publish signed message
* `GET /v1/blobs/{cid}` → fetch blob (with local cache)
* `GET /v1/subscribe?topic=` (SSE/websocket for local services)
* `GET /v1/checkpoints/latest` → last checkpoint JSON

## state & storage

* Peer book, topic subscriptions
* LRU caches for blobs & checkpoints
* Greylist/ratelimit tables

## dependencies

* libp2p
* local fullnode for blob persistence

## flows

* receive event from wallet → validate size/sig presence → gossip → (optional) forward to fullnode/lognode append queues
* blob fetch via DHT → persist to fullnode → serve via HTTP

## SLOs & limits

* publish to mesh: P95 < 50 ms
* blob fetch (warm): P95 < 150 ms, (cold): < 2 s
* per-peer quotas: msgs/min per topic; total bytes/sec

## security

* Drop oversize, unsigned, malformed messages
* Backoff/greylist bursty peers; per-topic quotas
* TLS for HTTP, noise for libp2p

## tests

* flood/abuse sims
* partition/reconnect
* DHT provider lookup correctness

---

# fullnode

## purpose

Durable blob store & proof server. Mirrors rules, checkpoints, and StatusLists. Serves inclusion/consistency proofs (proxy to lognode).

## responsibilities

* Persist event blobs by CID (RocksDB/FS)
* Index by `(type, from, to, epoch)`
* Mirror: `status/{issuer}/{epoch}`, `rules/active`, checkpoints
* HTTP: blobs, proofs, mirrors

## external API

* `POST /v1/events` → validate basic schema, persist blob, enqueue log append
  resp: `{cid, queued:true}`
* `GET /v1/blobs/{cid}` → raw bytes
* `GET /v1/proofs/inclusion?cid=&epoch=`
* `GET /v1/proofs/consistency?from=&to=`
* `GET /v1/status/{issuer}/{epoch}`
* `GET /v1/checkpoints/{epoch}`

## storage

* RocksDB:

  * `events:{cid} -> bytes`
  * `idx:to:{did}:{epoch} -> cid[]`
  * `idx:from:{did}:{epoch} -> cid[]`
  * `status:{issuer}:{epoch} -> bitmapCID`
  * `checkpoint:{epoch} -> json`
* FS for large evidence blobs

## dependencies

* p2p-gateway
* lognode (append & proofs)
* rules-registry & issuers (HTTP mirror)

## flows

1. on `POST /events`: persist → forward reference to lognode
2. on blob miss: ask gateway (DHT), then persist
3. on proof request: call lognode, cache result, return

## SLOs

* write+queue event: P95 < 150 ms
* blob read warm: P95 < 40 ms
* inclusion proof: P95 < 250 ms (cached)

## security

* Validate JSON shape & size before persist
* Quotas per origin; signature presence check (deep verify optional)
* DoS guards on proof endpoints

## tests

* crash/restart durability
* heavy read/write mixed load
* index correctness, pagination

---

# lognode (transparency log)

## purpose

Append-only Merkle log of event references (CIDs + headers) with inclusion/consistency proofs (CT-style).

## responsibilities

* Trillian personality: QueueLeaves, SignedTreeHead, proofs
* Append micro-batches (100–1000 events)
* Expose tree size, roots; retain history

## external API

* `POST /v1/log/append` → `[{cid, header_hash}]` → `{leafHashes, treeSize, sth}`
* `GET /v1/log/inclusion?cid=&size=` → audit path
* `GET /v1/log/consistency?from=&to=` → proof path
* `GET /v1/log/sth` → latest SignedTreeHead

## storage

* Trillian DB (MySQL/Postgres)
* Audit logs for append calls

## dependencies

* fullnode (producer)
* checkpointor (consumes STHs)
* p2p-gateway (announce root updates optional)

## flows

* batch append from fullnode queue → return receipts
* serve proofs to fullnodes/scorers/verifiers

## SLOs

* batch append: P95 < 300 ms
* inclusion proof: P95 < 150 ms
* consistency proof: P95 < 150 ms

## security

* Accept only well-formed event refs (hash+minimal header); no PII
* Append auth (mTLS from registered fullnodes)
* Rate limits; integrity audits (random rehash)

## tests

* proof correctness (known vectors)
* consistency across sizes
* replay/idempotence on duplicate cids

---

# checkpointor

## purpose

Periodically aggregates log roots into **checkpoints** and co-signs with a rotating committee (BLS threshold). Publishes to gossip & mirrors.

## responsibilities

* Fetch latest STH/root from lognode(s)
* Collect partial BLS signatures from committee
* Produce `Checkpoint{root, epoch, signers, sig}`
* Gossip + HTTP publish; time-lock windows

## external API

* `GET /v1/checkpoints/latest`
* `GET /v1/checkpoints/{epoch}`
* Committee internal:

  * `POST /v1/partials` (members submit partial sigs)
  * `GET /v1/tasks/current` (root, epoch to sign)

## storage

* Checkpoint registry (RocksDB/FS)
* Committee membership (current/next), VRF snapshots

## dependencies

* lognode(s)
* p2p-gateway
* rules-registry (for time-lock & version pin)

## flows

1. every Δt (e.g., 10 min), read STH → create signing task
2. collect ≥t partials → aggregate → publish
3. pin checkpoint to mirrors; gossip on `checkpoints/*`

## SLOs

* issuance jitter ±2 min from schedule
* availability P99.9 for last 24h of checkpoints

## security

* Rotate committee monthly via VRF/drand seed
* Verify all partials vs same root; slash/downgrade equivocators (reputation event)
* mTLS for committee RPC

## tests

* wrong-root partials (byzantine) rejection
* missed epoch recovery
* duplicate epoch prevention

---

# scorer

## purpose

Deterministically compute trust scores from events, VCs, ruleset, and revocations; bundle verifiable proofs for relying apps.

## responsibilities

* Ingest new events & checkpoints
* Maintain per-DID state (K/A/V/R/T terms) with decay
* Compute `S_i^c = αK + βA + γ*sqrt(Σ min(Sj,cap)*qj) − δR + τT`
* Bundle **inclusion**/**consistency** proofs + ruleset commit
* Optional zk threshold proof generation (later)

## external API

* `GET /v1/scores?did=&ctx=` → ScoreRecord + proofs
* `POST /v1/recompute?did=` (admin/dev only)
* `GET /v1/factors?did=&ctx=` (debug; commitments by default)

## storage

* RocksDB:

  * `state:{did}:{ctx}` → `{K,A,V,R,T,S,updatedAt}`
  * `graph:{did}:{ctx}:in` → `[{from, weight, ts}]`
  * `issuer_rep:{issuer}` → weight
  * `ruleset_current` → signed JSON
  * `checkpoint_current` → JSON

## dependencies

* fullnode (events, proofs), lognode (proofs)
* rules-registry, checkpointor
* optionally issuers (StatusList bitmaps via fullnode mirrors)

## flows

* On new checkpoint: pull appended CIDs since last size → update affected DIDs → recompute → cache ScoreRecords
* On `GET /scores`: read cached; if stale beyond policy, lazily refresh

## SLOs

* recompute latency after checkpoint: P95 < 5 min for impacted DIDs
* score fetch (cached): P95 < 60 ms

## security

* Deterministic only from inputs + pinned ruleset
* No raw PII; factors exposed as commitments unless debug-mode
* Guard against graph spam with per-sender caps + decay

## tests

* fixture vectors for S under controlled graphs
* monotonicity & clipping properties
* replay determinism across reorders/checkpoints
* decay half-life correctness

---

## cross-cutting: rules-registry (brief)

* signed JSON: weights, caps, budgets, decay, adjudication params, `validFrom`, `timeLockDays`
* endpoints:

  * `GET /v1/rules/active`
  * `GET /v1/rules/{id}`
* clients **pin** version; scorer embeds `ruleset.id` + hash in score outputs.

---

## suggested repo layout (mono)

```
/cmd
  /walletd
  /p2p-gateway
  /fullnode
  /lognode
  /checkpointor
  /scorer
/internal
  /events        // schemas, canonicalization, signing
  /didvc         // vc-go adapters, sd-jwt helpers
  /p2p           // libp2p host, topics, quotas
  /store         // rocksdb wrappers
  /log           // trillian client wrapper
  /score         // algorithms, decay, diversity
  /proof         // bundling; zk (later)
/api
  /http          // OpenAPI specs, handlers
/deploy
  /compose       // local stack
  /helm          // k8s charts
/tests
  /e2e           // synthetic network sims
```

---

## milestone cutlines (what “done” means)

* **wallet v1**: create/sign vouch; budget enforced; publish via gateway; fetch score bundle
* **p2p-gateway v1**: gossip pub/sub; CID fetch; blob cache; rate limits
* **fullnode v1**: persist events; mirror checkpoints; proxy proofs
* **lognode v1**: append & serve proofs (Trillian)
* **checkpointor v1a**: single-signer schedule; **v1b**: BLS threshold
* **scorer v1**: deterministic compute, decay, diversity (minimal), bundle proofs

---

want me to expand any one of these (e.g., **wallet** or **scorer**) into a full OpenAPI + Go interface skeleton next?
