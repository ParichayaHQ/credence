package vc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TrustFramework defines a trust framework with policies and rules
type TrustFramework struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Version        string                 `json:"version"`
	Description    string                 `json:"description"`
	Issuer         string                 `json:"issuer"`
	Created        time.Time              `json:"created"`
	Updated        time.Time              `json:"updated"`
	Policies       []TrustPolicy          `json:"policies"`
	TrustedIssuers []TrustedIssuer        `json:"trustedIssuers"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TrustPolicy defines trust rules and constraints
type TrustPolicy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "issuer", "credential", "presentation", "subject"
	Rules       []TrustRule       `json:"rules"`
	Actions     []string          `json:"actions"` // "accept", "reject", "review"
	Priority    int               `json:"priority"`
	Description string            `json:"description,omitempty"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
}

// TrustRule defines specific validation rules
type TrustRule struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // "allowlist", "blocklist", "schema", "expiry", "revocation"
	Field       string            `json:"field,omitempty"`
	Operator    string            `json:"operator"` // "equals", "contains", "regex", "exists", "before", "after"
	Value       interface{}       `json:"value,omitempty"`
	Required    bool              `json:"required"`
	Description string            `json:"description,omitempty"`
}

// TrustedIssuer defines trust configuration for an issuer
type TrustedIssuer struct {
	DID                string                 `json:"did"`
	Name               string                 `json:"name,omitempty"`
	TrustLevel         TrustLevel            `json:"trustLevel"`
	AllowedCredTypes   []string              `json:"allowedCredentialTypes,omitempty"`
	RestrictedCredTypes []string             `json:"restrictedCredentialTypes,omitempty"`
	ValidFrom          *time.Time            `json:"validFrom,omitempty"`
	ValidUntil         *time.Time            `json:"validUntil,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// TrustLevel represents the trust level for an entity
type TrustLevel string

const (
	TrustLevelHigh     TrustLevel = "high"
	TrustLevelMedium   TrustLevel = "medium" 
	TrustLevelLow      TrustLevel = "low"
	TrustLevelUntrusted TrustLevel = "untrusted"
)

// TrustFrameworkEngine processes trust decisions
type TrustFrameworkEngine struct {
	frameworks  map[string]*TrustFramework
	policyCache map[string]*PolicyDecision
}

// PolicyDecision represents the result of policy evaluation
type PolicyDecision struct {
	Decision        string                 `json:"decision"` // "accept", "reject", "review"
	Framework       string                 `json:"framework"`
	PolicyID        string                 `json:"policyId"`
	Reason          string                 `json:"reason"`
	Confidence      float64                `json:"confidence"`
	MatchedRules    []string               `json:"matchedRules"`
	Violations      []PolicyViolation      `json:"violations,omitempty"`
	EvaluatedAt     time.Time              `json:"evaluatedAt"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	RuleID      string `json:"ruleId"`
	Field       string `json:"field"`
	Expected    string `json:"expected"`
	Actual      string `json:"actual"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Description string `json:"description"`
}

// NewTrustFrameworkEngine creates a new trust framework engine
func NewTrustFrameworkEngine() *TrustFrameworkEngine {
	return &TrustFrameworkEngine{
		frameworks:  make(map[string]*TrustFramework),
		policyCache: make(map[string]*PolicyDecision),
	}
}

// LoadFramework loads a trust framework into the engine
func (tfe *TrustFrameworkEngine) LoadFramework(framework *TrustFramework) error {
	if framework.ID == "" {
		return fmt.Errorf("framework ID is required")
	}

	// Validate framework structure
	if err := tfe.validateFramework(framework); err != nil {
		return fmt.Errorf("framework validation failed: %w", err)
	}

	tfe.frameworks[framework.ID] = framework
	return nil
}

// GetFramework retrieves a trust framework by ID
func (tfe *TrustFrameworkEngine) GetFramework(frameworkID string) (*TrustFramework, error) {
	framework, exists := tfe.frameworks[frameworkID]
	if !exists {
		return nil, fmt.Errorf("trust framework not found: %s", frameworkID)
	}
	return framework, nil
}

// ListFrameworks returns all loaded trust frameworks
func (tfe *TrustFrameworkEngine) ListFrameworks() []*TrustFramework {
	frameworks := make([]*TrustFramework, 0, len(tfe.frameworks))
	for _, framework := range tfe.frameworks {
		frameworks = append(frameworks, framework)
	}
	return frameworks
}

// EvaluateCredential evaluates a credential against trust framework policies
func (tfe *TrustFrameworkEngine) EvaluateCredential(
	ctx context.Context,
	credential *VerifiableCredential,
	frameworkID string,
) (*PolicyDecision, error) {

	framework, err := tfe.GetFramework(frameworkID)
	if err != nil {
		return nil, err
	}

	// Create policy decision
	decision := &PolicyDecision{
		Framework:    frameworkID,
		Decision:     "reject", // Default to reject
		EvaluatedAt:  time.Now(),
		MatchedRules: make([]string, 0),
		Violations:   make([]PolicyViolation, 0),
		Metadata:     make(map[string]interface{}),
	}

	// Evaluate issuer trust
	issuerDecision := tfe.evaluateIssuerTrust(credential, framework)
	if issuerDecision.Decision == "reject" {
		decision.Decision = "reject"
		decision.Reason = "Issuer not trusted: " + issuerDecision.Reason
		decision.PolicyID = issuerDecision.PolicyID
		decision.Violations = append(decision.Violations, issuerDecision.Violations...)
		return decision, nil
	}

	// Evaluate policies
	highestPriorityDecision := decision
	highestPriority := -1

	for _, policy := range framework.Policies {
		if policy.Type == "credential" || policy.Type == "issuer" {
			policyDecision := tfe.evaluatePolicy(credential, &policy, framework)
			
			// Use highest priority policy decision
			if policy.Priority > highestPriority {
				highestPriority = policy.Priority
				highestPriorityDecision = policyDecision
			}

			decision.MatchedRules = append(decision.MatchedRules, policyDecision.MatchedRules...)
			decision.Violations = append(decision.Violations, policyDecision.Violations...)
		}
	}

	// Use highest priority decision
	if highestPriorityDecision != decision {
		decision.Decision = highestPriorityDecision.Decision
		decision.Reason = highestPriorityDecision.Reason
		decision.PolicyID = highestPriorityDecision.PolicyID
		decision.Confidence = highestPriorityDecision.Confidence
	}

	// Default to accept if no explicit reject
	if decision.Decision == "reject" && len(decision.Violations) == 0 {
		decision.Decision = "accept"
		decision.Reason = "No policy violations found"
	}

	return decision, nil
}

// evaluateIssuerTrust evaluates issuer trustworthiness
func (tfe *TrustFrameworkEngine) evaluateIssuerTrust(
	credential *VerifiableCredential,
	framework *TrustFramework,
) *PolicyDecision {

	decision := &PolicyDecision{
		Framework:   framework.ID,
		Decision:    "reject",
		EvaluatedAt: time.Now(),
		Reason:      "Issuer not found in trusted issuers list",
		Violations:  make([]PolicyViolation, 0),
	}

	issuerID := extractIssuerID(credential.Issuer)
	
	// Check trusted issuers list
	for _, trustedIssuer := range framework.TrustedIssuers {
		if trustedIssuer.DID == issuerID {
			// Check validity period
			now := time.Now()
			if trustedIssuer.ValidFrom != nil && now.Before(*trustedIssuer.ValidFrom) {
				decision.Violations = append(decision.Violations, PolicyViolation{
					RuleID:      "issuer-validity-from",
					Field:       "issuer",
					Expected:    fmt.Sprintf("valid from %s", trustedIssuer.ValidFrom.Format(time.RFC3339)),
					Actual:      fmt.Sprintf("current time %s", now.Format(time.RFC3339)),
					Severity:    "high",
					Description: "Issuer trust not yet valid",
				})
				return decision
			}

			if trustedIssuer.ValidUntil != nil && now.After(*trustedIssuer.ValidUntil) {
				decision.Violations = append(decision.Violations, PolicyViolation{
					RuleID:      "issuer-validity-until",
					Field:       "issuer", 
					Expected:    fmt.Sprintf("valid until %s", trustedIssuer.ValidUntil.Format(time.RFC3339)),
					Actual:      fmt.Sprintf("current time %s", now.Format(time.RFC3339)),
					Severity:    "high",
					Description: "Issuer trust has expired",
				})
				return decision
			}

			// Check credential type restrictions
			if len(trustedIssuer.RestrictedCredTypes) > 0 {
				for _, restrictedType := range trustedIssuer.RestrictedCredTypes {
					for _, credType := range credential.Type {
						if credType == restrictedType {
							decision.Violations = append(decision.Violations, PolicyViolation{
								RuleID:      "issuer-restricted-type",
								Field:       "type",
								Expected:    "not " + restrictedType,
								Actual:      credType,
								Severity:    "medium",
								Description: "Credential type is restricted for this issuer",
							})
							return decision
						}
					}
				}
			}

			// Issuer is trusted
			decision.Decision = "accept"
			decision.Reason = fmt.Sprintf("Issuer trusted with level: %s", trustedIssuer.TrustLevel)
			decision.Confidence = tfe.getTrustLevelConfidence(trustedIssuer.TrustLevel)
			return decision
		}
	}

	return decision
}

// evaluatePolicy evaluates a credential against a specific policy
func (tfe *TrustFrameworkEngine) evaluatePolicy(
	credential *VerifiableCredential,
	policy *TrustPolicy,
	framework *TrustFramework,
) *PolicyDecision {

	decision := &PolicyDecision{
		Framework:    framework.ID,
		PolicyID:     policy.ID,
		Decision:     "accept", // Default to accept for policy evaluation
		EvaluatedAt:  time.Now(),
		MatchedRules: make([]string, 0),
		Violations:   make([]PolicyViolation, 0),
		Confidence:   1.0,
	}

	// Evaluate each rule
	for _, rule := range policy.Rules {
		if !tfe.evaluateRule(credential, &rule, decision) {
			// Rule failed
			if rule.Required {
				decision.Decision = "reject"
				decision.Reason = fmt.Sprintf("Required rule failed: %s", rule.ID)
				return decision
			}
		} else {
			decision.MatchedRules = append(decision.MatchedRules, rule.ID)
		}
	}

	// Determine final decision based on actions
	if len(decision.Violations) > 0 {
		if contains(policy.Actions, "reject") {
			decision.Decision = "reject"
			decision.Reason = "Policy violations detected"
		} else if contains(policy.Actions, "review") {
			decision.Decision = "review"
			decision.Reason = "Policy violations require review"
		}
	}

	return decision
}

// evaluateRule evaluates a credential against a specific rule
func (tfe *TrustFrameworkEngine) evaluateRule(
	credential *VerifiableCredential,
	rule *TrustRule,
	decision *PolicyDecision,
) bool {

	switch rule.Type {
	case "allowlist":
		return tfe.evaluateAllowlistRule(credential, rule, decision)
	case "blocklist":
		return tfe.evaluateBlocklistRule(credential, rule, decision)
	case "schema":
		return tfe.evaluateSchemaRule(credential, rule, decision)
	case "expiry":
		return tfe.evaluateExpiryRule(credential, rule, decision)
	case "revocation":
		return tfe.evaluateRevocationRule(credential, rule, decision)
	default:
		// Unknown rule type, skip
		return true
	}
}

// evaluateAllowlistRule evaluates allowlist rules
func (tfe *TrustFrameworkEngine) evaluateAllowlistRule(
	credential *VerifiableCredential,
	rule *TrustRule,
	decision *PolicyDecision,
) bool {

	field := tfe.getCredentialField(credential, rule.Field)
	if field == nil {
		return !rule.Required
	}

	allowedValues, ok := rule.Value.([]interface{})
	if !ok {
		return false
	}

	fieldStr := fmt.Sprintf("%v", field)
	for _, allowed := range allowedValues {
		if fmt.Sprintf("%v", allowed) == fieldStr {
			return true
		}
	}

	decision.Violations = append(decision.Violations, PolicyViolation{
		RuleID:      rule.ID,
		Field:       rule.Field,
		Expected:    fmt.Sprintf("one of: %v", allowedValues),
		Actual:      fieldStr,
		Severity:    "medium",
		Description: "Value not in allowlist",
	})

	return false
}

// evaluateBlocklistRule evaluates blocklist rules
func (tfe *TrustFrameworkEngine) evaluateBlocklistRule(
	credential *VerifiableCredential,
	rule *TrustRule,
	decision *PolicyDecision,
) bool {

	field := tfe.getCredentialField(credential, rule.Field)
	if field == nil {
		return true // If field doesn't exist, blocklist doesn't apply
	}

	blockedValues, ok := rule.Value.([]interface{})
	if !ok {
		return true
	}

	fieldStr := fmt.Sprintf("%v", field)
	for _, blocked := range blockedValues {
		if fmt.Sprintf("%v", blocked) == fieldStr {
			decision.Violations = append(decision.Violations, PolicyViolation{
				RuleID:      rule.ID,
				Field:       rule.Field,
				Expected:    fmt.Sprintf("not: %v", blocked),
				Actual:      fieldStr,
				Severity:    "high",
				Description: "Value is blocklisted",
			})
			return false
		}
	}

	return true
}

// evaluateSchemaRule evaluates schema validation rules
func (tfe *TrustFrameworkEngine) evaluateSchemaRule(
	credential *VerifiableCredential,
	rule *TrustRule,
	decision *PolicyDecision,
) bool {
	// Schema validation would be integrated with the schema validator
	// For now, assume it passes
	return true
}

// evaluateExpiryRule evaluates credential expiry rules
func (tfe *TrustFrameworkEngine) evaluateExpiryRule(
	credential *VerifiableCredential,
	rule *TrustRule,
	decision *PolicyDecision,
) bool {

	if credential.ExpirationDate == "" {
		if rule.Required {
			decision.Violations = append(decision.Violations, PolicyViolation{
				RuleID:      rule.ID,
				Field:       "expirationDate",
				Expected:    "expiration date required",
				Actual:      "missing",
				Severity:    "medium",
				Description: "Credential must have expiration date",
			})
			return false
		}
		return true
	}

	expiryTime, err := time.Parse(time.RFC3339, credential.ExpirationDate)
	if err != nil {
		decision.Violations = append(decision.Violations, PolicyViolation{
			RuleID:      rule.ID,
			Field:       "expirationDate",
			Expected:    "valid RFC3339 date",
			Actual:      credential.ExpirationDate,
			Severity:    "high",
			Description: "Invalid expiration date format",
		})
		return false
	}

	if time.Now().After(expiryTime) {
		decision.Violations = append(decision.Violations, PolicyViolation{
			RuleID:      rule.ID,
			Field:       "expirationDate",
			Expected:    "not expired",
			Actual:      "expired on " + expiryTime.Format(time.RFC3339),
			Severity:    "critical",
			Description: "Credential has expired",
		})
		return false
	}

	return true
}

// evaluateRevocationRule evaluates revocation status rules
func (tfe *TrustFrameworkEngine) evaluateRevocationRule(
	credential *VerifiableCredential,
	rule *TrustRule,
	decision *PolicyDecision,
) bool {
	// Revocation checking would integrate with status list checking
	// For now, assume not revoked
	return true
}

// Helper methods

func (tfe *TrustFrameworkEngine) validateFramework(framework *TrustFramework) error {
	if framework.Name == "" {
		return fmt.Errorf("framework name is required")
	}
	if framework.Version == "" {
		return fmt.Errorf("framework version is required")
	}
	return nil
}

func (tfe *TrustFrameworkEngine) getCredentialField(credential *VerifiableCredential, fieldPath string) interface{} {
	// Convert credential to map and traverse path
	credBytes, _ := json.Marshal(credential)
	var credMap map[string]interface{}
	json.Unmarshal(credBytes, &credMap)

	parts := strings.Split(fieldPath, ".")
	var current interface{} = credMap

	for _, part := range parts {
		if currMap, ok := current.(map[string]interface{}); ok {
			current = currMap[part]
		} else {
			return nil
		}
	}

	return current
}

func (tfe *TrustFrameworkEngine) getTrustLevelConfidence(level TrustLevel) float64 {
	switch level {
	case TrustLevelHigh:
		return 0.9
	case TrustLevelMedium:
		return 0.7
	case TrustLevelLow:
		return 0.5
	case TrustLevelUntrusted:
		return 0.1
	default:
		return 0.5
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractIssuerID(issuer interface{}) string {
	switch iss := issuer.(type) {
	case string:
		return iss
	case map[string]interface{}:
		if id, exists := iss["id"]; exists {
			return fmt.Sprintf("%v", id)
		}
	}
	return ""
}