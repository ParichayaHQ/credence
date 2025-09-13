package vc

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PresentationDefinition represents a DIF Presentation Exchange presentation definition
type PresentationDefinition struct {
	ID      string                  `json:"id"`
	Name    string                  `json:"name,omitempty"`
	Purpose string                  `json:"purpose,omitempty"`
	Format  *ClaimFormat           `json:"format,omitempty"`
	InputDescriptors []InputDescriptor `json:"input_descriptors"`
}

// InputDescriptor describes requirements for input credentials
type InputDescriptor struct {
	ID          string       `json:"id"`
	Name        string       `json:"name,omitempty"`
	Purpose     string       `json:"purpose,omitempty"`
	Group       []string     `json:"group,omitempty"`
	Format      *ClaimFormat `json:"format,omitempty"`
	Constraints *Constraints `json:"constraints,omitempty"`
}

// Constraints defines filtering and path constraints for credentials
type Constraints struct {
	LimitDisclosure *LimitDisclosure `json:"limit_disclosure,omitempty"`
	Fields          []Field          `json:"fields,omitempty"`
	SubjectIsIssuer *SubjectIsIssuer `json:"subject_is_issuer,omitempty"`
}

// Field defines a field constraint with JSONPath selectors
type Field struct {
	Path      []string      `json:"path"`
	ID        string        `json:"id,omitempty"`
	Purpose   string        `json:"purpose,omitempty"`
	Name      string        `json:"name,omitempty"`
	Filter    *Filter       `json:"filter,omitempty"`
	Predicate *Predicate    `json:"predicate,omitempty"`
	Intent    *IntentToRetain `json:"intent_to_retain,omitempty"`
}

// Filter defines JSON Schema constraints for field values
type Filter struct {
	Type        string      `json:"type,omitempty"`
	Format      string      `json:"format,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Minimum     interface{} `json:"minimum,omitempty"`
	Maximum     interface{} `json:"maximum,omitempty"`
	MinLength   int         `json:"minLength,omitempty"`
	MaxLength   int         `json:"maxLength,omitempty"`
	Const       interface{} `json:"const,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
	Not         *Filter     `json:"not,omitempty"`
}

// Predicate defines logic-based constraints
type Predicate string

// IntentToRetain indicates whether the verifier intends to retain field data
type IntentToRetain bool

// LimitDisclosure controls selective disclosure
type LimitDisclosure string

// SubjectIsIssuer indicates the credential subject must be the issuer
type SubjectIsIssuer string

// ClaimFormat specifies supported credential formats
type ClaimFormat struct {
	JWT     *JWTFormat     `json:"jwt,omitempty"`
	JSONLD  *JSONLDFormat  `json:"jwt_vc,omitempty"`
	LDP     *LDPFormat     `json:"ldp_vc,omitempty"`
	SDJWT   *SDJWTFormat   `json:"sd-jwt,omitempty"`
}

// JWTFormat specifies JWT format constraints
type JWTFormat struct {
	Alg []string `json:"alg,omitempty"`
}

// JSONLDFormat specifies JSON-LD format constraints  
type JSONLDFormat struct {
	Alg   []string `json:"alg,omitempty"`
	ProofType []string `json:"proof_type,omitempty"`
}

// LDPFormat specifies Linked Data Proof format constraints
type LDPFormat struct {
	ProofType []string `json:"proof_type,omitempty"`
}

// SDJWTFormat specifies Selective Disclosure JWT format constraints
type SDJWTFormat struct {
	Alg []string `json:"alg,omitempty"`
	KbJWTAlg []string `json:"kb-jwt_alg,omitempty"`
}

// PresentationSubmission represents a submission against a presentation definition
type PresentationSubmission struct {
	ID              string                    `json:"id"`
	DefinitionID    string                    `json:"definition_id"`
	DescriptorMap   []DescriptorMap          `json:"descriptor_map"`
}

// DescriptorMap maps credentials to input descriptors
type DescriptorMap struct {
	ID         string                 `json:"id"`
	Format     string                 `json:"format"`
	Path       string                 `json:"path"`
	PathNested *DescriptorMap        `json:"path_nested,omitempty"`
}

// PresentationDefinitionProcessor handles presentation definition evaluation
type PresentationDefinitionProcessor struct {
	// Could add configuration or dependencies here
}

// NewPresentationDefinitionProcessor creates a new processor
func NewPresentationDefinitionProcessor() *PresentationDefinitionProcessor {
	return &PresentationDefinitionProcessor{}
}

