package sdk

import (
	"context"
	"log"
	"reflect"
)

// CreateHybridEndorService creates an EndorService from a hybrid service with automatic initialization
// This is the main function developers should use - it handles initialization transparently
func CreateHybridEndorService[T ResourceInstanceInterface](
	resourceId string,
	description string,
	configurator func(*EndorHybridService[T]) error,
) EndorService {
	// Create the hybrid service
	hybridService := NewEndorHybridService[T](resourceId, description)

	// Apply custom configuration if provided
	if configurator != nil {
		if err := configurator(hybridService); err != nil {
			log.Printf("Warning: Failed to configure hybrid service %s: %v", resourceId, err)
		}
	}

	// Initialize the service automatically
	ctx := context.Background()
	err := hybridService.InitializeHybridService(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize hybrid service %s: %v. Using static schema only.", resourceId, err)

		// Fallback: initialize with static schema only
		err = hybridService.initializeWithStaticSchemaOnly()
		if err != nil {
			panic(err) // This should never happen
		}
	}

	// Return the EndorService
	return hybridService.GetEndorService()
}

// CreateSimpleHybridEndorService creates a simple hybrid EndorService with no customization
func CreateSimpleHybridEndorService[T ResourceInstanceInterface](resourceId, description string) EndorService {
	return CreateHybridEndorService[T](resourceId, description, nil)
}

// initializeWithStaticSchemaOnly initializes the service with only the static schema (fallback)
func (h *EndorHybridService[T]) initializeWithStaticSchemaOnly() error {
	var zeroT T
	h.rootSchema = NewSchemaByType(reflect.TypeOf(zeroT))

	autogenerateID := true
	h.repository = NewResourceInstanceRepository[T](h.resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	h.BaseService = h.buildService()
	h.initialized = true
	return nil
}
