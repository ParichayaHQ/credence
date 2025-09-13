package events

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	// DID regex pattern for did:key format (accepting base64url for now)
	didKeyRegex = regexp.MustCompile(`^did:key:z[A-Za-z0-9_-]{32,}$`)
	
	// DID web regex pattern for did:web format  
	didWebRegex = regexp.MustCompile(`^did:web:[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)

	// Context values
	validContexts = map[string]bool{
		"general":  true,
		"commerce": true,
		"hiring":   true,
	}

	// Event types
	validEventTypes = map[EventType]bool{
		EventTypeVouch:              true,
		EventTypeReport:             true,
		EventTypeAppeal:             true,
		EventTypeRevocationAnnounce: true,
	}

	// Validator instance
	validate *validator.Validate
)

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("did", validateDID)
	validate.RegisterValidation("base64", validateBase64)
	validate.RegisterValidation("ctx", validateContext)
	validate.RegisterValidation("eventtype", validateEventType)
}

// validateDID validates DID format
func validateDID(fl validator.FieldLevel) bool {
	did := fl.Field().String()
	return IsValidDID(did)
}

// validateBase64 validates base64 encoding
func validateBase64(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

// validateContext validates context values
func validateContext(fl validator.FieldLevel) bool {
	ctx := fl.Field().String()
	return validContexts[ctx]
}

// validateEventType validates event types
func validateEventType(fl validator.FieldLevel) bool {
	eventType := EventType(fl.Field().String())
	return validEventTypes[eventType]
}

// IsValidDID checks if a string is a valid DID
func IsValidDID(did string) bool {
	if did == "" {
		return false
	}

	// Support did:key and did:web formats
	return didKeyRegex.MatchString(did) || didWebRegex.MatchString(did)
}

// IsValidContext checks if a context is valid
func IsValidContext(ctx string) bool {
	return validContexts[ctx]
}

// IsValidEventType checks if an event type is valid
func IsValidEventType(eventType EventType) bool {
	return validEventTypes[eventType]
}

// IsValidBase64 checks if a string is valid base64
func IsValidBase64(str string) bool {
	if str == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

// ValidateEvent validates an event using struct tags and custom rules
func ValidateEvent(event *Event) error {
	if event == nil {
		return ErrInvalidEventStructure
	}

	// Basic struct validation
	if err := validate.Struct(event); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Custom validation rules
	if err := event.Validate(); err != nil {
		return err
	}

	// Additional semantic validations
	if err := validateEventSemantics(event); err != nil {
		return err
	}

	return nil
}

// validateEventSemantics performs semantic validation beyond struct tags
func validateEventSemantics(event *Event) error {
	// Check timestamp is not too far in the future (allow 5 minute clock skew)
	if event.IssuedAt.After(time.Now().Add(5 * time.Minute)) {
		return fmt.Errorf("event timestamp too far in future: %w", ErrInvalidEventStructure)
	}

	// Check timestamp is not too old (allow up to 24 hours for offline signing)
	if event.IssuedAt.Before(time.Now().Add(-24 * time.Hour)) {
		return fmt.Errorf("event timestamp too old: %w", ErrInvalidEventStructure)
	}

	// Validate nonce length (should be at least 12 bytes when decoded)
	if event.Nonce != "" {
		decoded, err := base64.StdEncoding.DecodeString(event.Nonce)
		if err != nil {
			return ErrInvalidNonce
		}
		if len(decoded) < 12 {
			return fmt.Errorf("nonce too short, must be at least 12 bytes: %w", ErrInvalidNonce)
		}
	}

	// Validate epoch format (should be YYYY-MM)
	if !isValidEpoch(event.Epoch) {
		return fmt.Errorf("invalid epoch format, expected YYYY-MM: %w", ErrInvalidEventStructure)
	}

	// Event-specific validations
	switch event.Type {
	case EventTypeVouch:
		return validateVouchEvent(event)
	case EventTypeReport:
		return validateReportEvent(event)
	case EventTypeAppeal:
		return validateAppealEvent(event)
	case EventTypeRevocationAnnounce:
		return validateRevocationEvent(event)
	}

	return nil
}

// isValidEpoch validates epoch format (YYYY-MM)
func isValidEpoch(epoch string) bool {
	if len(epoch) != 7 {
		return false
	}
	
	parts := strings.Split(epoch, "-")
	if len(parts) != 2 {
		return false
	}

	// Basic format check - more detailed validation could be added
	year := parts[0]
	month := parts[1]
	
	if len(year) != 4 || len(month) != 2 {
		return false
	}

	// Could add actual date parsing here for stricter validation
	return true
}

// validateVouchEvent validates vouch-specific rules
func validateVouchEvent(event *Event) error {
	// Vouch cannot be to self
	if event.From == event.To {
		return fmt.Errorf("cannot vouch for self: %w", ErrInvalidEventStructure)
	}

	return nil
}

// validateReportEvent validates report-specific rules  
func validateReportEvent(event *Event) error {
	// Report cannot be against self
	if event.From == event.To {
		return fmt.Errorf("cannot report self: %w", ErrInvalidEventStructure)
	}

	return nil
}

// validateAppealEvent validates appeal-specific rules
func validateAppealEvent(event *Event) error {
	// Basic validation - case_id format could be validated here
	return nil
}

// validateRevocationEvent validates revocation announcement rules
func validateRevocationEvent(event *Event) error {
	// RevocationAnnounce should not have a 'To' field
	if event.To != "" {
		return fmt.Errorf("revocation announce should not have 'to' field: %w", ErrInvalidEventStructure)
	}

	return nil
}

// ValidateSignableEvent validates an event before signing
func ValidateSignableEvent(event *SignableEvent) error {
	if event == nil {
		return ErrInvalidEventStructure
	}

	// Convert to Event for validation (without signature)
	fullEvent := &Event{
		Type:       event.Type,
		From:       event.From,
		To:         event.To,
		Context:    event.Context,
		Epoch:      event.Epoch,
		PayloadCID: event.PayloadCID,
		Nonce:      event.Nonce,
		IssuedAt:   event.IssuedAt,
	}

	return ValidateEvent(fullEvent)
}