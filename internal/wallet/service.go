package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
)

// Service provides a high-level API for wallet operations
// This wraps the core wallet with additional HTTP service functionality
type Service struct {
	wallet Wallet
	config *Config
}

// Config for the wallet service
type Config struct {
	DataDir string
	// Add other service-specific configuration
}

// Status represents the wallet status
type Status struct {
	Locked       bool   `json:"locked"`
	KeysCount    int    `json:"keysCount"`
	DIDsCount    int    `json:"didsCount"`
	CredentialsCount int `json:"credentialsCount"`
}

// Define standard errors
var (
	ErrKeyNotFound        = NewWalletError(ErrorKeyNotFound, "key not found")
	ErrDIDNotFound        = NewWalletError(ErrorDIDNotFound, "DID not found")  
	ErrCredentialNotFound = NewWalletError(ErrorCredentialNotFound, "credential not found")
	ErrEventNotFound      = NewWalletError("event_not_found", "event not found")
)

// NewService creates a new wallet service
func NewService(config *Config) (*Service, error) {
	if config == nil {
		config = &Config{}
	}

	// Create wallet configuration
	walletConfig := DefaultWalletConfig()
	walletConfig.StorageType = "file"
	walletConfig.StoragePath = config.DataDir

	// Create storage
	storage := NewInMemoryStorage() // TODO: Use file storage when available

	// Create key manager
	keyManager := did.NewDefaultKeyManager()

	// Create wallet
	wallet, err := NewDefaultWallet(walletConfig, storage, keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &Service{
		wallet: wallet,
		config: config,
	}, nil
}

// Close closes the service and releases resources
func (s *Service) Close() error {
	// TODO: Implement proper cleanup if needed
	return nil
}

// Key Management

func (s *Service) GenerateKey(keyType string) (interface{}, error) {
	var kt did.KeyType
	switch keyType {
	case "Ed25519":
		kt = did.KeyTypeEd25519
	case "Secp256k1":
		kt = did.KeyTypeSecp256k1
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}

	return s.wallet.GenerateKey(kt)
}

func (s *Service) ListKeys() (interface{}, error) {
	return s.wallet.ListKeys()
}

func (s *Service) GetKey(keyID string) (interface{}, error) {
	return s.wallet.GetKey(keyID)
}

func (s *Service) DeleteKey(keyID string) error {
	return s.wallet.DeleteKey(keyID)
}

// DID Management

func (s *Service) CreateDID(keyID, method string) (interface{}, error) {
	return s.wallet.CreateDID(keyID, method)
}

func (s *Service) ListDIDs() (interface{}, error) {
	return s.wallet.ListDIDs()
}

func (s *Service) GetDID(did string) (interface{}, error) {
	return s.wallet.GetDID(did)
}

func (s *Service) ResolveDID(did string) (interface{}, error) {
	return s.wallet.ResolveDID(did)
}

// Credential Management

func (s *Service) StoreCredential(credential interface{}, metadata map[string]interface{}) (string, error) {
	// Convert credential to VerifiableCredential
	var vcred *vc.VerifiableCredential
	
	// Handle different input formats
	switch cred := credential.(type) {
	case *vc.VerifiableCredential:
		vcred = cred
	case map[string]interface{}:
		// Convert from map
		credBytes, err := json.Marshal(cred)
		if err != nil {
			return "", fmt.Errorf("failed to marshal credential: %w", err)
		}
		
		vcred = &vc.VerifiableCredential{}
		if err := json.Unmarshal(credBytes, vcred); err != nil {
			return "", fmt.Errorf("failed to unmarshal credential: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported credential type: %T", credential)
	}

	record, err := s.wallet.StoreCredential(vcred)
	if err != nil {
		return "", err
	}

	// Add metadata if provided
	if metadata != nil {
		record.Metadata = metadata
	}

	return record.ID, nil
}

func (s *Service) ListCredentials(filter map[string]interface{}) (interface{}, error) {
	// Convert filter map to CredentialFilter
	credFilter := &CredentialFilter{}
	
	if issuer, ok := filter["issuer"].(string); ok {
		credFilter.Issuer = issuer
	}
	if subject, ok := filter["subject"].(string); ok {
		credFilter.Subject = subject
	}
	// Add other filter fields as needed

	return s.wallet.ListCredentials(credFilter)
}

func (s *Service) GetCredential(credentialID string) (interface{}, error) {
	return s.wallet.GetCredential(credentialID)
}

func (s *Service) DeleteCredential(credentialID string) error {
	return s.wallet.DeleteCredential(credentialID)
}

// Event represents a trust event (vouch or report)
type Event struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // "vouch" or "report"
	From       string                 `json:"from"`       // DID of the issuer
	To         string                 `json:"to"`         // DID of the subject
	Context    string                 `json:"context"`    // Context for the event
	PayloadCID string                 `json:"payloadCID,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
	Signature  string                 `json:"signature,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Event Management

func (s *Service) CreateEvent(eventType, from, to, context, payloadCID string) (interface{}, error) {
	// Validate event type
	if eventType != "vouch" && eventType != "report" {
		return nil, fmt.Errorf("invalid event type: %s. Must be 'vouch' or 'report'", eventType)
	}

	// Validate DIDs
	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to DIDs are required")
	}

	if context == "" {
		return nil, fmt.Errorf("context is required")
	}

	// Create event
	event := &Event{
		ID:         fmt.Sprintf("event-%d", getCurrentTimestamp()),
		Type:       eventType,
		From:       from,
		To:         to,
		Context:    context,
		PayloadCID: payloadCID,
		Timestamp:  getCurrentTimestamp(),
	}

	// Get existing events
	events, err := s.getAllEvents()
	if err != nil {
		events = []*Event{} // Start with empty list if error
	}

	// Add new event
	events = append(events, event)

	// Store updated events list
	if err := s.storeAllEvents(events); err != nil {
		return nil, fmt.Errorf("failed to store event: %w", err)
	}

	return event, nil
}

func (s *Service) ListEvents(filter map[string]interface{}) (interface{}, error) {
	events, err := s.getAllEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var filteredEvents []*Event
	for _, event := range events {
		if s.matchesEventFilter(event, filter) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

func (s *Service) GetEvent(eventID string) (interface{}, error) {
	events, err := s.getAllEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	for _, event := range events {
		if event.ID == eventID {
			return event, nil
		}
	}

	return nil, ErrEventNotFound
}

// Helper methods for event storage
func (s *Service) getAllEvents() ([]*Event, error) {
	defaultWallet, ok := s.wallet.(*DefaultWallet)
	if !ok {
		return nil, fmt.Errorf("wallet type not supported for events")
	}

	data, err := defaultWallet.storage.GetMetadata("wallet_events")
	if err != nil {
		return []*Event{}, nil // Return empty list if no events exist
	}

	eventsJSON, ok := data.(string)
	if !ok {
		return []*Event{}, nil
	}

	var events []*Event
	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return events, nil
}

func (s *Service) storeAllEvents(events []*Event) error {
	defaultWallet, ok := s.wallet.(*DefaultWallet)
	if !ok {
		return fmt.Errorf("wallet type not supported for events")
	}

	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	return defaultWallet.storage.SetMetadata("wallet_events", string(eventsJSON))
}

// Helper function to match events against filters
func (s *Service) matchesEventFilter(event *Event, filter map[string]interface{}) bool {
	if eventType, ok := filter["type"].(string); ok && event.Type != eventType {
		return false
	}
	if from, ok := filter["from"].(string); ok && event.From != from {
		return false
	}
	if to, ok := filter["to"].(string); ok && event.To != to {
		return false
	}
	if context, ok := filter["context"].(string); ok && event.Context != context {
		return false
	}
	return true
}

// TrustScore represents a calculated trust score
type TrustScore struct {
	DID         string                 `json:"did"`
	Context     string                 `json:"context"`
	Score       float64                `json:"score"`
	VouchCount  int                    `json:"vouchCount"`
	ReportCount int                    `json:"reportCount"`
	LastUpdated int64                  `json:"lastUpdated"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Trust Score Management

func (s *Service) GetTrustScore(did, context string) (interface{}, error) {
	if did == "" {
		return nil, fmt.Errorf("DID is required")
	}

	// Calculate trust score based on events
	score, err := s.calculateTrustScore(did, context)
	if err != nil {
		return nil, err
	}

	return score, nil
}

func (s *Service) ListTrustScores(did string) (interface{}, error) {
	// Get all contexts for the DID
	contexts, err := s.getContextsForDID(did)
	if err != nil {
		return nil, err
	}

	var scores []*TrustScore
	if did != "" {
		// Get scores for specific DID across all contexts
		for _, context := range contexts {
			score, err := s.calculateTrustScore(did, context)
			if err == nil {
				scores = append(scores, score)
			}
		}
	} else {
		// Get all trust scores if no specific DID
		allDIDs, err := s.getAllDIDsWithEvents()
		if err != nil {
			return nil, err
		}

		for _, eventDID := range allDIDs {
			for _, context := range contexts {
				score, err := s.calculateTrustScore(eventDID, context)
				if err == nil {
					scores = append(scores, score)
				}
			}
		}
	}

	return scores, nil
}

// calculateTrustScore calculates the trust score for a DID in a specific context
func (s *Service) calculateTrustScore(did, context string) (*TrustScore, error) {
	// Get events for this DID
	filter := map[string]interface{}{
		"to": did,
	}
	if context != "" {
		filter["context"] = context
	}

	eventsInterface, err := s.ListEvents(filter)
	if err != nil {
		return nil, err
	}

	events, ok := eventsInterface.([]*Event)
	if !ok {
		return nil, fmt.Errorf("failed to cast events")
	}

	// Count vouches and reports
	vouchCount := 0
	reportCount := 0
	
	for _, event := range events {
		if event.Type == "vouch" {
			vouchCount++
		} else if event.Type == "report" {
			reportCount++
		}
	}

	// Calculate score using a simple algorithm
	// Score = (vouches - reports) / (vouches + reports + 1) * 100
	// Normalized to 0-100 range
	totalEvents := vouchCount + reportCount
	score := 50.0 // Neutral score
	
	if totalEvents > 0 {
		ratio := float64(vouchCount-reportCount) / float64(totalEvents+1)
		score = 50.0 + (ratio * 50.0)
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return &TrustScore{
		DID:         did,
		Context:     context,
		Score:       score,
		VouchCount:  vouchCount,
		ReportCount: reportCount,
		LastUpdated: getCurrentTimestamp(),
	}, nil
}

// getContextsForDID gets all contexts that have events for a specific DID
func (s *Service) getContextsForDID(did string) ([]string, error) {
	filter := map[string]interface{}{}
	if did != "" {
		filter["to"] = did
	}

	eventsInterface, err := s.ListEvents(filter)
	if err != nil {
		return nil, err
	}

	events, ok := eventsInterface.([]*Event)
	if !ok {
		return nil, fmt.Errorf("failed to cast events")
	}

	contextSet := make(map[string]bool)
	for _, event := range events {
		if event.Context != "" {
			contextSet[event.Context] = true
		}
	}

	contexts := make([]string, 0, len(contextSet))
	for context := range contextSet {
		contexts = append(contexts, context)
	}

	// Add empty context to include overall scores
	if len(contexts) == 0 {
		contexts = append(contexts, "")
	}

	return contexts, nil
}

// getAllDIDsWithEvents gets all DIDs that appear in events
func (s *Service) getAllDIDsWithEvents() ([]string, error) {
	eventsInterface, err := s.ListEvents(map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	events, ok := eventsInterface.([]*Event)
	if !ok {
		return nil, fmt.Errorf("failed to cast events")
	}

	didSet := make(map[string]bool)
	for _, event := range events {
		if event.To != "" {
			didSet[event.To] = true
		}
	}

	dids := make([]string, 0, len(didSet))
	for did := range didSet {
		dids = append(dids, did)
	}

	return dids, nil
}

// Presentation Definition Operations

func (s *Service) EvaluatePresentationDefinition(definition interface{}, credentialIDs []string) (interface{}, error) {
	// Parse presentation definition
	var presDefBytes []byte
	var err error

	switch def := definition.(type) {
	case string:
		presDefBytes = []byte(def)
	case map[string]interface{}:
		presDefBytes, err = json.Marshal(def)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal presentation definition: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported presentation definition type: %T", definition)
	}

	var presDef vc.PresentationDefinition
	if err := json.Unmarshal(presDefBytes, &presDef); err != nil {
		return nil, fmt.Errorf("failed to parse presentation definition: %w", err)
	}

	// Get credentials from wallet
	var credentials []*vc.VerifiableCredential
	if len(credentialIDs) > 0 {
		// Use specific credentials
		for _, credID := range credentialIDs {
			record, err := s.wallet.GetCredential(credID)
			if err != nil {
				continue // Skip missing credentials
			}
			credentials = append(credentials, record.Credential)
		}
	} else {
		// Use all credentials
		records, err := s.wallet.ListCredentials(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list credentials: %w", err)
		}
		
		for _, record := range records {
			credentials = append(credentials, record.Credential)
		}
	}

	// Evaluate credentials against presentation definition
	processor := vc.NewPresentationDefinitionProcessor()
	result, err := processor.EvaluateCredentials(&presDef, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate presentation definition: %w", err)
	}

	return result, nil
}

func (s *Service) CreatePresentationSubmission(definition interface{}, matchedCredentialIDs []string) (interface{}, error) {
	// Parse presentation definition
	var presDefBytes []byte
	var err error

	switch def := definition.(type) {
	case string:
		presDefBytes = []byte(def)
	case map[string]interface{}:
		presDefBytes, err = json.Marshal(def)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal presentation definition: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported presentation definition type: %T", definition)
	}

	var presDef vc.PresentationDefinition
	if err := json.Unmarshal(presDefBytes, &presDef); err != nil {
		return nil, fmt.Errorf("failed to parse presentation definition: %w", err)
	}

	// Get matched credentials
	var credentials []*vc.VerifiableCredential
	for _, credID := range matchedCredentialIDs {
		record, err := s.wallet.GetCredential(credID)
		if err != nil {
			return nil, fmt.Errorf("failed to get credential %s: %w", credID, err)
		}
		credentials = append(credentials, record.Credential)
	}

	// Create matches (simplified - in a full implementation would re-evaluate)
	processor := vc.NewPresentationDefinitionProcessor()
	var matches []*vc.CredentialMatch

	for i, credential := range credentials {
		// For simplicity, create a match for each credential against first input descriptor
		if len(presDef.InputDescriptors) > 0 {
			match := &vc.CredentialMatch{
				InputDescriptorID: presDef.InputDescriptors[0].ID,
				CredentialIndex:   i,
				Credential:        credential,
			}
			matches = append(matches, match)
		}
	}

	// Create submission
	submission := processor.CreateSubmission(&presDef, matches)
	return submission, nil
}

// Advanced Verification Operations

func (s *Service) VerifyCredentialsBatch(credentialIDs []string, options map[string]interface{}) (interface{}, error) {
	// Get credentials from wallet
	var credentials []*vc.VerifiableCredential
	for _, credID := range credentialIDs {
		record, err := s.wallet.GetCredential(credID)
		if err != nil {
			continue // Skip missing credentials
		}
		credentials = append(credentials, record.Credential)
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("no valid credentials found")
	}

	// Create workflow options
	workflowOptions := &vc.WorkflowOptions{
		Concurrency:     5,
		FailFast:        false,
		ValidateSchemas: true,
		CheckStatus:     true,
	}

	// Parse options
	if opt, exists := options["concurrency"]; exists {
		if concurrency, ok := opt.(float64); ok {
			workflowOptions.Concurrency = int(concurrency)
		}
	}

	if opt, exists := options["failFast"]; exists {
		if failFast, ok := opt.(bool); ok {
			workflowOptions.FailFast = failFast
		}
	}

	if opt, exists := options["trustFramework"]; exists {
		if framework, ok := opt.(string); ok {
			workflowOptions.TrustFramework = framework
		}
	}

	// Create key manager and resolver (simplified for demo)
	keyManager := did.NewDefaultKeyManager()
	var resolver did.MultiResolver = nil // Placeholder for demo
	
	// Create advanced workflow
	workflow := vc.NewAdvancedVerificationWorkflow(vc.NewDefaultCredentialVerifier(
		keyManager,
		resolver,
	))

	// Perform batch verification
	result, err := workflow.VerifyBatch(context.Background(), credentials, workflowOptions)
	if err != nil {
		return nil, fmt.Errorf("batch verification failed: %w", err)
	}

	return result, nil
}

func (s *Service) CreateVerificationWorkflow(workflowID string, steps []interface{}) (interface{}, error) {
	// Convert interface steps to verification steps
	var verificationSteps []vc.VerificationStep
	
	for i, stepInterface := range steps {
		stepBytes, err := json.Marshal(stepInterface)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal step %d: %w", i, err)
		}

		var step vc.VerificationStep
		if err := json.Unmarshal(stepBytes, &step); err != nil {
			return nil, fmt.Errorf("failed to parse step %d: %w", i, err)
		}

		verificationSteps = append(verificationSteps, step)
	}

	// Create key manager and resolver (simplified for demo)
	keyManager := did.NewDefaultKeyManager()
	var resolver did.MultiResolver = nil // Placeholder for demo
	
	// Create advanced workflow
	workflow := vc.NewAdvancedVerificationWorkflow(vc.NewDefaultCredentialVerifier(
		keyManager,
		resolver,
	))

	// Create multi-step flow
	flow := workflow.CreateMultiStepFlow(workflowID, verificationSteps)
	
	return flow, nil
}

func (s *Service) ExecuteVerificationWorkflow(workflowID string, inputs map[string]interface{}) (interface{}, error) {
	// In a real implementation, workflows would be stored and retrieved
	// For now, create a simple workflow for demonstration
	
	steps := []vc.VerificationStep{
		{
			ID:            "step-1",
			Name:          "Credential Verification",
			Type:          "credential",
			Configuration: map[string]interface{}{
				"validateSchema":  true,
				"trustFramework": "default",
			},
		},
	}

	// Create key manager and resolver (simplified for demo)
	keyManager := did.NewDefaultKeyManager()
	var resolver did.MultiResolver = nil // Placeholder for demo
	
	// Create advanced workflow
	workflow := vc.NewAdvancedVerificationWorkflow(vc.NewDefaultCredentialVerifier(
		keyManager,
		resolver,
	))

	flow := workflow.CreateMultiStepFlow(workflowID, steps)

	// Execute the flow
	err := workflow.ExecuteMultiStepFlow(context.Background(), flow, inputs)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	return flow, nil
}

// Wallet Operations

func (s *Service) Lock(password string) error {
	return s.wallet.Lock(password)
}

func (s *Service) Unlock(password string) error {
	return s.wallet.Unlock(password)
}

func (s *Service) GetStatus() interface{} {
	isLocked := s.wallet.IsLocked()
	
	// Get counts (simplified)
	keysCount := 0
	didsCount := 0
	credentialsCount := 0
	
	if !isLocked {
		if keys, err := s.wallet.ListKeys(); err == nil {
			keysCount = len(keys)
		}
		if dids, err := s.wallet.ListDIDs(); err == nil {
			didsCount = len(dids)
		}
		if creds, err := s.wallet.ListCredentials(nil); err == nil {
			credentialsCount = len(creds)
		}
	}

	return &Status{
		Locked:           isLocked,
		KeysCount:        keysCount,
		DIDsCount:        didsCount,
		CredentialsCount: credentialsCount,
	}
}

// Helper functions

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}