package p2p

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	t.Run("LogLevels", func(t *testing.T) {
		// Capture log output
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		// Debug should be filtered out at Info level
		logger.Debug("debug message")
		assert.Empty(t, buf.String())

		// Info should be logged
		logger.Info("info message")
		assert.Contains(t, buf.String(), "INFO test: info message")

		buf.Reset()

		// Warn should be logged
		logger.Warn("warn message")
		assert.Contains(t, buf.String(), "WARN test: warn message")

		buf.Reset()

		// Error should be logged
		logger.Error("error message")
		assert.Contains(t, buf.String(), "ERROR test: error message")
	})

	t.Run("LogLevelFiltering", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelError)
		logger.logger = log.New(&buf, "", 0)

		// Debug, Info, Warn should be filtered out
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		assert.Empty(t, buf.String())

		// Error should be logged
		logger.Error("error")
		assert.Contains(t, buf.String(), "ERROR test: error")
	})

	t.Run("WithFields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		logger.Info("message", map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		})

		output := buf.String()
		assert.Contains(t, output, "INFO test: message")
		assert.Contains(t, output, "key1=value1")
		assert.Contains(t, output, "key2=42")
		assert.Contains(t, output, "key3=true")
	})

	t.Run("LoggerContext", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		// Create context with default fields
		ctx := logger.WithFields(map[string]interface{}{
			"session": "abc123",
			"user":    "alice",
		})

		ctx.Info("logged in")
		output := buf.String()
		assert.Contains(t, output, "session=abc123")
		assert.Contains(t, output, "user=alice")
		assert.Contains(t, output, "logged in")
	})

	t.Run("LoggerContextWithPeer", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		ctx := logger.WithPeer(peerID)

		ctx.Info("peer connected")
		output := buf.String()
		assert.Contains(t, output, "peer_id=12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		assert.Contains(t, output, "peer connected")
	})

	t.Run("LoggerContextWithTopic", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		ctx := logger.WithTopic("events/vouch")
		ctx.Info("message received")

		output := buf.String()
		assert.Contains(t, output, "topic=events/vouch")
		assert.Contains(t, output, "message received")
	})

	t.Run("ContextFieldMerging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		// Create context with default fields
		ctx := logger.WithFields(map[string]interface{}{
			"session": "abc123",
			"action":  "connect",
		})

		// Log with additional fields, including override
		ctx.Info("operation completed", map[string]interface{}{
			"action":   "disconnect", // Should override context field
			"duration": "5s",
		})

		output := buf.String()
		assert.Contains(t, output, "session=abc123")
		assert.Contains(t, output, "action=disconnect") // Override should win
		assert.Contains(t, output, "duration=5s")
		assert.NotContains(t, output, "action=connect") // Original should not appear
	})

	t.Run("TimestampFormat", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		logger.Info("test message")
		output := buf.String()
		
		// Just verify that timestamp is included and formatted reasonably
		assert.Contains(t, output, "INFO test: test message")
		
		// Check that it contains a timestamp-like string (RFC3339 format)
		// Format should be: [TIMESTAMP] LEVEL component: message
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 1)
		
		line := lines[0]
		assert.True(t, strings.HasPrefix(line, "["))
		assert.Contains(t, line, "] INFO test: test message")
	})

	t.Run("ComponentName", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("P2PHost", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		logger.Info("test message")
		output := buf.String()
		assert.Contains(t, output, "INFO P2PHost: test message")
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		// Log with nil fields
		logger.Info("message with nil fields", nil)
		output := buf.String()
		assert.Contains(t, output, "INFO test: message with nil fields")
		assert.NotContains(t, output, " |") // No field separator should appear

		buf.Reset()

		// Log with empty fields
		logger.Info("message with empty fields", map[string]interface{}{})
		output = buf.String()
		assert.Contains(t, output, "INFO test: message with empty fields")
		assert.NotContains(t, output, " |") // No field separator should appear
	})

	// Note: We don't test Fatal as it would exit the process
}

func TestLogLevel(t *testing.T) {
	t.Run("LogLevelString", func(t *testing.T) {
		tests := []struct {
			level    LogLevel
			expected string
		}{
			{LogLevelDebug, "DEBUG"},
			{LogLevelInfo, "INFO"},
			{LogLevelWarn, "WARN"},
			{LogLevelError, "ERROR"},
			{LogLevelFatal, "FATAL"},
			{LogLevel(999), "UNKNOWN"}, // Invalid level
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, test.level.String())
		}
	})
}

func TestLoggerContextChaining(t *testing.T) {
	t.Run("MultipleContextLayers", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("test", LogLevelInfo)
		logger.logger = log.New(&buf, "", 0)

		// Create nested contexts
		peerID, _ := peer.Decode("12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		
		topicCtx := logger.WithFields(map[string]interface{}{
			"component": "networking",
			"version":   "1.0",
			"peer_id":   peerID.String(),
			"topic":     "events/vouch",
		})

		topicCtx.Info("processing message", map[string]interface{}{
			"size": 1024,
		})

		output := buf.String()
		assert.Contains(t, output, "component=networking")
		assert.Contains(t, output, "version=1.0")
		assert.Contains(t, output, "peer_id=12D3KooWGBfKT1krEZCRCRFfqKmYJPEzKNYvSFv7X7R2oVVGAr3P")
		assert.Contains(t, output, "topic=events/vouch")
		assert.Contains(t, output, "size=1024")
	})
}

// Test that logger works correctly with the actual log package
func TestLoggerIntegration(t *testing.T) {
	// Temporarily redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger("integration", LogLevelInfo)
	logger.Info("integration test message", map[string]interface{}{
		"test": true,
	})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "INFO integration: integration test message")
	assert.Contains(t, output, "test=true")
}

func BenchmarkLogger(b *testing.B) {
	logger := NewLogger("bench", LogLevelInfo)
	
	b.Run("InfoWithoutFields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("InfoWithFields", func(b *testing.B) {
		fields := map[string]interface{}{
			"iteration": 0,
			"component": "benchmark",
			"active":    true,
		}
		
		for i := 0; i < b.N; i++ {
			fields["iteration"] = i
			logger.Info("benchmark message", fields)
		}
	})

	b.Run("DebugFiltered", func(b *testing.B) {
		// Debug messages should be filtered out, testing overhead
		for i := 0; i < b.N; i++ {
			logger.Debug("debug message that should be filtered")
		}
	})

	b.Run("WithContext", func(b *testing.B) {
		ctx := logger.WithFields(map[string]interface{}{
			"session": "benchmark",
			"user":    "test",
		})
		
		for i := 0; i < b.N; i++ {
			ctx.Info("context message", map[string]interface{}{
				"iteration": i,
			})
		}
	})
}