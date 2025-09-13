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
	"github.com/ParichayaHQ/credence/internal/events"
	"github.com/ParichayaHQ/credence/internal/store"
)

// FullNodeServer wraps the full node storage with HTTP API
type FullNodeServer struct {
	fullNode *store.FullNode
	config   *ServerConfig
	server   *http.Server
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address     string        `json:"address"`
	Port        int           `json:"port"`
	ReadTimeout time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:      "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func main() {
	// Load configuration
	storeConfig := store.DefaultConfig()
	serverConfig := DefaultServerConfig()
	
	// Override with environment variables if present
	if addr := os.Getenv("FULLNODE_ADDRESS"); addr != "" {
		serverConfig.Address = addr
	}
	if portStr := os.Getenv("FULLNODE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			serverConfig.Port = port
		}
	}
	if dbPath := os.Getenv("ROCKSDB_PATH"); dbPath != "" {
		storeConfig.RocksDB.Path = dbPath
	}
	if blobPath := os.Getenv("BLOB_PATH"); blobPath != "" {
		storeConfig.BlobStore.FSPath = blobPath
	}
	
	// Initialize full node
	fullNode, err := store.NewFullNode(storeConfig)
	if err != nil {
		log.Fatalf("Failed to initialize full node: %v", err)
	}
	defer fullNode.Close()
	
	// Create server
	server := &FullNodeServer{
		fullNode: fullNode,
		config:   serverConfig,
	}
	
	// Start HTTP server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Server stopped")
}

// Start starts the HTTP server
func (s *FullNodeServer) Start() error {
	router := s.setupRoutes()
	
	addr := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	
	log.Printf("Starting full node server on %s", addr)
	
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	
	return nil
}

// Shutdown gracefully shuts down the server
func (s *FullNodeServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// setupRoutes configures HTTP routes
func (s *FullNodeServer) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// API v1 routes
	v1 := r.PathPrefix("/v1").Subrouter()
	
	// Health check
	r.HandleFunc("/health", s.handleHealth).Methods("GET")
	r.HandleFunc("/stats", s.handleStats).Methods("GET")
	
	// Blob storage endpoints
	v1.HandleFunc("/blobs/{cid}", s.handleGetBlob).Methods("GET")
	v1.HandleFunc("/blobs", s.handleStoreBlob).Methods("POST")
	v1.HandleFunc("/blobs/{cid}", s.handleDeleteBlob).Methods("DELETE")
	v1.HandleFunc("/blobs/{cid}/exists", s.handleBlobExists).Methods("GET")
	
	// Event storage endpoints
	v1.HandleFunc("/events", s.handleStoreEvent).Methods("POST")
	v1.HandleFunc("/events/{cid}", s.handleGetEvent).Methods("GET")
	v1.HandleFunc("/events/by-did/{did}", s.handleGetEventsByDID).Methods("GET")
	v1.HandleFunc("/events/by-type/{type}", s.handleGetEventsByType).Methods("GET")
	
	// Checkpoint endpoints
	v1.HandleFunc("/checkpoints", s.handleStoreCheckpoint).Methods("POST")
	v1.HandleFunc("/checkpoints/{epoch}", s.handleGetCheckpoint).Methods("GET")
	v1.HandleFunc("/checkpoints/latest", s.handleGetLatestCheckpoint).Methods("GET")
	v1.HandleFunc("/checkpoints", s.handleListCheckpoints).Methods("GET")
	
	// Status list endpoints  
	v1.HandleFunc("/status/{issuer}/{epoch}", s.handleStoreStatusList).Methods("POST")
	v1.HandleFunc("/status/{issuer}/{epoch}", s.handleGetStatusList).Methods("GET")
	
	// Combined operations
	v1.HandleFunc("/events/with-blob", s.handleStoreEventWithBlob).Methods("POST")
	v1.HandleFunc("/events/{cid}/with-blob", s.handleGetEventWithBlob).Methods("GET")
	
	// Add middleware
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)
	
	return r
}

// Health check endpoint
func (s *FullNodeServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "fullnode",
	})
}

// Statistics endpoint
func (s *FullNodeServer) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.fullNode.GetStorageStats()
	json.NewEncoder(w).Encode(stats)
}

// Blob storage handlers
func (s *FullNodeServer) handleStoreBlob(w http.ResponseWriter, r *http.Request) {
	data, err := readRequestBody(w, r, 16*1024*1024) // 16MB limit
	if err != nil {
		return
	}
	
	cid, err := s.fullNode.Store(r.Context(), data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store blob: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"cid": cid,
	})
}

func (s *FullNodeServer) handleGetBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]
	
	data, err := s.fullNode.Get(r.Context(), cid)
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "Blob not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get blob: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (s *FullNodeServer) handleDeleteBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]
	
	err := s.fullNode.Delete(r.Context(), cid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete blob: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func (s *FullNodeServer) handleBlobExists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]
	
	exists, err := s.fullNode.Has(r.Context(), cid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check blob existence: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"exists": exists,
	})
}

// Event storage handlers
func (s *FullNodeServer) handleStoreEvent(w http.ResponseWriter, r *http.Request) {
	var event events.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	if err := s.fullNode.StoreEvent(r.Context(), &event); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store event: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Generate CID for response  
	eventData, _ := json.Marshal(event)
	cid, _ := events.GenerateCIDFromJSON(eventData)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"cid": cid,
	})
}

