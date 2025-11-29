package sdk

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk/middleware"
)

// testMiddleware is a simple middleware for integration testing
type testMiddleware struct {
	name   string
	called []string
}

func (t *testMiddleware) Before(ctx interface{}) error {
	t.called = append(t.called, t.name+"_before")
	return nil
}

func (t *testMiddleware) After(ctx interface{}, response interface{}) error {
	t.called = append(t.called, t.name+"_after")
	return nil
}

func TestEndorService_WithMiddleware(t *testing.T) {
	// Create a basic service
	methods := map[string]EndorServiceAction{
		"test": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[string], error) {
				data := "test response"
				return NewResponseBuilder[string]().AddData(&data).Build(), nil
			},
			"Test action",
		),
	}

	service := NewEndorService("test", "Test service", methods)

	// Create test middleware
	middleware1 := &testMiddleware{name: "middleware1"}
	middleware2 := &testMiddleware{name: "middleware2"}

	// Decorate service with middleware
	decoratedService := service.WithMiddleware(middleware1, middleware2)

	// Verify decorated service preserves original interface
	if decoratedService.GetResource() != "test" {
		t.Errorf("Expected resource 'test', got '%s'", decoratedService.GetResource())
	}

	if decoratedService.GetDescription() != "Test service" {
		t.Errorf("Expected description 'Test service', got '%s'", decoratedService.GetDescription())
	}

	// Verify decorated methods are returned
	decoratedMethods := decoratedService.GetMethods()
	if len(decoratedMethods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(decoratedMethods))
	}

	if _, exists := decoratedMethods["test"]; !exists {
		t.Error("Expected 'test' method to exist in decorated service")
	}

	// Test that the decorated action is different from original (wrapped)
	originalMethods := service.GetMethods()
	originalAction := originalMethods["test"]
	decoratedAction := decoratedMethods["test"]

	// They should have different types (original vs decorated)
	if originalAction == decoratedAction {
		t.Error("Decorated action should be different from original action")
	}

	// Verify options are preserved
	originalOptions := originalAction.GetOptions()
	decoratedOptions := decoratedAction.GetOptions()

	if originalOptions.Description != decoratedOptions.Description {
		t.Error("Decorated action should preserve original options")
	}
}

func TestEndorService_WithMiddleware_EmptyChain(t *testing.T) {
	// Test that WithMiddleware works with no middleware
	methods := map[string]EndorServiceAction{
		"test": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[string], error) {
				data := "test"
				return NewResponseBuilder[string]().AddData(&data).Build(), nil
			},
			"Test action",
		),
	}

	service := NewEndorService("test", "Test service", methods)

	// Decorate with empty middleware chain
	decoratedService := service.WithMiddleware()

	// Should still work and preserve interface
	if decoratedService.GetResource() != "test" {
		t.Errorf("Expected resource 'test', got '%s'", decoratedService.GetResource())
	}

	decoratedMethods := decoratedService.GetMethods()
	if len(decoratedMethods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(decoratedMethods))
	}
}

func TestMiddleware_InterfaceCompliance(t *testing.T) {
	// Test that built-in middleware implement the interface correctly

	// Test AuthMiddleware
	authDeps := middleware.AuthMiddlewareDependencies{
		Config:      NewDefaultConfigProvider(),
		Logger:      NewDefaultLogger(),
		RequireAuth: false,
	}
	authMiddleware := middleware.NewAuthMiddleware(authDeps)

	// Should implement MiddlewareInterface
	var _ middleware.MiddlewareInterface = authMiddleware

	// Test LoggingMiddleware
	logDeps := middleware.LoggingMiddlewareDependencies{
		Logger:           NewDefaultLogger(),
		IncludeHeaders:   false,
		IncludePayload:   false,
		CorrelationIDKey: "x-correlation-id",
	}
	logMiddleware := middleware.NewLoggingMiddleware(logDeps)

	// Should implement MiddlewareInterface
	var _ middleware.MiddlewareInterface = logMiddleware

	// Test MetricsMiddleware
	metricsDeps := middleware.MetricsMiddlewareDependencies{
		Logger:             NewDefaultLogger(),
		Config:             NewDefaultConfigProvider(),
		RequestCountMetric: "requests_total",
		DurationMetric:     "request_duration",
		ErrorCountMetric:   "request_errors",
		EnabledMetricTypes: []string{"count", "duration", "errors"},
	}
	metricsMiddleware := middleware.NewMetricsMiddleware(metricsDeps)

	// Should implement MiddlewareInterface
	var _ middleware.MiddlewareInterface = metricsMiddleware
}
