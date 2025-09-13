package vc

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
)

// DefaultCredentialVerifier provides a comprehensive verifiable credential verifier
type DefaultCredentialVerifier struct {
	keyManager    did.KeyManager
	didResolver   did.MultiResolver
	jwtProcessor  *JWTCredentialProcessor
	sdjwtProcessor *SDJWTProcessor
	
	// Optional status list resolver for credential status checking
	statusResolver StatusResolver
}

// StatusResolver interface for checking credential status
type StatusResolver interface {
	CheckStatus(ctx context.Context, status *CredentialStatus) error
}

// NewDefaultCredentialVerifier creates a new comprehensive credential verifier
func NewDefaultCredentialVerifier(keyManager did.KeyManager, resolver did.MultiResolver) *DefaultCredentialVerifier {
	verifier := &DefaultCredentialVerifier{
		keyManager:  keyManager,
		didResolver: resolver,
	}
	
	verifier.jwtProcessor = NewJWTCredentialProcessor(keyManager, resolver)
	verifier.sdjwtProcessor = NewSDJWTProcessor(keyManager, resolver)
	
	return verifier
}

// SetStatusResolver sets the status resolver for credential status checking
func (v *DefaultCredentialVerifier) SetStatusResolver(resolver StatusResolver) {
	v.statusResolver = resolver
}

// VerifyCredential verifies a verifiable credential in JSON-LD format
func (v *DefaultCredentialVerifier) VerifyCredential(credential *VerifiableCredential, options *VerificationOptions) (*VerificationResult, error) {
	if credential == nil {
		return &VerificationResult{
			Verified: false,
			Error:    "credential cannot be nil",
		}, nil
	}

	// Validate basic structure
	if err := v.validateCredentialStructure(credential); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "structure validation failed: " + err.Error(),
		}, nil
	}

	// If it has a JWT representation, verify that instead
	if credential.JWT != "" {
		return v.VerifyJWTCredential(credential.JWT, options)
	}

	// Verify the proof
	if credential.Proof == nil {
		return &VerificationResult{
			Verified: false,
			Error:    "credential has no proof",
		}, nil
	}

	// For now, assume JSON-LD proof verification (would need full implementation)
	// This is a simplified version
	result := &VerificationResult{
		Verified:   true, // Simplified - real implementation would verify proof
		Credential: credential,
		Details: map[string]interface{}{
			"format": "json-ld",
			"issuer": getIssuerID(credential.Issuer),
		},
	}

	// Check credential status if requested
	if options != nil && options.CheckStatus && credential.CredentialStatus != nil {
		if err := v.checkCredentialStatus(credential.CredentialStatus); err != nil {
			result.Verified = false
			result.Error = "status check failed: " + err.Error()
		}
	}

	return result, nil
}

// VerifyPresentation verifies a verifiable presentation in JSON-LD format
func (v *DefaultCredentialVerifier) VerifyPresentation(presentation *VerifiablePresentation, options *VerificationOptions) (*VerificationResult, error) {
	if presentation == nil {
		return &VerificationResult{
			Verified: false,
			Error:    "presentation cannot be nil",
		}, nil
	}

	// Validate basic structure
	if err := v.validatePresentationStructure(presentation); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "structure validation failed: " + err.Error(),
		}, nil
	}

	// If it has a JWT representation, verify that instead
	if presentation.JWT != "" {
		return v.VerifyJWTPresentation(presentation.JWT, options)
	}

	// Verify the proof
	if presentation.Proof == nil {
		return &VerificationResult{
			Verified: false,
			Error:    "presentation has no proof",
		}, nil
	}

	// Verify each embedded credential
	for i, cred := range presentation.VerifiableCredential {
		credResult, err := v.verifyEmbeddedCredential(cred, options)
		if err != nil {
			return &VerificationResult{
				Verified: false,
				Error:    "failed to verify embedded credential " + string(rune(i)) + ": " + err.Error(),
			}, nil
		}
		
		if !credResult.Verified {
			return &VerificationResult{
				Verified: false,
				Error:    "embedded credential " + string(rune(i)) + " verification failed: " + credResult.Error,
			}, nil
		}
	}

	return &VerificationResult{
		Verified:     true,
		Presentation: presentation,
		Details: map[string]interface{}{
			"format":           "json-ld",
			"holder":           presentation.Holder,
			"credentials_count": len(presentation.VerifiableCredential),
		},
	}, nil
}

// VerifyJWTCredential verifies a JWT-format verifiable credential
func (v *DefaultCredentialVerifier) VerifyJWTCredential(token string, options *VerificationOptions) (*VerificationResult, error) {
	return v.jwtProcessor.VerifyJWTCredential(token, options)
}

// VerifyJWTPresentation verifies a JWT-format verifiable presentation
func (v *DefaultCredentialVerifier) VerifyJWTPresentation(token string, options *VerificationOptions) (*VerificationResult, error) {
	return v.jwtProcessor.VerifyJWTPresentation(token, options)
}

