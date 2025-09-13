package vc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
)

// SDJWTProcessor handles Selective Disclosure JWT operations
type SDJWTProcessor struct {
	keyManager did.KeyManager
	resolver   did.MultiResolver
}

// NewSDJWTProcessor creates a new SD-JWT processor
func NewSDJWTProcessor(keyManager did.KeyManager, resolver did.MultiResolver) *SDJWTProcessor {
	return &SDJWTProcessor{
		keyManager: keyManager,
		resolver:   resolver,
	}
}

// CreateSDJWT creates a Selective Disclosure JWT from a credential template
func (p *SDJWTProcessor) CreateSDJWT(template *CredentialTemplate, options *IssuanceOptions, privateKey interface{}) (string, error) {
	if template == nil {
		return "", NewVCError(ErrorInvalidCredential, "template cannot be nil")
	}

	if options == nil {
		return "", NewVCError(ErrorInvalidCredential, "options cannot be nil")
	}

	// Extract selectively disclosable claims
	claims, disclosures, err := p.processSelectiveDisclosure(template.CredentialSubject, template.SelectivelyDisclosable, options.SaltGenerator)
	if err != nil {
		return "", err
	}

	// Create the SD-JWT claims
	sdjwtClaims := map[string]interface{}{
		"iss":   getIssuerID(template.Issuer),
		"iat":   getCurrentTime(),
		"_sd":   p.createSDHashes(disclosures),
		"cnf":   p.createConfirmationClaim(options.RequireKeyBinding),
		"vc":    template,
	}

	// Add additional standard claims
	if subject := getCredentialSubjectID(template.CredentialSubject); subject != "" {
		sdjwtClaims["sub"] = subject
	}

	if template.ExpirationDate != "" {
		if expTime, err := parseTimeToUnix(template.ExpirationDate); err == nil {
			sdjwtClaims["exp"] = expTime
		}
	}

	// Add non-disclosable claims
	for key, value := range claims {
		if !p.isSelectivelyDisclosable(key, template.SelectivelyDisclosable) {
			sdjwtClaims[key] = value
		}
	}

	// Add additional claims
	if options.AdditionalClaims != nil {
		for key, value := range options.AdditionalClaims {
			sdjwtClaims[key] = value
		}
	}

	// Create header
	header := map[string]interface{}{
		"alg": options.Algorithm,
		"typ": "SD-JWT",
	}

	if options.KeyID != "" {
		header["kid"] = options.KeyID
	}

	// Create the JWT
	jwt, err := p.signJWT(header, sdjwtClaims, privateKey)
	if err != nil {
		return "", err
	}

	// Combine JWT with disclosures
	result := jwt
	for _, disclosure := range disclosures {
		result += "~" + disclosure.Encoded
	}

	// Add empty key binding placeholder
	result += "~"

	return result, nil
}

// ParseSDJWT parses an SD-JWT string into its components
func (p *SDJWTProcessor) ParseSDJWT(sdjwt string) (*SDJWTCredential, error) {
	parts := strings.Split(sdjwt, "~")
	if len(parts) < 2 {
		return nil, NewVCError(ErrorInvalidJWT, "SD-JWT must have at least JWT and empty key binding")
	}

	jwt := parts[0]
	disclosureParts := parts[1 : len(parts)-1] // Exclude last part (key binding)
	keyBindingJWT := parts[len(parts)-1]

	// Parse JWT
	jwtParts := strings.Split(jwt, ".")
	if len(jwtParts) != 3 {
		return nil, NewVCError(ErrorInvalidJWT, "JWT must have 3 parts")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(jwtParts[0])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode header", err.Error())
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse header", err.Error())
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(jwtParts[1])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode claims", err.Error())
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse claims", err.Error())
	}

	// Parse disclosures
	disclosures := make([]Disclosure, 0, len(disclosureParts))
	for _, disclosurePart := range disclosureParts {
		if disclosurePart == "" {
			continue
		}

		disclosure, err := p.parseDisclosure(disclosurePart)
		if err != nil {
			return nil, err
		}

		disclosures = append(disclosures, *disclosure)
	}

	// Parse key binding if present
	var keyBinding *KeyBinding
	if keyBindingJWT != "" {
		keyBinding, err = p.parseKeyBinding(keyBindingJWT)
		if err != nil {
			return nil, err
		}
	}

	return &SDJWTCredential{
		JWT:         jwt,
		Disclosures: disclosures,
		KeyBinding:  keyBinding,
		Header:      header,
		Claims:      claims,
	}, nil
}

