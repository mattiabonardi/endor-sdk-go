package debug

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DebugMode contains debug configuration settings
type DebugMode struct {
	// Enabled indicates whether debug mode is active
	Enabled bool

	// LogLevel controls the verbosity of debug logging
	LogLevel LogLevel

	// TraceRequests enables request-level tracing
	TraceRequests bool

	// TraceDependencies enables dependency resolution tracing
	TraceDependencies bool

	// TraceLifecycle enables service lifecycle event tracing
	TraceLifecycle bool

	// CollectMetrics enables performance metrics collection
	CollectMetrics bool

	// MetricsBufferSize controls the size of the metrics buffer
	MetricsBufferSize int

	// LogOutputFile specifies the file for debug logs (empty = stdout)
	LogOutputFile string
}

// LogLevel represents the logging verbosity level
type LogLevel int

const (
	LogLevelOff   LogLevel = iota // No logging
	LogLevelError                 // Only errors
	LogLevelWarn                  // Errors and warnings
	LogLevelInfo                  // Errors, warnings, and info
	LogLevelDebug                 // All messages including debug
	LogLevelTrace                 // Most verbose, includes traces
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelOff:
		return "OFF"
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// DebugLogger handles all debug logging functionality
type DebugLogger struct {
	config *DebugMode
	logger *log.Logger

	// Dependencies tracking
	dependencyTrace map[string]*DependencyTrace
	traceMutex      sync.RWMutex

	// Lifecycle events
	lifecycleEvents []LifecycleEvent
	lifecycleMutex  sync.RWMutex

	// Performance metrics
	metrics      []PerformanceMetric
	metricsMutex sync.RWMutex
}

// DependencyTrace tracks dependency resolution for a specific trace ID
type DependencyTrace struct {
	TraceID         string
	RequestID       string
	StartTime       time.Time
	ResolutionSteps []ResolutionStep
	TotalDuration   time.Duration
	Success         bool
	Error           error
}

// ResolutionStep represents a single step in dependency resolution
type ResolutionStep struct {
	StepID          string
	ServiceType     string
	Dependencies    []string
	StartTime       time.Time
	Duration        time.Duration
	Success         bool
	Error           error
	CacheHit        bool
	ConstructorInfo string
}

// LifecycleEvent represents a service lifecycle event
type LifecycleEvent struct {
	EventID     string
	TraceID     string
	ServiceName string
	ServiceType string
	EventType   LifecycleEventType
	Timestamp   time.Time
	Duration    time.Duration
	Success     bool
	Error       error
	Metadata    map[string]interface{}
}

// LifecycleEventType represents the type of lifecycle event
type LifecycleEventType string

const (
	EventServiceRegistration   LifecycleEventType = "ServiceRegistration"
	EventServiceConstruction   LifecycleEventType = "ServiceConstruction"
	EventServiceInitialization LifecycleEventType = "ServiceInitialization"
	EventServiceStartup        LifecycleEventType = "ServiceStartup"
	EventServiceShutdown       LifecycleEventType = "ServiceShutdown"
	EventServiceDisposal       LifecycleEventType = "ServiceDisposal"
	EventMiddlewareExecution   LifecycleEventType = "MiddlewareExecution"
	EventHTTPRequest           LifecycleEventType = "HTTPRequest"
	EventDatabaseOperation     LifecycleEventType = "DatabaseOperation"
)

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	MetricID    string
	TraceID     string
	Name        string
	Type        MetricType
	Value       float64
	Unit        string
	Timestamp   time.Time
	ServiceName string
	Operation   string
	Tags        map[string]string
}

// MetricType represents the type of performance metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "Counter"
	MetricTypeGauge     MetricType = "Gauge"
	MetricTypeHistogram MetricType = "Histogram"
	MetricTypeTiming    MetricType = "Timing"
)

// NewDebugLogger creates a new debug logger with the specified configuration
func NewDebugLogger(config *DebugMode) (*DebugLogger, error) {
	if config == nil {
		config = &DebugMode{
			Enabled:           false,
			LogLevel:          LogLevelInfo,
			MetricsBufferSize: 1000,
		}
	}

	var logger *log.Logger
	if config.LogOutputFile != "" {
		file, err := os.OpenFile(config.LogOutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open debug log file: %w", err)
		}
		logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	} else {
		logger = log.New(os.Stdout, "[ENDOR-DEBUG] ", log.LstdFlags|log.Lmicroseconds)
	}

	return &DebugLogger{
		config:          config,
		logger:          logger,
		dependencyTrace: make(map[string]*DependencyTrace),
		lifecycleEvents: make([]LifecycleEvent, 0, 100),
		metrics:         make([]PerformanceMetric, 0, config.MetricsBufferSize),
	}, nil
}