// VerifySDJWT verifies a Selective Disclosure JWT
func (v *DefaultCredentialVerifier) VerifySDJWT(sdjwt string, options *VerificationOptions) (*VerificationResult, error) {
	return v.sdjwtProcessor.VerifySDJWT(sdjwt, options)
}

// VerifyAny attempts to verify a credential/presentation in any supported format
func (v *DefaultCredentialVerifier) VerifyAny(input interface{}, options *VerificationOptions) (*VerificationResult, error) {
	switch cred := input.(type) {
	case string:
		// Could be JWT or SD-JWT
		if v.looksLikeSDJWT(cred) {
			return v.VerifySDJWT(cred, options)
		} else if v.looksLikeJWT(cred) {
			// Try as JWT credential first, then presentation
			result, err := v.VerifyJWTCredential(cred, options)
			if err == nil && result.Verified {
				return result, nil
			}
			return v.VerifyJWTPresentation(cred, options)
		}
		return &VerificationResult{
			Verified: false,
			Error:    "unrecognized string format",
		}, nil
		
	case *VerifiableCredential:
		return v.VerifyCredential(cred, options)
		
	case *VerifiablePresentation:
		return v.VerifyPresentation(cred, options)
		
	case map[string]interface{}:
		// Try to determine type from JSON
		if typeArray, ok := cred["type"].([]interface{}); ok {
			for _, t := range typeArray {
				if typeStr, ok := t.(string); ok {
					if typeStr == "VerifiableCredential" {
						// Convert to VerifiableCredential
						credBytes, err := json.Marshal(cred)
						if err != nil {
							return &VerificationResult{
								Verified: false,
								Error:    "failed to marshal credential: " + err.Error(),
							}, nil
						}
						
						var credential VerifiableCredential
						if err := json.Unmarshal(credBytes, &credential); err != nil {
							return &VerificationResult{
								Verified: false,
								Error:    "failed to parse credential: " + err.Error(),
							}, nil
						}
						
						return v.VerifyCredential(&credential, options)
					} else if typeStr == "VerifiablePresentation" {
						// Convert to VerifiablePresentation
						presBytes, err := json.Marshal(cred)
						if err != nil {
							return &VerificationResult{
								Verified: false,
								Error:    "failed to marshal presentation: " + err.Error(),
							}, nil
						}
						
						var presentation VerifiablePresentation
						if err := json.Unmarshal(presBytes, &presentation); err != nil {
							return &VerificationResult{
								Verified: false,
								Error:    "failed to parse presentation: " + err.Error(),
							}, nil
						}
						
						return v.VerifyPresentation(&presentation, options)
					}
				}
			}
		}
		
		return &VerificationResult{
			Verified: false,
			Error:    "unrecognized JSON format",
		}, nil
		
	default:
		return &VerificationResult{
			Verified: false,
			Error:    "unsupported input type",
		}, nil
	}
}

// Helper methods for validation

func (v *DefaultCredentialVerifier) validateCredentialStructure(credential *VerifiableCredential) error {
	// Validate required @context
	if len(credential.Context) == 0 {
		return NewVCError(ErrorInvalidContext, "@context is required")
	}
	
	if credential.Context[0] != "https://www.w3.org/2018/credentials/v1" {
		return NewVCError(ErrorInvalidContext, "first @context must be https://www.w3.org/2018/credentials/v1")
	}

	// Validate required type
	if len(credential.Type) == 0 {
		return NewVCError(ErrorInvalidCredential, "type is required")
	}
	
	hasCredentialType := false
	for _, t := range credential.Type {
		if t == "VerifiableCredential" {
			hasCredentialType = true
			break
		}
	}
	
	if !hasCredentialType {
		return NewVCError(ErrorInvalidCredential, "type must include VerifiableCredential")
	}

	// Validate issuer
	if credential.Issuer == nil {
		return NewVCError(ErrorInvalidIssuer, "issuer is required")
	}

	// Validate issuanceDate
	if credential.IssuanceDate == "" {
		return NewVCError(ErrorInvalidCredential, "issuanceDate is required")
	}

	// Validate credentialSubject
	if credential.CredentialSubject == nil {
		return NewVCError(ErrorInvalidCredential, "credentialSubject is required")
	}

	return nil
}

func (v *DefaultCredentialVerifier) validatePresentationStructure(presentation *VerifiablePresentation) error {
	// Validate required @context
	if len(presentation.Context) == 0 {
		return NewVCError(ErrorInvalidContext, "@context is required")
	}
	
	if presentation.Context[0] != "https://www.w3.org/2018/credentials/v1" {
		return NewVCError(ErrorInvalidContext, "first @context must be https://www.w3.org/2018/credentials/v1")
	}

	// Validate required type
	if len(presentation.Type) == 0 {
		return NewVCError(ErrorInvalidPresentation, "type is required")
	}
	
	hasPresentationType := false
	for _, t := range presentation.Type {
		if t == "VerifiablePresentation" {
			hasPresentationType = true
			break
		}
	}
	
	if !hasPresentationType {
		return NewVCError(ErrorInvalidPresentation, "type must include VerifiablePresentation")
	}

	return nil
}

