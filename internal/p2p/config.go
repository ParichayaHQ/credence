package p2p

import (
	"time"

	"github.com/multiformats/go-multiaddr"
)

// Config represents P2P node configuration
type Config struct {
	// Host configuration
	ListenAddrs   []multiaddr.Multiaddr `json:"listen_addrs"`
	BootstrapPeers []multiaddr.Multiaddr `json:"bootstrap_peers"`
	
	// Gossipsub parameters (v1.1)
	GossipsubConfig GossipsubConfig `json:"gossipsub"`
	
	// DHT configuration
	DHTConfig DHTConfig `json:"dht"`
	
	// Rate limiting
	RateLimit RateLimitConfig `json:"rate_limit"`
	
	// Cache configuration
	CacheConfig CacheConfig `json:"cache"`
	
	// Anti-abuse settings
	AntiAbuse AntiAbuseConfig `json:"anti_abuse"`
}

// GossipsubConfig contains gossipsub-specific settings
type GossipsubConfig struct {
	// Mesh parameters
	MeshN     int `json:"mesh_n"`      // Target mesh size (default: 8)
	MeshNLow  int `json:"mesh_n_low"`  // Low watermark (default: 5)
	MeshNHigh int `json:"mesh_n_high"` // High watermark (default: 12)
	
	// Timing parameters
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // default: 1s
	
	// Score parameters
	EnableScoring bool `json:"enable_scoring"` // default: true
}

// DHTConfig contains DHT-specific settings
type DHTConfig struct {
	BootstrapTimeout time.Duration `json:"bootstrap_timeout"` // default: 30s
	Mode            string        `json:"mode"`              // "client", "server", "auto"
	ProtocolPrefix  string        `json:"protocol_prefix"`   // default: "/credence"
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	// Per-peer limits
	PeerMsgPerMin   int           `json:"peer_msg_per_min"`   // default: 60
	PeerBytesPerSec int           `json:"peer_bytes_per_sec"` // default: 1024
	
	// Global limits
	GlobalMsgPerSec int `json:"global_msg_per_sec"` // default: 1000
	
	// Burst allowance
	BurstMultiplier float64 `json:"burst_multiplier"` // default: 2.0
	
	// Cleanup interval
	CleanupInterval time.Duration `json:"cleanup_interval"` // default: 1m
}

// CacheConfig contains caching settings
type CacheConfig struct {
	// Blob cache
	BlobCacheSize    int           `json:"blob_cache_size"`     // default: 1000
	BlobCacheTTL     time.Duration `json:"blob_cache_ttl"`      // default: 10m
	
	// Checkpoint cache
	CheckpointCacheSize int           `json:"checkpoint_cache_size"` // default: 100
	CheckpointCacheTTL  time.Duration `json:"checkpoint_cache_ttl"`  // default: 1h
	
	// Peer info cache
	PeerCacheSize int           `json:"peer_cache_size"` // default: 10000
	PeerCacheTTL  time.Duration `json:"peer_cache_ttl"`  // default: 30m
}

// AntiAbuseConfig contains anti-abuse settings
type AntiAbuseConfig struct {
	// Greylist settings
	GreylistDuration time.Duration `json:"greylist_duration"` // default: 10m
	GreylistThreshold int          `json:"greylist_threshold"` // default: 10 violations
	
	// Message validation
	MaxMessageSize int `json:"max_message_size"` // default: 16KB
	
	// Proof-of-work for reports (optional)
	EnablePoW    bool `json:"enable_pow"`     // default: false
	PoWDifficulty int  `json:"pow_difficulty"` // default: 20 (bits)
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddrs: []multiaddr.Multiaddr{},
		BootstrapPeers: []multiaddr.Multiaddr{},
		
		GossipsubConfig: GossipsubConfig{
			MeshN:             8,
			MeshNLow:          5,
			MeshNHigh:         12,
			HeartbeatInterval: time.Second,
			EnableScoring:     true,
		},
		
		DHTConfig: DHTConfig{
			BootstrapTimeout: 30 * time.Second,
			Mode:            "auto",
			ProtocolPrefix:  "/credence",
		},
		
		RateLimit: RateLimitConfig{
			PeerMsgPerMin:   60,
			PeerBytesPerSec: 1024,
			GlobalMsgPerSec: 1000,
			BurstMultiplier: 2.0,
			CleanupInterval: time.Minute,
		},
		
		CacheConfig: CacheConfig{
			BlobCacheSize:       1000,
			BlobCacheTTL:        10 * time.Minute,
			CheckpointCacheSize: 100,
			CheckpointCacheTTL:  time.Hour,
			PeerCacheSize:       10000,
			PeerCacheTTL:        30 * time.Minute,
		},
		
		AntiAbuse: AntiAbuseConfig{
			GreylistDuration:  10 * time.Minute,
			GreylistThreshold: 10,
			MaxMessageSize:    16 * 1024, // 16KB
			EnablePoW:         false,
			PoWDifficulty:     20,
		},
	}
}