// StartDependencyTrace begins tracking dependency resolution for a trace ID
func (dl *DebugLogger) StartDependencyTrace(traceID, requestID string) {
	if !dl.config.Enabled || !dl.config.TraceDependencies {
		return
	}

	dl.traceMutex.Lock()
	defer dl.traceMutex.Unlock()

	dl.dependencyTrace[traceID] = &DependencyTrace{
		TraceID:         traceID,
		RequestID:       requestID,
		StartTime:       time.Now(),
		ResolutionSteps: make([]ResolutionStep, 0),
	}

	if dl.config.LogLevel >= LogLevelTrace {
		dl.logger.Printf("[TRACE] Started dependency trace %s (request: %s)", traceID, requestID)
	}
}

// AddResolutionStep adds a dependency resolution step to the trace
func (dl *DebugLogger) AddResolutionStep(traceID, serviceType string, dependencies []string,
	duration time.Duration, success bool, err error, cacheHit bool, constructorInfo string) {

	if !dl.config.Enabled || !dl.config.TraceDependencies {
		return
	}

	dl.traceMutex.Lock()
	defer dl.traceMutex.Unlock()

	trace, exists := dl.dependencyTrace[traceID]
	if !exists {
		return
	}

	stepID := uuid.New().String()
	step := ResolutionStep{
		StepID:          stepID,
		ServiceType:     serviceType,
		Dependencies:    dependencies,
		StartTime:       time.Now().Add(-duration),
		Duration:        duration,
		Success:         success,
		Error:           err,
		CacheHit:        cacheHit,
		ConstructorInfo: constructorInfo,
	}

	trace.ResolutionSteps = append(trace.ResolutionSteps, step)

	if dl.config.LogLevel >= LogLevelDebug {
		status := "SUCCESS"
		if err != nil {
			status = fmt.Sprintf("ERROR: %v", err)
		}
		cacheStatus := ""
		if cacheHit {
			cacheStatus = " [CACHE HIT]"
		}

		dl.logger.Printf("[DEBUG] [%s] Resolved %s -> deps: %v, duration: %v, status: %s%s",
			traceID, serviceType, dependencies, duration, status, cacheStatus)
	}
}

// EndDependencyTrace completes a dependency resolution trace
func (dl *DebugLogger) EndDependencyTrace(traceID string, success bool, err error) {
	if !dl.config.Enabled || !dl.config.TraceDependencies {
		return
	}

	dl.traceMutex.Lock()
	defer dl.traceMutex.Unlock()

	trace, exists := dl.dependencyTrace[traceID]
	if !exists {
		return
	}

	trace.TotalDuration = time.Since(trace.StartTime)
	trace.Success = success
	trace.Error = err

	if dl.config.LogLevel >= LogLevelInfo {
		status := "SUCCESS"
		if err != nil {
			status = fmt.Sprintf("FAILED: %v", err)
		}

		dl.logger.Printf("[INFO] [%s] Dependency resolution completed: %d steps, duration: %v, status: %s",
			traceID, len(trace.ResolutionSteps), trace.TotalDuration, status)
	}
}

// GetDependencyTrace retrieves a dependency trace by trace ID
func (dl *DebugLogger) GetDependencyTrace(traceID string) (*DependencyTrace, bool) {
	dl.traceMutex.RLock()
	defer dl.traceMutex.RUnlock()

	trace, exists := dl.dependencyTrace[traceID]
	return trace, exists
}

// LogLifecycleEvent records a service lifecycle event
func (dl *DebugLogger) LogLifecycleEvent(serviceName, serviceType string, eventType LifecycleEventType,
	duration time.Duration, success bool, err error, metadata map[string]interface{}) {

	if !dl.config.Enabled || !dl.config.TraceLifecycle {
		return
	}

	eventID := uuid.New().String()
	traceID := getTraceIDFromContext() // Get from current context if available

	event := LifecycleEvent{
		EventID:     eventID,
		TraceID:     traceID,
		ServiceName: serviceName,
		ServiceType: serviceType,
		EventType:   eventType,
		Timestamp:   time.Now(),
		Duration:    duration,
		Success:     success,
		Error:       err,
		Metadata:    metadata,
	}

	dl.lifecycleMutex.Lock()
	dl.lifecycleEvents = append(dl.lifecycleEvents, event)

	// Keep only the last 1000 events to prevent memory issues
	if len(dl.lifecycleEvents) > 1000 {
		dl.lifecycleEvents = dl.lifecycleEvents[len(dl.lifecycleEvents)-1000:]
	}
	dl.lifecycleMutex.Unlock()

	if dl.config.LogLevel >= LogLevelInfo {
		status := "SUCCESS"
		if err != nil {
			status = fmt.Sprintf("FAILED: %v", err)
		}

		dl.logger.Printf("[INFO] [%s] Lifecycle event: %s.%s (%s), duration: %v, status: %s",
			traceID, serviceName, eventType, serviceType, duration, status)
	}
}