// EvaluateCredentials evaluates credentials against a presentation definition
func (pdp *PresentationDefinitionProcessor) EvaluateCredentials(
	definition *PresentationDefinition,
	credentials []*VerifiableCredential,
) (*EvaluationResult, error) {
	
	result := &EvaluationResult{
		DefinitionID:   definition.ID,
		EvaluatedAt:    time.Now(),
		Matches:        make([]*CredentialMatch, 0),
		Warnings:       make([]string, 0),
		Errors:         make([]string, 0),
	}

	// Evaluate each input descriptor
	for _, descriptor := range definition.InputDescriptors {
		matches := pdp.evaluateInputDescriptor(&descriptor, credentials)
		result.Matches = append(result.Matches, matches...)
		
		if len(matches) == 0 {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("No credentials match input descriptor '%s'", descriptor.ID))
		}
	}

	result.Valid = len(result.Errors) == 0

	return result, nil
}

// EvaluationResult contains the result of evaluating credentials against a definition
type EvaluationResult struct {
	Valid         bool                `json:"valid"`
	DefinitionID  string             `json:"definition_id"`
	Matches       []*CredentialMatch `json:"matches"`
	Warnings      []string           `json:"warnings,omitempty"`
	Errors        []string           `json:"errors,omitempty"`
	EvaluatedAt   time.Time          `json:"evaluated_at"`
}

// CredentialMatch represents a credential that matches an input descriptor
type CredentialMatch struct {
	InputDescriptorID string                 `json:"input_descriptor_id"`
	CredentialIndex   int                    `json:"credential_index"`
	Credential        *VerifiableCredential  `json:"credential"`
	MatchedFields     []FieldMatch          `json:"matched_fields,omitempty"`
}

// FieldMatch represents a field that matched constraints
type FieldMatch struct {
	FieldID    string      `json:"field_id,omitempty"`
	Path       string      `json:"path"`
	Value      interface{} `json:"value"`
	Satisfied  bool        `json:"satisfied"`
}

