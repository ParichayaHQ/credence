package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
)

// IssuerService provides credential issuance functionality
type IssuerService struct {
	wallet           Wallet
	didResolver      did.MultiResolver
	credentialIssuer vc.CredentialIssuer
}

// NewIssuerService creates a new issuer service
func NewIssuerService(wallet Wallet, resolver did.MultiResolver, issuer vc.CredentialIssuer) *IssuerService {
	return &IssuerService{
		wallet:           wallet,
		didResolver:      resolver,
		credentialIssuer: issuer,
	}
}

// IssueCredential issues a new verifiable credential
func (is *IssuerService) IssueCredential(ctx context.Context, request *IssuanceRequest) (*vc.VerifiableCredential, error) {
	if request == nil {
		return nil, NewWalletError(ErrorInvalidCredential, "issuance request cannot be nil")
	}

	// Validate issuance request
	if err := is.validateIssuanceRequest(request); err != nil {
		return nil, err
	}

	// Get issuer DID document to ensure we can sign
	result, err := is.didResolver.Resolve(ctx, request.Issuer, nil)
	if err != nil || result.DIDResolutionMetadata.Error != "" {
		errorMsg := "failed to resolve issuer DID"
		if err != nil {
			errorMsg += ": " + err.Error()
		} else {
			errorMsg += ": " + result.DIDResolutionMetadata.ErrorMessage
		}
		return nil, NewWalletErrorWithDetails(ErrorDIDNotFound, errorMsg, "")
	}

	// Check if we have the signing key in our wallet
	_, err = is.wallet.GetKey(request.SigningKeyID)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorKeyNotFound, 
			"signing key not found in wallet", request.SigningKeyID)
	}

	// For now, skip status list creation since we don't have statuslist manager
	// In a full implementation, this would create revocation entries

	// Build credential template for issuer
	template := &vc.CredentialTemplate{
		Context:           request.Context,
		Type:              request.Type,
		Issuer:            request.Issuer,
		CredentialSubject: request.CredentialSubject,
	}

	if request.ExpirationDate != nil {
		template.ExpirationDate = request.ExpirationDate.Format(time.RFC3339)
	}

	if request.ID != "" {
		// Note: CredentialTemplate doesn't have ID field in the vc package
		// This would need to be handled by the issuer implementation
	}

	// Create issuance options
	issueOptions := &vc.IssuanceOptions{
		KeyID:     request.SigningKeyID,
		Algorithm: request.Algorithm,
	}

	// Add additional claims if needed
	if issueOptions.AdditionalClaims == nil {
		issueOptions.AdditionalClaims = make(map[string]interface{})
	}
	if request.Challenge != "" {
		issueOptions.AdditionalClaims["challenge"] = request.Challenge
	}
	if request.Domain != "" {
		issueOptions.AdditionalClaims["domain"] = request.Domain
	}

	// Issue the credential using the credential issuer
	issuedCredential, err := is.credentialIssuer.IssueCredential(template, issueOptions)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCryptoError, 
			"failed to issue credential", err.Error())
	}

	// Store the issued credential in wallet if requested
	if request.StoreInWallet {
		_, err := is.wallet.StoreCredential(issuedCredential)
		if err != nil {
			// Log warning but don't fail the issuance
			// In production, you might want to handle this differently
			fmt.Printf("Warning: failed to store issued credential in wallet: %v\n", err)
		}
	}

	return issuedCredential, nil
}

// RevokeCredential revokes a previously issued credential
func (is *IssuerService) RevokeCredential(ctx context.Context, credentialID string, reason string) error {
	if credentialID == "" {
		return NewWalletError(ErrorInvalidCredential, "credential ID cannot be empty")
	}

	// Get the credential to find its status list information
	credRecord, err := is.wallet.GetCredential(credentialID)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorCredentialNotFound, 
			"credential not found", credentialID)
	}

	credential := credRecord.Credential
	if credential.CredentialStatus == nil {
		return NewWalletError(ErrorInvalidCredential, "credential does not support revocation")
	}

	// For now, skip status list index parsing since CredentialStatus doesn't have StatusListIndex field
	// In a full implementation, this would parse the index from the status list
	_ = 0 // statusListIndex would be used here

	// For now, skip status list update since we don't have statuslist manager
	// In a full implementation, this would update the revocation status

	// Update credential status in wallet
	credRecord.Status = CredentialStatusRevoked
	if credRecord.Metadata == nil {
		credRecord.Metadata = make(map[string]interface{})
	}
	credRecord.Metadata["revocationReason"] = reason
	credRecord.Metadata["revokedAt"] = time.Now().Format(time.RFC3339)

	// Store updated credential record
	err = is.wallet.(*DefaultWallet).storage.StoreCredential(credRecord)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorStorageError, 
			"failed to update credential status", err.Error())
	}

	return nil
}

