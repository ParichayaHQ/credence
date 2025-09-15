package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ParichayaHQ/credence/internal/wallet"
)

// Server represents the HTTP server for the wallet service
type Server struct {
	walletService *wallet.Service
	router        *mux.Router
}

// NewServer creates a new HTTP server instance
func NewServer(walletService *wallet.Service) *Server {
	s := &Server{
		walletService: walletService,
		router:        mux.NewRouter(),
	}
	
	s.setupRoutes()
	s.setupMiddleware()
	
	return s
}

// Router returns the configured HTTP router
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	// API version prefix
	api := s.router.PathPrefix("/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Key management
	keyRouter := api.PathPrefix("/keys").Subrouter()
	keyRouter.HandleFunc("", s.handleListKeys).Methods("GET")
	keyRouter.HandleFunc("/generate", s.handleGenerateKey).Methods("POST")
	keyRouter.HandleFunc("/{keyId}", s.handleGetKey).Methods("GET")
	keyRouter.HandleFunc("/{keyId}", s.handleDeleteKey).Methods("DELETE")

	// DID management
	didRouter := api.PathPrefix("/dids").Subrouter()
	didRouter.HandleFunc("", s.handleListDIDs).Methods("GET")
	didRouter.HandleFunc("/create", s.handleCreateDID).Methods("POST")
	didRouter.HandleFunc("/resolve", s.handleResolveDID).Methods("POST")
	didRouter.HandleFunc("/{did:.*}", s.handleGetDID).Methods("GET")

	// Credential management
	credRouter := api.PathPrefix("/credentials").Subrouter()
	credRouter.HandleFunc("", s.handleListCredentials).Methods("GET", "POST")
	credRouter.HandleFunc("/{credentialId}", s.handleGetCredential).Methods("GET")
	credRouter.HandleFunc("/{credentialId}", s.handleDeleteCredential).Methods("DELETE")

	// Event management (vouches/reports)
	eventRouter := api.PathPrefix("/events").Subrouter()
	eventRouter.HandleFunc("", s.handleListEvents).Methods("GET", "POST")
	eventRouter.HandleFunc("/{eventId}", s.handleGetEvent).Methods("GET")

	// Trust scores
	scoreRouter := api.PathPrefix("/scores").Subrouter()
	scoreRouter.HandleFunc("", s.handleListTrustScores).Methods("GET")
	scoreRouter.HandleFunc("/{did:.*}", s.handleGetTrustScore).Methods("GET")

	// Presentation definitions
	presDefRouter := api.PathPrefix("/presentation-definitions").Subrouter()
	presDefRouter.HandleFunc("/evaluate", s.handleEvaluatePresentationDefinition).Methods("POST")
	presDefRouter.HandleFunc("/submissions", s.handleCreatePresentationSubmission).Methods("POST")

	// Wallet operations
	walletRouter := api.PathPrefix("/wallet").Subrouter()
	walletRouter.HandleFunc("/lock", s.handleLockWallet).Methods("POST")
	walletRouter.HandleFunc("/unlock", s.handleUnlockWallet).Methods("POST")
	walletRouter.HandleFunc("/status", s.handleWalletStatus).Methods("GET")
}

func (s *Server) setupMiddleware() {
	// CORS middleware
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:*", "https://localhost:*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"}),
		handlers.AllowCredentials(),
	)

	// Apply CORS to all routes
	s.router.Use(corsHandler)

	// Request logging middleware
	s.router.Use(s.loggingMiddleware)

	// Error handling middleware
	s.router.Use(s.errorHandlingMiddleware)

	// Content type middleware
	s.router.Use(s.contentTypeMiddleware)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}

func (s *Server) errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// writeResponse writes a JSON response
func (s *Server) writeResponse(w http.ResponseWriter, statusCode int, data interface{}, err error) {
	resp := Response{
		Success:   err == nil,
		Data:      data,
		Timestamp: time.Now(),
	}

	if err != nil {
		resp.Error = err.Error()
	}

	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, statusCode int, err error) {
	s.writeResponse(w, statusCode, nil, err)
}

// parseJSON parses JSON request body
func (s *Server) parseJSON(r *http.Request, v interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("content-type must be application/json")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}