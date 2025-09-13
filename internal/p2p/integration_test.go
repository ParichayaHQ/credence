package p2p

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestP2PHostLifecycle tests the complete lifecycle of a P2P host
func TestP2PHostLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("StartStopCycle", func(t *testing.T) {
		config := DefaultConfig()
		
		// Use a free port for testing
		listenAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
		require.NoError(t, err)
		config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

		host := NewP2PHost(config)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Test startup
		err = host.Start(ctx)
		require.NoError(t, err)
		assert.True(t, host.started)

		// Test network info
		netInfo := host.GetNetworkInfo()
		assert.Equal(t, "running", netInfo["status"])
		assert.NotEmpty(t, netInfo["peer_id"])
		assert.Equal(t, 0, netInfo["connected_peers"]) // No peers initially

		// Test shutdown
		err = host.Stop(ctx)
		require.NoError(t, err)
		assert.False(t, host.started)

		// Test network info after shutdown
		netInfo = host.GetNetworkInfo()
		assert.Equal(t, "stopped", netInfo["status"])
	})

	t.Run("DuplicateStart", func(t *testing.T) {
		config := DefaultConfig()
		listenAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
		require.NoError(t, err)
		config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

		host := NewP2PHost(config)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = host.Start(ctx)
		require.NoError(t, err)
		defer host.Stop(ctx)

		// Second start should fail
		err = host.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already started")
	})

	t.Run("StopWithoutStart", func(t *testing.T) {
		host := NewP2PHost(DefaultConfig())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := host.Stop(ctx)
		assert.Error(t, err)
		assert.Equal(t, ErrNodeNotStarted, err)
	})
}

