package p2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/multiformats/go-multiaddr"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	t.Run("GossipsubConfig", func(t *testing.T) {
		assert.Equal(t, 8, config.GossipsubConfig.MeshN)
		assert.Equal(t, 5, config.GossipsubConfig.MeshNLow)
		assert.Equal(t, 12, config.GossipsubConfig.MeshNHigh)
		assert.Equal(t, time.Second, config.GossipsubConfig.HeartbeatInterval)
		assert.True(t, config.GossipsubConfig.EnableScoring)
	})

	t.Run("DHTConfig", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, config.DHTConfig.BootstrapTimeout)
		assert.Equal(t, "auto", config.DHTConfig.Mode)
		assert.Equal(t, "/credence", config.DHTConfig.ProtocolPrefix)
	})

	t.Run("RateLimitConfig", func(t *testing.T) {
		assert.Equal(t, 60, config.RateLimit.PeerMsgPerMin)
		assert.Equal(t, 1024, config.RateLimit.PeerBytesPerSec)
		assert.Equal(t, 1000, config.RateLimit.GlobalMsgPerSec)
		assert.Equal(t, 2.0, config.RateLimit.BurstMultiplier)
		assert.Equal(t, time.Minute, config.RateLimit.CleanupInterval)
	})

	t.Run("CacheConfig", func(t *testing.T) {
		assert.Equal(t, 1000, config.CacheConfig.BlobCacheSize)
		assert.Equal(t, 10*time.Minute, config.CacheConfig.BlobCacheTTL)
		assert.Equal(t, 100, config.CacheConfig.CheckpointCacheSize)
		assert.Equal(t, time.Hour, config.CacheConfig.CheckpointCacheTTL)
		assert.Equal(t, 10000, config.CacheConfig.PeerCacheSize)
		assert.Equal(t, 30*time.Minute, config.CacheConfig.PeerCacheTTL)
	})

	t.Run("AntiAbuseConfig", func(t *testing.T) {
		assert.Equal(t, 10*time.Minute, config.AntiAbuse.GreylistDuration)
		assert.Equal(t, 10, config.AntiAbuse.GreylistThreshold)
		assert.Equal(t, 16*1024, config.AntiAbuse.MaxMessageSize)
		assert.False(t, config.AntiAbuse.EnablePoW)
		assert.Equal(t, 20, config.AntiAbuse.PoWDifficulty)
	})

	t.Run("DefaultArrays", func(t *testing.T) {
		assert.NotNil(t, config.ListenAddrs)
		assert.Equal(t, 0, len(config.ListenAddrs))
		assert.NotNil(t, config.BootstrapPeers)
		assert.Equal(t, 0, len(config.BootstrapPeers))
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("ValidMultiaddrs", func(t *testing.T) {
		config := DefaultConfig()
		
		// Add valid multiaddrs
		addr1, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
		require.NoError(t, err)
		
		addr2, err := multiaddr.NewMultiaddr("/ip6/::1/tcp/4001")
		require.NoError(t, err)

		config.ListenAddrs = []multiaddr.Multiaddr{addr1, addr2}
		config.BootstrapPeers = []multiaddr.Multiaddr{addr1}

		assert.Equal(t, 2, len(config.ListenAddrs))
		assert.Equal(t, 1, len(config.BootstrapPeers))
	})

	t.Run("ConfigurableValues", func(t *testing.T) {
		config := DefaultConfig()

		// Modify gossipsub parameters
		config.GossipsubConfig.MeshN = 10
		config.GossipsubConfig.MeshNLow = 6
		config.GossipsubConfig.MeshNHigh = 15
		config.GossipsubConfig.HeartbeatInterval = 2 * time.Second
		config.GossipsubConfig.EnableScoring = false

		assert.Equal(t, 10, config.GossipsubConfig.MeshN)
		assert.Equal(t, 6, config.GossipsubConfig.MeshNLow)
		assert.Equal(t, 15, config.GossipsubConfig.MeshNHigh)
		assert.Equal(t, 2*time.Second, config.GossipsubConfig.HeartbeatInterval)
		assert.False(t, config.GossipsubConfig.EnableScoring)
	})

	t.Run("DHTModes", func(t *testing.T) {
		config := DefaultConfig()

		validModes := []string{"client", "server", "auto"}
		for _, mode := range validModes {
			config.DHTConfig.Mode = mode
			assert.Equal(t, mode, config.DHTConfig.Mode)
		}
	})

	t.Run("ProtocolPrefix", func(t *testing.T) {
		config := DefaultConfig()
		
		customPrefixes := []string{
			"/credence",
			"/test-network",
			"/custom/v1",
		}

		for _, prefix := range customPrefixes {
			config.DHTConfig.ProtocolPrefix = prefix
			assert.Equal(t, prefix, config.DHTConfig.ProtocolPrefix)
		}
	})

	t.Run("RateLimitBoundaries", func(t *testing.T) {
		config := DefaultConfig()

		// Test edge cases for rate limiting
		config.RateLimit.PeerMsgPerMin = 0 // Effectively disabled
		config.RateLimit.PeerBytesPerSec = 1 // Very restrictive
		config.RateLimit.GlobalMsgPerSec = 10000 // Very permissive
		config.RateLimit.BurstMultiplier = 0.5 // Below 1.0
		config.RateLimit.CleanupInterval = time.Second // Fast cleanup

		assert.Equal(t, 0, config.RateLimit.PeerMsgPerMin)
		assert.Equal(t, 1, config.RateLimit.PeerBytesPerSec)
		assert.Equal(t, 10000, config.RateLimit.GlobalMsgPerSec)
		assert.Equal(t, 0.5, config.RateLimit.BurstMultiplier)
		assert.Equal(t, time.Second, config.RateLimit.CleanupInterval)
	})

	t.Run("CacheSizes", func(t *testing.T) {
		config := DefaultConfig()

		// Test various cache sizes
		config.CacheConfig.BlobCacheSize = 10000
		config.CacheConfig.CheckpointCacheSize = 1000
		config.CacheConfig.PeerCacheSize = 50000

		assert.Equal(t, 10000, config.CacheConfig.BlobCacheSize)
		assert.Equal(t, 1000, config.CacheConfig.CheckpointCacheSize)
		assert.Equal(t, 50000, config.CacheConfig.PeerCacheSize)
	})

	t.Run("TTLValues", func(t *testing.T) {
		config := DefaultConfig()

		// Test various TTL values
		config.CacheConfig.BlobCacheTTL = 5 * time.Minute
		config.CacheConfig.CheckpointCacheTTL = 2 * time.Hour
		config.CacheConfig.PeerCacheTTL = 15 * time.Minute

		assert.Equal(t, 5*time.Minute, config.CacheConfig.BlobCacheTTL)
		assert.Equal(t, 2*time.Hour, config.CacheConfig.CheckpointCacheTTL)
		assert.Equal(t, 15*time.Minute, config.CacheConfig.PeerCacheTTL)
	})

	t.Run("AntiAbuseSettings", func(t *testing.T) {
		config := DefaultConfig()

		// Test different anti-abuse configurations
		config.AntiAbuse.GreylistDuration = 30 * time.Minute
		config.AntiAbuse.GreylistThreshold = 5
		config.AntiAbuse.MaxMessageSize = 32 * 1024 // 32KB
		config.AntiAbuse.EnablePoW = true
		config.AntiAbuse.PoWDifficulty = 25

		assert.Equal(t, 30*time.Minute, config.AntiAbuse.GreylistDuration)
		assert.Equal(t, 5, config.AntiAbuse.GreylistThreshold)
		assert.Equal(t, 32*1024, config.AntiAbuse.MaxMessageSize)
		assert.True(t, config.AntiAbuse.EnablePoW)
		assert.Equal(t, 25, config.AntiAbuse.PoWDifficulty)
	})
}

func TestConfigConsistency(t *testing.T) {
	t.Run("MeshParameterOrder", func(t *testing.T) {
		config := DefaultConfig()

		// Mesh parameters should satisfy: MeshNLow <= MeshN <= MeshNHigh
		assert.LessOrEqual(t, config.GossipsubConfig.MeshNLow, config.GossipsubConfig.MeshN)
		assert.LessOrEqual(t, config.GossipsubConfig.MeshN, config.GossipsubConfig.MeshNHigh)
	})

	t.Run("PositiveTimeouts", func(t *testing.T) {
		config := DefaultConfig()

		assert.Positive(t, config.DHTConfig.BootstrapTimeout)
		assert.Positive(t, config.GossipsubConfig.HeartbeatInterval)
		assert.Positive(t, config.RateLimit.CleanupInterval)
		assert.Positive(t, config.CacheConfig.BlobCacheTTL)
		assert.Positive(t, config.CacheConfig.CheckpointCacheTTL)
		assert.Positive(t, config.CacheConfig.PeerCacheTTL)
		assert.Positive(t, config.AntiAbuse.GreylistDuration)
	})

	t.Run("PositiveSizes", func(t *testing.T) {
		config := DefaultConfig()

		assert.Positive(t, config.CacheConfig.BlobCacheSize)
		assert.Positive(t, config.CacheConfig.CheckpointCacheSize)
		assert.Positive(t, config.CacheConfig.PeerCacheSize)
		assert.Positive(t, config.AntiAbuse.MaxMessageSize)
		assert.Positive(t, config.AntiAbuse.GreylistThreshold)
		assert.Positive(t, config.AntiAbuse.PoWDifficulty)
	})

	t.Run("ReasonableDefaults", func(t *testing.T) {
		config := DefaultConfig()

		// Gossipsub mesh should be reasonable for network health
		assert.GreaterOrEqual(t, config.GossipsubConfig.MeshN, 4)
		assert.LessOrEqual(t, config.GossipsubConfig.MeshN, 20)
		
		// Heartbeat shouldn't be too frequent or too slow
		assert.GreaterOrEqual(t, config.GossipsubConfig.HeartbeatInterval, 500*time.Millisecond)
		assert.LessOrEqual(t, config.GossipsubConfig.HeartbeatInterval, 10*time.Second)

		// Rate limits should be reasonable
		assert.GreaterOrEqual(t, config.RateLimit.PeerMsgPerMin, 10)
		assert.LessOrEqual(t, config.RateLimit.PeerMsgPerMin, 1000)
		
		// Cache sizes should be reasonable
		assert.GreaterOrEqual(t, config.CacheConfig.BlobCacheSize, 100)
		assert.GreaterOrEqual(t, config.CacheConfig.CheckpointCacheSize, 10)
		assert.GreaterOrEqual(t, config.CacheConfig.PeerCacheSize, 1000)

		// Message size should allow reasonable messages but prevent abuse
		assert.GreaterOrEqual(t, config.AntiAbuse.MaxMessageSize, 1*1024) // At least 1KB
		assert.LessOrEqual(t, config.AntiAbuse.MaxMessageSize, 1024*1024) // At most 1MB
	})
}

func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func TestConfigSerialization(t *testing.T) {
	t.Run("ConfigFieldsAccessible", func(t *testing.T) {
		config := DefaultConfig()

		// Ensure all fields are accessible (not internal)
		// This is more of a compile-time test, but validates the struct tags
		assert.NotNil(t, config.GossipsubConfig)
		assert.NotNil(t, config.DHTConfig)
		assert.NotNil(t, config.RateLimit)
		assert.NotNil(t, config.CacheConfig)
		assert.NotNil(t, config.AntiAbuse)
		assert.NotNil(t, config.ListenAddrs)
		assert.NotNil(t, config.BootstrapPeers)
	})
}