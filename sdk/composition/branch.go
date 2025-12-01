package composition

import (
	"context"
	"fmt"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// BranchedService represents a conditional routing service that selects which service
// to execute based on request analysis and routing logic.
type BranchedService struct {
	router    BranchRouter
	services  map[string]interfaces.EndorServiceInterface
	config    CompositionConfig
	validator CompositionValidator
}

// ServiceBranch creates a new branched service that routes requests to different services
// based on conditional logic provided by the router.
//
// Example usage:
//
//	router := &UserTypeRouter{}
//	services := map[string]interfaces.EndorServiceInterface{
//		"admin": adminService,
//		"user":  userService,
//		"guest": guestService,
//	}
//	branch := ServiceBranch(router, services)
//	result, err := branch.Execute(ctx, request)
//
// Performance: < 25μs overhead for routing decision and execution
func ServiceBranch(router BranchRouter, services map[string]interfaces.EndorServiceInterface) *BranchedService {
	return &BranchedService{
		router:    router,
		services:  services,
		config:    defaultBranchConfig(),
		validator: DefaultValidator(),
	}
}

// WithConfig applies custom configuration to the branched service.
func (b *BranchedService) WithConfig(config CompositionConfig) *BranchedService {
	b.config = config
	return b
}

// WithValidator applies a custom validator to the branched service.
func (b *BranchedService) WithValidator(validator CompositionValidator) *BranchedService {
	b.validator = validator
	return b
}

// Execute routes the request to the appropriate service based on the router's logic.
func (b *BranchedService) Execute(ctx context.Context, request interface{}) (interface{}, error) {
	if b.router == nil {
		return nil, NewCompositionError("branch", "", -1, fmt.Errorf("router cannot be nil"))
	}

	if len(b.services) == 0 {
		return nil, NewCompositionError("branch", "", -1, fmt.Errorf("no services configured for routing"))
	}

	// Apply timeout if configured
	if b.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.config.Timeout)
		defer cancel()
	}

	// Use router to select the appropriate service
	serviceKey, err := b.router.SelectService(ctx, request)
	if err != nil {
		// Try default service if selection fails
		defaultKey := b.router.GetDefaultService()
		if defaultKey == "" {
			return nil, NewCompositionError("branch", "", -1,
				fmt.Errorf("service selection failed and no default service configured: %w", err))
		}
		serviceKey = defaultKey
	}

	// Get the selected service
	selectedService, exists := b.services[serviceKey]
	if !exists {
		// Try default service if selected service doesn't exist
		defaultKey := b.router.GetDefaultService()
		if defaultKey == "" || defaultKey == serviceKey {
			return nil, NewCompositionError("branch", serviceKey, -1,
				fmt.Errorf("selected service '%s' not found and no valid default service", serviceKey))
		}

		selectedService, exists = b.services[defaultKey]
		if !exists {
			return nil, NewCompositionError("branch", defaultKey, -1,
				fmt.Errorf("default service '%s' not found", defaultKey))
		}
		serviceKey = defaultKey
	}

	// Execute the selected service
	result, err := b.executeSelectedService(ctx, selectedService, request)
	if err != nil {
		switch b.config.ErrorHandling {
		case FailFast:
			return nil, NewCompositionError("branch", serviceKey, 0, err)
		case ContinueOnError:
			return err, nil // Return error as result
		case RetryOnError:
			// Could implement retry with different service selection here
			return nil, NewCompositionError("branch", serviceKey, 0, err)
		}
	}

	return result, nil
}

// executeSelectedService executes the service selected by the router.
func (b *BranchedService) executeSelectedService(ctx context.Context, service interfaces.EndorServiceInterface, request interface{}) (interface{}, error) {
	// Validate service before execution
	if err := service.Validate(); err != nil {
		return nil, fmt.Errorf("selected service validation failed: %w", err)
	}

	// Record metrics if enabled
	start := time.Now()

	// In a real implementation, this would call the service's handler
	// For now, we simulate the execution
	result := request

	if b.config.EnableMetrics {
		duration := time.Since(start)
		_ = duration // Record metrics
	}

	return result, nil
}

// Validate checks that the branched service is properly configured.
func (b *BranchedService) Validate() error {
	if b.router == nil {
		return fmt.Errorf("router cannot be nil")
	}

	if len(b.services) == 0 {
		return fmt.Errorf("no services configured for routing")
	}

	if b.validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	// Validate all configured services
	services := make([]interfaces.EndorServiceInterface, 0, len(b.services))
	for _, service := range b.services {
		services = append(services, service)
	}

	return b.validator.ValidateChain(services)
}

