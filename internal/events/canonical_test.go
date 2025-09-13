package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanonicalizeJSON(t *testing.T) {
	t.Run("BasicCanonicalization", func(t *testing.T) {
		data := map[string]interface{}{
			"z_last":  "should be last",
			"a_first": "should be first",
			"number":  42,
			"boolean": true,
		}

		canonical, err := CanonicalizeJSON(data)
		require.NoError(t, err)

		// Should be sorted by keys
		expected := `{"a_first":"should be first","boolean":true,"number":42,"z_last":"should be last"}`
		assert.Equal(t, expected, string(canonical))
	})

	t.Run("EmptyValues", func(t *testing.T) {
		data := map[string]interface{}{
			"keep":       "value",
			"empty_str":  "",
			"nil_value":  nil,
			"empty_slice": []interface{}{},
		}

		canonical, err := CanonicalizeJSON(data)
		require.NoError(t, err)

		// Empty values should be omitted
		expected := `{"keep":"value"}`
		assert.Equal(t, expected, string(canonical))
	})

	t.Run("NestedObjects", func(t *testing.T) {
		data := map[string]interface{}{
			"outer": map[string]interface{}{
				"z_inner": "last",
				"a_inner": "first",
			},
			"simple": "value",
		}

		canonical, err := CanonicalizeJSON(data)
		require.NoError(t, err)

		expected := `{"outer":{"a_inner":"first","z_inner":"last"},"simple":"value"}`
		assert.Equal(t, expected, string(canonical))
	})
}

func TestCanonicalizeEvent(t *testing.T) {
	t.Run("DeterministicCanonicalization", func(t *testing.T) {
		event := &SignableEvent{
			Type:     EventTypeVouch,
			From:     "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			To:       "did:key:z6MkfrQeRZzJWdHaFLqv3iRPVYUV3CwcMV8H6pCKwmXQ7zK7",
			Context:  "commerce",
			Epoch:    "2025-09",
			Nonce:    "dGVzdC1ub25jZQ==",
			IssuedAt: time.Date(2025, 9, 13, 12, 0, 0, 0, time.UTC),
		}

		canonical1, err := CanonicalizeEvent(event)
		require.NoError(t, err)

		canonical2, err := CanonicalizeEvent(event)
		require.NoError(t, err)

		assert.Equal(t, canonical1, canonical2, "Canonicalization should be deterministic")
	})

	t.Run("EventWithoutOptionalFields", func(t *testing.T) {
		event := &SignableEvent{
			Type:     EventTypeRevocationAnnounce,
			From:     "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			Context:  "general",
			Epoch:    "2025-09",
			Nonce:    "dGVzdC1ub25jZQ==",
			IssuedAt: time.Date(2025, 9, 13, 12, 0, 0, 0, time.UTC),
		}

		canonical, err := CanonicalizeEvent(event)
		require.NoError(t, err)
		assert.NotEmpty(t, canonical)

		// Should not contain empty "to" field
		assert.NotContains(t, string(canonical), `"to":""`)
	})
}

func TestValidateCanonicalJSON(t *testing.T) {
	t.Run("ValidCanonicalJSON", func(t *testing.T) {
		canonical := []byte(`{"a":"first","z":"last"}`)
		err := ValidateCanonicalJSON(canonical)
		assert.NoError(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalid := []byte(`{"invalid": json}`)
		err := ValidateCanonicalJSON(invalid)
		assert.Error(t, err)
	})

	t.Run("NonCanonicalJSON", func(t *testing.T) {
		// Keys not sorted
		nonCanonical := []byte(`{"z":"last","a":"first"}`)
		err := ValidateCanonicalJSON(nonCanonical)
		assert.Error(t, err)
	})

	t.Run("EmptyData", func(t *testing.T) {
		err := ValidateCanonicalJSON([]byte{})
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidEventStructure, err)
	})
}