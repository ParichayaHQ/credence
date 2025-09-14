package vc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SchemaValidator provides credential schema validation capabilities
type SchemaValidator struct {
	client       *http.Client
	schemaCache  map[string]*SchemaDocument
	cacheTimeout time.Duration
}

// SchemaDocument represents a credential schema
type SchemaDocument struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Author      string                 `json:"author,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	CachedAt    time.Time              `json:"-"`
}

// SchemaValidationResult represents the result of schema validation
type SchemaValidationResult struct {
	Valid       bool                   `json:"valid"`
	SchemaID    string                 `json:"schemaId"`
	Errors      []SchemaValidationError `json:"errors,omitempty"`
	ValidatedAt time.Time              `json:"validatedAt"`
}

// SchemaValidationError represents a schema validation error
type SchemaValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		schemaCache:  make(map[string]*SchemaDocument),
		cacheTimeout: time.Hour, // Cache schemas for 1 hour
	}
}

// ValidateCredentialSchema validates a credential against its schema(s)
func (sv *SchemaValidator) ValidateCredentialSchema(credential *VerifiableCredential) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
		Errors:      make([]SchemaValidationError, 0),
	}

	if credential.CredentialSchema == nil || len(credential.CredentialSchema) == 0 {
		// No schema to validate against - this is not necessarily an error
		return result, nil
	}

	// Validate against each schema
	for _, schema := range credential.CredentialSchema {
		schemaResult, err := sv.validateAgainstSchema(credential, &schema)
		if err != nil {
			return nil, fmt.Errorf("failed to validate against schema %s: %w", schema.ID, err)
		}

		if !schemaResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, schemaResult.Errors...)
		}

		if result.SchemaID == "" {
			result.SchemaID = schema.ID
		}
	}

	return result, nil
}

// validateAgainstSchema validates a credential against a specific schema
func (sv *SchemaValidator) validateAgainstSchema(credential *VerifiableCredential, schema *CredentialSchema) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{
		Valid:       true,
		SchemaID:    schema.ID,
		ValidatedAt: time.Now(),
		Errors:      make([]SchemaValidationError, 0),
	}

	// Fetch schema document
	schemaDoc, err := sv.fetchSchema(schema.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema: %w", err)
	}

	// Validate credential type matches schema
	if err := sv.validateCredentialType(credential, schemaDoc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, SchemaValidationError{
			Field:   "type",
			Message: err.Error(),
			Code:    "type_mismatch",
		})
	}

	// Validate credential subject properties
	if err := sv.validateCredentialSubject(credential, schemaDoc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, SchemaValidationError{
			Field:   "credentialSubject",
			Message: err.Error(),
			Code:    "subject_validation_failed",
		})
	}

	// Validate required fields
	if err := sv.validateRequiredFields(credential, schemaDoc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, SchemaValidationError{
			Field:   "credentialSubject",
			Message: err.Error(),
			Code:    "required_field_missing",
		})
	}

	return result, nil
}

// fetchSchema fetches a schema document by ID
func (sv *SchemaValidator) fetchSchema(schemaID string) (*SchemaDocument, error) {
	// Check cache first
	if cached, exists := sv.schemaCache[schemaID]; exists {
		if time.Since(cached.CachedAt) < sv.cacheTimeout {
			return cached, nil
		}
		// Cache expired, remove it
		delete(sv.schemaCache, schemaID)
	}

	// Handle different schema types
	switch {
	case strings.HasPrefix(schemaID, "http://") || strings.HasPrefix(schemaID, "https://"):
		return sv.fetchHTTPSchema(schemaID)
	case strings.HasPrefix(schemaID, "did:"):
		return sv.fetchDIDSchema(schemaID)
	default:
		return sv.getBuiltinSchema(schemaID)
	}
}

// fetchHTTPSchema fetches a schema from HTTP(S) URL
func (sv *SchemaValidator) fetchHTTPSchema(url string) (*SchemaDocument, error) {
	resp, err := sv.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("schema fetch failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema response: %w", err)
	}

	var schema SchemaDocument
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	schema.CachedAt = time.Now()
	sv.schemaCache[url] = &schema

	return &schema, nil
}

// fetchDIDSchema fetches a schema from a DID URL (placeholder implementation)
func (sv *SchemaValidator) fetchDIDSchema(didURL string) (*SchemaDocument, error) {
	// TODO: Implement DID-based schema resolution
	// This would require DID resolution and schema retrieval from DID documents
	return nil, fmt.Errorf("DID-based schema resolution not yet implemented")
}

// getBuiltinSchema returns builtin schemas for common credential types
func (sv *SchemaValidator) getBuiltinSchema(schemaID string) (*SchemaDocument, error) {
	builtinSchemas := map[string]*SchemaDocument{
		"BasicCredentialSchema": {
			ID:          "BasicCredentialSchema",
			Type:        "JsonSchemaValidator2018",
			Name:        "Basic Credential Schema",
			Description: "Basic schema for general verifiable credentials",
			Version:     "1.0",
			Properties: map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the credential subject",
				},
			},
			Required: []string{"id"},
		},
		"PersonCredentialSchema": {
			ID:          "PersonCredentialSchema", 
			Type:        "JsonSchemaValidator2018",
			Name:        "Person Credential Schema",
			Description: "Schema for person-related credentials",
			Version:     "1.0",
			Properties: map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "DID or unique identifier for the person",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Full name of the person",
				},
				"firstName": map[string]interface{}{
					"type":        "string", 
					"description": "Given name of the person",
				},
				"lastName": map[string]interface{}{
					"type":        "string",
					"description": "Family name of the person",
				},
				"email": map[string]interface{}{
					"type":        "string",
					"format":      "email",
					"description": "Email address of the person",
				},
				"birthDate": map[string]interface{}{
					"type":        "string",
					"format":      "date",
					"description": "Birth date in YYYY-MM-DD format",
				},
			},
			Required: []string{"id", "name"},
		},
	}

	if schema, exists := builtinSchemas[schemaID]; exists {
		schema.CachedAt = time.Now()
		return schema, nil
	}

	return nil, fmt.Errorf("unknown builtin schema: %s", schemaID)
}

// validateCredentialType validates that the credential type matches the schema
func (sv *SchemaValidator) validateCredentialType(credential *VerifiableCredential, schema *SchemaDocument) error {
	// Basic type validation - ensure VerifiableCredential is present
	hasVCType := false
	for _, t := range credential.Type {
		if t == "VerifiableCredential" {
			hasVCType = true
			break
		}
	}

	if !hasVCType {
		return fmt.Errorf("credential must include VerifiableCredential type")
	}

	// Additional type validation could be added based on schema requirements
	return nil
}

// validateCredentialSubject validates the credential subject against schema properties
func (sv *SchemaValidator) validateCredentialSubject(credential *VerifiableCredential, schema *SchemaDocument) error {
	if credential.CredentialSubject == nil {
		return fmt.Errorf("credential subject is required")
	}

	// Convert credential subject to map for validation
	var subjectMap map[string]interface{}
	
	switch subject := credential.CredentialSubject.(type) {
	case map[string]interface{}:
		subjectMap = subject
	case string:
		// Handle string subject (just a DID)
		subjectMap = map[string]interface{}{"id": subject}
	default:
		// Try to marshal and unmarshal to convert to map
		subjectBytes, err := json.Marshal(subject)
		if err != nil {
			return fmt.Errorf("failed to marshal credential subject for validation")
		}
		
		if err := json.Unmarshal(subjectBytes, &subjectMap); err != nil {
			return fmt.Errorf("failed to convert credential subject to map for validation")
		}
	}

	// Validate each property against schema properties
	if schema.Properties != nil {
		for propName, propSchema := range schema.Properties {
			if value, exists := subjectMap[propName]; exists {
				if err := sv.validateProperty(propName, value, propSchema); err != nil {
					return fmt.Errorf("property %s validation failed: %w", propName, err)
				}
			}
		}
	}

	return nil
}

// validateRequiredFields validates that all required fields are present
func (sv *SchemaValidator) validateRequiredFields(credential *VerifiableCredential, schema *SchemaDocument) error {
	if len(schema.Required) == 0 {
		return nil
	}

	// Convert credential subject to map
	var subjectMap map[string]interface{}
	switch subject := credential.CredentialSubject.(type) {
	case map[string]interface{}:
		subjectMap = subject
	case string:
		subjectMap = map[string]interface{}{"id": subject}
	default:
		subjectBytes, err := json.Marshal(subject)
		if err != nil {
			return fmt.Errorf("failed to marshal credential subject")
		}
		if err := json.Unmarshal(subjectBytes, &subjectMap); err != nil {
			return fmt.Errorf("failed to convert credential subject to map")
		}
	}

	// Check each required field
	for _, required := range schema.Required {
		if _, exists := subjectMap[required]; !exists {
			return fmt.Errorf("required field '%s' is missing", required)
		}
	}

	return nil
}

// validateProperty validates a single property against its schema
func (sv *SchemaValidator) validateProperty(name string, value interface{}, propSchema interface{}) error {
	schemaMap, ok := propSchema.(map[string]interface{})
	if !ok {
		return nil // Skip validation if schema is not in expected format
	}

	// Type validation
	if expectedType, exists := schemaMap["type"]; exists {
		if err := sv.validatePropertyType(name, value, expectedType); err != nil {
			return err
		}
	}

	// Format validation
	if format, exists := schemaMap["format"]; exists {
		if err := sv.validatePropertyFormat(name, value, format); err != nil {
			return err
		}
	}

	return nil
}

// validatePropertyType validates a property's type
func (sv *SchemaValidator) validatePropertyType(name string, value interface{}, expectedType interface{}) error {
	typeStr, ok := expectedType.(string)
	if !ok {
		return nil
	}

	switch typeStr {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("property %s must be a string", name)
		}
	case "number":
		switch value.(type) {
		case int, int64, float64:
			// Valid number types
		default:
			return fmt.Errorf("property %s must be a number", name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("property %s must be a boolean", name)
		}
	case "array":
		// Check if it's a slice
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			// Valid array types
		default:
			return fmt.Errorf("property %s must be an array", name)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("property %s must be an object", name)
		}
	}

	return nil
}

// validatePropertyFormat validates a property's format
func (sv *SchemaValidator) validatePropertyFormat(name string, value interface{}, format interface{}) error {
	formatStr, ok := format.(string)
	if !ok {
		return nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil // Format validation only applies to strings
	}

	switch formatStr {
	case "email":
		if !strings.Contains(valueStr, "@") {
			return fmt.Errorf("property %s must be a valid email address", name)
		}
	case "date":
		if _, err := time.Parse("2006-01-02", valueStr); err != nil {
			return fmt.Errorf("property %s must be a valid date in YYYY-MM-DD format", name)
		}
	case "date-time":
		if _, err := time.Parse(time.RFC3339, valueStr); err != nil {
			return fmt.Errorf("property %s must be a valid date-time in RFC3339 format", name)
		}
	case "uri":
		if !strings.HasPrefix(valueStr, "http://") && !strings.HasPrefix(valueStr, "https://") && !strings.HasPrefix(valueStr, "did:") {
			return fmt.Errorf("property %s must be a valid URI", name)
		}
	}

	return nil
}

// SetCacheTimeout sets the schema cache timeout
func (sv *SchemaValidator) SetCacheTimeout(timeout time.Duration) {
	sv.cacheTimeout = timeout
}

// ClearCache clears the schema cache
func (sv *SchemaValidator) ClearCache() {
	sv.schemaCache = make(map[string]*SchemaDocument)
}