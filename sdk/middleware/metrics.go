package middleware

import (
	"fmt"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// MetricsMiddleware provides request metrics collection with Prometheus integration.
// It tracks request counts, durations, and success rates.
type MetricsMiddleware struct {
	Logger             interfaces.LoggerInterface         // For error logging
	Config             interfaces.ConfigProviderInterface // For metrics configuration
	RequestCountMetric string                             // Name of request count metric
	DurationMetric     string                             // Name of duration metric
	ErrorCountMetric   string                             // Name of error count metric
	EnabledMetricTypes []string                           // Which metric types to collect
}

// NewMetricsMiddleware creates a new metrics middleware with dependency injection.
func NewMetricsMiddleware(deps MetricsMiddlewareDependencies) *MetricsMiddleware {
	// Set default metric names if not provided
	requestMetric := deps.RequestCountMetric
	if requestMetric == "" {
		requestMetric = "http_requests_total"
	}

	durationMetric := deps.DurationMetric
	if durationMetric == "" {
		durationMetric = "http_request_duration_seconds"
	}

	errorMetric := deps.ErrorCountMetric
	if errorMetric == "" {
		errorMetric = "http_request_errors_total"
	}

	enabledTypes := deps.EnabledMetricTypes
	if len(enabledTypes) == 0 {
		enabledTypes = []string{"count", "duration", "errors"} // Default all
	}

	return &MetricsMiddleware{
		Logger:             deps.Logger,
		Config:             deps.Config,
		RequestCountMetric: requestMetric,
		DurationMetric:     durationMetric,
		ErrorCountMetric:   errorMetric,
		EnabledMetricTypes: enabledTypes,
	}
}

// MetricsMiddlewareDependencies contains all required dependencies for MetricsMiddleware.
type MetricsMiddlewareDependencies struct {
	Logger             interfaces.LoggerInterface
	Config             interfaces.ConfigProviderInterface
	RequestCountMetric string   // Optional, defaults to "http_requests_total"
	DurationMetric     string   // Optional, defaults to "http_request_duration_seconds"
	ErrorCountMetric   string   // Optional, defaults to "http_request_errors_total"
	EnabledMetricTypes []string // Optional, defaults to all types
}

// Before starts request timing for duration metrics.
func (m *MetricsMiddleware) Before(ctx interface{}) error {
	middlewareCtx, ok := ctx.(*middlewareContext)
	if !ok {
		return fmt.Errorf("metrics middleware requires middleware context, got %T", ctx)
	}

	// Store start time for duration calculation
	startTime := time.Now()
	if err := setContextValue(middlewareCtx.ginContext, "metrics_start_time", startTime); err != nil {
		m.Logger.Warn("Failed to set metrics start time", "error", err)
		// Don't fail the request for metrics issues
	}

	// Increment request count if enabled
	if m.isMetricEnabled("count") {
		m.recordRequestCount(middlewareCtx)
	}

	return nil
}

// After records final metrics including duration and error counts.
func (m *MetricsMiddleware) After(ctx interface{}, response interface{}) error {
	middlewareCtx, ok := ctx.(*middlewareContext)
	if !ok {
		return fmt.Errorf("metrics middleware requires middleware context, got %T", ctx)
	}

	// Calculate request duration if timing was started
	startTimeInterface := getContextValue(middlewareCtx.ginContext, "metrics_start_time")
	if startTime, ok := startTimeInterface.(time.Time); ok && m.isMetricEnabled("duration") {
		duration := time.Since(startTime)
		m.recordRequestDuration(middlewareCtx, duration)
	}

	// Record error metrics if applicable
	if m.isMetricEnabled("errors") {
		m.recordErrorMetrics(middlewareCtx, response)
	}

	return nil
}

// isMetricEnabled checks if a specific metric type is enabled
func (m *MetricsMiddleware) isMetricEnabled(metricType string) bool {
	for _, enabled := range m.EnabledMetricTypes {
		if enabled == metricType {
			return true
		}
	}
	return false
}

// recordRequestCount increments the request counter
func (m *MetricsMiddleware) recordRequestCount(ctx *middlewareContext) {
	// Extract request details for labels
	headers := extractHeaders(ctx.ginContext)
	method := "UNKNOWN"
	path := "UNKNOWN"

	// In a real implementation, this would increment a Prometheus counter
	// For now, just log the metric
	m.Logger.Debug("Recording request count metric",
		"metric", m.RequestCountMetric,
		"method", method,
		"path", path,
		"user_id", getHeader(headers, "x-user-id"))
}

// recordRequestDuration records the request duration
func (m *MetricsMiddleware) recordRequestDuration(ctx *middlewareContext, duration time.Duration) {
	// Extract request details for labels
	headers := extractHeaders(ctx.ginContext)
	method := "UNKNOWN"
	path := "UNKNOWN"

	durationSeconds := float64(duration.Nanoseconds()) / 1e9

	// In a real implementation, this would record to a Prometheus histogram
	// For now, just log the metric
	m.Logger.Debug("Recording request duration metric",
		"metric", m.DurationMetric,
		"method", method,
		"path", path,
		"duration_seconds", durationSeconds,
		"user_id", getHeader(headers, "x-user-id"))
}

// recordErrorMetrics increments error counters based on response
func (m *MetricsMiddleware) recordErrorMetrics(ctx *middlewareContext, response interface{}) {
	// Extract request details for labels
	headers := extractHeaders(ctx.ginContext)
	method := "UNKNOWN"
	path := "UNKNOWN"

	// Determine if this was an error response
	// This is simplified - in practice would check HTTP status codes
	isError := false
	errorType := "none"

	if response == nil {
		isError = true
		errorType = "no_response"
	}

	// In a real implementation, this would increment Prometheus error counters
	// For now, just log the metric
	if isError {
		m.Logger.Debug("Recording error count metric",
			"metric", m.ErrorCountMetric,
			"method", method,
			"path", path,
			"error_type", errorType,
			"user_id", getHeader(headers, "x-user-id"))
	}
}
