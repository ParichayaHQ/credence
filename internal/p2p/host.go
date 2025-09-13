package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

// P2PHost manages the libp2p host and associated services
type P2PHost struct {
	config *Config
	logger *Logger
	
	// Core libp2p components
	host     host.Host
	dht      *dht.IpfsDHT
	pubsub   *pubsub.PubSub
	
	// Topic management
	topics        *TopicManager
	subscriptions map[string]*pubsub.Subscription
	subMutex      sync.RWMutex
	
	// Rate limiting and anti-abuse
	rateLimiter *RateLimiter
	
	// Caches
	blobCache       *LRUCache
	checkpointCache *LRUCache
	peerCache       *LRUCache
	
	// State management
	started bool
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewP2PHost creates a new P2P host
func NewP2PHost(config *Config) *P2PHost {
	if config == nil {
		config = DefaultConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	logger := NewLogger("P2PHost", LogLevelInfo)
	
	return &P2PHost{
		config:        config,
		logger:        logger,
		topics:        NewTopicManager(),
		subscriptions: make(map[string]*pubsub.Subscription),
		rateLimiter:   NewRateLimiter(&config.RateLimit, &config.AntiAbuse),
		blobCache:     NewLRUCache(config.CacheConfig.BlobCacheSize, config.CacheConfig.BlobCacheTTL),
		checkpointCache: NewLRUCache(config.CacheConfig.CheckpointCacheSize, config.CacheConfig.CheckpointCacheTTL),
		peerCache:     NewLRUCache(config.CacheConfig.PeerCacheSize, config.CacheConfig.PeerCacheTTL),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start initializes and starts the P2P host
func (p *P2PHost) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.started {
		return ErrNodeAlreadyStarted
	}
	
	p.logger.Info("Starting P2P host", map[string]interface{}{
		"listen_addrs": len(p.config.ListenAddrs),
		"dht_mode":     p.config.DHTConfig.Mode,
	})
	
	// Create libp2p host options
	opts := []libp2p.Option{
		libp2p.ListenAddrs(p.config.ListenAddrs...),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	}
	
	// Create the libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		p.logger.Error("Failed to create libp2p host", map[string]interface{}{"error": err})
		return NewP2PError("create_host", err)
	}
	p.host = h
	
	p.logger.Info("LibP2P host created", map[string]interface{}{
		"peer_id":      h.ID().String(),
		"listen_addrs": len(h.Addrs()),
	})
	
	// Initialize DHT
	if err := p.initDHT(ctx); err != nil {
		p.logger.Error("Failed to initialize DHT", map[string]interface{}{"error": err})
		h.Close()
		return NewP2PError("init_dht", err)
	}
	
	// Initialize PubSub
	if err := p.initPubSub(ctx); err != nil {
		p.logger.Error("Failed to initialize pubsub", map[string]interface{}{"error": err})
		h.Close()
		return NewP2PError("init_pubsub", err)
	}
	
	// Bootstrap the node
	if err := p.bootstrap(ctx); err != nil {
		p.logger.Warn("Failed to bootstrap", map[string]interface{}{"error": err})
		// Don't fail startup on bootstrap failure
	}
	
	p.started = true
	
	// Subscribe to core topics
	if err := p.subscribeToTopics(ctx); err != nil {
		p.logger.Error("Failed to subscribe to core topics", map[string]interface{}{"error": err})
		h.Close()
		p.started = false
		return NewP2PError("subscribe_topics", err)
	}
	p.logger.Info("P2P host started successfully", map[string]interface{}{
		"peer_id":         p.host.ID().String(),
		"listen_addrs":    len(p.host.Addrs()),
		"subscribed_topics": len(p.subscriptions),
	})
	return nil
}

// Stop shuts down the P2P host
func (p *P2PHost) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if !p.started {
		return ErrNodeNotStarted
	}
	
	p.logger.Info("Stopping P2P host", map[string]interface{}{
		"peer_id": p.host.ID().String(),
		"subscriptions": len(p.subscriptions),
	})
	
	// Close subscriptions
	p.subMutex.Lock()
	for topic, sub := range p.subscriptions {
		p.logger.Debug("Closing subscription", map[string]interface{}{"topic": topic})
		sub.Cancel()
		delete(p.subscriptions, topic)
	}
	p.subMutex.Unlock()
	
	// Close components
	if p.rateLimiter != nil {
		p.logger.Debug("Closing rate limiter")
		p.rateLimiter.Close()
	}
	
	if p.dht != nil {
		p.logger.Debug("Closing DHT")
		if err := p.dht.Close(); err != nil {
			p.logger.Warn("Error closing DHT", map[string]interface{}{"error": err})
		}
	}
	
	if p.host != nil {
		p.logger.Debug("Closing libp2p host")
		if err := p.host.Close(); err != nil {
			p.logger.Warn("Error closing host", map[string]interface{}{"error": err})
		}
	}
	
	p.cancel()
	p.started = false
	p.logger.Info("P2P host stopped successfully")
	return nil
}

// initDHT initializes the Kademlia DHT
func (p *P2PHost) initDHT(ctx context.Context) error {
	var mode dht.ModeOpt
	switch p.config.DHTConfig.Mode {
	case "client":
		mode = dht.ModeClient
	case "server":
		mode = dht.ModeServer
	default:
		mode = dht.ModeAuto
	}
	
	kadDHT, err := dht.New(ctx, p.host,
		dht.Mode(mode),
		dht.ProtocolPrefix(protocol.ID(p.config.DHTConfig.ProtocolPrefix)),
	)
	if err != nil {
		return err
	}
	
	p.dht = kadDHT
	return nil
}

// initPubSub initializes the GossipSub system
func (p *P2PHost) initPubSub(ctx context.Context) error {
	// Configure GossipSub options
	opts := []pubsub.Option{
		pubsub.WithFloodPublish(false), // Disable flood publish except for checkpoints
		pubsub.WithMessageSigning(true),
		// Message ID function for deduplication would be added here
		// pubsub.WithMessageIdFn(...) - signature varies by libp2p version
		// Peer scoring would be configured here with application-specific functions
		// For now, use default behavior without custom scoring
	}
	
	// Note: GossipSub parameters are now set via options at initialization
	// The libp2p pubsub library doesn't expose these as settable global variables
	// in newer versions. They would be configured through NewGossipSubWithConfig
	
	// Create PubSub instance
	ps, err := pubsub.NewGossipSub(ctx, p.host, opts...)
	if err != nil {
		return err
	}
	
	p.pubsub = ps
	return nil
}

// bootstrap connects to bootstrap peers
func (p *P2PHost) bootstrap(ctx context.Context) error {
	if len(p.config.BootstrapPeers) == 0 {
		return nil // No bootstrap peers configured
	}
	
	// Connect to bootstrap peers
	for _, addr := range p.config.BootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue // Skip invalid addresses
		}
		
		// Connect with timeout
		connCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := p.host.Connect(connCtx, *pi); err != nil {
			cancel()
			continue // Skip failed connections
		}
		cancel()
	}
	
	// Bootstrap the DHT
	return p.dht.Bootstrap(ctx)
}