// VerifySDJWT verifies a Selective Disclosure JWT
func (p *SDJWTProcessor) VerifySDJWT(sdjwt string, options *VerificationOptions) (*VerificationResult, error) {
	// Parse SD-JWT
	sdCredential, err := p.ParseSDJWT(sdjwt)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    err.Error(),
		}, nil
	}

	// Get the issuer's public key
	issuer, ok := sdCredential.Claims["iss"].(string)
	if !ok {
		return &VerificationResult{
			Verified: false,
			Error:    "missing or invalid issuer claim",
		}, nil
	}

	publicKey, err := p.resolvePublicKey(issuer, sdCredential.Header.KeyID)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "failed to resolve issuer key: " + err.Error(),
		}, nil
	}

	// Verify JWT signature
	if err := p.verifyJWTSignature(sdCredential.JWT, publicKey); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "JWT signature verification failed: " + err.Error(),
		}, nil
	}

	// Verify disclosure hashes
	if err := p.verifyDisclosureHashes(sdCredential); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "disclosure verification failed: " + err.Error(),
		}, nil
	}

	// Verify key binding if present
	if sdCredential.KeyBinding != nil {
		if err := p.verifyKeyBinding(sdCredential, options); err != nil {
			return &VerificationResult{
				Verified: false,
				Error:    "key binding verification failed: " + err.Error(),
			}, nil
		}
	}

	// Validate time claims
	if err := p.validateTimeClaimsFromMap(sdCredential.Claims, options); err != nil {
		return &VerificationResult{
			Verified: false,
			Error:    "time validation failed: " + err.Error(),
		}, nil
	}

	// Reconstruct the full credential with disclosed claims
	reconstructedClaims := p.reconstructClaims(sdCredential)

	return &VerificationResult{
		Verified:        true,
		SDJWTCredential: sdCredential,
		Details: map[string]interface{}{
			"issuer":               issuer,
			"algorithm":            sdCredential.Header.Algorithm,
			"disclosed_claims":     reconstructedClaims,
			"disclosure_count":     len(sdCredential.Disclosures),
			"key_binding_present":  sdCredential.KeyBinding != nil,
		},
	}, nil
}

// CreateKeyBindingJWT creates a key binding JWT for holder verification
func (p *SDJWTProcessor) CreateKeyBindingJWT(sdjwt string, audience, nonce string, holderKey interface{}) (string, error) {
	// Parse the original SD-JWT to get the confirmation claim
	parts := strings.Split(sdjwt, "~")
	if len(parts) < 1 {
		return "", NewVCError(ErrorInvalidJWT, "invalid SD-JWT format")
	}

	jwt := parts[0]
	jwtParts := strings.Split(jwt, ".")
	if len(jwtParts) != 3 {
		return "", NewVCError(ErrorInvalidJWT, "JWT must have 3 parts")
	}

	// Create key binding claims
	claims := map[string]interface{}{
		"aud":   audience,
		"nonce": nonce,
		"iat":   getCurrentTime(),
		"sd_hash": p.hashString(jwt), // Hash of the SD-JWT
	}

	// Create header
	header := map[string]interface{}{
		"alg": "EdDSA", // Adjust based on key type
		"typ": "kb+jwt",
	}

	return p.signJWT(header, claims, holderKey)
}

// Helper methods

func (p *SDJWTProcessor) processSelectiveDisclosure(credentialSubject interface{}, selectiveFields []string, saltGenerator func() string) (map[string]interface{}, []Disclosure, error) {
	if saltGenerator == nil {
		saltGenerator = generateDefaultSalt
	}

	claims := make(map[string]interface{})
	var disclosures []Disclosure

	// Convert credential subject to map
	subjectMap, err := p.interfaceToMap(credentialSubject)
	if err != nil {
		return nil, nil, err
	}

	// Process each field
	for key, value := range subjectMap {
		if p.isSelectivelyDisclosable(key, selectiveFields) {
			// Create disclosure
			salt := saltGenerator()
			disclosure := Disclosure{
				Salt:  salt,
				Claim: key,
				Value: value,
			}

			// Encode disclosure
			disclosureArray := []interface{}{salt, key, value}
			disclosureBytes, err := json.Marshal(disclosureArray)
			if err != nil {
				return nil, nil, NewVCErrorWithDetails(ErrorInvalidCredential, "failed to encode disclosure", err.Error())
			}

			disclosure.Encoded = base64.RawURLEncoding.EncodeToString(disclosureBytes)
			disclosures = append(disclosures, disclosure)
		} else {
			claims[key] = value
		}
	}

	return claims, disclosures, nil
}

