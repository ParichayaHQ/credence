package vc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
)

// JWTCredentialProcessor handles JWT-format verifiable credentials
type JWTCredentialProcessor struct {
	keyManager did.KeyManager
	resolver   did.MultiResolver
}

// NewJWTCredentialProcessor creates a new JWT credential processor
func NewJWTCredentialProcessor(keyManager did.KeyManager, resolver did.MultiResolver) *JWTCredentialProcessor {
	return &JWTCredentialProcessor{
		keyManager: keyManager,
		resolver:   resolver,
	}
}

// ParseJWTCredential parses a JWT credential string into a structured format
func (p *JWTCredentialProcessor) ParseJWTCredential(token string) (*JWTCredential, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, NewVCError(ErrorInvalidJWT, "JWT must have 3 parts")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode header", err.Error())
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse header", err.Error())
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode claims", err.Error())
	}

	var jwtCred JWTCredential
	if err := json.Unmarshal(claimsBytes, &jwtCred); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse claims", err.Error())
	}

	jwtCred.Header = header
	jwtCred.Token = token

	return &jwtCred, nil
}

// ParseJWTPresentation parses a JWT presentation string into a structured format
func (p *JWTCredentialProcessor) ParseJWTPresentation(token string) (*JWTPresentation, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, NewVCError(ErrorInvalidJWT, "JWT must have 3 parts")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode header", err.Error())
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse header", err.Error())
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode claims", err.Error())
	}

	var jwtPres JWTPresentation
	if err := json.Unmarshal(claimsBytes, &jwtPres); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse claims", err.Error())
	}

	jwtPres.Header = header
	jwtPres.Token = token

	return &jwtPres, nil
}

// CreateJWTCredential creates a JWT credential from a template
func (p *JWTCredentialProcessor) CreateJWTCredential(template *CredentialTemplate, options *IssuanceOptions, privateKey interface{}) (string, error) {
	if template == nil {
		return "", NewVCError(ErrorInvalidCredential, "template cannot be nil")
	}

	if options == nil {
		return "", NewVCError(ErrorInvalidCredential, "options cannot be nil")
	}

	// Create the VC payload
	now := time.Now()
	vc := &VerifiableCredential{
		Context:           template.Context,
		Type:              template.Type,
		Issuer:            template.Issuer,
		IssuanceDate:      now.Format(time.RFC3339),
		CredentialSubject: template.CredentialSubject,
		ExpirationDate:    template.ExpirationDate,
		CredentialStatus:  template.CredentialStatus,
		CredentialSchema:  template.CredentialSchema,
	}

	// Create JWT claims
	claims := map[string]interface{}{
		"iss": getIssuerID(template.Issuer),
		"iat": now.Unix(),
		"vc":  vc,
	}

	// Add subject if present
	if subject := getCredentialSubjectID(template.CredentialSubject); subject != "" {
		claims["sub"] = subject
	}

	// Add expiration if present
	if template.ExpirationDate != "" {
		if expTime, err := time.Parse(time.RFC3339, template.ExpirationDate); err == nil {
			claims["exp"] = expTime.Unix()
		}
	}

	// Add additional claims
	if options.AdditionalClaims != nil {
		for key, value := range options.AdditionalClaims {
			claims[key] = value
		}
	}

	// Create header
	header := map[string]interface{}{
		"alg": options.Algorithm,
		"typ": "JWT",
	}

	if options.KeyID != "" {
		header["kid"] = options.KeyID
	}

	return p.signJWT(header, claims, privateKey)
}

// CreateJWTPresentation creates a JWT presentation from credentials
func (p *JWTCredentialProcessor) CreateJWTPresentation(credentials []interface{}, options *PresentationOptions, privateKey interface{}) (string, error) {
	if options == nil {
		return "", NewVCError(ErrorInvalidPresentation, "options cannot be nil")
	}

	// Create the VP payload
	now := time.Now()
	vp := &VerifiablePresentation{
		Context:              []string{"https://www.w3.org/2018/credentials/v1"},
		Type:                 []string{"VerifiablePresentation"},
		Holder:               options.Holder,
		VerifiableCredential: credentials,
	}

	// Create JWT claims
	claims := map[string]interface{}{
		"iss": options.Holder,
		"iat": now.Unix(),
		"vp":  vp,
	}

	// Add audience if present
	if options.Domain != "" {
		claims["aud"] = options.Domain
	}

	// Add nonce/challenge
	if options.Challenge != "" {
		claims["nonce"] = options.Challenge
	}

	// Add additional claims
	if options.AdditionalClaims != nil {
		for key, value := range options.AdditionalClaims {
			claims[key] = value
		}
	}

	// Create header
	header := map[string]interface{}{
		"alg": options.Algorithm,
		"typ": "JWT",
	}

	if options.KeyID != "" {
		header["kid"] = options.KeyID
	}

	return p.signJWT(header, claims, privateKey)
}

