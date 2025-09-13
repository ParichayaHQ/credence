package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

const (
	// MaxEventSize is the maximum allowed size for an event in bytes
	MaxEventSize = 16 * 1024 // 16KB
)

// CanonicalizeJSON converts any struct to canonical JSON representation
// This ensures deterministic serialization for hashing and signing
func CanonicalizeJSON(data interface{}) ([]byte, error) {
	// First marshal to get the basic JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("initial marshal failed: %w", err)
	}

	// Check size limit
	if len(jsonBytes) > MaxEventSize {
		return nil, ErrEventTooLarge
	}

	// Parse into generic interface to sort keys
	var generic interface{}
	if err := json.Unmarshal(jsonBytes, &generic); err != nil {
		return nil, fmt.Errorf("unmarshal for canonicalization failed: %w", err)
	}

	// Canonicalize the data structure
	canonical := canonicalizeValue(generic)

	// Marshal with no escaping and compact format
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	if err := encoder.Encode(canonical); err != nil {
		return nil, fmt.Errorf("canonical marshal failed: %w", err)
	}

	// Remove the trailing newline that Encode adds
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// canonicalizeValue recursively canonicalizes a value
func canonicalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return canonicalizeObject(v)
	case []interface{}:
		return canonicalizeArray(v)
	case string, float64, bool, nil:
		return v
	default:
		// For any other type, convert to string representation
		return fmt.Sprintf("%v", v)
	}
}

// canonicalizeObject canonicalizes a JSON object by sorting keys
func canonicalizeObject(obj map[string]interface{}) map[string]interface{} {
	if obj == nil {
		return nil
	}

	// Get sorted keys
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build new object with sorted keys and canonicalized values
	result := make(map[string]interface{})
	for _, k := range keys {
		// Skip empty values to maintain consistency
		if obj[k] != nil && !isEmpty(obj[k]) {
			result[k] = canonicalizeValue(obj[k])
		}
	}

	return result
}

// canonicalizeArray canonicalizes a JSON array
func canonicalizeArray(arr []interface{}) []interface{} {
	if arr == nil {
		return nil
	}

	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[i] = canonicalizeValue(v)
	}

	return result
}

// isEmpty checks if a value is considered empty for canonicalization
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// CanonicalizeEvent canonicalizes an event for signing or hashing
func CanonicalizeEvent(event *SignableEvent) ([]byte, error) {
	if event == nil {
		return nil, ErrInvalidEventStructure
	}

	return CanonicalizeJSON(event)
}

// CanonicalizeEventWithSignature canonicalizes a signed event
func CanonicalizeEventWithSignature(event *Event) ([]byte, error) {
	if event == nil {
		return nil, ErrInvalidEventStructure
	}

	return CanonicalizeJSON(event)
}

// ValidateCanonicalJSON validates that JSON bytes are in canonical form
func ValidateCanonicalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidEventStructure
	}

	if len(data) > MaxEventSize {
		return ErrEventTooLarge
	}

	// Parse the JSON
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Re-canonicalize and compare
	canonical, err := CanonicalizeJSON(parsed)
	if err != nil {
		return fmt.Errorf("re-canonicalization failed: %w", err)
	}

	if !bytes.Equal(data, canonical) {
		return ErrCanonicalizationFailed
	}

	return nil
}