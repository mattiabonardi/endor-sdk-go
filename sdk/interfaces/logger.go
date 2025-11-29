// Package interfaces provides logger interfaces for endor-sdk-go framework.
// These interfaces enable dependency injection, testing with mocks, and flexible logging
// implementations while maintaining consistent logging patterns across the framework.
package interfaces

// LoggerInterface defines the contract for logging in the endor-sdk-go framework.
// This interface abstracts logging operations, enabling testing with different
// log levels and outputs while supporting both structured and simple logging patterns.
//
// The interface follows standard logging practices while making logging testable
// and mockable for unit testing scenarios where log verification is needed.
//
// Example usage:
//
//	// Production usage with concrete implementation
//	logger := sdk.NewLogger() // implements LoggerInterface
//	logger.Info("Service started", "port", "8080", "service", "endor-api")
//	logger.Error("Database connection failed", "error", err, "database", "mongodb")
//
//	// Testing usage with mock logger
//	mockLogger := &MockLogger{}
//	mockLogger.On("Info").Return()
//	mockLogger.On("Error", "message", mock.Any, "error", mock.Any).Return()
//
//	// Test logger for verification
//	testLogger := &TestLogger{}
//	service := NewEndorServiceWithDeps(repo, config, testLogger, ctx)
//	// ... perform operations
//	testLogger.AssertInfoCalled(t, "Expected log message")
type LoggerInterface interface {
	// Debug logs a debug-level message with optional key-value pairs.
	// Debug messages are typically used for detailed diagnostic information
	// that is only of interest when diagnosing problems.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info-level message with optional key-value pairs.
	// Info messages are typically used for general operational entries
	// about what's happening inside the application.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning-level message with optional key-value pairs.
	// Warning messages are typically used for events that should be looked into
	// but don't necessarily represent errors.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error-level message with optional key-value pairs.
	// Error messages are typically used for events that indicate something
	// went wrong but the application can continue to operate.
	Error(msg string, keysAndValues ...interface{})

	// Fatal logs a fatal-level message with optional key-value pairs.
	// Fatal messages indicate severe error conditions that require
	// immediate attention and may cause application termination.
	Fatal(msg string, keysAndValues ...interface{})

	// With creates a new logger instance with additional key-value pairs
	// that will be included in all subsequent log entries from that logger.
	// This is useful for adding context like request IDs, user IDs, etc.
	With(keysAndValues ...interface{}) LoggerInterface

	// WithName creates a new logger instance with a specific name/component identifier.
	// This helps organize log entries by component or service area.
	WithName(name string) LoggerInterface
}

// StructuredLoggerInterface extends LoggerInterface for implementations that support
// structured logging with rich context and metadata.
type StructuredLoggerInterface interface {
	LoggerInterface

	// LogWithContext logs a message with rich contextual information.
	// This method supports complex structured data and metadata.
	LogWithContext(level LogLevel, msg string, context map[string]interface{})

	// SetLevel configures the minimum log level for this logger instance.
	// Messages below this level will be filtered out.
	SetLevel(level LogLevel)

	// GetLevel returns the current minimum log level for this logger.
	GetLevel() LogLevel
}

// LogLevel represents the severity level of a log entry.
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