// subscribeToTopics subscribes to core system topics
func (p *P2PHost) subscribeToTopics(ctx context.Context) error {
	coreTopics := p.topics.GetCoreTopics()
	
	for _, topic := range coreTopics {
		if err := p.Subscribe(ctx, topic); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", topic, err)
		}
	}
	
	return nil
}

// Subscribe subscribes to a topic
func (p *P2PHost) Subscribe(ctx context.Context, topic string) error {
	if !p.started {
		return ErrNodeNotStarted
	}
	
	if !p.topics.IsValidTopic(topic) {
		p.logger.Warn("Invalid topic subscription attempt", map[string]interface{}{"topic": topic})
		return NewP2PError("subscribe", ErrInvalidTopic).WithTopic(topic)
	}
	
	p.subMutex.Lock()
	defer p.subMutex.Unlock()
	
	// Check if already subscribed
	if _, exists := p.subscriptions[topic]; exists {
		p.logger.Debug("Already subscribed to topic", map[string]interface{}{"topic": topic})
		return nil
	}
	
	p.logger.Info("Subscribing to topic", map[string]interface{}{"topic": topic})
	
	// Subscribe to topic
	sub, err := p.pubsub.Subscribe(topic)
	if err != nil {
		p.logger.Error("Failed to subscribe to topic", map[string]interface{}{
			"topic": topic,
			"error": err,
		})
		return NewP2PError("subscribe", err).WithTopic(topic)
	}
	
	p.subscriptions[topic] = sub
	
	// Start message handler for this subscription
	go p.handleTopicMessages(ctx, topic, sub)
	
	p.logger.Debug("Successfully subscribed to topic", map[string]interface{}{"topic": topic})
	return nil
}