func (p *SDJWTProcessor) createSDHashes(disclosures []Disclosure) []string {
	hashes := make([]string, len(disclosures))
	for i, disclosure := range disclosures {
		hashes[i] = p.hashString(disclosure.Encoded)
	}
	return hashes
}

func (p *SDJWTProcessor) createConfirmationClaim(requireKeyBinding bool) map[string]interface{} {
	if !requireKeyBinding {
		return nil
	}

	// For simplicity, return a placeholder confirmation claim
	// In a real implementation, this would include the holder's public key or reference
	return map[string]interface{}{
		"jwk": map[string]interface{}{
			"kty": "OKP",
			"crv": "Ed25519",
		},
	}
}

func (p *SDJWTProcessor) parseDisclosure(encoded string) (*Disclosure, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode disclosure", err.Error())
	}

	var disclosureArray []interface{}
	if err := json.Unmarshal(decoded, &disclosureArray); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse disclosure", err.Error())
	}

	if len(disclosureArray) != 3 {
		return nil, NewVCError(ErrorInvalidJWT, "disclosure must have 3 elements")
	}

	salt, ok := disclosureArray[0].(string)
	if !ok {
		return nil, NewVCError(ErrorInvalidJWT, "disclosure salt must be string")
	}

	claim, ok := disclosureArray[1].(string)
	if !ok {
		return nil, NewVCError(ErrorInvalidJWT, "disclosure claim must be string")
	}

	return &Disclosure{
		Salt:    salt,
		Claim:   claim,
		Value:   disclosureArray[2],
		Encoded: encoded,
	}, nil
}

func (p *SDJWTProcessor) parseKeyBinding(jwt string) (*KeyBinding, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, NewVCError(ErrorInvalidJWT, "key binding JWT must have 3 parts")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode key binding header", err.Error())
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse key binding header", err.Error())
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to decode key binding claims", err.Error())
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, NewVCErrorWithDetails(ErrorInvalidJWT, "failed to parse key binding claims", err.Error())
	}

	return &KeyBinding{
		JWT:    jwt,
		Header: header,
		Claims: claims,
	}, nil
}

func (p *SDJWTProcessor) verifyDisclosureHashes(credential *SDJWTCredential) error {
	// Get the _sd array from claims
	sdHashes, ok := credential.Claims["_sd"].([]interface{})
	if !ok {
		return NewVCError(ErrorInvalidCredential, "missing or invalid _sd claim")
	}

	// Convert to string array
	expectedHashes := make([]string, len(sdHashes))
	for i, hash := range sdHashes {
		if hashStr, ok := hash.(string); ok {
			expectedHashes[i] = hashStr
		} else {
			return NewVCError(ErrorInvalidCredential, "invalid hash in _sd array")
		}
	}

	// Calculate hashes for disclosures
	actualHashes := p.createSDHashes(credential.Disclosures)

	// Verify all disclosure hashes are in the _sd array
	for _, actualHash := range actualHashes {
		found := false
		for _, expectedHash := range expectedHashes {
			if actualHash == expectedHash {
				found = true
				break
			}
		}
		if !found {
			return NewVCError(ErrorInvalidCredential, "disclosure hash not found in _sd array")
		}
	}

	return nil
}

