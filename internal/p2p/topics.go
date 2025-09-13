package p2p

import (
	"fmt"
	"regexp"
	"strings"
)

// Topic names as defined in the architecture
const (
	// Event topics
	TopicEventVouch  = "events/vouch"
	TopicEventReport = "events/report" 
	TopicEventAppeal = "events/appeal"
	
	// Revocation topics (pattern-based)
	TopicRevocationPrefix = "revocations/"
	
	// Rules topics
	TopicRulesActive = "rules/active"
	
	// Checkpoint topics
	TopicCheckpointsEpoch = "checkpoints/epoch"
	
	// Blob topics (pattern-based)
	TopicBlobPrefix = "blobs/"
)

var (
	// Valid topic patterns
	eventTopicRegex      = regexp.MustCompile(`^events/(vouch|report|appeal)$`)
	revocationTopicRegex = regexp.MustCompile(`^revocations/[a-zA-Z0-9._-]+$`)
	rulesTopicRegex      = regexp.MustCompile(`^rules/(active|[a-zA-Z0-9._-]+)$`)
	checkpointTopicRegex = regexp.MustCompile(`^checkpoints/(epoch|[0-9]+)$`)
	blobTopicRegex       = regexp.MustCompile(`^blobs/[a-zA-Z0-9]+$`)
)

// TopicManager manages topic subscriptions and validation
type TopicManager struct {
	validTopics map[string]bool
}

// NewTopicManager creates a new topic manager
func NewTopicManager() *TopicManager {
	return &TopicManager{
		validTopics: make(map[string]bool),
	}
}

// IsValidTopic checks if a topic name is valid
func (tm *TopicManager) IsValidTopic(topic string) bool {
	if topic == "" {
		return false
	}
	
	// Check against known patterns
	switch {
	case eventTopicRegex.MatchString(topic):
		return true
	case revocationTopicRegex.MatchString(topic):
		return true
	case rulesTopicRegex.MatchString(topic):
		return true
	case checkpointTopicRegex.MatchString(topic):
		return true
	case blobTopicRegex.MatchString(topic):
		return true
	default:
		return false
	}
}

// GetTopicType returns the category of a topic
func (tm *TopicManager) GetTopicType(topic string) string {
	switch {
	case strings.HasPrefix(topic, "events/"):
		return "event"
	case strings.HasPrefix(topic, "revocations/"):
		return "revocation"
	case strings.HasPrefix(topic, "rules/"):
		return "rules"
	case strings.HasPrefix(topic, "checkpoints/"):
		return "checkpoint"
	case strings.HasPrefix(topic, "blobs/"):
		return "blob"
	default:
		return "unknown"
	}
}

// GetEventTopics returns all event-related topics
func (tm *TopicManager) GetEventTopics() []string {
	return []string{
		TopicEventVouch,
		TopicEventReport,
		TopicEventAppeal,
	}
}

// GetCoreTopics returns the core system topics
func (tm *TopicManager) GetCoreTopics() []string {
	return []string{
		TopicEventVouch,
		TopicEventReport,
		TopicEventAppeal,
		TopicRulesActive,
		TopicCheckpointsEpoch,
	}
}

// ValidateTopicMessage performs basic validation on a topic message
func (tm *TopicManager) ValidateTopicMessage(topic string, data []byte) error {
	if !tm.IsValidTopic(topic) {
		return fmt.Errorf("invalid topic: %s", topic)
	}
	
	if len(data) == 0 {
		return fmt.Errorf("empty message data")
	}
	
	// Check size limits based on topic type
	topicType := tm.GetTopicType(topic)
	maxSize := tm.getMaxMessageSize(topicType)
	
	if len(data) > maxSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(data), maxSize)
	}
	
	return nil
}

// getMaxMessageSize returns the maximum message size for a topic type
func (tm *TopicManager) getMaxMessageSize(topicType string) int {
	switch topicType {
	case "event":
		return 16 * 1024 // 16KB for events
	case "checkpoint":
		return 8 * 1024  // 8KB for checkpoints
	case "rules":
		return 32 * 1024 // 32KB for rulesets
	case "revocation":
		return 4 * 1024  // 4KB for revocations
	case "blob":
		return 1024 * 1024 // 1MB for blobs (though these should be fetched separately)
	default:
		return 16 * 1024 // Default 16KB
	}
}

// GetTopicPriority returns the priority level for a topic (higher = more important)
func (tm *TopicManager) GetTopicPriority(topic string) int {
	switch {
	case strings.HasPrefix(topic, "checkpoints/"):
		return 10 // Highest priority - consensus critical
	case strings.HasPrefix(topic, "rules/"):
		return 9  // High priority - system parameters
	case strings.HasPrefix(topic, "events/"):
		return 5  // Medium priority - user events
	case strings.HasPrefix(topic, "revocations/"):
		return 7  // High priority - security critical
	case strings.HasPrefix(topic, "blobs/"):
		return 1  // Lowest priority - content
	default:
		return 3  // Default medium-low priority
	}
}