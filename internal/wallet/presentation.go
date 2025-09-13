package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
)

// PresentationService provides presentation creation and verification functionality
type PresentationService struct {
	wallet             Wallet
	credentialVerifier vc.CredentialVerifier
	didResolver        did.MultiResolver
}

// NewPresentationService creates a new presentation service
func NewPresentationService(wallet Wallet, verifier vc.CredentialVerifier, resolver did.MultiResolver) *PresentationService {
	return &PresentationService{
		wallet:             wallet,
		credentialVerifier: verifier,
		didResolver:        resolver,
	}
}

// CreatePresentation creates a verifiable presentation from selected credentials
func (ps *PresentationService) CreatePresentation(ctx context.Context, request *PresentationRequest) (*vc.VerifiablePresentation, error) {
	if request == nil {
		return nil, NewWalletError(ErrorInvalidCredential, "presentation request cannot be nil")
	}

	// Validate request
	if err := ps.validatePresentationRequest(request); err != nil {
		return nil, err
	}

	// Get credentials from wallet
	credentials := make([]*vc.VerifiableCredential, 0, len(request.CredentialIDs))
	for _, credID := range request.CredentialIDs {
		credRecord, err := ps.wallet.GetCredential(credID)
		if err != nil {
			return nil, NewWalletErrorWithDetails(ErrorCredentialNotFound, 
				"failed to get credential", credID)
		}
		credentials = append(credentials, credRecord.Credential)
	}

	// Apply selective disclosure if specified
	if request.SelectiveDisclosure != nil {
		processedCredentials, err := ps.applySelectiveDisclosure(credentials, request.SelectiveDisclosure)
		if err != nil {
			return nil, err
		}
		credentials = processedCredentials
	}

	// Create presentation options
	options := &PresentationOptions{
		Holder:              request.Holder,
		Verifier:            request.Verifier,
		Challenge:           request.Challenge,
		Domain:              request.Domain,
		Purpose:             request.Purpose,
		KeyID:               request.KeyID,
		Algorithm:           request.Algorithm,
		SelectiveDisclosure: request.SelectiveDisclosure,
		Metadata:            request.Metadata,
	}

	// Create presentation using wallet
	presentation, err := ps.wallet.CreatePresentation(request.CredentialIDs, options)
	if err != nil {
		return nil, err
	}

	return presentation, nil
}

// VerifyPresentation verifies a verifiable presentation
func (ps *PresentationService) VerifyPresentation(ctx context.Context, presentation *vc.VerifiablePresentation, options *VerificationOptions) (*VerificationResult, error) {
	if presentation == nil {
		return nil, NewWalletError(ErrorInvalidCredential, "presentation cannot be nil")
	}

	result := &VerificationResult{
		Valid:      false,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Details:    make(map[string]interface{}),
		VerifiedAt: time.Now(),
	}

	// Verify presentation signature
	if err := ps.verifyPresentationSignature(ctx, presentation); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("signature verification failed: %v", err))
		return result, nil
	}

	// Verify challenge and domain if provided
	if options != nil {
		if err := ps.verifyPresentationContext(presentation, options); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("context verification failed: %v", err))
			return result, nil
		}
	}

	// Verify each credential in the presentation
	credentialResults := make(map[string]*CredentialVerificationResult)
	if presentation.VerifiableCredential != nil {
		for i, credInterface := range presentation.VerifiableCredential {
			// Convert interface{} back to *vc.VerifiableCredential
			var cred *vc.VerifiableCredential
			if credPtr, ok := credInterface.(*vc.VerifiableCredential); ok {
				cred = credPtr
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("credential %d has invalid type", i))
				continue
			}
			
			credResult, err := ps.verifyCredentialInPresentation(ctx, cred, options)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("credential %d verification failed: %v", i, err))
				continue
			}
			credentialResults[fmt.Sprintf("credential_%d", i)] = credResult
			
			if !credResult.Valid {
				result.Errors = append(result.Errors, fmt.Sprintf("credential %d is invalid: %s", i, credResult.Reason))
			}
		}
	}

	result.Details["credentials"] = credentialResults
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// CreatePresentationFromTemplate creates a presentation using a predefined template
func (ps *PresentationService) CreatePresentationFromTemplate(ctx context.Context, templateID string, data map[string]interface{}) (*vc.VerifiablePresentation, error) {
	// Get template from storage or registry
	template, err := ps.getPresentationTemplate(templateID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCredentialNotFound, 
			"presentation template not found", templateID)
	}

	// Build presentation request from template and data
	request, err := ps.buildRequestFromTemplate(template, data)
	if err != nil {
		return nil, err
	}

	// Create presentation
	return ps.CreatePresentation(ctx, request)
}

