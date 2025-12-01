package composition

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// MergedService represents a composition that executes multiple services in parallel
// and merges their results using a configurable merging strategy.
type MergedService struct {
	services  []interfaces.EndorServiceInterface
	merger    ResultMerger
	config    CompositionConfig
	validator CompositionValidator
}

// ServiceMerger creates a new merged service that executes multiple services in parallel
// and combines their results using the provided merger strategy.
//
// Example usage:
//
//	merger := &FirstWinsMerger{}
//	mergedService := ServiceMerger([]interfaces.EndorServiceInterface{
//		serviceA, serviceB, serviceC,
//	}, merger)
//	result, err := mergedService.Execute(ctx, request)
//
// Performance: Parallel execution with configurable concurrency limits
func ServiceMerger(services []interfaces.EndorServiceInterface, merger ResultMerger) *MergedService {
	return &MergedService{
		services:  services,
		merger:    merger,
		config:    defaultMergerConfig(),
		validator: DefaultValidator(),
	}
}

// WithConfig applies custom configuration to the merged service.
func (m *MergedService) WithConfig(config CompositionConfig) *MergedService {
	m.config = config
	return m
}

// WithValidator applies a custom validator to the merged service.
func (m *MergedService) WithValidator(validator CompositionValidator) *MergedService {
	m.validator = validator
	return m
}

// Execute runs all services in parallel and merges their results.
func (m *MergedService) Execute(ctx context.Context, request interface{}) (interface{}, error) {
	if len(m.services) == 0 {
		return nil, NewCompositionError("merger", "", -1, fmt.Errorf("cannot execute merger with no services"))
	}

	if m.merger == nil {
		return nil, NewCompositionError("merger", "", -1, fmt.Errorf("merger cannot be nil"))
	}

	// Apply timeout if configured
	timeout := m.config.Timeout
	if m.merger.GetTimeout() > 0 {
		timeout = m.merger.GetTimeout()
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Determine concurrency level
	maxConcurrency := m.config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = len(m.services)
	}

	// Execute services with controlled concurrency
	results, err := m.executeServicesParallel(ctx, request, maxConcurrency)
	if err != nil {
		return nil, err
	}

	// Merge the results
	mergedResult, err := m.merger.MergeResults(ctx, request, results)
	if err != nil {
		return nil, NewCompositionError("merger", "", -1, fmt.Errorf("result merging failed: %w", err))
	}

	return mergedResult, nil
}

// serviceResult represents the result of executing a single service.
type serviceResult struct {
	Index  int
	Result interface{}
	Error  error
}

// executeServicesParallel executes services in parallel with controlled concurrency.
func (m *MergedService) executeServicesParallel(ctx context.Context, request interface{}, maxConcurrency int) ([]interface{}, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(chan serviceResult, len(m.services))

	// Start a goroutine for each service
	var wg sync.WaitGroup
	for i, service := range m.services {
		wg.Add(1)
		go func(index int, svc interfaces.EndorServiceInterface) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute the service
			result, err := m.executeService(ctx, svc, request)

			// Send result
			results <- serviceResult{
				Index:  index,
				Result: result,
				Error:  err,
			}
		}(i, service)
	}

	// Start a goroutine to close results channel when all services complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	serviceResults := make([]interface{}, len(m.services))
	var errors []error
	completedCount := 0

	for result := range results {
		serviceResults[result.Index] = result.Result

		if result.Error != nil {
			errors = append(errors, NewCompositionError("merger",
				m.services[result.Index].GetResource(), result.Index, result.Error))

			// Handle error based on strategy
			switch m.config.ErrorHandling {
			case FailFast:
				return nil, errors[len(errors)-1]
			case ContinueOnError:
				// Continue collecting results
			case RetryOnError:
				// Could implement retry logic here
			}
		}

		completedCount++

		// Check if we can return early (for strategies that don't need all results)
		if !m.merger.ShouldWaitForAll() && completedCount >= 1 && result.Error == nil {
			return serviceResults[:completedCount], nil
		}
	}

	// Check if we had any successful results
	if len(errors) > 0 && len(errors) == len(m.services) {
		return nil, NewCompositionError("merger", "", -1,
			fmt.Errorf("all services failed: %d errors", len(errors)))
	}

	return serviceResults, nil
}

// executeService executes a single service within the merger.
func (m *MergedService) executeService(ctx context.Context, service interfaces.EndorServiceInterface, request interface{}) (interface{}, error) {
	// Validate service before execution
	if err := service.Validate(); err != nil {
		return nil, fmt.Errorf("service validation failed: %w", err)
	}

	// Record metrics if enabled
	start := time.Now()

	// In a real implementation, this would call the service's handler
	// For now, we simulate the execution
	result := request

	if m.config.EnableMetrics {
		duration := time.Since(start)
		_ = duration // Record metrics
	}

	return result, nil
}

// Validate checks that the merged service is properly configured.
func (m *MergedService) Validate() error {
	if len(m.services) == 0 {
		return fmt.Errorf("cannot validate merger with no services")
	}

	if m.merger == nil {
		return fmt.Errorf("merger cannot be nil")
	}

	if m.validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	return m.validator.ValidateChain(m.services)
}

