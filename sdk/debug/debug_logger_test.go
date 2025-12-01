package debug

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDebugLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  *DebugMode
		wantErr bool
	}{
		{
			name:    "nil config creates default",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config",
			config: &DebugMode{
				Enabled:           true,
				LogLevel:          LogLevelDebug,
				TraceDependencies: true,
				TraceLifecycle:    true,
				CollectMetrics:    true,
				MetricsBufferSize: 500,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewDebugLogger(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.NotNil(t, logger.config)
				assert.NotNil(t, logger.logger)
				assert.NotNil(t, logger.dependencyTrace)
				assert.NotNil(t, logger.lifecycleEvents)
				assert.NotNil(t, logger.metrics)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelOff, "OFF"},
		{LogLevelError, "ERROR"},
		{LogLevelWarn, "WARN"},
		{LogLevelInfo, "INFO"},
		{LogLevelDebug, "DEBUG"},
		{LogLevelTrace, "TRACE"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestDependencyTracing(t *testing.T) {
	config := &DebugMode{
		Enabled:           true,
		LogLevel:          LogLevelTrace,
		TraceDependencies: true,
		MetricsBufferSize: 100,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	traceID := "test-trace-123"
	requestID := "req-456"

	// Start dependency trace
	logger.StartDependencyTrace(traceID, requestID)

	// Verify trace was created
	trace, exists := logger.GetDependencyTrace(traceID)
	assert.True(t, exists)
	assert.Equal(t, traceID, trace.TraceID)
	assert.Equal(t, requestID, trace.RequestID)
	assert.NotZero(t, trace.StartTime)
	assert.Empty(t, trace.ResolutionSteps)

	// Add resolution steps
	logger.AddResolutionStep(traceID, "UserService", []string{"UserRepository", "Logger"},
		50*time.Millisecond, true, nil, false, "UserService constructor")
	logger.AddResolutionStep(traceID, "UserRepository", []string{"Database"},
		30*time.Millisecond, true, nil, true, "UserRepository constructor")

	// Verify resolution steps were added
	trace, _ = logger.GetDependencyTrace(traceID)
	assert.Len(t, trace.ResolutionSteps, 2)

	step1 := trace.ResolutionSteps[0]
	assert.Equal(t, "UserService", step1.ServiceType)
	assert.Equal(t, []string{"UserRepository", "Logger"}, step1.Dependencies)
	assert.Equal(t, 50*time.Millisecond, step1.Duration)
	assert.True(t, step1.Success)
	assert.Nil(t, step1.Error)
	assert.False(t, step1.CacheHit)
	assert.Equal(t, "UserService constructor", step1.ConstructorInfo)

	step2 := trace.ResolutionSteps[1]
	assert.Equal(t, "UserRepository", step2.ServiceType)
	assert.True(t, step2.CacheHit)

	// End trace
	logger.EndDependencyTrace(traceID, true, nil)

	// Verify trace completion
	trace, _ = logger.GetDependencyTrace(traceID)
	assert.True(t, trace.Success)
	assert.Nil(t, trace.Error)
	assert.Greater(t, trace.TotalDuration, time.Duration(0))
}

func TestDependencyTracing_Disabled(t *testing.T) {
	config := &DebugMode{
		Enabled:           false, // Disabled
		TraceDependencies: false,
		MetricsBufferSize: 100,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	traceID := "test-trace-123"
	requestID := "req-456"

	// Operations should be no-ops when disabled
	logger.StartDependencyTrace(traceID, requestID)

	trace, exists := logger.GetDependencyTrace(traceID)
	assert.False(t, exists)
	assert.Nil(t, trace)

	logger.AddResolutionStep(traceID, "UserService", []string{"UserRepository"},
		50*time.Millisecond, true, nil, false, "UserService constructor")
	logger.EndDependencyTrace(traceID, true, nil)

	trace, exists = logger.GetDependencyTrace(traceID)
	assert.False(t, exists)
	assert.Nil(t, trace)
}

func TestLifecycleEventTracking(t *testing.T) {
	config := &DebugMode{
		Enabled:           true,
		LogLevel:          LogLevelInfo,
		TraceLifecycle:    true,
		MetricsBufferSize: 100,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	// Log some lifecycle events
	logger.LogLifecycleEvent("UserService", "EndorService", EventServiceRegistration,
		10*time.Millisecond, true, nil, map[string]interface{}{"version": "1.0"})
	logger.LogLifecycleEvent("UserService", "EndorService", EventServiceConstruction,
		50*time.Millisecond, true, nil, nil)
	logger.LogLifecycleEvent("OrderService", "EndorHybridService", EventServiceInitialization,
		100*time.Millisecond, false, assert.AnError, map[string]interface{}{"attempts": 3})

	// Since we don't have trace ID context in tests, we'll check all events
	allEvents := logger.lifecycleEvents
	assert.Len(t, allEvents, 3)

	// Check first event
	event1 := allEvents[0]
	assert.Equal(t, "UserService", event1.ServiceName)
	assert.Equal(t, "EndorService", event1.ServiceType)
	assert.Equal(t, EventServiceRegistration, event1.EventType)
	assert.Equal(t, 10*time.Millisecond, event1.Duration)
	assert.True(t, event1.Success)
	assert.Nil(t, event1.Error)
	assert.Equal(t, "1.0", event1.Metadata["version"])
	assert.NotZero(t, event1.Timestamp)
	assert.NotEmpty(t, event1.EventID)

	// Check error event
	event3 := allEvents[2]
	assert.Equal(t, "OrderService", event3.ServiceName)
	assert.Equal(t, EventServiceInitialization, event3.EventType)
	assert.False(t, event3.Success)
	assert.Equal(t, assert.AnError, event3.Error)
	assert.Equal(t, 3, event3.Metadata["attempts"])
}

func TestPerformanceMetrics(t *testing.T) {
	config := &DebugMode{
		Enabled:           true,
		LogLevel:          LogLevelDebug,
		CollectMetrics:    true,
		MetricsBufferSize: 50,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	// Record some metrics
	logger.RecordMetric("request_duration", MetricTypeTiming, 150.5, "ms",
		"UserService", "CreateUser", map[string]string{"method": "POST"})
	logger.RecordMetric("active_connections", MetricTypeGauge, 42.0, "count",
		"DatabaseService", "Connect", nil)
	logger.RecordMetric("requests_total", MetricTypeCounter, 1.0, "count",
		"APIService", "HandleRequest", map[string]string{"endpoint": "/api/users"})

	// Check metrics were recorded
	allMetrics := logger.metrics
	assert.Len(t, allMetrics, 3)

	// Check first metric
	metric1 := allMetrics[0]
	assert.Equal(t, "request_duration", metric1.Name)
	assert.Equal(t, MetricTypeTiming, metric1.Type)
	assert.Equal(t, 150.5, metric1.Value)
	assert.Equal(t, "ms", metric1.Unit)
	assert.Equal(t, "UserService", metric1.ServiceName)
	assert.Equal(t, "CreateUser", metric1.Operation)
	assert.Equal(t, "POST", metric1.Tags["method"])
	assert.NotZero(t, metric1.Timestamp)
	assert.NotEmpty(t, metric1.MetricID)

	// Check gauge metric
	metric2 := allMetrics[1]
	assert.Equal(t, "active_connections", metric2.Name)
	assert.Equal(t, MetricTypeGauge, metric2.Type)
	assert.Equal(t, 42.0, metric2.Value)

	// Check counter metric
	metric3 := allMetrics[2]
	assert.Equal(t, "requests_total", metric3.Name)
	assert.Equal(t, MetricTypeCounter, metric3.Type)
	assert.Equal(t, 1.0, metric3.Value)
}

func TestLogging(t *testing.T) {
	config := &DebugMode{
		Enabled:           true,
		LogLevel:          LogLevelDebug,
		MetricsBufferSize: 100,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	traceID := "test-trace-123"

	// Test different log levels
	logger.Debug(traceID, "Debug message: %s", "test")
	logger.Info(traceID, "Info message: %d", 42)
	logger.Warn(traceID, "Warning message")
	logger.Error(traceID, "Error message: %v", assert.AnError)

	// Test log level filtering
	config.LogLevel = LogLevelWarn
	logger.config = config

	logger.Debug(traceID, "This should not be logged")
	logger.Info(traceID, "This should not be logged")
	logger.Warn(traceID, "This should be logged")
	logger.Error(traceID, "This should be logged")

	// No direct way to test log output without capturing it,
	// but we can test that the methods don't panic
}

func TestDebugReport(t *testing.T) {
	config := &DebugMode{
		Enabled:           true,
		LogLevel:          LogLevelTrace,
		TraceDependencies: true,
		TraceLifecycle:    true,
		CollectMetrics:    true,
		MetricsBufferSize: 100,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	traceID := "test-trace-123"
	requestID := "req-456"

	// Simulate a complete request trace
	logger.StartDependencyTrace(traceID, requestID)
	logger.AddResolutionStep(traceID, "UserService", []string{"UserRepository"},
		50*time.Millisecond, true, nil, false, "UserService constructor")
	logger.AddResolutionStep(traceID, "UserRepository", []string{"Database"},
		30*time.Millisecond, true, nil, true, "UserRepository constructor")
	logger.EndDependencyTrace(traceID, true, nil)

	// Manually add lifecycle events with the specific trace ID for testing
	// (In real usage, this would come from context)
	event1 := LifecycleEvent{
		EventID:     "event-1",
		TraceID:     traceID,
		ServiceName: "UserService",
		ServiceType: "EndorService",
		EventType:   EventServiceConstruction,
		Timestamp:   time.Now(),
		Duration:    75 * time.Millisecond,
		Success:     true,
		Error:       nil,
		Metadata:    nil,
	}
	event2 := LifecycleEvent{
		EventID:     "event-2",
		TraceID:     traceID,
		ServiceName: "UserService",
		ServiceType: "EndorService",
		EventType:   EventHTTPRequest,
		Timestamp:   time.Now(),
		Duration:    200 * time.Millisecond,
		Success:     true,
		Error:       nil,
		Metadata:    map[string]interface{}{"method": "POST"},
	}

	logger.lifecycleMutex.Lock()
	logger.lifecycleEvents = append(logger.lifecycleEvents, event1, event2)
	logger.lifecycleMutex.Unlock()

	// Manually add metrics with the specific trace ID for testing
	metric1 := PerformanceMetric{
		MetricID:    "metric-1",
		TraceID:     traceID,
		Name:        "request_duration",
		Type:        MetricTypeTiming,
		Value:       200.0,
		Unit:        "ms",
		Timestamp:   time.Now(),
		ServiceName: "UserService",
		Operation:   "CreateUser",
		Tags:        nil,
	}
	metric2 := PerformanceMetric{
		MetricID:    "metric-2",
		TraceID:     traceID,
		Name:        "database_queries",
		Type:        MetricTypeCounter,
		Value:       3.0,
		Unit:        "count",
		Timestamp:   time.Now(),
		ServiceName: "UserRepository",
		Operation:   "FindUser",
		Tags:        nil,
	}

	logger.metricsMutex.Lock()
	logger.metrics = append(logger.metrics, metric1, metric2)
	logger.metricsMutex.Unlock()

	// Generate report
	report := logger.GenerateDebugReport(traceID)

	// Verify report structure
	assert.Equal(t, traceID, report.TraceID)
	assert.NotZero(t, report.GeneratedAt)
	assert.True(t, report.HasDependencyTrace)
	assert.NotNil(t, report.DependencyTrace)
	assert.Len(t, report.LifecycleEvents, 2)
	assert.Len(t, report.Metrics, 2)

	// Verify summary
	summary := report.Summary
	assert.Equal(t, 2, summary.LifecycleEventCount)
	assert.Equal(t, 2, summary.MetricCount)
	assert.Equal(t, 0, summary.ErrorCount)
	assert.Greater(t, summary.DependencyResolutionTime, time.Duration(0))
	assert.Greater(t, summary.TotalDuration, time.Duration(0))
}

func TestLoadDebugModeFromEnv(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"ENDOR_DEBUG_ENABLED":      os.Getenv("ENDOR_DEBUG_ENABLED"),
		"ENDOR_TRACE_REQUESTS":     os.Getenv("ENDOR_TRACE_REQUESTS"),
		"ENDOR_TRACE_DEPENDENCIES": os.Getenv("ENDOR_TRACE_DEPENDENCIES"),
		"ENDOR_TRACE_LIFECYCLE":    os.Getenv("ENDOR_TRACE_LIFECYCLE"),
		"ENDOR_COLLECT_METRICS":    os.Getenv("ENDOR_COLLECT_METRICS"),
		"ENDOR_LOG_LEVEL":          os.Getenv("ENDOR_LOG_LEVEL"),
		"ENDOR_DEBUG_LOG_FILE":     os.Getenv("ENDOR_DEBUG_LOG_FILE"),
	}

	// Clean up after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("ENDOR_DEBUG_ENABLED", "true")
	os.Setenv("ENDOR_TRACE_REQUESTS", "true")
	os.Setenv("ENDOR_TRACE_DEPENDENCIES", "true")
	os.Setenv("ENDOR_TRACE_LIFECYCLE", "true")
	os.Setenv("ENDOR_COLLECT_METRICS", "true")
	os.Setenv("ENDOR_LOG_LEVEL", "DEBUG")
	os.Setenv("ENDOR_DEBUG_LOG_FILE", "/tmp/debug.log")

	config := LoadDebugModeFromEnv()

	assert.True(t, config.Enabled)
	assert.True(t, config.TraceRequests)
	assert.True(t, config.TraceDependencies)
	assert.True(t, config.TraceLifecycle)
	assert.True(t, config.CollectMetrics)
	assert.Equal(t, LogLevelDebug, config.LogLevel)
	assert.Equal(t, "/tmp/debug.log", config.LogOutputFile)
	assert.Equal(t, 1000, config.MetricsBufferSize)

	// Test different log levels
	testLevels := map[string]LogLevel{
		"OFF":     LogLevelOff,
		"ERROR":   LogLevelError,
		"WARN":    LogLevelWarn,
		"INFO":    LogLevelInfo,
		"DEBUG":   LogLevelDebug,
		"TRACE":   LogLevelTrace,
		"INVALID": LogLevelInfo, // Should default to INFO
	}

	for envValue, expected := range testLevels {
		os.Setenv("ENDOR_LOG_LEVEL", envValue)
		config := LoadDebugModeFromEnv()
		assert.Equal(t, expected, config.LogLevel, "Failed for log level: %s", envValue)
	}
}

func TestGlobalDebugLogger(t *testing.T) {
	// Test that global logger is initialized
	logger1 := GetGlobalDebugLogger()
	assert.NotNil(t, logger1)

	// Test that it returns the same instance (singleton)
	logger2 := GetGlobalDebugLogger()
	assert.Same(t, logger1, logger2)
}

func TestNewTraceID(t *testing.T) {
	traceID1 := NewTraceID()
	traceID2 := NewTraceID()

	assert.NotEmpty(t, traceID1)
	assert.NotEmpty(t, traceID2)
	assert.NotEqual(t, traceID1, traceID2, "Trace IDs should be unique")

	// Should be valid UUIDs
	assert.Len(t, traceID1, 36, "Should be UUID format")
	assert.Len(t, traceID2, 36, "Should be UUID format")
}

func TestMetricsBufferManagement(t *testing.T) {
	config := &DebugMode{
		Enabled:           true,
		CollectMetrics:    true,
		MetricsBufferSize: 3, // Small buffer for testing
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	// Add more metrics than buffer size
	for i := 0; i < 5; i++ {
		logger.RecordMetric("test_metric", MetricTypeCounter, float64(i), "count",
			"TestService", "TestOperation", nil)
	}

	// Should only keep the last 3 metrics
	assert.Len(t, logger.metrics, 3)

	// Should have metrics with values 2, 3, 4 (the last 3)
	values := make([]float64, len(logger.metrics))
	for i, metric := range logger.metrics {
		values[i] = metric.Value
	}
	assert.Equal(t, []float64{2.0, 3.0, 4.0}, values)
}

func TestLifecycleEventsBufferManagement(t *testing.T) {
	config := &DebugMode{
		Enabled:        true,
		TraceLifecycle: true,
	}

	logger, err := NewDebugLogger(config)
	assert.NoError(t, err)

	// Add many lifecycle events to test buffer management
	// The implementation keeps only the last 1000 events
	for i := 0; i < 1005; i++ {
		logger.LogLifecycleEvent("TestService", "EndorService", EventHTTPRequest,
			time.Millisecond, true, nil, nil)
	}

	// Should only keep the last 1000 events
	assert.Len(t, logger.lifecycleEvents, 1000)
}
