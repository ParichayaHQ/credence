package score

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// HTTPService provides HTTP endpoints for the trust scoring engine
type HTTPService struct {
	engine        Engine
	budgetManager BudgetManager
	graphAnalyzer GraphAnalyzer
	enforcer      *BudgetEnforcer
	config        *ScoreConfig
	server        *http.Server
}

// NewHTTPService creates a new HTTP service for trust scoring
func NewHTTPService(
	engine Engine,
	budgetManager BudgetManager,
	graphAnalyzer GraphAnalyzer,
	config *ScoreConfig,
	port int,
) *HTTPService {
	service := &HTTPService{
		engine:        engine,
		budgetManager: budgetManager,
		graphAnalyzer: graphAnalyzer,
		config:        config,
	}
	
	if budgetManager != nil {
		service.enforcer = NewBudgetEnforcer(budgetManager, config)
	}
	
	router := service.setupRoutes()
	
	// Add CORS support
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	
	service.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      c.Handler(router),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return service
}

// setupRoutes configures HTTP routes
func (s *HTTPService) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// API prefix
	api := r.PathPrefix("/api/v1").Subrouter()
	
	// Score computation endpoints
	api.HandleFunc("/score/{did}", s.handleGetScore).Methods("GET")
	api.HandleFunc("/score/{did}/recompute", s.handleRecomputeScore).Methods("POST")
	api.HandleFunc("/scores/batch", s.handleBatchScore).Methods("POST")
	api.HandleFunc("/score/{did}/factors", s.handleGetFactors).Methods("GET")
	api.HandleFunc("/score/{did}/proof", s.handleGetProof).Methods("GET")
	api.HandleFunc("/proof/verify", s.handleVerifyProof).Methods("POST")
	
	// Budget management endpoints
	api.HandleFunc("/budget/{did}", s.handleGetBudget).Methods("GET")
	api.HandleFunc("/budget/{did}/spend", s.handleSpendBudget).Methods("POST")
	api.HandleFunc("/budget/{did}/refill", s.handleRefillBudget).Methods("POST")
	api.HandleFunc("/budget/{did}/utilization", s.handleGetBudgetUtilization).Methods("GET")
	
	// Graph analysis endpoints
	api.HandleFunc("/analysis/collusion", s.handleDetectCollusion).Methods("GET")
	api.HandleFunc("/analysis/diversity/{did}", s.handleGetDiversity).Methods("GET")
	api.HandleFunc("/analysis/dense-subgraphs", s.handleGetDenseSubgraphs).Methods("GET")
	
	// Configuration endpoints
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET")
	api.HandleFunc("/config", s.handleUpdateConfig).Methods("PUT")
	
	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	return r
}