// GetLifecycleEvents returns all lifecycle events for a trace ID
func (dl *DebugLogger) GetLifecycleEvents(traceID string) []LifecycleEvent {
	dl.lifecycleMutex.RLock()
	defer dl.lifecycleMutex.RUnlock()

	var events []LifecycleEvent
	for _, event := range dl.lifecycleEvents {
		if event.TraceID == traceID {
			events = append(events, event)
		}
	}

	return events
}

// RecordMetric records a performance metric
func (dl *DebugLogger) RecordMetric(name string, metricType MetricType, value float64,
	unit, serviceName, operation string, tags map[string]string) {

	if !dl.config.Enabled || !dl.config.CollectMetrics {
		return
	}

	metricID := uuid.New().String()
	traceID := getTraceIDFromContext()

	metric := PerformanceMetric{
		MetricID:    metricID,
		TraceID:     traceID,
		Name:        name,
		Type:        metricType,
		Value:       value,
		Unit:        unit,
		Timestamp:   time.Now(),
		ServiceName: serviceName,
		Operation:   operation,
		Tags:        tags,
	}

	dl.metricsMutex.Lock()
	defer dl.metricsMutex.Unlock()

	dl.metrics = append(dl.metrics, metric)

	// Keep only the last N metrics to prevent memory issues
	if len(dl.metrics) > dl.config.MetricsBufferSize {
		dl.metrics = dl.metrics[len(dl.metrics)-dl.config.MetricsBufferSize:]
	}

	if dl.config.LogLevel >= LogLevelDebug {
		dl.logger.Printf("[DEBUG] [%s] Metric: %s = %f %s (service: %s, operation: %s)",
			traceID, name, value, unit, serviceName, operation)
	}
}