// Publish publishes a message to a topic
func (p *P2PHost) Publish(ctx context.Context, topic string, data []byte) error {
	if !p.started {
		return ErrNodeNotStarted
	}
	
	// Validate topic and message
	if err := p.topics.ValidateTopicMessage(topic, data); err != nil {
		p.logger.Warn("Invalid message for topic", map[string]interface{}{
			"topic": topic,
			"error": err,
			"data_size": len(data),
		})
		return NewP2PError("publish", err).WithTopic(topic).WithContext("data_size", len(data))
	}
	
	p.logger.Debug("Publishing message", map[string]interface{}{
		"topic": topic,
		"data_size": len(data),
	})
	
	// Publish message
	if err := p.pubsub.Publish(topic, data); err != nil {
		p.logger.Error("Failed to publish message", map[string]interface{}{
			"topic": topic,
			"error": err,
			"data_size": len(data),
		})
		return NewP2PError("publish", err).WithTopic(topic).WithContext("data_size", len(data))
	}
	
	return nil
}

// GetNetworkInfo returns information about the network state
func (p *P2PHost) GetNetworkInfo() map[string]interface{} {
	if !p.started {
		return map[string]interface{}{"status": "stopped"}
	}
	
	peers := p.host.Network().Peers()
	
	return map[string]interface{}{
		"status":           "running",
		"peer_id":         p.host.ID().String(),
		"connected_peers": len(peers),
		"listen_addrs":    p.host.Addrs(),
		"topics":          len(p.subscriptions),
		"rate_limit_stats": p.rateLimiter.GetStats(),
	}
}

// handleTopicMessages handles incoming messages for a topic
func (p *P2PHost) handleTopicMessages(ctx context.Context, topic string, sub *pubsub.Subscription) {
	logger := p.logger.WithTopic(topic)
	logger.Info("Started message handler for topic")
	
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in topic message handler", map[string]interface{}{
				"panic": r,
			})
		}
		logger.Info("Message handler stopped for topic")
	}()
	
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Context cancelled, stopping message handler")
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // Context cancelled
				}
				logger.Error("Error receiving message", map[string]interface{}{"error": err})
				continue
			}
			
			msgLogger := p.logger.WithPeer(msg.ReceivedFrom)
			
			// Rate limit check
			if !p.rateLimiter.AllowMessage(msg.ReceivedFrom, topic, len(msg.Data)) {
				msgLogger.Debug("Message rate limited", map[string]interface{}{
					"data_size": len(msg.Data),
				})
				continue
			}
			
			// Validate message format
			if err := p.topics.ValidateTopicMessage(topic, msg.Data); err != nil {
				msgLogger.Warn("Invalid message format", map[string]interface{}{
					"error": err,
					"data_size": len(msg.Data),
				})
				continue
			}
			
			msgLogger.Debug("Processing message", map[string]interface{}{
				"data_size": len(msg.Data),
			})
			
			// Process message based on topic type
			p.processMessage(ctx, topic, msg)
		}
	}
}