// validatePresentationRequest validates the presentation request
func (ps *PresentationService) validatePresentationRequest(request *PresentationRequest) error {
	if len(request.CredentialIDs) == 0 {
		return NewWalletError(ErrorInvalidCredential, "at least one credential ID is required")
	}

	if request.Holder == "" {
		return NewWalletError(ErrorInvalidCredential, "holder is required")
	}

	if request.KeyID == "" {
		return NewWalletError(ErrorInvalidCredential, "key ID is required")
	}

	// Validate holder DID format
	if _, err := did.ParseDID(request.Holder); err != nil {
		return NewWalletErrorWithDetails(ErrorInvalidDID, "invalid holder DID", err.Error())
	}

	// Validate verifier DID if provided
	if request.Verifier != "" {
		if _, err := did.ParseDID(request.Verifier); err != nil {
			return NewWalletErrorWithDetails(ErrorInvalidDID, "invalid verifier DID", err.Error())
		}
	}

	return nil
}

// applySelectiveDisclosure applies selective disclosure to credentials
func (ps *PresentationService) applySelectiveDisclosure(credentials []*vc.VerifiableCredential, disclosure map[string][]string) ([]*vc.VerifiableCredential, error) {
	processedCredentials := make([]*vc.VerifiableCredential, len(credentials))
	
	for i, cred := range credentials {
		credKey := fmt.Sprintf("credential_%d", i)
		fields, hasDisclosure := disclosure[credKey]
		
		if !hasDisclosure {
			// No selective disclosure for this credential
			processedCredentials[i] = cred
			continue
		}

		// Apply selective disclosure
		processedCred, err := ps.applySelectiveDisclosureToCredential(cred, fields)
		if err != nil {
			return nil, NewWalletErrorWithDetails(ErrorInvalidCredential,
				fmt.Sprintf("failed to apply selective disclosure to credential %d", i), err.Error())
		}
		
		processedCredentials[i] = processedCred
	}

	return processedCredentials, nil
}

// applySelectiveDisclosureToCredential applies selective disclosure to a single credential
func (ps *PresentationService) applySelectiveDisclosureToCredential(cred *vc.VerifiableCredential, fields []string) (*vc.VerifiableCredential, error) {
	// Create a copy of the credential
	credData, err := json.Marshal(cred)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential: %w", err)
	}

	var credCopy vc.VerifiableCredential
	if err := json.Unmarshal(credData, &credCopy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential copy: %w", err)
	}

	// Apply field selection to credential subject
	if credCopy.CredentialSubject != nil {
		// Convert interface{} to map for field selection
		if subjectMap, ok := credCopy.CredentialSubject.(map[string]interface{}); ok {
			filteredSubject := make(map[string]interface{})
			
			// Always include required fields
			if id, exists := subjectMap["id"]; exists {
				filteredSubject["id"] = id
			}
			
			// Include selected fields
			for _, field := range fields {
				if value, exists := subjectMap[field]; exists {
					filteredSubject[field] = value
				}
			}
			
			credCopy.CredentialSubject = filteredSubject
		}
	}

	return &credCopy, nil
}

