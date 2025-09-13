package p2p

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// LogLevel represents logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging for P2P components
type Logger struct {
	component string
	level     LogLevel
	logger    *log.Logger
}

// NewLogger creates a new logger for a component
func NewLogger(component string, level LogLevel) *Logger {
	return &Logger{
		component: component,
		level:     level,
		logger:    log.New(os.Stdout, "", 0),
	}
}

// shouldLog checks if message should be logged at current level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// formatMessage formats log message with timestamp and component
func (l *Logger) formatMessage(level LogLevel, msg string, fields map[string]interface{}) string {
	timestamp := time.Now().Format(time.RFC3339)
	formatted := fmt.Sprintf("[%s] %s %s: %s", 
		timestamp, level.String(), l.component, msg)
	
	// Add structured fields
	if len(fields) > 0 {
		formatted += " |"
		for key, value := range fields {
			formatted += fmt.Sprintf(" %s=%v", key, value)
		}
	}
	
	return formatted
}

// Debug logs debug message with optional fields
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	if !l.shouldLog(LogLevelDebug) {
		return
	}
	
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	
	l.logger.Println(l.formatMessage(LogLevelDebug, msg, fieldMap))
}

// Info logs info message with optional fields
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	if !l.shouldLog(LogLevelInfo) {
		return
	}
	
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	
	l.logger.Println(l.formatMessage(LogLevelInfo, msg, fieldMap))
}

// Warn logs warning message with optional fields
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	if !l.shouldLog(LogLevelWarn) {
		return
	}
	
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	
	l.logger.Println(l.formatMessage(LogLevelWarn, msg, fieldMap))
}

// Error logs error message with optional fields
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	if !l.shouldLog(LogLevelError) {
		return
	}
	
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	
	l.logger.Println(l.formatMessage(LogLevelError, msg, fieldMap))
}

// Fatal logs fatal message and exits
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	
	l.logger.Println(l.formatMessage(LogLevelFatal, msg, fieldMap))
	os.Exit(1)
}

// WithFields creates logger context with default fields
func (l *Logger) WithFields(fields map[string]interface{}) *LoggerContext {
	return &LoggerContext{
		logger: l,
		fields: fields,
	}
}

// WithPeer creates logger context with peer information
func (l *Logger) WithPeer(peerID peer.ID) *LoggerContext {
	return l.WithFields(map[string]interface{}{
		"peer_id": peerID.String(),
	})
}

// WithTopic creates logger context with topic information
func (l *Logger) WithTopic(topic string) *LoggerContext {
	return l.WithFields(map[string]interface{}{
		"topic": topic,
	})
}

// LoggerContext provides logging with predefined fields
type LoggerContext struct {
	logger *Logger
	fields map[string]interface{}
}

// mergeFields merges context fields with additional fields
func (lc *LoggerContext) mergeFields(additional map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// Copy context fields
	for k, v := range lc.fields {
		merged[k] = v
	}
	
	// Add additional fields (overwrites context fields if same key)
	for k, v := range additional {
		merged[k] = v
	}
	
	return merged
}

// Debug logs debug message with context fields
func (lc *LoggerContext) Debug(msg string, fields ...map[string]interface{}) {
	var additional map[string]interface{}
	if len(fields) > 0 {
		additional = fields[0]
	}
	
	merged := lc.mergeFields(additional)
	lc.logger.Debug(msg, merged)
}

// Info logs info message with context fields
func (lc *LoggerContext) Info(msg string, fields ...map[string]interface{}) {
	var additional map[string]interface{}
	if len(fields) > 0 {
		additional = fields[0]
	}
	
	merged := lc.mergeFields(additional)
	lc.logger.Info(msg, merged)
}

// Warn logs warning message with context fields
func (lc *LoggerContext) Warn(msg string, fields ...map[string]interface{}) {
	var additional map[string]interface{}
	if len(fields) > 0 {
		additional = fields[0]
	}
	
	merged := lc.mergeFields(additional)
	lc.logger.Warn(msg, merged)
}

// Error logs error message with context fields
func (lc *LoggerContext) Error(msg string, fields ...map[string]interface{}) {
	var additional map[string]interface{}
	if len(fields) > 0 {
		additional = fields[0]
	}
	
	merged := lc.mergeFields(additional)
	lc.logger.Error(msg, merged)
}

// Fatal logs fatal message with context fields and exits
func (lc *LoggerContext) Fatal(msg string, fields ...map[string]interface{}) {
	var additional map[string]interface{}
	if len(fields) > 0 {
		additional = fields[0]
	}
	
	merged := lc.mergeFields(additional)
	lc.logger.Fatal(msg, merged)
}