// GetServices returns the services in the merger for introspection.
func (m *MergedService) GetServices() []interfaces.EndorServiceInterface {
	// Return a copy to prevent external modification
	result := make([]interfaces.EndorServiceInterface, len(m.services))
	copy(result, m.services)
	return result
}

// GetMerger returns the current result merger.
func (m *MergedService) GetMerger() ResultMerger {
	return m.merger
}

// GetConfig returns the current configuration of the merged service.
func (m *MergedService) GetConfig() CompositionConfig {
	return m.config
}

// defaultMergerConfig returns the default configuration for service mergers.
func defaultMergerConfig() CompositionConfig {
	return CompositionConfig{
		Timeout:        30 * time.Second,
		MaxConcurrency: 10, // Allow parallel execution
		ErrorHandling:  ContinueOnError,
		EnableMetrics:  true,
		ValidationMode: Strict,
	}
}

// FirstWinsMerger returns the first successful result received.
type FirstWinsMerger struct {
	Timeout time.Duration
}

// MergeResults returns the first non-nil result.
func (f *FirstWinsMerger) MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error) {
	for _, result := range results {
		if result != nil {
			return result, nil
		}
	}

	// If no results, return the request (for our simple implementation)
	if len(results) > 0 {
		return request, nil
	}

	return nil, fmt.Errorf("no results to merge")
}

// ShouldWaitForAll returns false since we only need the first result.
func (f *FirstWinsMerger) ShouldWaitForAll() bool {
	return false
}

// GetTimeout returns the configured timeout.
func (f *FirstWinsMerger) GetTimeout() time.Duration {
	return f.Timeout
}

// AllResultsMerger collects all results into a slice.
type AllResultsMerger struct {
	Timeout time.Duration
}

// MergeResults combines all results into a single slice.
func (a *AllResultsMerger) MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error) {
	return results, nil
}

// ShouldWaitForAll returns true since we need all results.
func (a *AllResultsMerger) ShouldWaitForAll() bool {
	return true
}

// GetTimeout returns the configured timeout.
func (a *AllResultsMerger) GetTimeout() time.Duration {
	return a.Timeout
}

// FunctionMerger uses a custom function to merge results.
type FunctionMerger struct {
	MergeFunc  func(context.Context, interface{}, []interface{}) (interface{}, error)
	WaitForAll bool
	Timeout    time.Duration
}

// MergeResults uses the configured merge function.
func (f *FunctionMerger) MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error) {
	if f.MergeFunc == nil {
		return results, nil
	}
	return f.MergeFunc(ctx, request, results)
}

// ShouldWaitForAll returns the configured wait strategy.
func (f *FunctionMerger) ShouldWaitForAll() bool {
	return f.WaitForAll
}

// GetTimeout returns the configured timeout.
func (f *FunctionMerger) GetTimeout() time.Duration {
	return f.Timeout
}

// MergerBuilder provides a fluent API for building complex service mergers.
type MergerBuilder struct {
	services []interfaces.EndorServiceInterface
	merger   ResultMerger
	config   CompositionConfig
}

// NewMergerBuilder creates a new builder for constructing service mergers.
func NewMergerBuilder() *MergerBuilder {
	return &MergerBuilder{
		services: make([]interfaces.EndorServiceInterface, 0),
		config:   defaultMergerConfig(),
	}
}

// AddService adds a service to the merger.
func (b *MergerBuilder) AddService(service interfaces.EndorServiceInterface) *MergerBuilder {
	b.services = append(b.services, service)
	return b
}

// AddServices adds multiple services to the merger.
func (b *MergerBuilder) AddServices(services ...interfaces.EndorServiceInterface) *MergerBuilder {
	b.services = append(b.services, services...)
	return b
}

// WithMerger sets the result merger strategy.
func (b *MergerBuilder) WithMerger(merger ResultMerger) *MergerBuilder {
	b.merger = merger
	return b
}

// WithFirstWins uses the first-wins merge strategy.
func (b *MergerBuilder) WithFirstWins(timeout time.Duration) *MergerBuilder {
	b.merger = &FirstWinsMerger{Timeout: timeout}
	return b
}

// WithAllResults uses the all-results merge strategy.
func (b *MergerBuilder) WithAllResults(timeout time.Duration) *MergerBuilder {
	b.merger = &AllResultsMerger{Timeout: timeout}
	return b
}

// WithTimeout sets the timeout for merger operations.
func (b *MergerBuilder) WithTimeout(timeout time.Duration) *MergerBuilder {
	b.config.Timeout = timeout
	return b
}

// WithConcurrency sets the maximum concurrency for parallel execution.
func (b *MergerBuilder) WithConcurrency(maxConcurrency int) *MergerBuilder {
	b.config.MaxConcurrency = maxConcurrency
	return b
}

// WithErrorHandling sets the error handling strategy for the merger.
func (b *MergerBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *MergerBuilder {
	b.config.ErrorHandling = strategy
	return b
}

// Build constructs the final merged service composition.
func (b *MergerBuilder) Build() *MergedService {
	if b.merger == nil {
		b.merger = &AllResultsMerger{Timeout: b.config.Timeout}
	}

	return &MergedService{
		services:  b.services,
		merger:    b.merger,
		config:    b.config,
		validator: DefaultValidator(),
	}
}
