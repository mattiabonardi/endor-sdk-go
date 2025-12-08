package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogType defines the output format of logs
type LogType string

const (
	StringLog LogType = "STRING"
	JSONLog   LogType = "JSON"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogContext holds contextual information for logging
type LogContext struct {
	UserID      string `json:"user_id"`
	UserSession string `json:"user_session"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// Config defines the logger configuration
type Config struct {
	LogType LogType
}

// Logger is the main logger instance
type Logger struct {
	config    Config
	context   LogContext
	stdLogger *log.Logger
}

// LogEntry represents a structured log entry for JSON output
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	UserID      string                 `json:"user_id,omitempty"`
	UserSession string                 `json:"user_session,omitempty"`
	Request     string                 `json:"resource"`
	Action      string                 `json:"action"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// NewLogger creates a new logger instance with the given configuration and context
func NewLogger(cfg Config, ctx LogContext) *Logger {
	return &Logger{
		config:    cfg,
		context:   ctx,
		stdLogger: log.New(os.Stdout, "", 0),
	}
}

// SetContext updates the logger's context
func (l *Logger) SetContext(ctx LogContext) {
	l.context = ctx
}

// GetContext returns the current logger context
func (l *Logger) GetContext() LogContext {
	return l.context
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, msg string, extra map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)

	switch l.config.LogType {
	case JSONLog:
		l.logJSON(timestamp, level, msg, extra)
	case StringLog:
		l.logString(timestamp, level, msg, extra)
	default:
		l.logString(timestamp, level, msg, extra)
	}
}

// logJSON outputs log in JSON format
func (l *Logger) logJSON(timestamp string, level LogLevel, msg string, extra map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   timestamp,
		Level:       level,
		Message:     msg,
		UserID:      l.context.UserID,
		UserSession: l.context.UserSession,
		Request:     l.context.Resource,
		Action:      l.context.Action,
		Extra:       extra,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		l.stdLogger.Printf("ERROR: Failed to marshal log entry: %v", err)
		return
	}

	l.stdLogger.Println(string(jsonData))
}

// logString outputs log in string format
func (l *Logger) logString(timestamp string, level LogLevel, msg string, extra map[string]interface{}) {
	output := fmt.Sprintf("[%s] %s | %s", timestamp, level, msg)

	if l.context.UserID != "" {
		output += fmt.Sprintf(" | user_id=%s", l.context.UserID)
	}

	if l.context.UserSession != "" {
		output += fmt.Sprintf(" | user_session=%s", l.context.UserSession)
	}

	if l.context.Resource != "" {
		output += fmt.Sprintf(" | resource=%s", l.context.Resource)
	}

	if l.context.Action != "" {
		output += fmt.Sprintf(" | action=%s", l.context.Action)
	}

	if len(extra) > 0 {
		for k, v := range extra {
			output += fmt.Sprintf(" | %s=%v", k, v)
		}
	}

	l.stdLogger.Println(output)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg, nil)
}

// DebugWithFields logs a debug message with additional fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(INFO, msg, nil)
}

// InfoWithFields logs an info message with additional fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WARN, msg, nil)
}

// WarnWithFields logs a warning message with additional fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(ERROR, msg, nil)
}

// ErrorWithFields logs an error message with additional fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields)
}

// ErrorWithStackTrace logs an error with stack trace
func (l *Logger) ErrorWithStackTrace(err error) {
	if err == nil {
		return
	}

	stackTrace := captureStackTrace(3) // Skip 3 frames: captureStackTrace, ErrorWithStackTrace, and the runtime

	fields := map[string]interface{}{
		"error":       err.Error(),
		"stack_trace": stackTrace,
	}

	l.log(ERROR, "Error occurred", fields)
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) []string {
	const maxStackDepth = 32
	pcs := make([]uintptr, maxStackDepth)
	n := runtime.Callers(skip, pcs)

	frames := runtime.CallersFrames(pcs[:n])
	var stackTrace []string

	for {
		frame, more := frames.Next()

		// Format: function at file:line
		trace := fmt.Sprintf("%s at %s:%d", frame.Function, shortenPath(frame.File), frame.Line)
		stackTrace = append(stackTrace, trace)

		if !more {
			break
		}
	}

	return stackTrace
}

// shortenPath shortens the file path to make it more readable
func shortenPath(path string) string {
	// Try to find common base paths and shorten them
	if idx := strings.LastIndex(path, "/src/"); idx != -1 {
		return path[idx+5:]
	}
	if idx := strings.LastIndex(path, "/pkg/mod/"); idx != -1 {
		return path[idx+9:]
	}
	// Return last 2 path components if possible
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return path
}