func (p *SDJWTProcessor) verifyKeyBinding(credential *SDJWTCredential, options *VerificationOptions) error {
	if credential.KeyBinding == nil {
		return nil // No key binding to verify
	}

	// Get the confirmation claim from the original JWT
	cnf, ok := credential.Claims["cnf"]
	if !ok {
		return NewVCError(ErrorInvalidProof, "missing confirmation claim")
	}

	// Extract public key from confirmation claim
	publicKey, err := p.extractPublicKeyFromConfirmation(cnf)
	if err != nil {
		return err
	}

	// Verify key binding JWT signature
	if err := p.verifyJWTSignature(credential.KeyBinding.JWT, publicKey); err != nil {
		return NewVCError(ErrorInvalidProof, "key binding signature verification failed: "+err.Error())
	}

	// Verify key binding claims
	if options != nil {
		if options.Challenge != "" {
			if nonce, ok := credential.KeyBinding.Claims["nonce"].(string); !ok || nonce != options.Challenge {
				return NewVCError(ErrorInvalidProof, "key binding nonce mismatch")
			}
		}

		if options.Domain != "" {
			if aud, ok := credential.KeyBinding.Claims["aud"].(string); !ok || aud != options.Domain {
				return NewVCError(ErrorInvalidProof, "key binding audience mismatch")
			}
		}
	}

	// Verify SD-JWT hash
	expectedHash := p.hashString(credential.JWT)
	if actualHash, ok := credential.KeyBinding.Claims["sd_hash"].(string); !ok || actualHash != expectedHash {
		return NewVCError(ErrorInvalidProof, "key binding SD-JWT hash mismatch")
	}

	return nil
}

func (p *SDJWTProcessor) reconstructClaims(credential *SDJWTCredential) map[string]interface{} {
	result := make(map[string]interface{})

	// Add non-selective claims
	for key, value := range credential.Claims {
		if key != "_sd" && key != "cnf" {
			result[key] = value
		}
	}

	// Add disclosed claims
	for _, disclosure := range credential.Disclosures {
		result[disclosure.Claim] = disclosure.Value
	}

	return result
}

func (p *SDJWTProcessor) isSelectivelyDisclosable(field string, selectiveFields []string) bool {
	for _, sf := range selectiveFields {
		if sf == field {
			return true
		}
	}
	return false
}

func (p *SDJWTProcessor) interfaceToMap(v interface{}) (map[string]interface{}, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		return val, nil
	case *CredentialSubject:
		result := make(map[string]interface{})
		if val.ID != "" {
			result["id"] = val.ID
		}
		for k, v := range val.Claims {
			result[k] = v
		}
		return result, nil
	default:
		// Use reflection to convert struct to map
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		err = json.Unmarshal(bytes, &result)
		return result, err
	}
}

func (p *SDJWTProcessor) hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func (p *SDJWTProcessor) extractPublicKeyFromConfirmation(cnf interface{}) (interface{}, error) {
	// This is a simplified implementation
	// In practice, the confirmation claim could contain various key representations
	return nil, NewVCError(ErrorInvalidProof, "key binding verification not fully implemented")
}

// Common helper functions

func generateDefaultSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func getCurrentTime() int64 {
	return time.Now().Unix()
}

func parseTimeToUnix(timeStr string) (int64, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Duplicate helper functions from jwt.go to avoid circular dependencies
func (p *SDJWTProcessor) signJWT(header, claims map[string]interface{}, privateKey interface{}) (string, error) {
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

func (p *SDJWTProcessor) verifyJWTSignature(token string, publicKey interface{}) error {
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

func (p *SDJWTProcessor) resolvePublicKey(didStr, keyID string) (interface{}, error) {
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
			return p.extractPublicKeyFromVM(&vm)
		}
	}

	return nil, NewVCError(ErrorInvalidIssuer, "verification method not found: "+methodID)
}

func (p *SDJWTProcessor) extractPublicKeyFromVM(vm *did.VerificationMethod) (interface{}, error) {
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

func (p *SDJWTProcessor) validateTimeClaimsFromMap(claims map[string]interface{}, options *VerificationOptions) error {
	var now time.Time
	if options != nil && options.Now != nil {
		now = *options.Now
	} else {
		now = time.Now()
	}

	// Check expiration
	if expInterface, ok := claims["exp"]; ok {
		if exp, ok := expInterface.(float64); ok && now.Unix() >= int64(exp) {
			return NewVCError(ErrorExpiredCredential, "credential has expired")
		}
	}

	// Check not before
	if nbfInterface, ok := claims["nbf"]; ok {
		if nbf, ok := nbfInterface.(float64); ok && now.Unix() < int64(nbf) {
			return NewVCError(ErrorInvalidCredential, "credential not yet valid")
		}
	}

	return nil
}