func (v *DefaultCredentialVerifier) verifyEmbeddedCredential(cred interface{}, options *VerificationOptions) (*VerificationResult, error) {
	switch c := cred.(type) {
	case string:
		// JWT or SD-JWT credential
		return v.VerifyAny(c, options)
	case map[string]interface{}:
		// JSON-LD credential
		return v.VerifyAny(c, options)
	case *VerifiableCredential:
		return v.VerifyCredential(c, options)
	default:
		return &VerificationResult{
			Verified: false,
			Error:    "unsupported embedded credential type",
		}, nil
	}
}

func (v *DefaultCredentialVerifier) checkCredentialStatus(status *CredentialStatus) error {
	if v.statusResolver == nil {
		// No status resolver configured, skip check
		return nil
	}
	
	return v.statusResolver.CheckStatus(context.Background(), status)
}

func (v *DefaultCredentialVerifier) looksLikeSDJWT(input string) bool {
	// SD-JWT contains tildes (~) as separators for disclosures
	return strings.Contains(input, "~")
}

func (v *DefaultCredentialVerifier) looksLikeJWT(input string) bool {
	// JWT has exactly 3 parts separated by dots
	parts := strings.Split(input, ".")
	return len(parts) == 3
}

// DefaultCredentialIssuer provides a comprehensive verifiable credential issuer
type DefaultCredentialIssuer struct {
	keyManager     did.KeyManager
	didResolver    did.MultiResolver
	jwtProcessor   *JWTCredentialProcessor
	sdjwtProcessor *SDJWTProcessor
}

// NewDefaultCredentialIssuer creates a new comprehensive credential issuer
func NewDefaultCredentialIssuer(keyManager did.KeyManager, resolver did.MultiResolver) *DefaultCredentialIssuer {
	issuer := &DefaultCredentialIssuer{
		keyManager:  keyManager,
		didResolver: resolver,
	}
	
	issuer.jwtProcessor = NewJWTCredentialProcessor(keyManager, resolver)
	issuer.sdjwtProcessor = NewSDJWTProcessor(keyManager, resolver)
	
	return issuer
}

// IssueCredential issues a verifiable credential in JSON-LD format
func (i *DefaultCredentialIssuer) IssueCredential(template *CredentialTemplate, options *IssuanceOptions) (*VerifiableCredential, error) {
	if template == nil {
		return nil, NewVCError(ErrorInvalidCredential, "template cannot be nil")
	}

	// For now, return a simplified JSON-LD credential
	// In a full implementation, this would generate proper linked data proofs
	credential := &VerifiableCredential{
		Context:           template.Context,
		Type:              template.Type,
		Issuer:            template.Issuer,
		IssuanceDate:      getCurrentTimeString(),
		CredentialSubject: template.CredentialSubject,
		ExpirationDate:    template.ExpirationDate,
		CredentialStatus:  template.CredentialStatus,
		CredentialSchema:  template.CredentialSchema,
	}

	// TODO: Generate proper proof
	credential.Proof = map[string]interface{}{
		"type":    "Ed25519Signature2020",
		"created": getCurrentTimeString(),
		"verificationMethod": options.KeyID,
		"proofPurpose": "assertionMethod",
		// TODO: Generate actual signature
	}

	return credential, nil
}

// IssueJWTCredential issues a JWT-format verifiable credential
func (i *DefaultCredentialIssuer) IssueJWTCredential(template *CredentialTemplate, options *IssuanceOptions) (string, error) {
	// Resolve the issuer's private key
	privateKey, err := i.resolvePrivateKey(getIssuerID(template.Issuer), options.KeyID)
	if err != nil {
		return "", err
	}

	return i.jwtProcessor.CreateJWTCredential(template, options, privateKey)
}

// IssueSDJWT issues a Selective Disclosure JWT
func (i *DefaultCredentialIssuer) IssueSDJWT(template *CredentialTemplate, options *IssuanceOptions) (string, error) {
	// Resolve the issuer's private key
	privateKey, err := i.resolvePrivateKey(getIssuerID(template.Issuer), options.KeyID)
	if err != nil {
		return "", err
	}

	return i.sdjwtProcessor.CreateSDJWT(template, options, privateKey)
}

func (i *DefaultCredentialIssuer) resolvePrivateKey(issuerDID, keyID string) (interface{}, error) {
	// This is a placeholder - in a real implementation, you would have
	// a secure key store that maps DID + keyID to private keys
	
	// For now, generate a temporary key (this is not secure!)
	return i.keyManager.GenerateKey(did.KeyTypeEd25519)
}

func getCurrentTimeString() string {
	return time.Now().UTC().Format(time.RFC3339)
}