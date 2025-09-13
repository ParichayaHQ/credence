package main

import (
	"context"
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
	"github.com/ParichayaHQ/credence/internal/consensus"
)

// CheckpointorServer wraps the checkpointor service with HTTP API
type CheckpointorServer struct {
	checkpointor consensus.Checkpointor
	config       *ServerConfig
	server       *http.Server
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
		Port:         8083,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func main() {
	// Load configuration
	checkpointorConfig := consensus.DefaultCheckpointorConfig()
	serverConfig := DefaultServerConfig()
	
	// Override with environment variables if present
	if addr := os.Getenv("CHECKPOINTOR_ADDRESS"); addr != "" {
		serverConfig.Address = addr
	}
	if portStr := os.Getenv("CHECKPOINTOR_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			serverConfig.Port = port
		}
	}
	
	// TODO: Initialize dependencies (BLS aggregator, VRF provider, log client, etc.)
	// For now, create with stub implementations
	store := consensus.NewMemoryCheckpointStore()
	logClient := &StubLogNodeClient{}
	publisher := &StubP2PPublisher{}
	blsAgg := &StubBLSAggregator{}
	vrfProvider := &StubVRFProvider{}
	
	// Create checkpointor service
	checkpointor := consensus.NewCheckpointor(
		checkpointorConfig,
		store,
		logClient,
		publisher,
		blsAgg,
		vrfProvider,
	)
	
	// Create server
	server := &CheckpointorServer{
		checkpointor: checkpointor,
		config:       serverConfig,
	}
	
	// Start checkpointor service
	ctx := context.Background()
	if err := checkpointor.Start(ctx); err != nil {
		log.Fatalf("Failed to start checkpointor: %v", err)
	}
	
	// Start HTTP server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down checkpointor...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Stop checkpointor service
	if err := checkpointor.Stop(shutdownCtx); err != nil {
		log.Printf("Checkpointor shutdown error: %v", err)
	}
	
	// Stop HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Checkpointor stopped")
}

// Start starts the HTTP server
func (s *CheckpointorServer) Start() error {
	router := s.setupRoutes()
	
	addr := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	
	log.Printf("Starting checkpointor server on %s", addr)
	
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	
	return nil
}

// Shutdown gracefully shuts down the server
func (s *CheckpointorServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// setupRoutes configures HTTP routes
func (s *CheckpointorServer) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Health check
	r.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// API v1 routes
	v1 := r.PathPrefix("/v1").Subrouter()
	
	// Checkpoint operations
	v1.HandleFunc("/checkpoints/latest", s.handleGetLatestCheckpoint).Methods("GET")
	v1.HandleFunc("/checkpoints/{epoch:[0-9]+}", s.handleGetCheckpoint).Methods("GET")
	v1.HandleFunc("/checkpoints/force", s.handleForceCheckpoint).Methods("POST")
	
	// Committee operations
	v1.HandleFunc("/partials", s.handleSubmitPartialSignature).Methods("POST")
	v1.HandleFunc("/tasks/current", s.handleGetCurrentTask).Methods("GET")
	
	// Add middleware
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)
	
	return r
}

// Health check endpoint
func (s *CheckpointorServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "checkpointor",
	})
}

// Get latest checkpoint
func (s *CheckpointorServer) handleGetLatestCheckpoint(w http.ResponseWriter, r *http.Request) {
	checkpoint, err := s.checkpointor.GetLatestCheckpoint(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get latest checkpoint: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkpoint)
}

// Get checkpoint by epoch
func (s *CheckpointorServer) handleGetCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epochStr := vars["epoch"]
	
	epoch, err := strconv.ParseInt(epochStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid epoch parameter", http.StatusBadRequest)
		return
	}
	
	checkpoint, err := s.checkpointor.GetCheckpoint(r.Context(), epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get checkpoint: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkpoint)
}

// Force checkpoint creation (admin/testing)
func (s *CheckpointorServer) handleForceCheckpoint(w http.ResponseWriter, r *http.Request) {
	checkpoint, err := s.checkpointor.ForceCheckpoint(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to force checkpoint: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(checkpoint)
}

// Submit partial signature (committee members)
func (s *CheckpointorServer) handleSubmitPartialSignature(w http.ResponseWriter, r *http.Request) {
	var partial consensus.PartialSignature
	
	if err := json.NewDecoder(r.Body).Decode(&partial); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	if err := s.checkpointor.SubmitPartialSignature(r.Context(), &partial); err != nil {
		http.Error(w, fmt.Sprintf("Failed to submit partial signature: %v", err), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
	})
}

// Get current signing task (committee members)
func (s *CheckpointorServer) handleGetCurrentTask(w http.ResponseWriter, r *http.Request) {
	task, err := s.checkpointor.GetCurrentTask(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get current task: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
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

// Stub implementations (to be replaced with real implementations)

type StubLogNodeClient struct{}

func (c *StubLogNodeClient) GetLatestSTH(ctx context.Context) (*consensus.SignedTreeHead, error) {
	return &consensus.SignedTreeHead{
		TreeSize:    100,
		RootHash:    []byte("stub-root-hash"),
		Timestamp:   time.Now(),
		Signature:   []byte("stub-signature"),
		SignerKeyID: "stub-key-id",
	}, nil
}

func (c *StubLogNodeClient) GetSTH(ctx context.Context, treeSize int64) (*consensus.SignedTreeHead, error) {
	return c.GetLatestSTH(ctx)
}

type StubP2PPublisher struct{}

func (p *StubP2PPublisher) PublishCheckpoint(ctx context.Context, checkpoint *consensus.Checkpoint) error {
	log.Printf("Published checkpoint epoch %d", checkpoint.Epoch)
	return nil
}

func (p *StubP2PPublisher) SubscribeCheckpoints(ctx context.Context) (<-chan *consensus.Checkpoint, error) {
	ch := make(chan *consensus.Checkpoint)
	return ch, nil
}

type StubBLSAggregator struct{}

func (a *StubBLSAggregator) VerifyPartialSignature(partial *consensus.PartialSignature, memberPublicKey []byte) error {
	return nil // Always valid for stub
}

func (a *StubBLSAggregator) AggregateSignatures(partials []consensus.PartialSignature, threshold int) ([]byte, error) {
	return []byte("stub-aggregated-signature"), nil
}

func (a *StubBLSAggregator) VerifyAggregatedSignature(checkpoint *consensus.Checkpoint, committeePublicKeys [][]byte) error {
	return nil // Always valid for stub
}

type StubVRFProvider struct{}

func (v *StubVRFProvider) GenerateProof(privateKey []byte, seed []byte) ([]byte, []byte, error) {
	return []byte("stub-proof"), []byte("stub-output"), nil
}

func (v *StubVRFProvider) VerifyProof(publicKey []byte, seed []byte, proof []byte, output []byte) error {
	return nil // Always valid for stub
}

func (v *StubVRFProvider) GetCurrentSeed(ctx context.Context) ([]byte, error) {
	return []byte("stub-vrf-seed"), nil
}