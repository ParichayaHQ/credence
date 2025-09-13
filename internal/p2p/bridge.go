package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
)

// HTTPBridge provides HTTP endpoints for internal services to interact with P2P
type HTTPBridge struct {
	p2pHost    *P2PHost
	server     *http.Server
	listenAddr string
	logger     *Logger
}

// PublishRequest represents a message publish request
type PublishRequest struct {
	Topic string `json:"topic" validate:"required"`
	Data  []byte `json:"data" validate:"required"`
}

// PublishResponse represents a message publish response
type PublishResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NetworkStatusResponse represents network status information
type NetworkStatusResponse struct {
	Status          string                 `json:"status"`
	PeerID          string                 `json:"peer_id"`
	ConnectedPeers  int                    `json:"connected_peers"`
	ListenAddrs     []string               `json:"listen_addrs"`
	Topics          []string               `json:"topics"`
	RateLimitStats  map[string]interface{} `json:"rate_limit_stats"`
	CacheStats      map[string]interface{} `json:"cache_stats"`
}

// ProvidersResponse represents the response for provider queries
type ProvidersResponse struct {
	Providers []string `json:"providers"`
	Count     int      `json:"count"`
}

// NewHTTPBridge creates a new HTTP bridge
func NewHTTPBridge(p2pHost *P2PHost, listenAddr string) *HTTPBridge {
	return &HTTPBridge{
		p2pHost:    p2pHost,
		listenAddr: listenAddr,
		logger:     NewLogger("HTTPBridge", LogLevelInfo),
	}
}

// Start starts the HTTP bridge server
func (b *HTTPBridge) Start(ctx context.Context) error {
	b.logger.Info("Starting HTTP bridge", map[string]interface{}{
		"listen_addr": b.listenAddr,
	})
	
	mux := http.NewServeMux()
	
	// Register endpoints with logging middleware
	mux.HandleFunc("/v1/publish", b.withLogging(b.handlePublish))
	mux.HandleFunc("/v1/blobs/", b.withLogging(b.handleBlobs))
	mux.HandleFunc("/v1/subscribe", b.withLogging(b.handleSubscribe))
	mux.HandleFunc("/v1/checkpoints/latest", b.withLogging(b.handleLatestCheckpoint))
	mux.HandleFunc("/v1/providers/", b.withLogging(b.handleProviders))
	mux.HandleFunc("/v1/connect", b.withLogging(b.handleConnect))
	mux.HandleFunc("/v1/status", b.withLogging(b.handleStatus))
	mux.HandleFunc("/health", b.withLogging(b.handleHealth))
	
	// Create server
	b.server = &http.Server{
		Addr:         b.listenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		b.logger.Info("HTTP bridge server listening", map[string]interface{}{
			"addr": b.listenAddr,
		})
		if err := b.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.logger.Error("HTTP bridge server error", map[string]interface{}{"error": err})
		}
	}()
	
	return nil
}

// Stop stops the HTTP bridge server
func (b *HTTPBridge) Stop(ctx context.Context) error {
	if b.server != nil {
		b.logger.Info("Stopping HTTP bridge server")
		if err := b.server.Shutdown(ctx); err != nil {
			b.logger.Error("Error stopping HTTP bridge server", map[string]interface{}{"error": err})
			return err
		}
		b.logger.Info("HTTP bridge server stopped successfully")
	}
	return nil
}

