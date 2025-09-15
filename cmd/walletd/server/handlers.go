package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ParichayaHQ/credence/internal/wallet"
)

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "walletd",
		"version":   "1.0.0",
		"timestamp": r.Context().Value("timestamp"),
	}
	
	s.writeResponse(w, http.StatusOK, health, nil)
}

// Key Management Handlers

type GenerateKeyRequest struct {
	KeyType string `json:"keyType"`
}

func (s *Server) handleGenerateKey(w http.ResponseWriter, r *http.Request) {
	var req GenerateKeyRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.KeyType == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("keyType is required"))
		return
	}

	key, err := s.walletService.GenerateKey(req.KeyType)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusCreated, key, nil)
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.walletService.ListKeys()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, keys, nil)
}

func (s *Server) handleGetKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["keyId"]

	if keyID == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("keyId is required"))
		return
	}

	key, err := s.walletService.GetKey(keyID)
	if err != nil {
		if err == wallet.ErrKeyNotFound {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.writeResponse(w, http.StatusOK, key, nil)
}

func (s *Server) handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["keyId"]

	if keyID == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("keyId is required"))
		return
	}

	err := s.walletService.DeleteKey(keyID)
	if err != nil {
		if err == wallet.ErrKeyNotFound {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.writeResponse(w, http.StatusOK, map[string]string{"message": "Key deleted successfully"}, nil)
}

// DID Management Handlers

type CreateDIDRequest struct {
	KeyID  string `json:"keyId"`
	Method string `json:"method"`
}

func (s *Server) handleCreateDID(w http.ResponseWriter, r *http.Request) {
	var req CreateDIDRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.KeyID == "" || req.Method == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("keyId and method are required"))
		return
	}

	did, err := s.walletService.CreateDID(req.KeyID, req.Method)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusCreated, did, nil)
}

func (s *Server) handleListDIDs(w http.ResponseWriter, r *http.Request) {
	dids, err := s.walletService.ListDIDs()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, dids, nil)
}

func (s *Server) handleGetDID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	didStr := vars["did"]

	if didStr == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("did is required"))
		return
	}

	did, err := s.walletService.GetDID(didStr)
	if err != nil {
		if err == wallet.ErrDIDNotFound {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.writeResponse(w, http.StatusOK, did, nil)
}

type ResolveDIDRequest struct {
	DID string `json:"did"`
}

func (s *Server) handleResolveDID(w http.ResponseWriter, r *http.Request) {
	var req ResolveDIDRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.DID == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("did is required"))
		return
	}

	didDocument, err := s.walletService.ResolveDID(req.DID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, didDocument, nil)
}

// Credential Management Handlers

