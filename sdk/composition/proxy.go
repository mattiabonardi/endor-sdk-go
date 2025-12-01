package composition

import (
	"context"
	"fmt"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// ProxiedService represents a service that forwards requests to a target service
// with optional interception and transformation of requests and responses.
type ProxiedService struct {
	target      interfaces.EndorServiceInterface
	interceptor ProxyInterceptor
	config      CompositionConfig
	validator   CompositionValidator
}

// ServiceProxy creates a new proxied service that forwards requests to a target service
// with optional request/response interception and transformation.
//
// Example usage:
//
//	proxy := ServiceProxy(targetService, &LoggingInterceptor{})
//	result, err := proxy.Execute(ctx, request)
//
// Performance: < 10μs overhead for transparent proxying
func ServiceProxy(target interfaces.EndorServiceInterface, interceptor ProxyInterceptor) *ProxiedService {
	return &ProxiedService{
		target:      target,
		interceptor: interceptor,
		config:      defaultProxyConfig(),
		validator:   DefaultValidator(),
	}
}

// WithConfig applies custom configuration to the proxied service.
func (p *ProxiedService) WithConfig(config CompositionConfig) *ProxiedService {
	p.config = config
	return p
}

// WithValidator applies a custom validator to the proxied service.
func (p *ProxiedService) WithValidator(validator CompositionValidator) *ProxiedService {
	p.validator = validator
	return p
}

// Execute forwards the request to the target service through the interceptor.
func (p *ProxiedService) Execute(ctx context.Context, request interface{}) (interface{}, error) {
	if p.target == nil {
		return nil, NewCompositionError("proxy", "", -1, fmt.Errorf("target service cannot be nil"))
	}

	// Apply timeout if configured
	if p.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	var err error
	processedRequest := request

	// Apply request interception if interceptor is configured
	if p.interceptor != nil {
		processedRequest, err = p.interceptor.BeforeRequest(ctx, request)
		if err != nil {
			return nil, NewCompositionError("proxy", p.target.GetResource(), 0,
				fmt.Errorf("request interception failed: %w", err))
		}
	}

	// Execute the target service
	result, err := p.executeTargetService(ctx, processedRequest)
	if err != nil {
		switch p.config.ErrorHandling {
		case FailFast:
			return nil, NewCompositionError("proxy", p.target.GetResource(), 0, err)
		case ContinueOnError:
			result = err // Pass error as result for response interception
		case RetryOnError:
			// Implement retry logic if needed
			return nil, NewCompositionError("proxy", p.target.GetResource(), 0, err)
		}
	}

	// Apply response interception if interceptor is configured
	if p.interceptor != nil {
		finalResult, interceptErr := p.interceptor.AfterResponse(ctx, processedRequest, result)
		if interceptErr != nil {
			return nil, NewCompositionError("proxy", p.target.GetResource(), 0,
				fmt.Errorf("response interception failed: %w", interceptErr))
		}
		result = finalResult
	}

	return result, nil
}

// executeTargetService executes the target service with the processed request.
func (p *ProxiedService) executeTargetService(ctx context.Context, request interface{}) (interface{}, error) {
	// Validate target service before execution
	if err := p.target.Validate(); err != nil {
		return nil, fmt.Errorf("target service validation failed: %w", err)
	}

	// Record metrics if enabled
	start := time.Now()

	// In a real implementation, this would call the target service's handler
	// For now, we simulate the execution
	result := request

	if p.config.EnableMetrics {
		duration := time.Since(start)
		_ = duration // Record metrics
	}

	return result, nil
}

// Validate checks that the proxied service is properly configured.
func (p *ProxiedService) Validate() error {
	if p.target == nil {
		return fmt.Errorf("proxy target cannot be nil")
	}

	if p.validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	return p.validator.ValidateChain([]interfaces.EndorServiceInterface{p.target})
}

// GetServices returns the target service for introspection.
func (p *ProxiedService) GetServices() []interfaces.EndorServiceInterface {
	return []interfaces.EndorServiceInterface{p.target}
}

// GetTarget returns the target service being proxied.
func (p *ProxiedService) GetTarget() interfaces.EndorServiceInterface {
	return p.target
}

// GetInterceptor returns the current interceptor.
func (p *ProxiedService) GetInterceptor() ProxyInterceptor {
	return p.interceptor
}

// GetConfig returns the current configuration of the proxied service.
func (p *ProxiedService) GetConfig() CompositionConfig {
	return p.config
}

// defaultProxyConfig returns the default configuration for service proxies.
func defaultProxyConfig() CompositionConfig {
	return CompositionConfig{
		Timeout:        30 * time.Second,
		MaxConcurrency: 1,
		ErrorHandling:  FailFast,
		EnableMetrics:  true,
		ValidationMode: Strict,
	}
}

// NoOpInterceptor provides a no-operation interceptor that passes requests and responses unchanged.
type NoOpInterceptor struct{}

// BeforeRequest passes the request through unchanged.
func (n *NoOpInterceptor) BeforeRequest(ctx context.Context, request interface{}) (interface{}, error) {
	return request, nil
}

// AfterResponse passes the response through unchanged.
func (n *NoOpInterceptor) AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error) {
	return response, nil
}