// evaluateInputDescriptor evaluates credentials against a single input descriptor
func (pdp *PresentationDefinitionProcessor) evaluateInputDescriptor(
	descriptor *InputDescriptor,
	credentials []*VerifiableCredential,
) []*CredentialMatch {
	
	var matches []*CredentialMatch

	for i, credential := range credentials {
		if pdp.credentialMatchesDescriptor(credential, descriptor) {
			match := &CredentialMatch{
				InputDescriptorID: descriptor.ID,
				CredentialIndex:   i,
				Credential:        credential,
				MatchedFields:     pdp.evaluateFieldConstraints(credential, descriptor),
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// credentialMatchesDescriptor checks if a credential matches basic descriptor requirements
func (pdp *PresentationDefinitionProcessor) credentialMatchesDescriptor(
	credential *VerifiableCredential,
	descriptor *InputDescriptor,
) bool {
	
	// Check format constraints
	if descriptor.Format != nil {
		if !pdp.checkFormatConstraints(credential, descriptor.Format) {
			return false
		}
	}

	// Check field constraints
	if descriptor.Constraints != nil && len(descriptor.Constraints.Fields) > 0 {
		for _, field := range descriptor.Constraints.Fields {
			if !pdp.evaluateFieldConstraint(credential, &field) {
				return false
			}
		}
	}

	return true
}

// checkFormatConstraints validates credential format against descriptor format requirements
func (pdp *PresentationDefinitionProcessor) checkFormatConstraints(
	credential *VerifiableCredential,
	format *ClaimFormat,
) bool {
	
	// Simple format checking - in a full implementation this would be more comprehensive
	if credential.JWT != "" {
		// JWT format credential
		return format.JWT != nil || format.JSONLD != nil
	}
	
	// JSON-LD format credential
	return format.JSONLD != nil || format.LDP != nil
}

// evaluateFieldConstraints evaluates all field constraints and returns matches
func (pdp *PresentationDefinitionProcessor) evaluateFieldConstraints(
	credential *VerifiableCredential,
	descriptor *InputDescriptor,
) []FieldMatch {
	
	var matches []FieldMatch

	if descriptor.Constraints == nil {
		return matches
	}

	for _, field := range descriptor.Constraints.Fields {
		for _, path := range field.Path {
			value, found := pdp.extractValueByPath(credential, path)
			if found {
				satisfied := pdp.evaluateFieldFilter(value, field.Filter)
				matches = append(matches, FieldMatch{
					FieldID:   field.ID,
					Path:      path,
					Value:     value,
					Satisfied: satisfied,
				})
			}
		}
	}

	return matches
}

// evaluateFieldConstraint evaluates a single field constraint
func (pdp *PresentationDefinitionProcessor) evaluateFieldConstraint(
	credential *VerifiableCredential,
	field *Field,
) bool {
	
	for _, path := range field.Path {
		value, found := pdp.extractValueByPath(credential, path)
		if found {
			if field.Filter != nil {
				return pdp.evaluateFieldFilter(value, field.Filter)
			}
			return true // Field exists and no filter constraints
		}
	}

	return false // No paths matched
}

// extractValueByPath extracts a value from credential using JSONPath-like selector
func (pdp *PresentationDefinitionProcessor) extractValueByPath(
	credential *VerifiableCredential,
	path string,
) (interface{}, bool) {
	
	// Convert credential to map for path traversal
	credMap, err := pdp.credentialToMap(credential)
	if err != nil {
		return nil, false
	}

	// Simple JSONPath-like implementation
	// In a full implementation, this would use a proper JSONPath library
	return pdp.traversePath(credMap, path)
}

// credentialToMap converts credential to map for path traversal
func (pdp *PresentationDefinitionProcessor) credentialToMap(
	credential *VerifiableCredential,
) (map[string]interface{}, error) {
	
	credBytes, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	var credMap map[string]interface{}
	if err := json.Unmarshal(credBytes, &credMap); err != nil {
		return nil, err
	}

	return credMap, nil
}

// traversePath implements simple JSONPath-like traversal
func (pdp *PresentationDefinitionProcessor) traversePath(
	data map[string]interface{},
	path string,
) (interface{}, bool) {
	
	// Handle JSONPath expressions like "$.credentialSubject.name"
	if strings.HasPrefix(path, "$.") {
		path = path[2:] // Remove "$."
	}

	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch curr := current.(type) {
		case map[string]interface{}:
			if val, exists := curr[part]; exists {
				current = val
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// evaluateFieldFilter evaluates a field value against filter constraints
func (pdp *PresentationDefinitionProcessor) evaluateFieldFilter(
	value interface{},
	filter *Filter,
) bool {
	
	if filter == nil {
		return true
	}

	// Type constraint
	if filter.Type != "" {
		if !pdp.checkValueType(value, filter.Type) {
			return false
		}
	}

	// Const constraint
	if filter.Const != nil {
		return pdp.valuesEqual(value, filter.Const)
	}

	// Enum constraint
	if len(filter.Enum) > 0 {
		for _, enumVal := range filter.Enum {
			if pdp.valuesEqual(value, enumVal) {
				return true
			}
		}
		return false
	}

	// Pattern constraint (for strings)
	if filter.Pattern != "" {
		if str, ok := value.(string); ok {
			// Simple pattern matching - in full implementation would use regex
			return strings.Contains(str, strings.Trim(filter.Pattern, "*"))
		}
		return false
	}

	return true
}

// checkValueType checks if value matches expected type
func (pdp *PresentationDefinitionProcessor) checkValueType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int64, float64:
			return true
		default:
			return false
		}
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			return true
		default:
			return false
		}
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true // Unknown type, assume valid
	}
}

// valuesEqual compares two values for equality
func (pdp *PresentationDefinitionProcessor) valuesEqual(a, b interface{}) bool {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)
	
	if aErr != nil || bErr != nil {
		return false
	}
	
	return string(aJSON) == string(bJSON)
}

// CreateSubmission creates a presentation submission for matched credentials
func (pdp *PresentationDefinitionProcessor) CreateSubmission(
	definition *PresentationDefinition,
	matches []*CredentialMatch,
) *PresentationSubmission {
	
	submission := &PresentationSubmission{
		ID:            fmt.Sprintf("submission-%d", time.Now().Unix()),
		DefinitionID:  definition.ID,
		DescriptorMap: make([]DescriptorMap, 0),
	}

	// Create descriptor map entries for matches
	for _, match := range matches {
		descriptorMap := DescriptorMap{
			ID:     match.InputDescriptorID,
			Format: pdp.getCredentialFormat(match.Credential),
			Path:   fmt.Sprintf("$.verifiableCredential[%d]", match.CredentialIndex),
		}
		
		submission.DescriptorMap = append(submission.DescriptorMap, descriptorMap)
	}

	return submission
}

// getCredentialFormat determines the format of a credential
func (pdp *PresentationDefinitionProcessor) getCredentialFormat(credential *VerifiableCredential) string {
	if credential.JWT != "" {
		return "jwt_vc"
	}
	return "ldp_vc" // Default to Linked Data Proof format
}