// SuspendCredential suspends a previously issued credential
func (is *IssuerService) SuspendCredential(ctx context.Context, credentialID string, reason string) error {
	if credentialID == "" {
		return NewWalletError(ErrorInvalidCredential, "credential ID cannot be empty")
	}

	// Get the credential
	credRecord, err := is.wallet.GetCredential(credentialID)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorCredentialNotFound, 
			"credential not found", credentialID)
	}

	// Update credential status in wallet
	credRecord.Status = CredentialStatusSuspended
	if credRecord.Metadata == nil {
		credRecord.Metadata = make(map[string]interface{})
	}
	credRecord.Metadata["suspensionReason"] = reason
	credRecord.Metadata["suspendedAt"] = time.Now().Format(time.RFC3339)

	// Store updated credential record
	err = is.wallet.(*DefaultWallet).storage.StoreCredential(credRecord)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorStorageError, 
			"failed to update credential status", err.Error())
	}

	return nil
}

// ReinstateCredential reinstates a suspended credential
func (is *IssuerService) ReinstateCredential(ctx context.Context, credentialID string) error {
	if credentialID == "" {
		return NewWalletError(ErrorInvalidCredential, "credential ID cannot be empty")
	}

	// Get the credential
	credRecord, err := is.wallet.GetCredential(credentialID)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorCredentialNotFound, 
			"credential not found", credentialID)
	}

	if credRecord.Status != CredentialStatusSuspended {
		return NewWalletError(ErrorInvalidCredential, "credential is not suspended")
	}

	// Update credential status in wallet
	credRecord.Status = CredentialStatusValid
	if credRecord.Metadata == nil {
		credRecord.Metadata = make(map[string]interface{})
	}
	credRecord.Metadata["reinstatedAt"] = time.Now().Format(time.RFC3339)
	// Remove suspension metadata
	delete(credRecord.Metadata, "suspensionReason")
	delete(credRecord.Metadata, "suspendedAt")

	// Store updated credential record
	err = is.wallet.(*DefaultWallet).storage.StoreCredential(credRecord)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorStorageError, 
			"failed to update credential status", err.Error())
	}

	return nil
}

// GetIssuedCredentials returns credentials issued by this service
func (is *IssuerService) GetIssuedCredentials(ctx context.Context, issuerDID string, filter *CredentialFilter) ([]*CredentialRecord, error) {
	if issuerDID == "" {
		return nil, NewWalletError(ErrorInvalidDID, "issuer DID cannot be empty")
	}

	// Create filter for issued credentials
	issuerFilter := &CredentialFilter{
		Issuer: issuerDID,
	}
	
	// Merge with provided filter
	if filter != nil {
		if filter.Subject != "" {
			issuerFilter.Subject = filter.Subject
		}
		if len(filter.Type) > 0 {
			issuerFilter.Type = filter.Type
		}
		if filter.Status != "" {
			issuerFilter.Status = filter.Status
		}
		if len(filter.Tags) > 0 {
			issuerFilter.Tags = filter.Tags
		}
		if filter.IssuedAfter != nil {
			issuerFilter.IssuedAfter = filter.IssuedAfter
		}
		if filter.IssuedBefore != nil {
			issuerFilter.IssuedBefore = filter.IssuedBefore
		}
		if filter.ExpiresAfter != nil {
			issuerFilter.ExpiresAfter = filter.ExpiresAfter
		}
		if filter.ExpiresBefore != nil {
			issuerFilter.ExpiresBefore = filter.ExpiresBefore
		}
		if filter.Limit > 0 {
			issuerFilter.Limit = filter.Limit
		}
		if filter.Offset > 0 {
			issuerFilter.Offset = filter.Offset
		}
	}

	return is.wallet.ListCredentials(issuerFilter)
}

// CreateCredentialTemplate creates a reusable credential template
func (is *IssuerService) CreateCredentialTemplate(template *CredentialTemplate) error {
	if template == nil {
		return NewWalletError(ErrorInvalidCredential, "template cannot be nil")
	}

	if err := is.validateCredentialTemplate(template); err != nil {
		return err
	}

	// Store template in wallet metadata
	key := fmt.Sprintf("template_%s", template.ID)
	err := is.wallet.(*DefaultWallet).storage.SetMetadata(key, template)
	if err != nil {
		return NewWalletErrorWithDetails(ErrorStorageError, 
			"failed to store template", err.Error())
	}

	return nil
}

// IssueCredentialFromTemplate issues a credential using a template
func (is *IssuerService) IssueCredentialFromTemplate(ctx context.Context, templateID string, data map[string]interface{}) (*vc.VerifiableCredential, error) {
	if templateID == "" {
		return nil, NewWalletError(ErrorInvalidCredential, "template ID cannot be empty")
	}

	// Get template
	key := fmt.Sprintf("template_%s", templateID)
	templateData, err := is.wallet.(*DefaultWallet).storage.GetMetadata(key)
	if err != nil {
		return nil, NewWalletErrorWithDetails(ErrorCredentialNotFound, 
			"template not found", templateID)
	}

	template, ok := templateData.(*CredentialTemplate)
	if !ok {
		return nil, NewWalletError(ErrorSerializationError, "invalid template format")
	}

	// Build issuance request from template
	request, err := is.buildRequestFromTemplate(template, data)
	if err != nil {
		return nil, err
	}

	return is.IssueCredential(ctx, request)
}