// Start starts the HTTP service
func (s *HTTPService) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the HTTP service
func (s *HTTPService) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleGetScore handles GET /api/v1/score/{did}
func (s *HTTPService) handleGetScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400 // Default to current epoch (day)
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	includeProof := r.URL.Query().Get("include_proof") == "true"
	includeFactors := r.URL.Query().Get("include_factors") == "true"
	
	score, err := s.engine.ComputeScore(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to compute score: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := &ComputeResponse{
		Score: score,
	}
	
	if includeFactors {
		factors, err := s.engine.GetFactors(r.Context(), did, context, epoch)
		if err == nil {
			response.Components = factors
		}
	}
	
	if includeProof {
		proof, err := s.engine.GetProof(r.Context(), score)
		if err == nil {
			response.Proof = proof
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRecomputeScore handles POST /api/v1/score/{did}/recompute
func (s *HTTPService) handleRecomputeScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	score, err := s.engine.RecomputeScore(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to recompute score: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := &ComputeResponse{
		Score: score,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleBatchScore handles POST /api/v1/scores/batch
func (s *HTTPService) handleBatchScore(w http.ResponseWriter, r *http.Request) {
	var request BatchComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	response := &BatchComputeResponse{
		Responses: make([]*ComputeResponse, len(request.Requests)),
		Errors:    make([]string, 0),
	}
	
	for i, req := range request.Requests {
		score, err := s.engine.ComputeScore(r.Context(), req.DID, req.Context, req.Epoch)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("DID %s: %v", req.DID, err))
			continue
		}
		
		computeResp := &ComputeResponse{
			Score: score,
		}
		
		if req.IncludeFactors {
			factors, err := s.engine.GetFactors(r.Context(), req.DID, req.Context, req.Epoch)
			if err == nil {
				computeResp.Components = factors
			}
		}
		
		if req.IncludeProof {
			proof, err := s.engine.GetProof(r.Context(), score)
			if err == nil {
				computeResp.Proof = proof
			}
		}
		
		response.Responses[i] = computeResp
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetFactors handles GET /api/v1/score/{did}/factors
func (s *HTTPService) handleGetFactors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	factors, err := s.engine.GetFactors(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get factors: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(factors)
}

// handleGetProof handles GET /api/v1/score/{did}/proof
func (s *HTTPService) handleGetProof(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	score, err := s.engine.ComputeScore(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to compute score: %v", err), http.StatusInternalServerError)
		return
	}
	
	proof, err := s.engine.GetProof(r.Context(), score)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate proof: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proof)
}

// handleVerifyProof handles POST /api/v1/proof/verify
func (s *HTTPService) handleVerifyProof(w http.ResponseWriter, r *http.Request) {
	var proof ScoreProof
	if err := json.NewDecoder(r.Body).Decode(&proof); err != nil {
		http.Error(w, fmt.Sprintf("Invalid proof format: %v", err), http.StatusBadRequest)
		return
	}
	
	err := s.engine.VerifyProof(r.Context(), &proof)
	result := map[string]interface{}{
		"valid": err == nil,
	}
	
	if err != nil {
		result["error"] = err.Error()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleGetBudget handles GET /api/v1/budget/{did}
func (s *HTTPService) handleGetBudget(w http.ResponseWriter, r *http.Request) {
	if s.budgetManager == nil {
		http.Error(w, "Budget manager not configured", http.StatusServiceUnavailable)
		return
	}
	
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	budget, err := s.budgetManager.GetBudget(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get budget: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

// handleSpendBudget handles POST /api/v1/budget/{did}/spend
func (s *HTTPService) handleSpendBudget(w http.ResponseWriter, r *http.Request) {
	if s.budgetManager == nil {
		http.Error(w, "Budget manager not configured", http.StatusServiceUnavailable)
		return
	}
	
	vars := mux.Vars(r)
	did := vars["did"]
	
	var request struct {
		Context string  `json:"context"`
		Epoch   int64   `json:"epoch"`
		Amount  float64 `json:"amount"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	if request.Context == "" {
		request.Context = "default"
	}
	
	if request.Epoch == 0 {
		request.Epoch = time.Now().Unix() / 86400
	}
	
	err := s.budgetManager.SpendBudget(r.Context(), did, request.Context, request.Epoch, request.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to spend budget: %v", err), http.StatusBadRequest)
		return
	}
	
	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully spent %.2f from budget", request.Amount),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleRefillBudget handles POST /api/v1/budget/{did}/refill
func (s *HTTPService) handleRefillBudget(w http.ResponseWriter, r *http.Request) {
	if s.budgetManager == nil {
		http.Error(w, "Budget manager not configured", http.StatusServiceUnavailable)
		return
	}
	
	vars := mux.Vars(r)
	did := vars["did"]
	
	var request struct {
		Context string  `json:"context"`
		Epoch   int64   `json:"epoch"`
		Score   float64 `json:"score"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	if request.Context == "" {
		request.Context = "default"
	}
	
	if request.Epoch == 0 {
		request.Epoch = time.Now().Unix() / 86400
	}
	
	err := s.budgetManager.RefillBudget(r.Context(), did, request.Context, request.Epoch, request.Score)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to refill budget: %v", err), http.StatusInternalServerError)
		return
	}
	
	result := map[string]interface{}{
		"success": true,
		"message": "Budget refilled successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleGetBudgetUtilization handles GET /api/v1/budget/{did}/utilization
func (s *HTTPService) handleGetBudgetUtilization(w http.ResponseWriter, r *http.Request) {
	if s.enforcer == nil {
		http.Error(w, "Budget enforcer not configured", http.StatusServiceUnavailable)
		return
	}
	
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	utilization, err := s.enforcer.GetBudgetUtilization(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get budget utilization: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utilization)
}

// handleDetectCollusion handles GET /api/v1/analysis/collusion
func (s *HTTPService) handleDetectCollusion(w http.ResponseWriter, r *http.Request) {
	if s.graphAnalyzer == nil {
		http.Error(w, "Graph analyzer not configured", http.StatusServiceUnavailable)
		return
	}
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	clusters, err := s.graphAnalyzer.DetectCollusion(r.Context(), context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to detect collusion: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// handleGetDiversity handles GET /api/v1/analysis/diversity/{did}
func (s *HTTPService) handleGetDiversity(w http.ResponseWriter, r *http.Request) {
	if s.graphAnalyzer == nil {
		http.Error(w, "Graph analyzer not configured", http.StatusServiceUnavailable)
		return
	}
	
	vars := mux.Vars(r)
	did := vars["did"]
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	diversity, err := s.graphAnalyzer.ComputeDiversity(r.Context(), did, context, epoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to compute diversity: %v", err), http.StatusInternalServerError)
		return
	}
	
	result := map[string]interface{}{
		"did":       did,
		"context":   context,
		"epoch":     epoch,
		"diversity": diversity,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleGetDenseSubgraphs handles GET /api/v1/analysis/dense-subgraphs
func (s *HTTPService) handleGetDenseSubgraphs(w http.ResponseWriter, r *http.Request) {
	if s.graphAnalyzer == nil {
		http.Error(w, "Graph analyzer not configured", http.StatusServiceUnavailable)
		return
	}
	
	context := r.URL.Query().Get("context")
	if context == "" {
		context = "default"
	}
	
	epochStr := r.URL.Query().Get("epoch")
	epoch := time.Now().Unix() / 86400
	if epochStr != "" {
		if e, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = e
		}
	}
	
	thresholdStr := r.URL.Query().Get("threshold")
	threshold := 0.7 // Default threshold
	if thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			threshold = t
		}
	}
	
	subgraphs, err := s.graphAnalyzer.GetDenseSubgraphs(r.Context(), context, epoch, threshold)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get dense subgraphs: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subgraphs)
}

// handleGetConfig handles GET /api/v1/config
func (s *HTTPService) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

// handleUpdateConfig handles PUT /api/v1/config
func (s *HTTPService) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig ScoreConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Invalid config format: %v", err), http.StatusBadRequest)
		return
	}
	
	// Validate config (basic validation)
	if newConfig.Factors.Alpha < 0 || newConfig.Factors.Beta < 0 || 
	   newConfig.Factors.Gamma < 0 || newConfig.Factors.Delta < 0 || 
	   newConfig.Factors.Tau < 0 {
		http.Error(w, "Factor weights must be non-negative", http.StatusBadRequest)
		return
	}
	
	// Update config
	s.config = &newConfig
	
	result := map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleHealth handles GET /api/v1/health
func (s *HTTPService) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"components": map[string]bool{
			"engine":         s.engine != nil,
			"budget_manager": s.budgetManager != nil,
			"graph_analyzer": s.graphAnalyzer != nil,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}