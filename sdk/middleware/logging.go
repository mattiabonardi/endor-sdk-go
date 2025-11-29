package middleware

import (
	"fmt"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// LoggingMiddleware provides request/response logging with correlation ID support.
// It logs request start, completion, and performance metrics.
type LoggingMiddleware struct {
	Logger           interfaces.LoggerInterface // For structured logging
	IncludeHeaders   bool                       // Whether to log request headers
	IncludePayload   bool                       // Whether to log request payload (be careful with sensitive data)
	CorrelationIDKey string                     // Header key for correlation ID (default: "x-correlation-id")
}

// NewLoggingMiddleware creates a new logging middleware with dependency injection.
func NewLoggingMiddleware(deps LoggingMiddlewareDependencies) *LoggingMiddleware {
	correlationKey := deps.CorrelationIDKey
	if correlationKey == "" {
		correlationKey = "x-correlation-id" // Default
	}

	return &LoggingMiddleware{
		Logger:           deps.Logger,
		IncludeHeaders:   deps.IncludeHeaders,
		IncludePayload:   deps.IncludePayload,
		CorrelationIDKey: correlationKey,
	}
}

// LoggingMiddlewareDependencies contains all required dependencies for LoggingMiddleware.
type LoggingMiddlewareDependencies struct {
	Logger           interfaces.LoggerInterface
	IncludeHeaders   bool
	IncludePayload   bool
	CorrelationIDKey string // Optional, defaults to "x-correlation-id"
}

// Before logs request start and sets up correlation ID.
func (l *LoggingMiddleware) Before(ctx interface{}) error {
	middlewareCtx, ok := ctx.(*middlewareContext)
	if !ok {
		return fmt.Errorf("logging middleware requires middleware context, got %T", ctx)
	}

	// Extract request information
	headers := extractHeaders(middlewareCtx.ginContext)
	if headers == nil {
		return fmt.Errorf("logging middleware could not access request headers")
	}

	// Get or generate correlation ID
	correlationID := getHeader(headers, l.CorrelationIDKey)
	if correlationID == "" {
		correlationID = generateCorrelationID()
		// Set correlation ID back to context for downstream use
		setContextValue(middlewareCtx.ginContext, "correlationID", correlationID)
	}

	// Log request start
	logFields := []interface{}{
		"event", "request_start",
		"correlation_id", correlationID,
		"user_id", getHeader(headers, "x-user-id"),
	}

	// Add headers if requested
	if l.IncludeHeaders {
		logFields = append(logFields, "headers", headers)
	}

	// Store start time for duration calculation
	setContextValue(middlewareCtx.ginContext, "request_start_time", time.Now())

	l.Logger.Info("Request started", logFields...)

	return nil
}

// After logs request completion with performance metrics.
func (l *LoggingMiddleware) After(ctx interface{}, response interface{}) error {
	middlewareCtx, ok := ctx.(*middlewareContext)
	if !ok {
		return fmt.Errorf("logging middleware requires middleware context, got %T", ctx)
	}

	// Extract request information
	headers := extractHeaders(middlewareCtx.ginContext)
	if headers == nil {
		return fmt.Errorf("logging middleware could not access request headers for completion logging")
	}

	// Get correlation ID and timing info
	correlationID := getContextValue(middlewareCtx.ginContext, "correlationID")
	if correlationID == nil {
		correlationID = getHeader(headers, l.CorrelationIDKey)
	}

	startTimeInterface := getContextValue(middlewareCtx.ginContext, "request_start_time")
	var duration time.Duration
	if startTime, ok := startTimeInterface.(time.Time); ok {
		duration = time.Since(startTime)
	}

	// Log request completion
	logFields := []interface{}{
		"event", "request_complete",
		"correlation_id", correlationID,
		"user_id", getHeader(headers, "x-user-id"),
		"duration_ms", float64(duration.Nanoseconds()) / 1e6,
	}

	// Add response information if available
	if response != nil && l.IncludePayload {
		logFields = append(logFields, "response_type", fmt.Sprintf("%T", response))
	}

	l.Logger.Info("Request completed", logFields...)

	return nil
}

// generateCorrelationID creates a simple correlation ID for request tracing.
// In production, this might use UUIDs or other more sophisticated approaches.
func generateCorrelationID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// getContextValue gets a value from the context using reflection
func getContextValue(ctx interface{}, key string) interface{} {
	// This would use reflection to call Get method if available (gin.Context.Get)
	// For now, return nil - this would be implemented with reflection like setContextValue
	return nil
}
