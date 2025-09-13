package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
)

// BlobManager handles blob storage, retrieval, and caching via DHT
type BlobManager struct {
	p2pHost *P2PHost
	cache   *LRUCache
	logger  *Logger
}

// NewBlobManager creates a new blob manager
func NewBlobManager(p2pHost *P2PHost, cacheSize int, cacheTTL time.Duration) *BlobManager {
	return &BlobManager{
		p2pHost: p2pHost,
		cache:   NewLRUCache(cacheSize, cacheTTL),
		logger:  NewLogger("BlobManager", LogLevelInfo),
	}
}

// Close closes the blob manager
func (bm *BlobManager) Close() {
	if bm.cache != nil {
		bm.cache.Close()
	}
}

// StoreBlob stores a blob locally and announces it to the DHT
func (bm *BlobManager) StoreBlob(ctx context.Context, data []byte) (cid.Cid, error) {
	if len(data) == 0 {
		bm.logger.Warn("Attempted to store empty blob")
		return cid.Undef, NewP2PError("store_blob", fmt.Errorf("empty blob data"))
	}

	bm.logger.Debug("Storing blob", map[string]interface{}{"size": len(data)})

	// Generate CID for the blob
	cidGen := &CIDGenerator{} // Would use the CID generator from internal/cid
	c, err := cidGen.GenerateFromBytes(data)
	if err != nil {
		bm.logger.Error("Failed to generate CID for blob", map[string]interface{}{
			"error": err,
			"size": len(data),
		})
		return cid.Undef, NewP2PError("store_blob", err).WithContext("size", len(data))
	}

	cidStr := c.String()
	bm.logger.Info("Generated CID for blob", map[string]interface{}{
		"cid": cidStr,
		"size": len(data),
	})

	// Store in cache
	bm.cache.Set(cidStr, data)

	// Announce to DHT that we can provide this content
	if err := bm.p2pHost.Provide(ctx, c); err != nil {
		// Log error but don't fail - we still have the content locally
		bm.logger.Warn("Failed to announce blob to DHT", map[string]interface{}{
			"cid": cidStr,
			"error": err,
		})
	} else {
		bm.logger.Debug("Successfully announced blob to DHT", map[string]interface{}{"cid": cidStr})
	}

	return c, nil
}

// GetBlob retrieves a blob by CID, first from cache, then from DHT
func (bm *BlobManager) GetBlob(ctx context.Context, c cid.Cid) ([]byte, error) {
	cidStr := c.String()
	bm.logger.Debug("Retrieving blob", map[string]interface{}{"cid": cidStr})

	// Check local cache first
	if data, found := bm.cache.Get(cidStr); found {
		bm.logger.Debug("Blob found in cache", map[string]interface{}{
			"cid": cidStr,
			"size": len(data),
		})
		return data, nil
	}

	bm.logger.Debug("Blob not in cache, searching DHT", map[string]interface{}{"cid": cidStr})

	// Find providers via DHT
	providers, err := bm.p2pHost.FindProviders(ctx, c)
	if err != nil {
		bm.logger.Error("Failed to find providers for blob", map[string]interface{}{
			"cid": cidStr,
			"error": err,
		})
		return nil, NewP2PError("get_blob", err).WithContext("cid", cidStr)
	}

	if len(providers) == 0 {
		bm.logger.Info("No providers found for blob", map[string]interface{}{"cid": cidStr})
		return nil, NewP2PError("get_blob", ErrProviderNotFound).WithContext("cid", cidStr)
	}

	bm.logger.Info("Found providers for blob", map[string]interface{}{
		"cid": cidStr,
		"provider_count": len(providers),
	})

	// Try to fetch from providers
	data, err := bm.fetchFromProviders(ctx, c, providers)
	if err != nil {
		bm.logger.Error("Failed to fetch blob from providers", map[string]interface{}{
			"cid": cidStr,
			"provider_count": len(providers),
			"error": err,
		})
		return nil, NewP2PError("get_blob", err).WithContext("cid", cidStr)
	}

	bm.logger.Info("Successfully fetched blob", map[string]interface{}{
		"cid": cidStr,
		"size": len(data),
	})

	// Cache the retrieved data
	bm.cache.Set(cidStr, data)

	return data, nil
}