// withLogging wraps HTTP handlers with request logging
func (b *HTTPBridge) withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer that captures status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		b.logger.Debug("HTTP request started", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		})
		
		// Call the actual handler
		defer func() {
			if rec := recover(); rec != nil {
				b.logger.Error("HTTP handler panic", map[string]interface{}{
					"panic":  rec,
					"method": r.Method,
					"path":   r.URL.Path,
				})
				http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next(rw, r)
		
		duration := time.Since(start)
		b.logger.Info("HTTP request completed", map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      rw.statusCode,
			"duration_ms": duration.Milliseconds(),
			"remote":      r.RemoteAddr,
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handlePublish handles message publishing
func (b *HTTPBridge) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse query parameters for topic
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "Topic parameter required", http.StatusBadRequest)
		return
	}
	
	// Read message data
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Publish message
	err = b.p2pHost.Publish(r.Context(), topic, body)
	
	response := PublishResponse{
		Success: err == nil,
	}
	
	if err != nil {
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleBlobs handles blob retrieval
func (b *HTTPBridge) handleBlobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract CID from path
	path := strings.TrimPrefix(r.URL.Path, "/v1/blobs/")
	if path == "" {
		http.Error(w, "CID required", http.StatusBadRequest)
		return
	}
	
	// Check cache first
	if data, found := b.p2pHost.blobCache.Get(path); found {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(data)
		return
	}
	
	// Parse CID
	c, err := cid.Parse(path)
	if err != nil {
		http.Error(w, "Invalid CID", http.StatusBadRequest)
		return
	}
	
	// Find providers
	providers, err := b.p2pHost.FindProviders(r.Context(), c)
	if err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}
	
	if len(providers) == 0 {
		http.Error(w, "No providers found", http.StatusNotFound)
		return
	}
	
	// For now, return provider information
	// In a full implementation, we would fetch the content from providers
	response := map[string]interface{}{
		"cid":       path,
		"providers": len(providers),
		"message":   "Content discovery successful (fetch not implemented)",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSubscribe handles topic subscription (Server-Sent Events)
func (b *HTTPBridge) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "Topic parameter required", http.StatusBadRequest)
		return
	}
	
	// Validate topic
	if !b.p2pHost.topics.IsValidTopic(topic) {
		http.Error(w, "Invalid topic", http.StatusBadRequest)
		return
	}
	
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Subscribe to topic if not already subscribed
	if err := b.p2pHost.Subscribe(r.Context(), topic); err != nil {
		http.Error(w, "Failed to subscribe", http.StatusInternalServerError)
		return
	}
	
	// For now, just send a confirmation
	// In a full implementation, we would stream messages
	fmt.Fprintf(w, "data: {\"type\":\"subscribed\",\"topic\":\"%s\"}\n\n", topic)
	
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	
	// Keep connection alive (simplified)
	select {
	case <-r.Context().Done():
		return
	case <-time.After(30 * time.Second):
		fmt.Fprintf(w, "data: {\"type\":\"keepalive\"}\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// handleLatestCheckpoint handles latest checkpoint retrieval
func (b *HTTPBridge) handleLatestCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Try to get latest checkpoint from cache
	if data, found := b.p2pHost.checkpointCache.Get("latest"); found {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}
	
	// No cached checkpoint available
	response := map[string]interface{}{
		"message": "No checkpoint available",
		"status":  "not_found",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}

// handleProviders handles provider queries
func (b *HTTPBridge) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract CID from path
	path := strings.TrimPrefix(r.URL.Path, "/v1/providers/")
	if path == "" {
		http.Error(w, "CID required", http.StatusBadRequest)
		return
	}
	
	// Parse CID
	c, err := cid.Parse(path)
	if err != nil {
		http.Error(w, "Invalid CID", http.StatusBadRequest)
		return
	}
	
	// Find providers
	providers, err := b.p2pHost.FindProviders(r.Context(), c)
	if err != nil {
		response := ProvidersResponse{
			Providers: []string{},
			Count:     0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Convert to string slice
	providerStrs := make([]string, len(providers))
	for i, p := range providers {
		providerStrs[i] = p.ID.String()
	}
	
	response := ProvidersResponse{
		Providers: providerStrs,
		Count:     len(providerStrs),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleConnect handles peer connection requests
func (b *HTTPBridge) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	addr := r.URL.Query().Get("addr")
	if addr == "" {
		http.Error(w, "Address parameter required", http.StatusBadRequest)
		return
	}
	
	// Parse multiaddress
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		http.Error(w, "Invalid multiaddress", http.StatusBadRequest)
		return
	}
	
	// Connect to peer
	err = b.p2pHost.ConnectToPeer(r.Context(), ma)
	
	response := map[string]interface{}{
		"success": err == nil,
		"address": addr,
	}
	
	if err != nil {
		response["error"] = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatus handles network status requests
func (b *HTTPBridge) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	netInfo := b.p2pHost.GetNetworkInfo()
	
	// Add cache stats
	cacheStats := map[string]interface{}{
		"blob_cache":       b.p2pHost.blobCache.Stats(),
		"checkpoint_cache": b.p2pHost.checkpointCache.Stats(),
		"peer_cache":       b.p2pHost.peerCache.Stats(),
	}
	
	response := NetworkStatusResponse{
		Status:         netInfo["status"].(string),
		CacheStats:     cacheStats,
		RateLimitStats: netInfo["rate_limit_stats"].(map[string]interface{}),
	}
	
	if status, ok := netInfo["status"].(string); ok && status == "running" {
		response.PeerID = netInfo["peer_id"].(string)
		response.ConnectedPeers = netInfo["connected_peers"].(int)
		
		// Convert listen addresses
		if addrs, ok := netInfo["listen_addrs"].([]string); ok {
			response.ListenAddrs = addrs
		}
		
		// Get subscribed topics
		b.p2pHost.subMutex.RLock()
		topics := make([]string, 0, len(b.p2pHost.subscriptions))
		for topic := range b.p2pHost.subscriptions {
			topics = append(topics, topic)
		}
		b.p2pHost.subMutex.RUnlock()
		response.Topics = topics
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles health check requests
func (b *HTTPBridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(time.Now()).String(), // Would track actual uptime
	}
	
	// Check if P2P host is running
	if !b.p2pHost.started {
		health["status"] = "unhealthy"
		health["error"] = "P2P host not started"
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}