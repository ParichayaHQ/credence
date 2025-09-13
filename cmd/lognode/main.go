package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	logpkg "github.com/ParichayaHQ/credence/internal/log"
)

// LogNodeServer wraps the transparency log with HTTP API
type LogNodeServer struct {
	transparencyLog logpkg.TransparencyLog
	config          *ServerConfig
	server          *http.Server
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address      string        `json:"address"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:      "0.0.0.0",
		Port:         8081,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func main() {
	// Load configuration
	logConfig := logpkg.DefaultConfig()
	serverConfig := DefaultServerConfig()
	
	// Override with environment variables if present
	if addr := os.Getenv("LOGNODE_ADDRESS"); addr != "" {
		serverConfig.Address = addr
	}
	if portStr := os.Getenv("LOGNODE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			serverConfig.Port = port
		}
	}
	if treeIDStr := os.Getenv("TREE_ID"); treeIDStr != "" {
		if treeID, err := strconv.ParseInt(treeIDStr, 10, 64); err == nil {
			logConfig.TreeID = treeID
		}
	}
	
	// Initialize transparency log
	transparencyLog, err := logpkg.NewMemoryTransparencyLog(logConfig)
	if err != nil {
		log.Fatalf("Failed to initialize transparency log: %v", err)
	}
	defer transparencyLog.Close()
	
	// Create server
	server := &LogNodeServer{
		transparencyLog: transparencyLog,
		config:          serverConfig,
	}
	
	// Start HTTP server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down log node...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Log node stopped")
}

// Start starts the HTTP server
func (s *LogNodeServer) Start() error {
	router := s.setupRoutes()
	
	addr := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	
	log.Printf("Starting log node server on %s", addr)
	
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	
	return nil
}

// Shutdown gracefully shuts down the server
func (s *LogNodeServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// setupRoutes configures HTTP routes
func (s *LogNodeServer) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// API v1 routes
	v1 := r.PathPrefix("/v1").Subrouter()
	
	// Health check and stats
	r.HandleFunc("/health", s.handleHealth).Methods("GET")
	r.HandleFunc("/stats", s.handleStats).Methods("GET")
	
	// Log operations
	v1.HandleFunc("/log/append", s.handleAppendLeaves).Methods("POST")
	v1.HandleFunc("/log/leaf", s.handleGetLeafByHash).Methods("GET")
	v1.HandleFunc("/log/leaves", s.handleGetLeavesByRange).Methods("GET")
	v1.HandleFunc("/log/inclusion", s.handleGetInclusionProof).Methods("GET")
	v1.HandleFunc("/log/consistency", s.handleGetConsistencyProof).Methods("GET")
	v1.HandleFunc("/log/sth", s.handleGetSignedTreeHead).Methods("GET")
	v1.HandleFunc("/log/size", s.handleGetTreeSize).Methods("GET")
	
	// Event reference helpers
	v1.HandleFunc("/events/append", s.handleAppendEventReferences).Methods("POST")
	
	// Add middleware
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)
	
	return r
}

// Health check endpoint
func (s *LogNodeServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "lognode",
		"tree_size": s.getTreeSizeUnsafe(),
	})
}

// Statistics endpoint
func (s *LogNodeServer) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Get stats from memory log if possible
	var stats interface{}
	if memLog, ok := s.transparencyLog.(*logpkg.MemoryTransparencyLog); ok {
		stats = memLog.GetStats()
	} else {
		// Fallback to basic stats
		treeSize, _ := s.transparencyLog.GetTreeSize(r.Context())
		stats = map[string]interface{}{
			"tree_size": treeSize,
			"backend":   "unknown",
		}
	}
	
	json.NewEncoder(w).Encode(stats)
}

// Append leaves to the log
func (s *LogNodeServer) handleAppendLeaves(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Leaves []logpkg.Leaf `json:"leaves"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	if len(request.Leaves) == 0 {
		http.Error(w, "No leaves provided", http.StatusBadRequest)
		return
	}
	
	result, err := s.transparencyLog.AppendLeaves(r.Context(), request.Leaves)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to append leaves: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// Get leaf by hash
func (s *LogNodeServer) handleGetLeafByHash(w http.ResponseWriter, r *http.Request) {
	hashParam := r.URL.Query().Get("hash")
	if hashParam == "" {
		http.Error(w, "Missing hash parameter", http.StatusBadRequest)
		return
	}
	
	leafHash, err := hex.DecodeString(hashParam)
	if err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}
	
	leaf, err := s.transparencyLog.GetLeafByHash(r.Context(), leafHash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get leaf: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaf)
}