// HasBlob checks if a blob is available locally
func (bm *BlobManager) HasBlob(c cid.Cid) bool {
	return bm.cache.Has(c.String())
}

// DeleteBlob removes a blob from local cache
func (bm *BlobManager) DeleteBlob(c cid.Cid) {
	bm.cache.Delete(c.String())
}

// GetBlobStats returns blob cache statistics
func (bm *BlobManager) GetBlobStats() map[string]interface{} {
	stats := bm.cache.Stats()
	stats["type"] = "blob_cache"
	return stats
}

// fetchFromProviders attempts to fetch content from provider peers
func (bm *BlobManager) fetchFromProviders(ctx context.Context, c cid.Cid, providers []peer.AddrInfo) ([]byte, error) {
	// In a full implementation, this would:
	// 1. Connect to provider peers
	// 2. Request the content using a custom protocol
	// 3. Verify the content matches the CID
	// 4. Return the content

	// For this core infrastructure implementation, we'll simulate the process
	if len(providers) == 0 {
		return nil, ErrProviderNotFound
	}

	// Create timeout context for fetching
	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Simulate content fetching delay
	select {
	case <-fetchCtx.Done():
		return nil, fmt.Errorf("fetch timeout")
	case <-time.After(100 * time.Millisecond):
		// Simulate successful fetch
	}

	// For now, return a placeholder indicating successful provider discovery
	// In reality, this would return the actual blob content
	return []byte(fmt.Sprintf("BLOB_CONTENT_FOR_%s_FROM_%d_PROVIDERS", c.String(), len(providers))), nil
}

// CIDGenerator is a placeholder for the CID generator from internal/cid
type CIDGenerator struct{}

// GenerateFromBytes generates a CID from bytes (placeholder implementation)
func (g *CIDGenerator) GenerateFromBytes(data []byte) (cid.Cid, error) {
	// This would use the actual CID generation from internal/cid
	// For now, create a simple CID
	return cid.Parse("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N") // Placeholder
}

// BlobRequest represents a request for blob content
type BlobRequest struct {
	CID       cid.Cid `json:"cid"`
	Requester peer.ID `json:"requester"`
	Priority  int     `json:"priority"`
	Timestamp time.Time `json:"timestamp"`
}

// BlobResponse represents a response with blob content
type BlobResponse struct {
	CID     cid.Cid `json:"cid"`
	Data    []byte  `json:"data"`
	Success bool    `json:"success"`
	Error   string  `json:"error,omitempty"`
}

// RequestQueue manages queued blob requests
type RequestQueue struct {
	requests []BlobRequest
	maxSize  int
}

// NewRequestQueue creates a new request queue
func NewRequestQueue(maxSize int) *RequestQueue {
	return &RequestQueue{
		requests: make([]BlobRequest, 0),
		maxSize:  maxSize,
	}
}

// Add adds a request to the queue
func (rq *RequestQueue) Add(req BlobRequest) {
	if len(rq.requests) >= rq.maxSize {
		// Remove oldest request
		rq.requests = rq.requests[1:]
	}
	rq.requests = append(rq.requests, req)
}

// GetNext returns the next request to process
func (rq *RequestQueue) GetNext() *BlobRequest {
	if len(rq.requests) == 0 {
		return nil
	}
	
	// Return highest priority, oldest first
	bestIdx := 0
	best := &rq.requests[0]
	
	for i := 1; i < len(rq.requests); i++ {
		req := &rq.requests[i]
		if req.Priority > best.Priority || 
		   (req.Priority == best.Priority && req.Timestamp.Before(best.Timestamp)) {
			bestIdx = i
			best = req
		}
	}
	
	// Remove from queue
	copy(rq.requests[bestIdx:], rq.requests[bestIdx+1:])
	rq.requests = rq.requests[:len(rq.requests)-1]
	
	return best
}

// Size returns the queue size
func (rq *RequestQueue) Size() int {
	return len(rq.requests)
}