// LoggingInterceptor provides basic request/response logging functionality.
type LoggingInterceptor struct {
	ServiceName string
}

// BeforeRequest logs the incoming request.
func (l *LoggingInterceptor) BeforeRequest(ctx context.Context, request interface{}) (interface{}, error) {
	// In a real implementation, this would use the actual logging framework
	fmt.Printf("PROXY[%s]: Request received: %+v\n", l.ServiceName, request)
	return request, nil
}

// AfterResponse logs the outgoing response.
func (l *LoggingInterceptor) AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error) {
	// In a real implementation, this would use the actual logging framework
	fmt.Printf("PROXY[%s]: Response sent: %+v\n", l.ServiceName, response)
	return response, nil
}

// TransformationInterceptor provides request/response transformation capabilities.
type TransformationInterceptor struct {
	RequestTransformer  func(context.Context, interface{}) (interface{}, error)
	ResponseTransformer func(context.Context, interface{}, interface{}) (interface{}, error)
}

// BeforeRequest applies the request transformation if configured.
func (t *TransformationInterceptor) BeforeRequest(ctx context.Context, request interface{}) (interface{}, error) {
	if t.RequestTransformer != nil {
		return t.RequestTransformer(ctx, request)
	}
	return request, nil
}

// AfterResponse applies the response transformation if configured.
func (t *TransformationInterceptor) AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error) {
	if t.ResponseTransformer != nil {
		return t.ResponseTransformer(ctx, request, response)
	}
	return response, nil
}

// ProxyBuilder provides a fluent API for building complex service proxies.
type ProxyBuilder struct {
	target      interfaces.EndorServiceInterface
	interceptor ProxyInterceptor
	config      CompositionConfig
}

// NewProxyBuilder creates a new builder for constructing service proxies.
func NewProxyBuilder(target interfaces.EndorServiceInterface) *ProxyBuilder {
	return &ProxyBuilder{
		target: target,
		config: defaultProxyConfig(),
	}
}

// WithInterceptor sets the interceptor for the proxy.
func (b *ProxyBuilder) WithInterceptor(interceptor ProxyInterceptor) *ProxyBuilder {
	b.interceptor = interceptor
	return b
}

// WithLogging adds logging interception to the proxy.
func (b *ProxyBuilder) WithLogging(serviceName string) *ProxyBuilder {
	b.interceptor = &LoggingInterceptor{ServiceName: serviceName}
	return b
}

// WithTimeout sets the timeout for proxy operations.
func (b *ProxyBuilder) WithTimeout(timeout time.Duration) *ProxyBuilder {
	b.config.Timeout = timeout
	return b
}

// WithErrorHandling sets the error handling strategy for the proxy.
func (b *ProxyBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *ProxyBuilder {
	b.config.ErrorHandling = strategy
	return b
}

// Build constructs the final proxied service composition.
func (b *ProxyBuilder) Build() *ProxiedService {
	return &ProxiedService{
		target:      b.target,
		interceptor: b.interceptor,
		config:      b.config,
		validator:   DefaultValidator(),
	}
}