// Get leaves by range
func (s *LogNodeServer) handleGetLeavesByRange(w http.ResponseWriter, r *http.Request) {
	startParam := r.URL.Query().Get("start")
	endParam := r.URL.Query().Get("end")
	
	if startParam == "" {
		http.Error(w, "Missing start parameter", http.StatusBadRequest)
		return
	}
	
	startSeq, err := strconv.ParseInt(startParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid start parameter", http.StatusBadRequest)
		return
	}
	
	endSeq := startSeq + 99 // Default to 100 leaves
	if endParam != "" {
		if parsed, err := strconv.ParseInt(endParam, 10, 64); err == nil {
			endSeq = parsed
		}
	}
	
	// Limit range size
	if endSeq-startSeq > 1000 {
		endSeq = startSeq + 1000
	}
	
	leaves, err := s.transparencyLog.GetLeavesByRange(r.Context(), startSeq, endSeq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get leaves: %v", err), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaves": leaves,
		"count":  len(leaves),
		"start":  startSeq,
		"end":    endSeq,
	})
}

// Get inclusion proof
func (s *LogNodeServer) handleGetInclusionProof(w http.ResponseWriter, r *http.Request) {
	hashParam := r.URL.Query().Get("hash")
	treeSizeParam := r.URL.Query().Get("tree_size")
	
	if hashParam == "" {
		http.Error(w, "Missing hash parameter", http.StatusBadRequest)
		return
	}
	
	leafHash, err := hex.DecodeString(hashParam)
	if err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}
	
	var treeSize int64
	if treeSizeParam != "" {
		if parsed, err := strconv.ParseInt(treeSizeParam, 10, 64); err == nil {
			treeSize = parsed
		}
	}
	
	proof, err := s.transparencyLog.GetInclusionProof(r.Context(), leafHash, treeSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get inclusion proof: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proof)
}

// Get consistency proof
func (s *LogNodeServer) handleGetConsistencyProof(w http.ResponseWriter, r *http.Request) {
	fromSizeParam := r.URL.Query().Get("from_size")
	toSizeParam := r.URL.Query().Get("to_size")
	
	if fromSizeParam == "" || toSizeParam == "" {
		http.Error(w, "Missing from_size or to_size parameter", http.StatusBadRequest)
		return
	}
	
	fromSize, err := strconv.ParseInt(fromSizeParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid from_size parameter", http.StatusBadRequest)
		return
	}
	
	toSize, err := strconv.ParseInt(toSizeParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid to_size parameter", http.StatusBadRequest)
		return
	}
	
	proof, err := s.transparencyLog.GetConsistencyProof(r.Context(), fromSize, toSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get consistency proof: %v", err), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proof)
}

// Get signed tree head
func (s *LogNodeServer) handleGetSignedTreeHead(w http.ResponseWriter, r *http.Request) {
	sth, err := s.transparencyLog.GetSignedTreeHead(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get signed tree head: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sth)
}

// Get tree size
func (s *LogNodeServer) handleGetTreeSize(w http.ResponseWriter, r *http.Request) {
	size, err := s.transparencyLog.GetTreeSize(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tree size: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"tree_size": size,
	})
}

// Helper endpoint to append event references
func (s *LogNodeServer) handleAppendEventReferences(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EventReferences []logpkg.EventReference `json:"event_references"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	if len(request.EventReferences) == 0 {
		http.Error(w, "No event references provided", http.StatusBadRequest)
		return
	}
	
	// Convert event references to leaves
	leaves := make([]logpkg.Leaf, len(request.EventReferences))
	for i, eventRef := range request.EventReferences {
		// Serialize event reference as leaf value
		leafValue, err := json.Marshal(eventRef)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to serialize event reference: %v", err), http.StatusInternalServerError)
			return
		}
		
		leaves[i] = logpkg.Leaf{
			LeafValue: leafValue,
			// Hash will be computed by the transparency log
		}
	}
	
	// Append to log
	result, err := s.transparencyLog.AppendLeaves(r.Context(), leaves)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to append event references: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"appended_count": len(request.EventReferences),
		"tree_size":      result.TreeSize,
		"leaf_indexes":   result.LeafIndexes,
		"root_hash":      hex.EncodeToString(result.RootHash),
	})
}

// Utility function to get tree size without error handling
func (s *LogNodeServer) getTreeSizeUnsafe() int64 {
	size, _ := s.transparencyLog.GetTreeSize(context.Background())
	return size
}

// Middleware functions
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}