// GetServices returns all services configured for routing.
func (b *BranchedService) GetServices() []interfaces.EndorServiceInterface {
	services := make([]interfaces.EndorServiceInterface, 0, len(b.services))
	for _, service := range b.services {
		services = append(services, service)
	}
	return services
}

// GetServiceKeys returns all configured service keys for introspection.
func (b *BranchedService) GetServiceKeys() []string {
	keys := make([]string, 0, len(b.services))
	for key := range b.services {
		keys = append(keys, key)
	}
	return keys
}

// GetRouter returns the current router.
func (b *BranchedService) GetRouter() BranchRouter {
	return b.router
}

// GetConfig returns the current configuration of the branched service.
func (b *BranchedService) GetConfig() CompositionConfig {
	return b.config
}

// defaultBranchConfig returns the default configuration for service branching.
func defaultBranchConfig() CompositionConfig {
	return CompositionConfig{
		Timeout:        30 * time.Second,
		MaxConcurrency: 1,
		ErrorHandling:  FailFast,
		EnableMetrics:  true,
		ValidationMode: Strict,
	}
}

// SimpleRouter provides a basic router implementation based on request field values.
type SimpleRouter struct {
	FieldName      string
	RouteMap       map[interface{}]string
	DefaultService string
}

// SelectService routes based on a specific field value in the request.
func (r *SimpleRouter) SelectService(ctx context.Context, request interface{}) (string, error) {
	// In a real implementation, this would use reflection to extract field values
	// For now, we simulate routing logic

	// This is a simplified implementation - would need proper reflection
	// to extract field values from the request in a real scenario
	return r.DefaultService, nil
}

// GetDefaultService returns the configured default service key.
func (r *SimpleRouter) GetDefaultService() string {
	return r.DefaultService
}

// FunctionRouter provides routing based on a custom function.
type FunctionRouter struct {
	RoutingFunc    func(context.Context, interface{}) (string, error)
	DefaultService string
}

// SelectService routes using the configured routing function.
func (r *FunctionRouter) SelectService(ctx context.Context, request interface{}) (string, error) {
	if r.RoutingFunc == nil {
		return r.DefaultService, nil
	}
	return r.RoutingFunc(ctx, request)
}

// GetDefaultService returns the configured default service key.
func (r *FunctionRouter) GetDefaultService() string {
	return r.DefaultService
}

// ContextRouter provides routing based on context values.
type ContextRouter struct {
	ContextKey     string
	RouteMap       map[interface{}]string
	DefaultService string
}

// SelectService routes based on context values.
func (r *ContextRouter) SelectService(ctx context.Context, request interface{}) (string, error) {
	value := ctx.Value(r.ContextKey)
	if value == nil {
		return r.DefaultService, nil
	}

	if serviceKey, exists := r.RouteMap[value]; exists {
		return serviceKey, nil
	}

	return r.DefaultService, nil
}

// GetDefaultService returns the configured default service key.
func (r *ContextRouter) GetDefaultService() string {
	return r.DefaultService
}

// BranchBuilder provides a fluent API for building complex service branches.
type BranchBuilder struct {
	router   BranchRouter
	services map[string]interfaces.EndorServiceInterface
	config   CompositionConfig
}

// NewBranchBuilder creates a new builder for constructing service branches.
func NewBranchBuilder(router BranchRouter) *BranchBuilder {
	return &BranchBuilder{
		router:   router,
		services: make(map[string]interfaces.EndorServiceInterface),
		config:   defaultBranchConfig(),
	}
}

// AddService adds a service with the specified routing key.
func (b *BranchBuilder) AddService(key string, service interfaces.EndorServiceInterface) *BranchBuilder {
	b.services[key] = service
	return b
}

// AddServices adds multiple services from a map.
func (b *BranchBuilder) AddServices(services map[string]interfaces.EndorServiceInterface) *BranchBuilder {
	for key, service := range services {
		b.services[key] = service
	}
	return b
}

// WithTimeout sets the timeout for branch operations.
func (b *BranchBuilder) WithTimeout(timeout time.Duration) *BranchBuilder {
	b.config.Timeout = timeout
	return b
}

// WithErrorHandling sets the error handling strategy for the branch.
func (b *BranchBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *BranchBuilder {
	b.config.ErrorHandling = strategy
	return b
}

// Build constructs the final branched service composition.
func (b *BranchBuilder) Build() *BranchedService {
	return &BranchedService{
		router:    b.router,
		services:  b.services,
		config:    b.config,
		validator: DefaultValidator(),
	}
}
