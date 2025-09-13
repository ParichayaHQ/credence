package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Benchmark P2P operations
func BenchmarkP2POperations(b *testing.B) {
	// Setup common test data
	config := DefaultConfig()
	rateLimiter := NewRateLimiter(&config.RateLimit, &config.AntiAbuse)
	defer rateLimiter.Close()

	topicManager := NewTopicManager()
	testData := []byte(`{"type":"vouch","from":"did:key:z123","to":"did:key:z456","context":"commerce"}`)
	
	b.Run("RateLimit_AllowMessage", func(b *testing.B) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rateLimiter.AllowMessage(peerID, "events/vouch", len(testData))
		}
	})

	b.Run("TopicManager_IsValidTopic", func(b *testing.B) {
		topics := []string{
			"events/vouch",
			"events/report",
			"blobs/QmHash123",
			"invalid/topic",
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			topic := topics[i%len(topics)]
			topicManager.IsValidTopic(topic)
		}
	})

	b.Run("TopicManager_ValidateTopicMessage", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			topicManager.ValidateTopicMessage("events/vouch", testData)
		}
	})

	b.Run("P2PError_Creation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := NewP2PError("test_operation", ErrRateLimited)
			_ = err.Error()
		}
	})

	b.Run("P2PError_WithContext", func(b *testing.B) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := NewP2PError("test_operation", ErrRateLimited).
				WithPeer(peerID).
				WithTopic("events/vouch").
				WithContext("size", len(testData))
			_ = err.Error()
		}
	})
}

func BenchmarkCacheOperations(b *testing.B) {
	cache := NewLRUCache(1000, time.Hour)
	defer cache.Close()

	testData := make([]byte, 1024) // 1KB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	b.Run("Cache_Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + string(rune(i%1000))
			cache.Set(key, testData)
		}
	})

	// Pre-populate cache for Get benchmark
	for i := 0; i < 1000; i++ {
		key := "key" + string(rune(i))
		cache.Set(key, testData)
	}

	b.Run("Cache_Get_Hit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + string(rune(i%1000))
			cache.Get(key)
		}
	})

	b.Run("Cache_Get_Miss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "miss" + string(rune(i%1000))
			cache.Get(key)
		}
	})

	b.Run("Cache_Has", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + string(rune(i%1000))
			cache.Has(key)
		}
	})

	b.Run("Cache_Mixed_Operations", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "mixed" + string(rune(i%1000))
			switch i % 4 {
			case 0:
				cache.Set(key, testData)
			case 1, 2:
				cache.Get(key)
			case 3:
				cache.Has(key)
			}
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	logger := NewLogger("benchmark", LogLevelInfo)

	b.Run("Logger_InfoWithoutFields", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("Logger_InfoWithFields", func(b *testing.B) {
		fields := map[string]interface{}{
			"iteration": 0,
			"component": "benchmark",
			"active":    true,
			"size":      1024,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fields["iteration"] = i
			logger.Info("benchmark message", fields)
		}
	})

	b.Run("Logger_DebugFiltered", func(b *testing.B) {
		// Debug messages should be filtered out at Info level
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("debug message that should be filtered")
		}
	})

	b.Run("Logger_WithContext", func(b *testing.B) {
		ctx := logger.WithFields(map[string]interface{}{
			"session": "benchmark",
			"user":    "test",
		})
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx.Info("context message", map[string]interface{}{
				"iteration": i,
			})
		}
	})

	b.Run("Logger_ErrorLevel", func(b *testing.B) {
		errorLogger := NewLogger("benchmark", LogLevelError)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// These should be filtered out
			errorLogger.Debug("debug")
			errorLogger.Info("info")
			errorLogger.Warn("warn")
			// Only this should be logged
			errorLogger.Error("error")
		}
	})
}

func BenchmarkNetworkOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping network benchmarks in short mode")
	}

	config := DefaultConfig()
	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

	b.Run("P2PHost_StartStop", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			host := NewP2PHost(config)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			b.StartTimer()
			
			host.Start(ctx)
			host.Stop(ctx)
			
			cancel()
		}
	})

	b.Run("NetworkInfo_Retrieval", func(b *testing.B) {
		host := NewP2PHost(config)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		host.Start(ctx)
		defer host.Stop(ctx)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			host.GetNetworkInfo()
		}
	})

	b.Run("Topic_Subscribe", func(b *testing.B) {
		host := NewP2PHost(config)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		host.Start(ctx)
		defer host.Stop(ctx)
		
		topics := []string{
			"events/vouch",
			"events/report",
			"events/appeal",
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			topic := topics[i%len(topics)]
			host.Subscribe(ctx, topic)
		}
	})

	b.Run("Message_Publish", func(b *testing.B) {
		host := NewP2PHost(config)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		host.Start(ctx)
		defer host.Stop(ctx)
		
		// Subscribe to topic first
		host.Subscribe(ctx, "events/vouch")
		
		testData := []byte(`{"type":"vouch","from":"did:key:z123","to":"did:key:z456"}`)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			host.Publish(ctx, "events/vouch", testData)
		}
	})
}

func BenchmarkBlobOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping blob benchmarks in short mode")
	}

	config := DefaultConfig()
	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

	host := NewP2PHost(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	host.Start(ctx)
	defer host.Stop(ctx)

	blobManager := NewBlobManager(host, 1000, 10*time.Minute)
	defer blobManager.Close()

	// Test data of various sizes
	testData1KB := make([]byte, 1024)
	testData10KB := make([]byte, 10*1024)
	testData100KB := make([]byte, 100*1024)

	for i := range testData1KB {
		testData1KB[i] = byte(i % 256)
	}
	for i := range testData10KB {
		testData10KB[i] = byte(i % 256)
	}
	for i := range testData100KB {
		testData100KB[i] = byte(i % 256)
	}

	b.Run("Blob_Store_1KB", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			blobManager.StoreBlob(ctx, testData1KB)
		}
	})

	b.Run("Blob_Store_10KB", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			blobManager.StoreBlob(ctx, testData10KB)
		}
	})

	b.Run("Blob_Store_100KB", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			blobManager.StoreBlob(ctx, testData100KB)
		}
	})

	// Pre-store some blobs for retrieval benchmarks
	storedCIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		cid, _ := blobManager.StoreBlob(ctx, testData1KB)
		storedCIDs[i] = cid.String()
	}

	b.Run("Blob_Has", func(b *testing.B) {
		cid, _ := blobManager.StoreBlob(ctx, testData1KB)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			blobManager.HasBlob(cid)
		}
	})

	b.Run("Blob_GetStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			blobManager.GetBlobStats()
		}
	})
}

func BenchmarkConfigOperations(b *testing.B) {
	b.Run("DefaultConfig_Creation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultConfig()
		}
	})

	b.Run("Config_Validation", func(b *testing.B) {
		config := DefaultConfig()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate validation operations
			_ = config.GossipsubConfig.MeshN >= config.GossipsubConfig.MeshNLow
			_ = config.GossipsubConfig.MeshN <= config.GossipsubConfig.MeshNHigh
			_ = config.CacheConfig.BlobCacheSize > 0
			_ = config.RateLimit.PeerMsgPerMin > 0
		}
	})
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("LRUCache_Allocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache := NewLRUCache(100, time.Hour)
			cache.Close()
		}
	})

	b.Run("Logger_Allocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewLogger("test", LogLevelInfo)
		}
	})

	b.Run("TopicManager_Allocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewTopicManager()
		}
	})

	b.Run("Error_Allocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewP2PError("test", ErrRateLimited)
		}
	})
}

// Concurrent operation benchmarks
func BenchmarkConcurrentOperations(b *testing.B) {
	cache := NewLRUCache(10000, time.Hour)
	defer cache.Close()

	testData := make([]byte, 1024)
	
	b.Run("Cache_ConcurrentReadWrite", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := "key" + string(rune(i%1000))
				if i%3 == 0 {
					cache.Set(key, testData)
				} else {
					cache.Get(key)
				}
				i++
			}
		})
	})

	rateLimiter := NewRateLimiter(&RateLimitConfig{
		PeerMsgPerMin:   1000,
		PeerBytesPerSec: 1024*1024,
		GlobalMsgPerSec: 10000,
		BurstMultiplier: 2.0,
		CleanupInterval: time.Minute,
	}, &AntiAbuseConfig{
		GreylistDuration:  10 * time.Minute,
		GreylistThreshold: 10,
		MaxMessageSize:    16 * 1024,
	})
	defer rateLimiter.Close()

	b.Run("RateLimit_ConcurrentCheck", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
			for pb.Next() {
				rateLimiter.AllowMessage(peerID, "events/vouch", 100)
			}
		})
	})

	logger := NewLogger("concurrent", LogLevelInfo)
	
	b.Run("Logger_ConcurrentLog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				logger.Info("concurrent message", map[string]interface{}{
					"iteration": i,
					"goroutine": "test",
				})
				i++
			}
		})
	})
}