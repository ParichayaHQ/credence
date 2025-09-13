package p2p

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// RateLimiter manages rate limiting for peers and topics
type RateLimiter struct {
	config    *RateLimitConfig
	antiAbuse *AntiAbuseConfig
	
	// Per-peer rate limiting
	peerLimits map[peer.ID]*PeerLimit
	peerMutex  sync.RWMutex
	
	// Global rate limiting
	globalCount int64
	globalReset time.Time
	globalMutex sync.Mutex
	
	// Cleanup ticker
	cleanup *time.Ticker
	done    chan struct{}
}

// PeerLimit tracks rate limiting data for a peer
type PeerLimit struct {
	// Message count tracking
	MessageCount int
	ByteCount    int64
	ResetTime    time.Time
	
	// Violation tracking
	Violations int
	LastViolation time.Time
	
	// Greylist status
	IsGreylisted bool
	GreylistUntil time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig, antiAbuse *AntiAbuseConfig) *RateLimiter {
	rl := &RateLimiter{
		config:     config,
		antiAbuse:  antiAbuse,
		peerLimits: make(map[peer.ID]*PeerLimit),
		done:       make(chan struct{}),
	}
	
	// Start cleanup routine
	rl.cleanup = time.NewTicker(config.CleanupInterval)
	go rl.cleanupRoutine()
	
	return rl
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	if rl.cleanup != nil {
		rl.cleanup.Stop()
	}
	close(rl.done)
}

// AllowMessage checks if a message from a peer should be allowed
func (rl *RateLimiter) AllowMessage(peerID peer.ID, topic string, size int) bool {
	// Check global rate limit first
	if !rl.checkGlobalLimit() {
		return false
	}
	
	// Check peer-specific limits
	return rl.checkPeerLimit(peerID, topic, size)
}

// checkGlobalLimit checks the global message rate limit
func (rl *RateLimiter) checkGlobalLimit() bool {
	rl.globalMutex.Lock()
	defer rl.globalMutex.Unlock()
	
	now := time.Now()
	
	// Reset counter if window has passed
	if now.After(rl.globalReset) {
		rl.globalCount = 0
		rl.globalReset = now.Add(time.Second)
	}
	
	// Check if we're within limits
	if rl.globalCount >= int64(rl.config.GlobalMsgPerSec) {
		return false
	}
	
	rl.globalCount++
	return true
}

// checkPeerLimit checks peer-specific rate limits
func (rl *RateLimiter) checkPeerLimit(peerID peer.ID, topic string, size int) bool {
	rl.peerMutex.Lock()
	defer rl.peerMutex.Unlock()
	
	limit, exists := rl.peerLimits[peerID]
	if !exists {
		limit = &PeerLimit{
			ResetTime: time.Now().Add(time.Minute),
		}
		rl.peerLimits[peerID] = limit
	}
	
	now := time.Now()
	
	// Check if peer is greylisted
	if limit.IsGreylisted && now.Before(limit.GreylistUntil) {
		return false
	} else if limit.IsGreylisted && now.After(limit.GreylistUntil) {
		// Remove from greylist
		limit.IsGreylisted = false
		limit.Violations = 0
	}
	
	// Reset counters if window has passed
	if now.After(limit.ResetTime) {
		limit.MessageCount = 0
		limit.ByteCount = 0
		limit.ResetTime = now.Add(time.Minute)
	}
	
	// Check message count limit
	if limit.MessageCount >= rl.config.PeerMsgPerMin {
		rl.recordViolation(peerID, limit, "message_count")
		return false
	}
	
	// Check byte count limit
	if limit.ByteCount+int64(size) > int64(rl.config.PeerBytesPerSec*60) { // Convert to per-minute
		rl.recordViolation(peerID, limit, "byte_count")
		return false
	}
	
	// Update counters
	limit.MessageCount++
	limit.ByteCount += int64(size)
	
	return true
}

// recordViolation records a rate limit violation for a peer
func (rl *RateLimiter) recordViolation(peerID peer.ID, limit *PeerLimit, violationType string) {
	limit.Violations++
	limit.LastViolation = time.Now()
	
	// Greylist if threshold exceeded
	if limit.Violations >= rl.antiAbuse.GreylistThreshold {
		limit.IsGreylisted = true
		limit.GreylistUntil = time.Now().Add(rl.antiAbuse.GreylistDuration)
	}
}

// IsGreylisted checks if a peer is greylisted
func (rl *RateLimiter) IsGreylisted(peerID peer.ID) bool {
	rl.peerMutex.RLock()
	defer rl.peerMutex.RUnlock()
	
	limit, exists := rl.peerLimits[peerID]
	if !exists {
		return false
	}
	
	return limit.IsGreylisted && time.Now().Before(limit.GreylistUntil)
}

// GetPeerStats returns rate limiting stats for a peer
func (rl *RateLimiter) GetPeerStats(peerID peer.ID) *PeerLimit {
	rl.peerMutex.RLock()
	defer rl.peerMutex.RUnlock()
	
	limit, exists := rl.peerLimits[peerID]
	if !exists {
		return nil
	}
	
	// Return a copy
	return &PeerLimit{
		MessageCount:  limit.MessageCount,
		ByteCount:     limit.ByteCount,
		ResetTime:     limit.ResetTime,
		Violations:    limit.Violations,
		LastViolation: limit.LastViolation,
		IsGreylisted:  limit.IsGreylisted,
		GreylistUntil: limit.GreylistUntil,
	}
}

// cleanupRoutine periodically cleans up old peer data
func (rl *RateLimiter) cleanupRoutine() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.cleanupOldPeers()
		case <-rl.done:
			return
		}
	}
}

// cleanupOldPeers removes data for peers that haven't been active
func (rl *RateLimiter) cleanupOldPeers() {
	rl.peerMutex.Lock()
	defer rl.peerMutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-time.Hour) // Remove peers inactive for 1 hour
	
	for peerID, limit := range rl.peerLimits {
		// Keep greylisted peers until greylist expires
		if limit.IsGreylisted && now.Before(limit.GreylistUntil) {
			continue
		}
		
		// Remove if reset time is old and no recent violations
		if limit.ResetTime.Before(cutoff) && 
		   (limit.LastViolation.IsZero() || limit.LastViolation.Before(cutoff)) {
			delete(rl.peerLimits, peerID)
		}
	}
}

// GetStats returns overall rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.peerMutex.RLock()
	rl.globalMutex.Lock()
	
	totalPeers := len(rl.peerLimits)
	greylistedPeers := 0
	
	for _, limit := range rl.peerLimits {
		if limit.IsGreylisted && time.Now().Before(limit.GreylistUntil) {
			greylistedPeers++
		}
	}
	
	globalCount := rl.globalCount
	
	rl.globalMutex.Unlock()
	rl.peerMutex.RUnlock()
	
	return map[string]interface{}{
		"total_peers":      totalPeers,
		"greylisted_peers": greylistedPeers,
		"global_msg_count": globalCount,
		"config": map[string]interface{}{
			"peer_msg_per_min":   rl.config.PeerMsgPerMin,
			"peer_bytes_per_sec": rl.config.PeerBytesPerSec,
			"global_msg_per_sec": rl.config.GlobalMsgPerSec,
		},
	}
}