// GetMetrics returns all metrics for a trace ID
func (dl *DebugLogger) GetMetrics(traceID string) []PerformanceMetric {
	dl.metricsMutex.RLock()
	defer dl.metricsMutex.RUnlock()

	var metrics []PerformanceMetric
	for _, metric := range dl.metrics {
		if metric.TraceID == traceID {
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

// Log writes a log message at the specified level
func (dl *DebugLogger) Log(level LogLevel, traceID, message string, args ...interface{}) {
	if !dl.config.Enabled || level > dl.config.LogLevel {
		return
	}

	formattedMsg := message
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(message, args...)
	}

	traceInfo := ""
	if traceID != "" {
		traceInfo = fmt.Sprintf("[%s] ", traceID)
	}

	dl.logger.Printf("[%s] %s%s", level.String(), traceInfo, formattedMsg)
}

// Debug logs a debug message
func (dl *DebugLogger) Debug(traceID, message string, args ...interface{}) {
	dl.Log(LogLevelDebug, traceID, message, args...)
}

// Info logs an info message
func (dl *DebugLogger) Info(traceID, message string, args ...interface{}) {
	dl.Log(LogLevelInfo, traceID, message, args...)
}

// Warn logs a warning message
func (dl *DebugLogger) Warn(traceID, message string, args ...interface{}) {
	dl.Log(LogLevelWarn, traceID, message, args...)
}

// Error logs an error message
func (dl *DebugLogger) Error(traceID, message string, args ...interface{}) {
	dl.Log(LogLevelError, traceID, message, args...)
}

// GenerateDebugReport generates a comprehensive debug report for a trace ID
func (dl *DebugLogger) GenerateDebugReport(traceID string) DebugReport {
	dependencyTrace, hasDependencyTrace := dl.GetDependencyTrace(traceID)
	lifecycleEvents := dl.GetLifecycleEvents(traceID)
	metrics := dl.GetMetrics(traceID)

	return DebugReport{
		TraceID:            traceID,
		GeneratedAt:        time.Now(),
		DependencyTrace:    dependencyTrace,
		LifecycleEvents:    lifecycleEvents,
		Metrics:            metrics,
		HasDependencyTrace: hasDependencyTrace,
		Summary:            generateReportSummary(dependencyTrace, lifecycleEvents, metrics),
	}
}

// DebugReport contains comprehensive debug information for a trace
type DebugReport struct {
	TraceID            string
	GeneratedAt        time.Time
	DependencyTrace    *DependencyTrace
	LifecycleEvents    []LifecycleEvent
	Metrics            []PerformanceMetric
	HasDependencyTrace bool
	Summary            ReportSummary
}

// ReportSummary provides a high-level summary of the debug report
type ReportSummary struct {
	TotalDuration            time.Duration
	DependencyResolutionTime time.Duration
	LifecycleEventCount      int
	MetricCount              int
	ErrorCount               int
	PerformanceIssues        []string
	Recommendations          []string
}

// generateReportSummary creates a summary of the debug information
func generateReportSummary(dependencyTrace *DependencyTrace, lifecycleEvents []LifecycleEvent,
	metrics []PerformanceMetric) ReportSummary {

	summary := ReportSummary{
		LifecycleEventCount: len(lifecycleEvents),
		MetricCount:         len(metrics),
		PerformanceIssues:   make([]string, 0),
		Recommendations:     make([]string, 0),
	}

	// Calculate total duration and dependency resolution time
	if dependencyTrace != nil {
		summary.DependencyResolutionTime = dependencyTrace.TotalDuration
		summary.TotalDuration = dependencyTrace.TotalDuration
	}

	// Count errors in lifecycle events
	for _, event := range lifecycleEvents {
		if !event.Success {
			summary.ErrorCount++
		}

		// Update total duration if we have a longer operation
		if event.Duration > summary.TotalDuration {
			summary.TotalDuration = event.Duration
		}
	}

	// Analyze performance issues
	if dependencyTrace != nil && summary.DependencyResolutionTime > 100*time.Millisecond {
		summary.PerformanceIssues = append(summary.PerformanceIssues,
			fmt.Sprintf("Slow dependency resolution: %v", summary.DependencyResolutionTime))
		summary.Recommendations = append(summary.Recommendations,
			"Consider optimizing dependency injection container or using caching")
	}

	// Check for excessive lifecycle events
	if summary.LifecycleEventCount > 50 {
		summary.PerformanceIssues = append(summary.PerformanceIssues,
			fmt.Sprintf("High number of lifecycle events: %d", summary.LifecycleEventCount))
		summary.Recommendations = append(summary.Recommendations,
			"Review service composition and consider reducing middleware chain")
	}

	// Check for errors
	if summary.ErrorCount > 0 {
		summary.PerformanceIssues = append(summary.PerformanceIssues,
			fmt.Sprintf("Found %d errors in lifecycle events", summary.ErrorCount))
		summary.Recommendations = append(summary.Recommendations,
			"Review error logs and fix underlying issues")
	}

	return summary
}

// Helper functions

// getTraceIDFromContext tries to extract a trace ID from the current context
// In a real implementation, this would integrate with your tracing system
func getTraceIDFromContext() string {
	// For now, return empty string - would integrate with context.Context in real use
	return ""
}

// NewTraceID generates a new trace ID
func NewTraceID() string {
	return uuid.New().String()
}

// LoadDebugModeFromEnv loads debug configuration from environment variables
func LoadDebugModeFromEnv() *DebugMode {
	config := &DebugMode{
		Enabled:           os.Getenv("ENDOR_DEBUG_ENABLED") == "true",
		TraceRequests:     os.Getenv("ENDOR_TRACE_REQUESTS") == "true",
		TraceDependencies: os.Getenv("ENDOR_TRACE_DEPENDENCIES") == "true",
		TraceLifecycle:    os.Getenv("ENDOR_TRACE_LIFECYCLE") == "true",
		CollectMetrics:    os.Getenv("ENDOR_COLLECT_METRICS") == "true",
		MetricsBufferSize: 1000,
		LogOutputFile:     os.Getenv("ENDOR_DEBUG_LOG_FILE"),
	}

	// Parse log level
	switch os.Getenv("ENDOR_LOG_LEVEL") {
	case "OFF":
		config.LogLevel = LogLevelOff
	case "ERROR":
		config.LogLevel = LogLevelError
	case "WARN":
		config.LogLevel = LogLevelWarn
	case "INFO":
		config.LogLevel = LogLevelInfo
	case "DEBUG":
		config.LogLevel = LogLevelDebug
	case "TRACE":
		config.LogLevel = LogLevelTrace
	default:
		config.LogLevel = LogLevelInfo
	}

	return config
}

// Global debug logger instance
var globalDebugLogger *DebugLogger
var globalDebugLoggerOnce sync.Once

// GetGlobalDebugLogger returns the global debug logger instance
func GetGlobalDebugLogger() *DebugLogger {
	globalDebugLoggerOnce.Do(func() {
		config := LoadDebugModeFromEnv()
		logger, err := NewDebugLogger(config)
		if err != nil {
			// Fallback to disabled logger
			logger, _ = NewDebugLogger(&DebugMode{Enabled: false})
		}
		globalDebugLogger = logger
	})

	return globalDebugLogger
}

// Context helpers for trace ID management

// ContextWithTraceID adds a trace ID to the context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "traceID", traceID)
}

// TraceIDFromContext extracts a trace ID from the context
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value("traceID").(string); ok {
		return traceID
	}
	return ""
}