type StoreCredentialRequest struct {
	Credential interface{}            `json:"credential"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (s *Server) handleListCredentials(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Parse query parameters for filtering
		query := r.URL.Query()
		filter := make(map[string]interface{})
		
		for key, values := range query {
			if len(values) > 0 {
				filter[key] = values[0]
			}
		}

		credentials, err := s.walletService.ListCredentials(filter)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}

		s.writeResponse(w, http.StatusOK, credentials, nil)

	case "POST":
		var req StoreCredentialRequest
		if err := s.parseJSON(r, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, err)
			return
		}

		if req.Credential == nil {
			s.writeError(w, http.StatusBadRequest, fmt.Errorf("credential is required"))
			return
		}

		credentialID, err := s.walletService.StoreCredential(req.Credential, req.Metadata)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}

		response := map[string]string{"credentialId": credentialID}
		s.writeResponse(w, http.StatusCreated, response, nil)
	}
}

func (s *Server) handleGetCredential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	credentialID := vars["credentialId"]

	if credentialID == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("credentialId is required"))
		return
	}

	credential, err := s.walletService.GetCredential(credentialID)
	if err != nil {
		if err == wallet.ErrCredentialNotFound {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.writeResponse(w, http.StatusOK, credential, nil)
}

func (s *Server) handleDeleteCredential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	credentialID := vars["credentialId"]

	if credentialID == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("credentialId is required"))
		return
	}

	err := s.walletService.DeleteCredential(credentialID)
	if err != nil {
		if err == wallet.ErrCredentialNotFound {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.writeResponse(w, http.StatusOK, map[string]string{"message": "Credential deleted successfully"}, nil)
}

// Event Management Handlers

type CreateEventRequest struct {
	Type       string `json:"type"`
	From       string `json:"from"`
	To         string `json:"to"`
	Context    string `json:"context"`
	PayloadCID string `json:"payloadCID,omitempty"`
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Parse query parameters for filtering
		query := r.URL.Query()
		filter := make(map[string]interface{})
		
		for key, values := range query {
			if len(values) > 0 {
				filter[key] = values[0]
			}
		}

		events, err := s.walletService.ListEvents(filter)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}

		s.writeResponse(w, http.StatusOK, events, nil)

	case "POST":
		var req CreateEventRequest
		if err := s.parseJSON(r, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, err)
			return
		}

		if req.Type == "" || req.From == "" || req.To == "" || req.Context == "" {
			s.writeError(w, http.StatusBadRequest, fmt.Errorf("type, from, to, and context are required"))
			return
		}

		event, err := s.walletService.CreateEvent(req.Type, req.From, req.To, req.Context, req.PayloadCID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}

		s.writeResponse(w, http.StatusCreated, event, nil)
	}
}

func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["eventId"]

	if eventID == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("eventId is required"))
		return
	}

	event, err := s.walletService.GetEvent(eventID)
	if err != nil {
		if err == wallet.ErrEventNotFound {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.writeResponse(w, http.StatusOK, event, nil)
}

// Trust Score Handlers

func (s *Server) handleListTrustScores(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	did := query.Get("did")

	scores, err := s.walletService.ListTrustScores(did)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, scores, nil)
}

func (s *Server) handleGetTrustScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	didStr := vars["did"]
	context := r.URL.Query().Get("context")

	if didStr == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("did is required"))
		return
	}

	score, err := s.walletService.GetTrustScore(didStr, context)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, score, nil)
}

// Wallet Operation Handlers

type LockWalletRequest struct {
	Password string `json:"password"`
}

type UnlockWalletRequest struct {
	Password string `json:"password"`
}

func (s *Server) handleLockWallet(w http.ResponseWriter, r *http.Request) {
	var req LockWalletRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Password == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("password is required"))
		return
	}

	err := s.walletService.Lock(req.Password)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, map[string]string{"message": "Wallet locked successfully"}, nil)
}

func (s *Server) handleUnlockWallet(w http.ResponseWriter, r *http.Request) {
	var req UnlockWalletRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Password == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("password is required"))
		return
	}

	err := s.walletService.Unlock(req.Password)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, map[string]string{"message": "Wallet unlocked successfully"}, nil)
}

func (s *Server) handleWalletStatus(w http.ResponseWriter, r *http.Request) {
	status := s.walletService.GetStatus()
	s.writeResponse(w, http.StatusOK, status, nil)
}

// Presentation Definition Handlers

type EvaluatePresentationDefinitionRequest struct {
	Definition    interface{} `json:"definition"`
	CredentialIDs []string    `json:"credentialIds,omitempty"`
}

func (s *Server) handleEvaluatePresentationDefinition(w http.ResponseWriter, r *http.Request) {
	var req EvaluatePresentationDefinitionRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Definition == nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("definition is required"))
		return
	}

	result, err := s.walletService.EvaluatePresentationDefinition(req.Definition, req.CredentialIDs)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, result, nil)
}

type CreatePresentationSubmissionRequest struct {
	Definition           interface{} `json:"definition"`
	MatchedCredentialIDs []string    `json:"matchedCredentialIds"`
}

func (s *Server) handleCreatePresentationSubmission(w http.ResponseWriter, r *http.Request) {
	var req CreatePresentationSubmissionRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Definition == nil || len(req.MatchedCredentialIDs) == 0 {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("definition and matchedCredentialIds are required"))
		return
	}

	submission, err := s.walletService.CreatePresentationSubmission(req.Definition, req.MatchedCredentialIDs)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusCreated, submission, nil)
}