// processMessage processes a received message
func (p *P2PHost) processMessage(ctx context.Context, topic string, msg *pubsub.Message) {
	logger := p.logger.WithTopic(topic)
	
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic processing message", map[string]interface{}{
				"panic": r,
				"data_size": len(msg.Data),
			})
		}
	}()
	
	// This is where we would forward messages to appropriate services
	// For now, we just cache certain types of messages
	
	topicType := p.topics.GetTopicType(topic)
	logger.Debug("Processing message by type", map[string]interface{}{
		"topic_type": topicType,
		"data_size": len(msg.Data),
	})
	
	switch topicType {
	case "checkpoint":
		// Cache checkpoint messages
		p.checkpointCache.Set(topic, msg.Data)
		logger.Debug("Cached checkpoint message", map[string]interface{}{
			"cache_key": topic,
		})
	case "blob":
		// Cache blob data
		if len(topic) > 6 { // "blobs/" prefix
			cidStr := topic[6:]
			p.blobCache.Set(cidStr, msg.Data)
			logger.Debug("Cached blob data", map[string]interface{}{
				"cid": cidStr,
			})
		}
	default:
		logger.Debug("No special processing for topic type")
	}
	
	// TODO: Forward to appropriate internal services via HTTP bridge
}

// FindProviders finds providers for a CID using the DHT
func (p *P2PHost) FindProviders(ctx context.Context, c cid.Cid) ([]peer.AddrInfo, error) {
	if !p.started {
		return nil, ErrNodeNotStarted
	}
	
	if p.dht == nil {
		return nil, ErrDHTNotReady
	}
	
	cidStr := c.String()
	p.logger.Debug("Finding providers for CID", map[string]interface{}{"cid": cidStr})
	
	// For now, simplified provider discovery - would need proper DHT integration
	// The exact DHT API varies by version, so this is a placeholder implementation
	p.logger.Debug("DHT provider discovery not fully implemented")
	
	// Return empty result for now - this would be replaced with actual DHT provider discovery
	var result []peer.AddrInfo
	// TODO: Implement actual DHT provider discovery
	// for provider := range providersCh {
	//     result = append(result, provider)
	//     if len(result) >= 20 { break }
	// }
	
	if len(result) == 0 {
		p.logger.Info("No providers found for CID", map[string]interface{}{"cid": cidStr})
		return nil, NewP2PError("find_providers", ErrProviderNotFound).WithContext("cid", cidStr)
	}
	
	p.logger.Info("Found providers for CID", map[string]interface{}{
		"cid": cidStr,
		"provider_count": len(result),
	})
	return result, nil
}

// Provide announces that this node can provide content for a CID
func (p *P2PHost) Provide(ctx context.Context, c cid.Cid) error {
	if !p.started {
		return ErrNodeNotStarted
	}
	
	if p.dht == nil {
		return ErrDHTNotReady
	}
	
	return p.dht.Provide(ctx, c, true)
}

// ConnectToPeer connects to a specific peer
func (p *P2PHost) ConnectToPeer(ctx context.Context, addr multiaddr.Multiaddr) error {
	if !p.started {
		return ErrNodeNotStarted
	}
	
	pi, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		p.logger.Warn("Invalid peer address", map[string]interface{}{
			"addr": addr.String(),
			"error": err,
		})
		return NewP2PError("connect_peer", err).WithContext("addr", addr.String())
	}
	
	p.logger.Info("Connecting to peer", map[string]interface{}{
		"peer_id": pi.ID.String(),
		"addr": addr.String(),
	})
	
	connCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	if err := p.host.Connect(connCtx, *pi); err != nil {
		p.logger.Error("Failed to connect to peer", map[string]interface{}{
			"peer_id": pi.ID.String(),
			"addr": addr.String(),
			"error": err,
		})
		return NewP2PError("connect_peer", err).WithPeer(pi.ID).WithContext("addr", addr.String())
	}
	
	p.logger.Info("Successfully connected to peer", map[string]interface{}{
		"peer_id": pi.ID.String(),
	})
	return nil
}