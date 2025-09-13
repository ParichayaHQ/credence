package events

import "errors"

var (
	// ErrInvalidEventType indicates an unknown or invalid event type
	ErrInvalidEventType = errors.New("invalid event type")

	// ErrInvalidEventStructure indicates the event structure is malformed
	ErrInvalidEventStructure = errors.New("invalid event structure")

	// ErrMissingRequiredField indicates a required field is missing
	ErrMissingRequiredField = errors.New("missing required field")

	// ErrInvalidSignature indicates the signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidDID indicates the DID format is invalid
	ErrInvalidDID = errors.New("invalid DID format")

	// ErrInvalidNonce indicates the nonce is invalid
	ErrInvalidNonce = errors.New("invalid nonce")

	// ErrInvalidContext indicates the context is invalid
	ErrInvalidContext = errors.New("invalid context")

	// ErrCanonicalizationFailed indicates canonicalization failed
	ErrCanonicalizationFailed = errors.New("canonicalization failed")

	// ErrEventTooLarge indicates the event exceeds size limits
	ErrEventTooLarge = errors.New("event too large")
)