// VerifyJWTCredential verifies a JWT credential
func (p *JWTCredentialProcessor) VerifyJWTCredential(token string, options *VerificationOptions) (*VerificationResult, error) {
	// Parse the JWT
	jwtCred, err := p.ParseJWTCredential(token)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    err.Error(),
		}, nil
	}

	// Get the issuer's public key
	issuerDID := jwtCred.Issuer
	publicKey, err := p.resolvePublicKey(issuerDID, jwtCred.Header.KeyID)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "failed to resolve issuer key: " + err.Error(),
		}, nil
	}

	// Verify the signature
	if err := p.verifyJWTSignature(token, publicKey); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "signature verification failed: " + err.Error(),
		}, nil
	}

	// Validate time claims
	if err := p.validateTimeClaims(jwtCred, options); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "time validation failed: " + err.Error(),
		}, nil
	}

	// Check credential status if requested
	if options != nil && options.CheckStatus && jwtCred.VC.CredentialStatus != nil {
		if err := p.checkCredentialStatus(jwtCred.VC.CredentialStatus); err != nil {
			return &VerificationResult{
				Verified: false,
				Error:    "status check failed: " + err.Error(),
			}, nil
		}
	}

	return &VerificationResult{
		Verified:        true,
		Credential:      jwtCred.VC,
		JWTCredential:   jwtCred,
		Details: map[string]interface{}{
			"issuer":    issuerDID,
			"algorithm": jwtCred.Header.Algorithm,
		},
	}, nil
}

// VerifyJWTPresentation verifies a JWT presentation
func (p *JWTCredentialProcessor) VerifyJWTPresentation(token string, options *VerificationOptions) (*VerificationResult, error) {
	// Parse the JWT
	jwtPres, err := p.ParseJWTPresentation(token)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    err.Error(),
		}, nil
	}

	// Get the holder's public key
	holderDID := jwtPres.Issuer
	publicKey, err := p.resolvePublicKey(holderDID, jwtPres.Header.KeyID)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "failed to resolve holder key: " + err.Error(),
		}, nil
	}

	// Verify the signature
	if err := p.verifyJWTSignature(token, publicKey); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "signature verification failed: " + err.Error(),
		}, nil
	}

	// Validate challenge/nonce
	if options != nil && options.Challenge != "" {
		if jwtPres.Nonce != options.Challenge {
			return &VerificationResult{
				Verified: false,
				Error:    "challenge/nonce mismatch",
			}, nil
		}
	}

	// Validate domain/audience
	if options != nil && options.Domain != "" {
		if !p.validateAudience(jwtPres.Audience, options.Domain) {
			return &VerificationResult{
				Verified: false,
				Error:    "domain/audience mismatch",
			}, nil
		}
	}

	// Validate time claims
	if err := p.validateTimeClaims(jwtPres, options); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "time validation failed: " + err.Error(),
		}, nil
	}

	return &VerificationResult{
		Verified:         true,
		Presentation:     jwtPres.VP,
		JWTPresentation:  jwtPres,
		Details: map[string]interface{}{
			"holder":    holderDID,
			"algorithm": jwtPres.Header.Algorithm,
		},
	}, nil
}

// Helper methods

func (p *JWTCredentialProcessor) signJWT(header, claims map[string]interface{}, privateKey interface{}) (string, error) {
	// Encode header
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", NewVCErrorWithDetails(ErrorInvalidJWT, "failed to encode header", err.Error())
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode claims
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", NewVCErrorWithDetails(ErrorInvalidJWT, "failed to encode claims", err.Error())
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signing input
	signingInput := headerB64 + "." + claimsB64

	// Sign
	signature, err := p.keyManager.Sign(privateKey, []byte(signingInput))
	if err != nil {
		return "", NewVCErrorWithDetails(ErrorInvalidSignature, "failed to sign JWT", err.Error())
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64, nil
}

func (p *JWTCredentialProcessor) verifyJWTSignature(token string, publicKey interface{}) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return NewVCError(ErrorInvalidJWT, "JWT must have 3 parts")
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return NewVCErrorWithDetails(ErrorInvalidSignature, "failed to decode signature", err.Error())
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	if !p.keyManager.Verify(publicKey, []byte(signingInput), signature) {
		return NewVCError(ErrorInvalidSignature, "signature verification failed")
	}

	return nil
}