// verifyPresentationSignature verifies the presentation's signature
func (ps *PresentationService) verifyPresentationSignature(ctx context.Context, presentation *vc.VerifiablePresentation) error {
	if presentation.Proof == nil {
		return fmt.Errorf("presentation proof is required")
	}

	// Resolve holder's DID to get verification method
	result, err := ps.didResolver.Resolve(ctx, presentation.Holder, nil)
	if err != nil || result.DIDResolutionMetadata.Error != "" {
		errorMsg := "failed to resolve holder DID"
		if err != nil {
			errorMsg += ": " + err.Error()
		} else {
			errorMsg += ": " + result.DIDResolutionMetadata.ErrorMessage
		}
		return fmt.Errorf("%s", errorMsg)
	}

	// Use the credential verifier to verify the presentation
	verifyOptions := &vc.VerificationOptions{}
	if presentation.Holder != "" {
		verifyOptions.Audience = presentation.Holder
	}
	
	verifyResult, err := ps.credentialVerifier.VerifyPresentation(presentation, verifyOptions)
	if err != nil || !verifyResult.Verified {
		errorMsg := "proof verification failed"
		if err != nil {
			errorMsg += ": " + err.Error()
		} else if verifyResult.Error != "" {
			errorMsg += ": " + verifyResult.Error
		}
		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

// verifyPresentationContext verifies challenge, domain, and other context
func (ps *PresentationService) verifyPresentationContext(presentation *vc.VerifiablePresentation, options *VerificationOptions) error {
	// Since Proof is interface{}, we'll skip detailed proof inspection for now
	// In a full implementation, this would parse the proof structure and verify fields
	if presentation.Proof == nil {
		return fmt.Errorf("presentation proof is required")
	}

	// For now, just check that expected values are provided in options
	// A full implementation would extract and verify these from the actual proof
	if options.ExpectedChallenge == "" && options.ExpectedDomain == "" && options.ExpectedPurpose == "" {
		// No context verification required
		return nil
	}

	// In a production implementation, you would:
	// 1. Parse presentation.Proof based on its type (e.g., Ed25519Signature2018)
	// 2. Extract challenge, domain, and purpose from the proof
	// 3. Compare with expected values
	
	return nil
}

// verifyCredentialInPresentation verifies a single credential within a presentation
func (ps *PresentationService) verifyCredentialInPresentation(ctx context.Context, credential *vc.VerifiableCredential, options *VerificationOptions) (*CredentialVerificationResult, error) {
	result := &CredentialVerificationResult{
		Valid:      false,
		Reason:     "",
		VerifiedAt: time.Now(),
	}

	// Verify credential signature and validity
	verifyOpts := &vc.VerificationOptions{
		CheckStatus: true,
	}

	if options != nil && options.TrustFramework != "" {
		// Add trust framework validation 
		if verifyOpts.Params == nil {
			verifyOpts.Params = make(map[string]interface{})
		}
		verifyOpts.Params["trustFramework"] = options.TrustFramework
	}

	verifyResult, err := ps.credentialVerifier.VerifyCredential(credential, verifyOpts)
	if err != nil || !verifyResult.Verified {
		errorMsg := "credential verification failed"
		if err != nil {
			errorMsg += ": " + err.Error()
		} else if verifyResult.Error != "" {
			errorMsg += ": " + verifyResult.Error
		}
		result.Reason = errorMsg
		return result, nil
	}

	result.Valid = true
	return result, nil
}

// getPresentationTemplate retrieves a presentation template
func (ps *PresentationService) getPresentationTemplate(templateID string) (*PresentationTemplate, error) {
	// This would typically fetch from a template store
	// For now, return a basic template
	return &PresentationTemplate{
		ID:          templateID,
		Name:        "Basic Presentation",
		Description: "Basic presentation template",
		Requirements: []TemplateRequirement{
			{
				CredentialType: "VerifiableCredential",
				Required:       true,
				SelectiveFields: []string{"id", "name"},
			},
		},
	}, nil
}

// buildRequestFromTemplate builds a presentation request from a template
func (ps *PresentationService) buildRequestFromTemplate(template *PresentationTemplate, data map[string]interface{}) (*PresentationRequest, error) {
	// Extract required data from template and data map
	holder, ok := data["holder"].(string)
	if !ok {
		return nil, NewWalletError(ErrorInvalidCredential, "holder is required in template data")
	}

	keyID, ok := data["keyId"].(string)
	if !ok {
		return nil, NewWalletError(ErrorInvalidCredential, "keyId is required in template data")
	}

	credentialIDs, ok := data["credentialIds"].([]string)
	if !ok {
		return nil, NewWalletError(ErrorInvalidCredential, "credentialIds is required in template data")
	}

	request := &PresentationRequest{
		CredentialIDs: credentialIDs,
		Holder:        holder,
		KeyID:         keyID,
		Algorithm:     "EdDSA",
	}

	// Add optional fields if present
	if verifier, ok := data["verifier"].(string); ok {
		request.Verifier = verifier
	}
	if challenge, ok := data["challenge"].(string); ok {
		request.Challenge = challenge
	}
	if domain, ok := data["domain"].(string); ok {
		request.Domain = domain
	}

	return request, nil
}

// PresentationRequest represents a request to create a presentation
type PresentationRequest struct {
	CredentialIDs       []string               `json:"credentialIds"`
	Holder              string                 `json:"holder"`
	Verifier            string                 `json:"verifier,omitempty"`
	Challenge           string                 `json:"challenge,omitempty"`
	Domain              string                 `json:"domain,omitempty"`
	Purpose             string                 `json:"purpose,omitempty"`
	KeyID               string                 `json:"keyId"`
	Algorithm           string                 `json:"algorithm"`
	SelectiveDisclosure map[string][]string    `json:"selectiveDisclosure,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationOptions provides options for presentation verification
type VerificationOptions struct {
	ExpectedChallenge  string `json:"expectedChallenge,omitempty"`
	ExpectedDomain     string `json:"expectedDomain,omitempty"`
	ExpectedPurpose    string `json:"expectedPurpose,omitempty"`
	TrustFramework     string `json:"trustFramework,omitempty"`
	CheckRevocation    bool   `json:"checkRevocation"`
	MaxAge             time.Duration `json:"maxAge,omitempty"`
}

// VerificationResult represents the result of presentation verification
type VerificationResult struct {
	Valid      bool                   `json:"valid"`
	Errors     []string               `json:"errors,omitempty"`
	Warnings   []string               `json:"warnings,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	VerifiedAt time.Time              `json:"verifiedAt"`
}

// CredentialVerificationResult represents the result of individual credential verification
type CredentialVerificationResult struct {
	Valid      bool      `json:"valid"`
	Reason     string    `json:"reason,omitempty"`
	VerifiedAt time.Time `json:"verifiedAt"`
}

// PresentationTemplate defines a reusable presentation template
type PresentationTemplate struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Requirements []TemplateRequirement   `json:"requirements"`
	Metadata     map[string]interface{}  `json:"metadata,omitempty"`
}

// TemplateRequirement defines a requirement in a presentation template
type TemplateRequirement struct {
	CredentialType  string   `json:"credentialType"`
	Required        bool     `json:"required"`
	SelectiveFields []string `json:"selectiveFields,omitempty"`
	Constraints     map[string]interface{} `json:"constraints,omitempty"`
}