package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicManager(t *testing.T) {
	tm := NewTopicManager()

	t.Run("ValidTopics", func(t *testing.T) {
		validTopics := []string{
			"events/vouch",
			"events/report", 
			"events/appeal",
			"revocations/announce",
			"rules/active",
			"checkpoints/epoch",
			"checkpoints/12345",
			"blobs/QmHash123",
		}

		for _, topic := range validTopics {
			assert.True(t, tm.IsValidTopic(topic), "Topic %s should be valid", topic)
		}
	})

	t.Run("InvalidTopics", func(t *testing.T) {
		invalidTopics := []string{
			"",                    // Empty
			"invalid",            // No category
			"events",             // Missing subtopic
			"events/",            // Empty subtopic
			"events/invalid",     // Invalid event type
			"invalid/topic",      // Invalid category
			"events/vouch/extra", // Too many parts
			"EVENTS/VOUCH",       // Wrong case
			"events vouch",       // Space instead of slash
			"events\\vouch",      // Backslash
		}

		for _, topic := range invalidTopics {
			assert.False(t, tm.IsValidTopic(topic), "Topic %s should be invalid", topic)
		}
	})

	t.Run("GetCoreTopics", func(t *testing.T) {
		coreTopics := tm.GetCoreTopics()
		
		expectedCoreTopics := []string{
			"events/vouch",
			"events/report",
			"events/appeal",
			"rules/active",
			"checkpoints/epoch",
		}

		assert.Equal(t, len(expectedCoreTopics), len(coreTopics))
		
		// Convert to map for easier checking
		topicMap := make(map[string]bool)
		for _, topic := range coreTopics {
			topicMap[topic] = true
		}

		for _, expected := range expectedCoreTopics {
			assert.True(t, topicMap[expected], "Core topic %s should be included", expected)
		}
	})

	t.Run("GetTopicType", func(t *testing.T) {
		testCases := []struct {
			topic        string
			expectedType string
		}{
			{"events/vouch", "event"},
			{"events/report", "event"},
			{"events/appeal", "event"},
			{"revocations/announce", "revocation"},
			{"rules/active", "rules"},
			{"checkpoints/epoch", "checkpoint"},
			{"blobs/QmHash123", "blob"},
			{"invalid/topic", "unknown"},
			{"", "unknown"},
		}

		for _, tc := range testCases {
			actualType := tm.GetTopicType(tc.topic)
			assert.Equal(t, tc.expectedType, actualType, 
				"Topic %s should have type %s, got %s", tc.topic, tc.expectedType, actualType)
		}
	})

	t.Run("ValidateTopicMessage", func(t *testing.T) {
		t.Run("ValidMessages", func(t *testing.T) {
			validCases := []struct {
				topic   string
				data    []byte
			}{
				{"events/vouch", []byte(`{"type":"vouch","from":"did:key:z123","to":"did:key:z456"}`)},
				{"events/report", make([]byte, 1024)}, // 1KB message
				{"blobs/QmHash", make([]byte, 15*1024)}, // 15KB message (under 16KB limit)
				{"checkpoints/epoch", []byte(`{"epoch":100,"root":"hash"}`)},
			}

			for _, tc := range validCases {
				err := tm.ValidateTopicMessage(tc.topic, tc.data)
				assert.NoError(t, err, "Valid message for topic %s should pass validation", tc.topic)
			}
		})

		t.Run("InvalidMessages", func(t *testing.T) {
			invalidCases := []struct {
				topic string
				data  []byte
				desc  string
			}{
				{"events/vouch", nil, "nil data"},
				{"events/vouch", []byte{}, "empty data"},
				{"events/vouch", make([]byte, 17*1024), "message too large (17KB)"},
				{"invalid/topic", []byte("test"), "invalid topic"},
			}

			for _, tc := range invalidCases {
				err := tm.ValidateTopicMessage(tc.topic, tc.data)
				assert.Error(t, err, "Invalid case should fail: %s", tc.desc)
			}
		})
	})

	t.Run("BlobTopics", func(t *testing.T) {
		blobTopics := []string{
			"blobs/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
			"blobs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			"blobs/QmHash123", // Simplified for testing
		}

		for _, topic := range blobTopics {
			assert.True(t, tm.IsValidTopic(topic), "Blob topic %s should be valid", topic)
			assert.Equal(t, "blob", tm.GetTopicType(topic))
		}
	})

	t.Run("EventSubtypes", func(t *testing.T) {
		eventTypes := []string{"vouch", "report", "appeal"}
		
		for _, eventType := range eventTypes {
			topic := "events/" + eventType
			assert.True(t, tm.IsValidTopic(topic), "Event topic %s should be valid", topic)
			assert.Equal(t, "event", tm.GetTopicType(topic))
		}
	})

	t.Run("CaseSensitivity", func(t *testing.T) {
		// Topics should be case-sensitive and only lowercase is valid
		caseCases := []struct {
			topic   string
			valid   bool
		}{
			{"events/vouch", true},   // Correct case
			{"Events/vouch", false},  // Capital E
			{"events/Vouch", false},  // Capital V
			{"EVENTS/VOUCH", false},  // All caps
			{"Events/Vouch", false},  // Mixed case
		}

		for _, tc := range caseCases {
			result := tm.IsValidTopic(tc.topic)
			assert.Equal(t, tc.valid, result, 
				"Topic %s case sensitivity: expected %v, got %v", tc.topic, tc.valid, result)
		}
	})

	t.Run("TopicParsing", func(t *testing.T) {
		// Test edge cases in topic parsing
		edgeCases := []struct {
			topic string
			valid bool
		}{
			{"events/vouch", true},
			{"events/vouch/", false}, // Trailing slash
			{"/events/vouch", false}, // Leading slash
			{"events//vouch", false}, // Double slash
			{"events/vouch ", false}, // Trailing space
			{" events/vouch", false}, // Leading space
			{"events\tvouch", false}, // Tab character
			{"events\nvouch", false}, // Newline character
		}

		for _, tc := range edgeCases {
			result := tm.IsValidTopic(tc.topic)
			assert.Equal(t, tc.valid, result, 
				"Edge case topic '%s': expected %v, got %v", tc.topic, tc.valid, result)
		}
	})

	t.Run("MessageSizeLimits", func(t *testing.T) {
		// Test message size validation
		maxSize := 16 * 1024 // 16KB as defined in config

		sizeCases := []struct {
			size  int
			valid bool
		}{
			{1, true},           // 1 byte
			{1024, true},        // 1KB
			{maxSize - 1, true}, // Just under limit
			{maxSize, true},     // Exactly at limit
			{maxSize + 1, false}, // Just over limit
			{32 * 1024, false},  // 32KB - definitely over
		}

		for _, tc := range sizeCases {
			data := make([]byte, tc.size)
			err := tm.ValidateTopicMessage("events/vouch", data)
			
			if tc.valid {
				assert.NoError(t, err, "Message of size %d should be valid", tc.size)
			} else {
				assert.Error(t, err, "Message of size %d should be invalid", tc.size)
			}
		}
	})
}

