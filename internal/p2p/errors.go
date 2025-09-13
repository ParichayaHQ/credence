package p2p

import (
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	// ErrNodeNotStarted indicates the P2P node is not started
	ErrNodeNotStarted = errors.New("p2p node not started")

	// ErrNodeAlreadyStarted indicates the P2P node is already started
	ErrNodeAlreadyStarted = errors.New("p2p node already started")

	// ErrInvalidTopic indicates an invalid topic name
	ErrInvalidTopic = errors.New("invalid topic name")

	// ErrTopicNotSubscribed indicates not subscribed to the topic
	ErrTopicNotSubscribed = errors.New("not subscribed to topic")

	// ErrMessageTooLarge indicates the message exceeds size limits
	ErrMessageTooLarge = errors.New("message too large")

	// ErrPeerNotFound indicates the peer was not found
	ErrPeerNotFound = errors.New("peer not found")

	// ErrRateLimited indicates the operation is rate limited
	ErrRateLimited = errors.New("rate limited")

	// ErrPeerGreylisted indicates the peer is greylisted
	ErrPeerGreylisted = errors.New("peer greylisted")

	// ErrInvalidMessage indicates the message format is invalid
	ErrInvalidMessage = errors.New("invalid message format")

	// ErrProviderNotFound indicates no providers found for CID
	ErrProviderNotFound = errors.New("no providers found for CID")

	// ErrDHTNotReady indicates the DHT is not ready
	ErrDHTNotReady = errors.New("DHT not ready")

	// ErrConnectionFailed indicates connection to peer failed
	ErrConnectionFailed = errors.New("connection to peer failed")

	// ErrCacheMiss indicates requested item not found in cache
	ErrCacheMiss = errors.New("cache miss")

	// ErrInvalidCID indicates invalid content identifier
	ErrInvalidCID = errors.New("invalid CID")

	// ErrSubscriptionClosed indicates pubsub subscription was closed
	ErrSubscriptionClosed = errors.New("subscription closed")

	// ErrBlobFetchTimeout indicates blob fetch operation timed out
	ErrBlobFetchTimeout = errors.New("blob fetch timeout")

	// ErrValidationFailed indicates message validation failed
	ErrValidationFailed = errors.New("message validation failed")

	// ErrNetworkNotReady indicates the network is not ready
	ErrNetworkNotReady = errors.New("network not ready")
)

// P2PError represents a P2P-specific error with context
type P2PError struct {
	Op      string            // Operation that failed
	Err     error            // Underlying error
	PeerID  *peer.ID         // Associated peer ID (if any)
	Topic   string           // Associated topic (if any)
	Context map[string]interface{} // Additional context
}

// Error implements the error interface
func (e *P2PError) Error() string {
	msg := fmt.Sprintf("p2p %s: %v", e.Op, e.Err)
	
	if e.PeerID != nil {
		msg += fmt.Sprintf(" (peer: %s)", e.PeerID.String())
	}
	
	if e.Topic != "" {
		msg += fmt.Sprintf(" (topic: %s)", e.Topic)
	}
	
	return msg
}

// Unwrap returns the underlying error
func (e *P2PError) Unwrap() error {
	return e.Err
}

// NewP2PError creates a new P2P error
func NewP2PError(op string, err error) *P2PError {
	return &P2PError{
		Op:      op,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithPeer adds peer context to the error
func (e *P2PError) WithPeer(peerID peer.ID) *P2PError {
	e.PeerID = &peerID
	return e
}

// WithTopic adds topic context to the error
func (e *P2PError) WithTopic(topic string) *P2PError {
	e.Topic = topic
	return e
}

// WithContext adds arbitrary context to the error
func (e *P2PError) WithContext(key string, value interface{}) *P2PError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for known retryable errors
	switch {
	case errors.Is(err, ErrConnectionFailed):
		return true
	case errors.Is(err, ErrBlobFetchTimeout):
		return true
	case errors.Is(err, ErrProviderNotFound):
		return true
	case errors.Is(err, ErrNetworkNotReady):
		return true
	default:
		return false
	}
}

// IsTemporary determines if an error is temporary
func IsTemporary(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for known temporary errors
	switch {
	case errors.Is(err, ErrRateLimited):
		return true
	case errors.Is(err, ErrBlobFetchTimeout):
		return true
	case errors.Is(err, ErrNetworkNotReady):
		return true
	default:
		return false
	}
}