// validateIssuanceRequest validates the issuance request
func (is *IssuerService) validateIssuanceRequest(request *IssuanceRequest) error {
	if request.Issuer == "" {
		return NewWalletError(ErrorInvalidCredential, "issuer is required")
	}

	if request.SigningKeyID == "" {
		return NewWalletError(ErrorInvalidCredential, "signing key ID is required")
	}

	if len(request.Type) == 0 {
		return NewWalletError(ErrorInvalidCredential, "credential type is required")
	}

	if request.CredentialSubject == nil {
		return NewWalletError(ErrorInvalidCredential, "credential subject is required")
	}

	// Validate issuer DID format
	if _, err := did.ParseDID(request.Issuer); err != nil {
		return NewWalletErrorWithDetails(ErrorInvalidDID, "invalid issuer DID", err.Error())
	}

	// Validate expiration date
	if request.ExpirationDate != nil && request.ExpirationDate.Before(request.IssuanceDate) {
		return NewWalletError(ErrorInvalidCredential, "expiration date must be after issuance date")
	}

	return nil
}

// validateCredentialTemplate validates a credential template
func (is *IssuerService) validateCredentialTemplate(template *CredentialTemplate) error {
	if template.ID == "" {
		return NewWalletError(ErrorInvalidCredential, "template ID is required")
	}

	if template.Name == "" {
		return NewWalletError(ErrorInvalidCredential, "template name is required")
	}

	if len(template.Type) == 0 {
		return NewWalletError(ErrorInvalidCredential, "template type is required")
	}

	if len(template.RequiredFields) == 0 {
		return NewWalletError(ErrorInvalidCredential, "template must have at least one required field")
	}

	return nil
}


// buildRequestFromTemplate builds an issuance request from a template
func (is *IssuerService) buildRequestFromTemplate(template *CredentialTemplate, data map[string]interface{}) (*IssuanceRequest, error) {
	// Validate required fields are provided
	for _, field := range template.RequiredFields {
		if _, exists := data[field]; !exists {
			return nil, NewWalletErrorWithDetails(ErrorInvalidCredential, 
				"required field missing", field)
		}
	}

	// Build credential subject from template and data
	credentialSubject := make(map[string]interface{})
	
	// Add template defaults
	for key, value := range template.DefaultValues {
		credentialSubject[key] = value
	}
	
	// Add provided data
	for key, value := range data {
		credentialSubject[key] = value
	}

	request := &IssuanceRequest{
		Context:           template.Context,
		Type:              template.Type,
		CredentialSubject: credentialSubject,
		IssuanceDate:      time.Now(),
	}

	// Extract required fields from data
	if issuer, ok := data["issuer"].(string); ok {
		request.Issuer = issuer
	} else {
		return nil, NewWalletError(ErrorInvalidCredential, "issuer is required in template data")
	}

	if keyID, ok := data["signingKeyId"].(string); ok {
		request.SigningKeyID = keyID
	} else {
		return nil, NewWalletError(ErrorInvalidCredential, "signingKeyId is required in template data")
	}

	// Optional fields
	if id, ok := data["id"].(string); ok {
		request.ID = id
	}
	if algorithm, ok := data["algorithm"].(string); ok {
		request.Algorithm = algorithm
	} else {
		request.Algorithm = "EdDSA" // Default
	}
	if expiration, ok := data["expirationDate"].(time.Time); ok {
		request.ExpirationDate = &expiration
	}
	if revocation, ok := data["enableRevocation"].(bool); ok {
		request.EnableRevocation = revocation
	}

	return request, nil
}

// parseStatusListIndex parses a status list index string to int
func parseStatusListIndex(indexStr string) (int, error) {
	var index int
	_, err := fmt.Sscanf(indexStr, "%d", &index)
	return index, err
}

// IssuanceRequest represents a credential issuance request
type IssuanceRequest struct {
	ID                string                 `json:"id,omitempty"`
	Context           []string               `json:"@context"`
	Type              []string               `json:"type"`
	Issuer            string                 `json:"issuer"`
	CredentialSubject map[string]interface{} `json:"credentialSubject"`
	IssuanceDate      time.Time              `json:"issuanceDate"`
	ExpirationDate    *time.Time             `json:"expirationDate,omitempty"`
	SigningKeyID      string                 `json:"signingKeyId"`
	Algorithm         string                 `json:"algorithm"`
	EnableRevocation  bool                   `json:"enableRevocation"`
	Challenge         string                 `json:"challenge,omitempty"`
	Domain            string                 `json:"domain,omitempty"`
	StoreInWallet     bool                   `json:"storeInWallet"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// CredentialTemplate defines a reusable credential template
type CredentialTemplate struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Context        []string               `json:"@context"`
	Type           []string               `json:"type"`
	RequiredFields []string               `json:"requiredFields"`
	OptionalFields []string               `json:"optionalFields"`
	DefaultValues  map[string]interface{} `json:"defaultValues"`
	Constraints    map[string]interface{} `json:"constraints,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// StatusListEntry represents an entry in a status list
type StatusListEntry struct {
	Index                int    `json:"index"`
	StatusListCredential string `json:"statusListCredential"`
}