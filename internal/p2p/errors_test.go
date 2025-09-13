package p2p

import (
	"errors"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

func TestP2PError(t *testing.T) {
	t.Run("BasicError", func(t *testing.T) {
		underlying := errors.New("connection failed")
		p2pErr := NewP2PError("connect", underlying)

		assert.Equal(t, "connect", p2pErr.Op)
		assert.Equal(t, underlying, p2pErr.Err)
		assert.Nil(t, p2pErr.PeerID)
		assert.Empty(t, p2pErr.Topic)
		assert.Equal(t, "p2p connect: connection failed", p2pErr.Error())
	})

	t.Run("WithPeer", func(t *testing.T) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		underlying := errors.New("peer unreachable")
		p2pErr := NewP2PError("connect", underlying).WithPeer(peerID)

		expectedMsg := "p2p connect: peer unreachable (peer: 12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P)"
		assert.Equal(t, expectedMsg, p2pErr.Error())
		assert.Equal(t, peerID, *p2pErr.PeerID)
	})

	t.Run("WithTopic", func(t *testing.T) {
		underlying := errors.New("invalid message")
		p2pErr := NewP2PError("publish", underlying).WithTopic("events/vouch")

		expectedMsg := "p2p publish: invalid message (topic: events/vouch)"
		assert.Equal(t, expectedMsg, p2pErr.Error())
		assert.Equal(t, "events/vouch", p2pErr.Topic)
	})

	t.Run("WithContext", func(t *testing.T) {
		underlying := errors.New("timeout")
		p2pErr := NewP2PError("fetch", underlying).
			WithContext("cid", "QmHash123").
			WithContext("size", 1024)

		assert.Equal(t, "QmHash123", p2pErr.Context["cid"])
		assert.Equal(t, 1024, p2pErr.Context["size"])
	})

	t.Run("FullContext", func(t *testing.T) {
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		underlying := errors.New("rate limited")
		p2pErr := NewP2PError("publish", underlying).
			WithPeer(peerID).
			WithTopic("events/report").
			WithContext("rate", "100/min")

		expectedMsg := "p2p publish: rate limited (peer: 12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P) (topic: events/report)"
		assert.Equal(t, expectedMsg, p2pErr.Error())
		assert.Equal(t, "100/min", p2pErr.Context["rate"])
	})

	t.Run("Unwrap", func(t *testing.T) {
		underlying := errors.New("original error")
		p2pErr := NewP2PError("test", underlying)

		assert.Equal(t, underlying, p2pErr.Unwrap())
		assert.True(t, errors.Is(p2pErr, underlying))
	})
}

func TestErrorClassification(t *testing.T) {
	t.Run("IsRetryable", func(t *testing.T) {
		// Retryable errors
		assert.True(t, IsRetryable(ErrConnectionFailed))
		assert.True(t, IsRetryable(ErrBlobFetchTimeout))
		assert.True(t, IsRetryable(ErrProviderNotFound))
		assert.True(t, IsRetryable(ErrNetworkNotReady))

		// Non-retryable errors
		assert.False(t, IsRetryable(ErrInvalidTopic))
		assert.False(t, IsRetryable(ErrInvalidMessage))
		assert.False(t, IsRetryable(ErrPeerGreylisted))
		assert.False(t, IsRetryable(ErrNodeAlreadyStarted))

		// Wrapped retryable errors
		wrapped := NewP2PError("connect", ErrConnectionFailed)
		assert.True(t, IsRetryable(wrapped))

		// Nil error
		assert.False(t, IsRetryable(nil))
	})

	t.Run("IsTemporary", func(t *testing.T) {
		// Temporary errors
		assert.True(t, IsTemporary(ErrRateLimited))
		assert.True(t, IsTemporary(ErrBlobFetchTimeout))
		assert.True(t, IsTemporary(ErrNetworkNotReady))

		// Non-temporary errors
		assert.False(t, IsTemporary(ErrInvalidTopic))
		assert.False(t, IsTemporary(ErrPeerGreylisted))
		assert.False(t, IsTemporary(ErrNodeNotStarted))

		// Wrapped temporary errors
		wrapped := NewP2PError("rate_limit", ErrRateLimited)
		assert.True(t, IsTemporary(wrapped))

		// Nil error
		assert.False(t, IsTemporary(nil))
	})

	t.Run("BothRetryableAndTemporary", func(t *testing.T) {
		// Some errors are both retryable and temporary
		assert.True(t, IsRetryable(ErrBlobFetchTimeout))
		assert.True(t, IsTemporary(ErrBlobFetchTimeout))

		assert.True(t, IsRetryable(ErrNetworkNotReady))
		assert.True(t, IsTemporary(ErrNetworkNotReady))
	})
}

func TestStandardErrors(t *testing.T) {
	t.Run("ErrorMessages", func(t *testing.T) {
		errorTests := []struct {
			err      error
			expected string
		}{
			{ErrNodeNotStarted, "p2p node not started"},
			{ErrNodeAlreadyStarted, "p2p node already started"},
			{ErrInvalidTopic, "invalid topic name"},
			{ErrTopicNotSubscribed, "not subscribed to topic"},
			{ErrMessageTooLarge, "message too large"},
			{ErrPeerNotFound, "peer not found"},
			{ErrRateLimited, "rate limited"},
			{ErrPeerGreylisted, "peer greylisted"},
			{ErrInvalidMessage, "invalid message format"},
			{ErrProviderNotFound, "no providers found for CID"},
			{ErrDHTNotReady, "DHT not ready"},
			{ErrConnectionFailed, "connection to peer failed"},
			{ErrCacheMiss, "cache miss"},
			{ErrInvalidCID, "invalid CID"},
			{ErrSubscriptionClosed, "subscription closed"},
			{ErrBlobFetchTimeout, "blob fetch timeout"},
			{ErrValidationFailed, "message validation failed"},
			{ErrNetworkNotReady, "network not ready"},
		}

		for _, test := range errorTests {
			assert.Equal(t, test.expected, test.err.Error())
		}
	})
}

func TestErrorWrappingBehavior(t *testing.T) {
	t.Run("ErrorsIs", func(t *testing.T) {
		base := ErrConnectionFailed
		wrapped := NewP2PError("connect_peer", base)
		doubleWrapped := NewP2PError("retry_connect", wrapped)

		assert.True(t, errors.Is(wrapped, base))
		assert.True(t, errors.Is(doubleWrapped, base))
		assert.True(t, errors.Is(doubleWrapped, wrapped))
	})

	t.Run("ErrorsAs", func(t *testing.T) {
		base := NewP2PError("test", ErrInvalidTopic)
		wrapped := NewP2PError("validate", base)

		var p2pErr *P2PError
		assert.True(t, errors.As(wrapped, &p2pErr))
		assert.Equal(t, "validate", p2pErr.Op)

		// Should find the inner P2PError too
		var innerP2pErr *P2PError
		assert.True(t, errors.As(wrapped, &innerP2pErr))
	})
}