func (p *JWTCredentialProcessor) resolvePublicKey(didStr, keyID string) (interface{}, error) {
	if p.resolver == nil {
		return nil, NewVCError(ErrorInvalidIssuer, "no DID resolver configured")
	}

	// Resolve the DID document
	result, err := p.resolver.Resolve(nil, didStr, nil)
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidIssuer, "failed to resolve DID", err.Error())
	}

	if result.DIDResolutionMetadata.Error != "" {
		return nil, NewVCError(ErrorInvalidIssuer, "DID resolution failed: "+result.DIDResolutionMetadata.Error)
	}

	if result.DIDDocument == nil {
		return nil, NewVCError(ErrorInvalidIssuer, "no DID document found")
	}

	// Find the verification method
	var methodID string
	if keyID != "" {
		methodID = keyID
	} else {
		// Use the first authentication method
		if len(result.DIDDocument.Authentication) > 0 {
			if authRef, ok := result.DIDDocument.Authentication[0].(string); ok {
				methodID = authRef
			}
		}
	}

	if methodID == "" {
		return nil, NewVCError(ErrorInvalidIssuer, "no verification method found")
	}

	// Find the verification method
	for _, vm := range result.DIDDocument.VerificationMethod {
		if vm.ID == methodID || "#"+strings.TrimPrefix(vm.ID, didStr) == methodID {
			return p.extractPublicKey(&vm)
		}
	}

	return nil, NewVCError(ErrorInvalidIssuer, "verification method not found: "+methodID)
}

func (p *JWTCredentialProcessor) extractPublicKey(vm *did.VerificationMethod) (interface{}, error) {
	if vm.PublicKeyMultibase != nil {
		// Decode multibase key
		decoded, err := MultibaseDecode(*vm.PublicKeyMultibase)
		if err != nil {
			return nil, NewVCErrorWithDetails(ErrorInvalidIssuer, "failed to decode multibase key", err.Error())
		}

		// For Ed25519, remove the multicodec prefix
		if len(decoded) >= 2 && decoded[0] == 0xed && decoded[1] == 0x01 {
			return decoded[2:], nil
		}

		return decoded, nil
	}

	if vm.PublicKeyJwk != nil {
		// Convert JWK to key
		return p.keyManager.JWKToKey(vm.PublicKeyJwk)
	}

	return nil, NewVCError(ErrorInvalidIssuer, "unsupported key format")
}

func (p *JWTCredentialProcessor) validateTimeClaims(claims interface{}, options *VerificationOptions) error {
	var now time.Time
	if options != nil && options.Now != nil {
		now = *options.Now
	} else {
		now = time.Now()
	}

	// Extract time claims based on type
	var exp, nbf *int64
	
	switch c := claims.(type) {
	case *JWTCredential:
		exp = c.ExpirationTime
		nbf = c.NotBefore
	case *JWTPresentation:
		exp = c.ExpirationTime
		nbf = c.NotBefore
	}

	// Check expiration
	if exp != nil && now.Unix() >= *exp {
		return NewVCError(ErrorExpiredCredential, "credential/presentation has expired")
	}

	// Check not before
	if nbf != nil && now.Unix() < *nbf {
		return NewVCError(ErrorInvalidCredential, "credential/presentation not yet valid")
	}

	return nil
}

func (p *JWTCredentialProcessor) validateAudience(audience interface{}, expectedDomain string) bool {
	if audience == nil {
		return expectedDomain == ""
	}

	switch aud := audience.(type) {
	case string:
		return aud == expectedDomain
	case []interface{}:
		for _, a := range aud {
			if s, ok := a.(string); ok && s == expectedDomain {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (p *JWTCredentialProcessor) checkCredentialStatus(status *CredentialStatus) error {
	// This is a placeholder - actual implementation would depend on the status method
	// For example, StatusList 2021, RevocationList2020, etc.
	
	// For now, just return nil (status is valid)
	return nil
}

func getIssuerID(issuer interface{}) string {
	switch iss := issuer.(type) {
	case string:
		return iss
	case *Issuer:
		return iss.ID
	case map[string]interface{}:
		if id, ok := iss["id"].(string); ok {
			return id
		}
	}
	return ""
}

func getCredentialSubjectID(subject interface{}) string {
	switch sub := subject.(type) {
	case *CredentialSubject:
		return sub.ID
	case map[string]interface{}:
		if id, ok := sub["id"].(string); ok {
			return id
		}
	}
	return ""
}

// Helper function to decode multibase (temporary implementation)
func MultibaseDecode(encoded string) ([]byte, error) {
	if len(encoded) == 0 {
		return nil, fmt.Errorf("empty multibase string")
	}

	// For now, delegate to the did package's encoding functions
	// In a real implementation, this would be a proper multibase decoder
	if strings.HasPrefix(encoded, "z") {
		// Use a placeholder base58 decode
		return []byte(encoded[1:]), nil
	}
	
	return []byte(encoded), nil
}