func (s *FullNodeServer) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]
	
	event, err := s.fullNode.GetEvent(r.Context(), cid)
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get event: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (s *FullNodeServer) handleGetEventsByDID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars["did"]
	
	// Parse query parameters
	direction := store.DirectionBoth // default
	if dirStr := r.URL.Query().Get("direction"); dirStr != "" {
		switch dirStr {
		case "from":
			direction = store.DirectionFrom
		case "to":
			direction = store.DirectionTo
		case "both":
			direction = store.DirectionBoth
		}
	}
	
	fromEpoch := r.URL.Query().Get("from_epoch")
	toEpoch := r.URL.Query().Get("to_epoch")
	
	events, err := s.fullNode.GetEventsByDID(r.Context(), did, direction, fromEpoch, toEpoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get events: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

func (s *FullNodeServer) handleGetEventsByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventType := vars["type"]
	
	fromEpoch := r.URL.Query().Get("from_epoch")
	toEpoch := r.URL.Query().Get("to_epoch")
	
	events, err := s.fullNode.GetEventsByType(r.Context(), eventType, fromEpoch, toEpoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get events: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// Checkpoint handlers
func (s *FullNodeServer) handleStoreCheckpoint(w http.ResponseWriter, r *http.Request) {
	var checkpoint store.Checkpoint
	if err := json.NewDecoder(r.Body).Decode(&checkpoint); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	if err := s.fullNode.StoreCheckpoint(r.Context(), &checkpoint); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store checkpoint: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
}

func (s *FullNodeServer) handleGetCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	epochStr := vars["epoch"]
	
	epoch, err := strconv.ParseInt(epochStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid epoch", http.StatusBadRequest)
		return
	}
	
	checkpoint, err := s.fullNode.GetCheckpoint(r.Context(), epoch)
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "Checkpoint not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get checkpoint: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkpoint)
}

func (s *FullNodeServer) handleGetLatestCheckpoint(w http.ResponseWriter, r *http.Request) {
	checkpoint, err := s.fullNode.GetLatestCheckpoint(r.Context())
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "No checkpoints found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get latest checkpoint: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkpoint)
}

func (s *FullNodeServer) handleListCheckpoints(w http.ResponseWriter, r *http.Request) {
	fromEpoch, _ := strconv.ParseInt(r.URL.Query().Get("from_epoch"), 10, 64)
	toEpoch, _ := strconv.ParseInt(r.URL.Query().Get("to_epoch"), 10, 64)
	
	if toEpoch == 0 {
		toEpoch = time.Now().Unix() // Default to current time
	}
	
	checkpoints, err := s.fullNode.ListCheckpoints(r.Context(), fromEpoch, toEpoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list checkpoints: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"checkpoints": checkpoints,
		"count":       len(checkpoints),
	})
}

// Status list handlers
func (s *FullNodeServer) handleStoreStatusList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	issuer := vars["issuer"]
	epoch := vars["epoch"]
	
	var payload struct {
		BitmapCID string `json:"bitmap_cid"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	if err := s.fullNode.StoreStatusList(r.Context(), issuer, epoch, payload.BitmapCID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store status list: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
}

func (s *FullNodeServer) handleGetStatusList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	issuer := vars["issuer"]
	epoch := vars["epoch"]
	
	bitmapCID, err := s.fullNode.GetStatusList(r.Context(), issuer, epoch)
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "Status list not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get status list: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"bitmap_cid": bitmapCID,
	})
}

// Combined operation handlers
func (s *FullNodeServer) handleStoreEventWithBlob(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form for event + blob
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}
	
	// Get event JSON
	eventJSON := r.FormValue("event")
	if eventJSON == "" {
		http.Error(w, "Missing event data", http.StatusBadRequest)
		return
	}
	
	var event events.Event
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid event JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	// Get blob data if present
	var blobData []byte
	if file, _, err := r.FormFile("blob"); err == nil {
		defer file.Close()
		blobData, err = readFromReader(file, 16*1024*1024) // 16MB limit
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read blob: %v", err), http.StatusBadRequest)
			return
		}
	}
	
	// Store both atomically
	blobCID, err := s.fullNode.StoreEventAndBlob(r.Context(), &event, blobData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store event and blob: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Generate event CID for response
	eventData, _ := json.Marshal(event)
	eventCID, _ := events.GenerateCIDFromJSON(eventData)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"event_cid": eventCID,
		"blob_cid":  blobCID,
	})
}

func (s *FullNodeServer) handleGetEventWithBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]
	
	event, blobData, err := s.fullNode.GetEventWithBlob(r.Context(), cid)
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get event with blob: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"event": event,
	}
	
	if blobData != nil {
		response["blob_size"] = len(blobData)
		// For JSON response, we could base64 encode small blobs
		// For large blobs, return a reference URL
		if len(blobData) < 1024*1024 { // 1MB threshold
			response["blob_data"] = blobData
		} else {
			response["blob_url"] = fmt.Sprintf("/v1/blobs/%s", event.PayloadCID)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// Utility functions
func readRequestBody(w http.ResponseWriter, r *http.Request, maxSize int64) ([]byte, error) {
	if r.ContentLength > maxSize {
		http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		return nil, fmt.Errorf("request too large")
	}
	
	limitedReader := http.MaxBytesReader(w, r.Body, maxSize)
	return readFromReader(limitedReader, maxSize)
}

func readFromReader(r interface{ Read([]byte) (int, error) }, maxSize int64) ([]byte, error) {
	data := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	
	for {
		n, err := r.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
			if int64(len(data)) > maxSize {
				return nil, fmt.Errorf("data too large")
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
	}
	
	return data, nil
}