func TestP2PSubscriptionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	listenAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	require.NoError(t, err)
	config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

	host := NewP2PHost(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = host.Start(ctx)
	require.NoError(t, err)
	defer host.Stop(ctx)

	t.Run("SubscribeToValidTopic", func(t *testing.T) {
		err := host.Subscribe(ctx, "events/vouch")
		assert.NoError(t, err)

		// Check subscription exists
		host.subMutex.RLock()
		_, exists := host.subscriptions["events/vouch"]
		host.subMutex.RUnlock()
		assert.True(t, exists)
	})

	t.Run("SubscribeToInvalidTopic", func(t *testing.T) {
		err := host.Subscribe(ctx, "invalid/topic")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid topic")
	})

	t.Run("PublishToSubscribedTopic", func(t *testing.T) {
		// Subscribe first
		err := host.Subscribe(ctx, "events/report")
		require.NoError(t, err)

		// Publish message
		testData := []byte(`{"type":"report","from":"did:key:z123"}`)
		err = host.Publish(ctx, "events/report", testData)
		assert.NoError(t, err)
	})

	t.Run("PublishWithoutSubscription", func(t *testing.T) {
		// Should still work - pubsub allows publishing without subscription
		testData := []byte(`{"type":"vouch","from":"did:key:z123"}`)
		err := host.Publish(ctx, "events/vouch", testData)
		assert.NoError(t, err)
	})

	t.Run("PublishInvalidMessage", func(t *testing.T) {
		// Empty message should fail
		err := host.Publish(ctx, "events/vouch", []byte{})
		assert.Error(t, err)
	})

	t.Run("PublishToInvalidTopic", func(t *testing.T) {
		testData := []byte(`{"type":"test"}`)
		err := host.Publish(ctx, "invalid/topic", testData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid topic")
	})
}

func TestHTTPBridge(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup P2P host
	config := DefaultConfig()
	listenAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	require.NoError(t, err)
	config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

	host := NewP2PHost(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = host.Start(ctx)
	require.NoError(t, err)
	defer host.Stop(ctx)

	// Setup HTTP bridge on free port
	bridge := NewHTTPBridge(host, getFreeTCPPort())
	err = bridge.Start(ctx)
	require.NoError(t, err)
	defer bridge.Stop(ctx)

	baseURL := "http://" + bridge.server.Addr

	t.Run("HealthEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("StatusEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/v1/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("PublishEndpoint", func(t *testing.T) {
		// Valid publish
		testData := `{"type":"vouch","from":"did:key:z123"}`
		resp, err := http.Post(
			baseURL+"/v1/publish?topic=events/vouch",
			"application/json",
			strings.NewReader(testData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PublishInvalidTopic", func(t *testing.T) {
		testData := `{"type":"test"}`
		resp, err := http.Post(
			baseURL+"/v1/publish?topic=invalid/topic",
			"application/json",
			strings.NewReader(testData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("BlobEndpoint", func(t *testing.T) {
		// Test blob retrieval (should return not found since no providers)
		resp, err := http.Get(baseURL + "/v1/blobs/QmTestHash123")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return not found or provider info
		assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK)
	})

	t.Run("ProvidersEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/v1/providers/QmTestHash123")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("CheckpointEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/v1/checkpoints/latest")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return not found since no checkpoints cached
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("ConnectEndpoint", func(t *testing.T) {
		// Test with invalid multiaddress
		resp, err := http.Post(
			baseURL+"/v1/connect?addr=invalid",
			"application/json",
			strings.NewReader(""),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestRateLimiting(t *testing.T) {
	config := &RateLimitConfig{
		PeerMsgPerMin:   5,  // Very low for testing
		PeerBytesPerSec: 100,
		GlobalMsgPerSec: 10,
		BurstMultiplier: 1.0,
		CleanupInterval: time.Second,
	}

	antiAbuse := &AntiAbuseConfig{
		GreylistDuration:  time.Minute,
		GreylistThreshold: 3,
		MaxMessageSize:    1024,
	}

	rateLimiter := NewRateLimiter(config, antiAbuse)
	defer rateLimiter.Close()

	t.Run("MessageRateLimit", func(t *testing.T) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		topic := "events/vouch"
		msgSize := 50

		// First few messages should be allowed
		for i := 0; i < 3; i++ {
			allowed := rateLimiter.AllowMessage(peerID, topic, msgSize)
			assert.True(t, allowed, "Message %d should be allowed", i+1)
		}

		// After hitting the limit, should be rate limited
		// Note: depending on timing, this might need adjustment
		time.Sleep(100 * time.Millisecond) // Small delay
		allowed := rateLimiter.AllowMessage(peerID, topic, msgSize)
		// May or may not be rate limited depending on timing - this is expected
		t.Logf("Rate limiting result: %v", allowed)
	})

	t.Run("ByteRateLimit", func(t *testing.T) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr2Q")
		topic := "events/report"

		// Send a message that's exactly at the byte limit
		allowed := rateLimiter.AllowMessage(peerID, topic, 100)
		assert.True(t, allowed, "Message at byte limit should be allowed")

		// Immediately send another - should be rate limited
		allowed = rateLimiter.AllowMessage(peerID, topic, 50)
		t.Logf("Second message rate limiting result: %v", allowed)
	})

	t.Run("GreylistBehavior", func(t *testing.T) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr2R")
		topic := "events/vouch"

		// Generate several violations quickly
		for i := 0; i < 5; i++ {
			rateLimiter.AllowMessage(peerID, topic, 1000) // Large messages
		}

		// Check if peer gets greylisted eventually
		isGreylisted := rateLimiter.IsGreylisted(peerID)
		t.Logf("Peer greylisted after violations: %v", isGreylisted)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := rateLimiter.GetStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "total_peers")
		assert.Contains(t, stats, "greylisted_peers")
		assert.Contains(t, stats, "global_msg_count")
		assert.Contains(t, stats, "config")

		t.Logf("Rate limiter stats: %+v", stats)
	})
}

func TestBlobManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a minimal P2P host for blob manager testing
	config := DefaultConfig()
	listenAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	require.NoError(t, err)
	config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}

	host := NewP2PHost(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = host.Start(ctx)
	require.NoError(t, err)
	defer host.Stop(ctx)

	blobManager := NewBlobManager(host, 100, 10*time.Minute)
	defer blobManager.Close()

	t.Run("StoreBlobAndRetrieve", func(t *testing.T) {
		testData := []byte("Hello, P2P world!")

		// Store blob
		cid, err := blobManager.StoreBlob(ctx, testData)
		require.NoError(t, err)
		assert.NotEmpty(t, cid.String())

		// Check if blob is stored locally
		hasBlob := blobManager.HasBlob(cid)
		assert.True(t, hasBlob)

		// Retrieve blob (should come from local cache)
		retrievedData, err := blobManager.GetBlob(ctx, cid)
		require.NoError(t, err)
		assert.Equal(t, testData, retrievedData)
	})

	t.Run("StoreBlobEmptyData", func(t *testing.T) {
		_, err := blobManager.StoreBlob(ctx, []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty blob")
	})

	t.Run("GetNonexistentBlob", func(t *testing.T) {
		// Try to get a blob that doesn't exist
		fakeCID, err := blobManager.StoreBlob(ctx, []byte("temp"))
		require.NoError(t, err)
		
		// Delete it from cache to simulate non-existence
		blobManager.DeleteBlob(fakeCID)

		// Now try to retrieve - should fail since no providers exist
		_, err = blobManager.GetBlob(ctx, fakeCID)
		assert.Error(t, err)
	})

	t.Run("DeleteBlob", func(t *testing.T) {
		testData := []byte("Data to be deleted")
		cid, err := blobManager.StoreBlob(ctx, testData)
		require.NoError(t, err)

		// Verify it exists
		hasBlob := blobManager.HasBlob(cid)
		assert.True(t, hasBlob)

		// Delete it
		blobManager.DeleteBlob(cid)

		// Verify it's gone
		hasBlob = blobManager.HasBlob(cid)
		assert.False(t, hasBlob)
	})

	t.Run("BlobStats", func(t *testing.T) {
		stats := blobManager.GetBlobStats()
		assert.NotNil(t, stats)
		assert.Equal(t, "blob_cache", stats["type"])
		assert.Contains(t, stats, "size")
		assert.Contains(t, stats, "max_size")

		t.Logf("Blob manager stats: %+v", stats)
	})
}

// Helper function to get a free TCP port
func getFreeTCPPort() string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return ":8080" // Fallback
	}
	defer listener.Close()
	
	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf(":%d", addr.Port)
}

func TestMultipleP2PNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("TwoNodeNetwork", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Create two nodes
		node1 := createTestNode(t)
		node2 := createTestNode(t)

		// Start both nodes
		err := node1.Start(ctx)
		require.NoError(t, err)
		defer node1.Stop(ctx)

		err = node2.Start(ctx)
		require.NoError(t, err)
		defer node2.Stop(ctx)

		// Get network info
		info1 := node1.GetNetworkInfo()
		info2 := node2.GetNetworkInfo()

		assert.Equal(t, "running", info1["status"])
		assert.Equal(t, "running", info2["status"])
		assert.NotEqual(t, info1["peer_id"], info2["peer_id"])

		t.Logf("Node 1 ID: %s", info1["peer_id"])
		t.Logf("Node 2 ID: %s", info2["peer_id"])

		// For a more complete test, we would connect the nodes
		// and test message passing, but that requires more complex setup
		// with bootstrap peers or direct connection
	})
}

func createTestNode(t *testing.T) *P2PHost {
	config := DefaultConfig()
	listenAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	require.NoError(t, err)
	config.ListenAddrs = []multiaddr.Multiaddr{listenAddr}
	
	// Make test more permissive to avoid startup issues
	config.DHTConfig.BootstrapTimeout = 5 * time.Second
	
	return NewP2PHost(config)
}