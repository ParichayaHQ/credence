package events

import (
	"time"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeVouch             EventType = "vouch"
	EventTypeReport            EventType = "report"
	EventTypeAppeal            EventType = "appeal"
	EventTypeRevocationAnnounce EventType = "revocation_announce"
)

// Event represents a canonical event in the system
type Event struct {
	Type       EventType `json:"type" validate:"required,oneof=vouch report appeal revocation_announce"`
	From       string    `json:"from" validate:"required,did"`
	To         string    `json:"to,omitempty" validate:"omitempty,did"`
	Context    string    `json:"ctx" validate:"required,oneof=general commerce hiring"`
	Epoch      string    `json:"epoch" validate:"required"`
	PayloadCID string    `json:"payloadCID,omitempty"`
	Nonce      string    `json:"nonce" validate:"required,base64"`
	IssuedAt   time.Time `json:"issuedAt" validate:"required"`
	Signature  string    `json:"sig,omitempty"`
}

// VouchEvent represents a vouch from one DID to another
type VouchEvent struct {
	Event
	WeightHint float64 `json:"weight_hint,omitempty" validate:"omitempty,min=0,max=1"`
}

// ReportEvent represents a report against a DID
type ReportEvent struct {
	Event
	ReasonCode  string `json:"reason_code" validate:"required"`
	EvidenceCID string `json:"evidenceCID,omitempty"`
}

// AppealEvent represents an appeal to a report
type AppealEvent struct {
	Event
	CaseID         string `json:"case_id" validate:"required"`
	NewEvidenceCID string `json:"new_evidenceCID,omitempty"`
}

// RevocationAnnounceEvent represents a revocation announcement
type RevocationAnnounceEvent struct {
	Event
	Issuer        string `json:"issuer" validate:"required,did"`
	StatusListURI string `json:"statuslistURI" validate:"required,uri"`
	BitmapCID     string `json:"bitmapCID" validate:"required"`
}

// EventHeader contains the basic event information for indexing
type EventHeader struct {
	Type     EventType `json:"type"`
	From     string    `json:"from"`
	To       string    `json:"to,omitempty"`
	Context  string    `json:"ctx"`
	Epoch    string    `json:"epoch"`
	IssuedAt time.Time `json:"issuedAt"`
}

// SignableEvent represents an event without signature for signing
type SignableEvent struct {
	Type       EventType `json:"type"`
	From       string    `json:"from"`
	To         string    `json:"to,omitempty"`
	Context    string    `json:"ctx"`
	Epoch      string    `json:"epoch"`
	PayloadCID string    `json:"payloadCID,omitempty"`
	Nonce      string    `json:"nonce"`
	IssuedAt   time.Time `json:"issuedAt"`
}

// ToSignable converts an Event to a SignableEvent (removes signature)
func (e *Event) ToSignable() *SignableEvent {
	return &SignableEvent{
		Type:       e.Type,
		From:       e.From,
		To:         e.To,
		Context:    e.Context,
		Epoch:      e.Epoch,
		PayloadCID: e.PayloadCID,
		Nonce:      e.Nonce,
		IssuedAt:   e.IssuedAt,
	}
}

// ToHeader extracts header information from an event
func (e *Event) ToHeader() *EventHeader {
	return &EventHeader{
		Type:     e.Type,
		From:     e.From,
		To:       e.To,
		Context:  e.Context,
		Epoch:    e.Epoch,
		IssuedAt: e.IssuedAt,
	}
}

// Validate checks if the event is valid according to basic rules
func (e *Event) Validate() error {
	// For RevocationAnnounce, 'To' field should be empty
	if e.Type == EventTypeRevocationAnnounce && e.To != "" {
		return ErrInvalidEventStructure
	}
	
	// For other event types, 'To' field is required
	if e.Type != EventTypeRevocationAnnounce && e.To == "" {
		return ErrMissingRequiredField
	}

	return nil
}