func TestTopicManagerConcurrency(t *testing.T) {
	tm := NewTopicManager()
	
	// Test concurrent access to topic validation
	t.Run("ConcurrentValidation", func(t *testing.T) {
		topics := []string{
			"events/vouch",
			"events/report",
			"blobs/QmHash123",
			"checkpoints/latest",
			"invalid/topic",
		}

		// Run validation concurrently
		done := make(chan bool, len(topics)*10)
		
		for i := 0; i < 10; i++ {
			for _, topic := range topics {
				go func(t string) {
					// Should not panic or race
					tm.IsValidTopic(t)
					tm.GetTopicType(t)
					done <- true
				}(topic)
			}
		}

		// Wait for all goroutines
		for i := 0; i < len(topics)*10; i++ {
			<-done
		}
	})

	t.Run("ConcurrentMessageValidation", func(t *testing.T) {
		data := []byte("test message")
		topics := []string{
			"events/vouch",
			"events/report",
			"blobs/QmHash123",
		}

		done := make(chan bool, len(topics)*5)

		for i := 0; i < 5; i++ {
			for _, topic := range topics {
				go func(t string) {
					// Should not panic or race
					tm.ValidateTopicMessage(t, data)
					done <- true
				}(topic)
			}
		}

		// Wait for all goroutines
		for i := 0; i < len(topics)*5; i++ {
			<-done
		}
	})
}

func BenchmarkTopicManager(b *testing.B) {
	tm := NewTopicManager()

	b.Run("IsValidTopic", func(b *testing.B) {
		topics := []string{
			"events/vouch",
			"events/report",
			"blobs/QmHash123",
			"invalid/topic",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			topic := topics[i%len(topics)]
			tm.IsValidTopic(topic)
		}
	})

	b.Run("GetTopicType", func(b *testing.B) {
		topics := []string{
			"events/vouch",
			"revocations/announce",
			"blobs/QmHash123",
			"checkpoints/latest",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			topic := topics[i%len(topics)]
			tm.GetTopicType(topic)
		}
	})

	b.Run("ValidateTopicMessage", func(b *testing.B) {
		data := []byte(`{"type":"vouch","from":"did:key:z123","to":"did:key:z456"}`)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tm.ValidateTopicMessage("events/vouch", data)
		}
	})

	b.Run("GetCoreTopics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tm.GetCoreTopics()
		}
	})
}