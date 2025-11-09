package sdk

import (
	"context"
	"fmt"
	"log"
)

// HybridServiceManager manages a collection of hybrid services
type HybridServiceManager struct {
	services map[string]*EndorHybridServiceWrapper
}

// EndorHybridServiceWrapper wraps any EndorHybridService to provide a common interface
type EndorHybridServiceWrapper struct {
	resourceId  string
	description string
	initializer func(context.Context) (EndorService, error)
}

// NewHybridServiceManager creates a new manager for hybrid services
func NewHybridServiceManager() *HybridServiceManager {
	return &HybridServiceManager{
		services: make(map[string]*EndorHybridServiceWrapper),
	}
}

// RegisterHybridService registers a hybrid service with the manager
func RegisterHybridService[T ResourceInstanceInterface](
	manager *HybridServiceManager,
	resourceId string,
	description string,
	configurator func(*EndorHybridService[T]) error,
) {
	wrapper := &EndorHybridServiceWrapper{
		resourceId:  resourceId,
		description: description,
		initializer: func(ctx context.Context) (EndorService, error) {
			// Create the hybrid service
			hybridService := NewEndorHybridService[T](resourceId, description)

			// Apply custom configuration if provided
			if configurator != nil {
				if err := configurator(hybridService); err != nil {
					return EndorService{}, fmt.Errorf("failed to configure hybrid service %s: %w", resourceId, err)
				}
			}

			// Initialize the service using the new approach
			err := hybridService.InitializeHybridService(ctx)
			if err != nil {
				log.Printf("Warning: Failed to initialize hybrid service %s: %v. Using static schema only.", resourceId, err)
				// Fallback to static schema only
				err = hybridService.initializeWithStaticSchemaOnly()
				if err != nil {
					return EndorService{}, fmt.Errorf("failed to fallback initialize service %s: %w", resourceId, err)
				}
			}

			// Return the EndorService
			return hybridService.GetEndorService(), nil
		},
	}

	manager.services[resourceId] = wrapper
}

// InitializeAll initializes all registered hybrid services
func (m *HybridServiceManager) InitializeAll(ctx context.Context) ([]EndorService, error) {
	var services []EndorService
	var errors []error

	for resourceId, wrapper := range m.services {
		service, err := wrapper.initializer(ctx)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to initialize service %s: %w", resourceId, err))
			continue
		}
		services = append(services, service)
		log.Printf("Successfully initialized hybrid service: %s", resourceId)
	}

	if len(errors) > 0 {
		// Return partial success with error details
		var errorMsg string
		for _, err := range errors {
			errorMsg += err.Error() + "; "
		}
		return services, fmt.Errorf("some services failed to initialize: %s", errorMsg)
	}

	return services, nil
}

// GetServiceNames returns all registered service names
func (m *HybridServiceManager) GetServiceNames() []string {
	var names []string
	for name := range m.services {
		names = append(names, name)
	}
	return names
}

// Helper function to create a simple hybrid service with minimal configuration
// Deprecated: Use CreateSimpleHybridEndorService from hybrid_service_factory.go instead
func CreateSimpleHybridService[T ResourceInstanceInterface](resourceId, description string) EndorService {
	// Use the new factory approach
	return CreateSimpleHybridEndorService[T](resourceId, description)
}

// Helper function for quick setup with custom handlers
// Deprecated: Use CreateHybridEndorService from hybrid_service_factory.go instead
func CreateCustomHybridService[T ResourceInstanceInterface](
	resourceId, description string,
	configurator func(*EndorHybridService[T]) error,
) EndorService {
	// Use the new factory approach
	return CreateHybridEndorService(resourceId, description, configurator)
}
