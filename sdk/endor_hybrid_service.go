package sdk

import (
	"context"
	"fmt"
	"reflect"
)

// HybridServiceInitializer interface to identify hybrid services that need initialization
type HybridServiceInitializer interface {
	InitializeHybridService(ctx context.Context) error
	IsInitialized() bool
	GetEndorService() EndorService
}

// EndorHybridService wraps EndorService to handle hybrid resources
// T: the static base model defined by the developer
type EndorHybridService[T ResourceInstanceInterface] struct {
	BaseService EndorService
	resource    string
	description string
	rootSchema  *RootSchema
	repository  *ResourceInstanceRepository[T]
	initialized bool

	// Customizable handlers that developers can override
	SchemaHandler   func(*EndorContext[NoPayload]) (*Response[any], error)
	InstanceHandler func(*EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error)
	ListHandler     func(*EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error)
	CreateHandler   func(*EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error)
	UpdateHandler   func(*EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error)
	DeleteHandler   func(*EndorContext[ReadInstanceDTO]) (*Response[any], error)
}

// NewEndorHybridService creates a new hybrid service
// T: the static base model that implements ResourceInstanceInterface
// resource: the resource identifier
// description: human-readable description
func NewEndorHybridService[T ResourceInstanceInterface](resource string, description string) *EndorHybridService[T] {
	service := &EndorHybridService[T]{
		resource:    resource,
		description: description,
		initialized: false,
	}

	// Set default handlers
	service.SchemaHandler = service.defaultSchema
	service.InstanceHandler = service.defaultInstance
	service.ListHandler = service.defaultList
	service.CreateHandler = service.defaultCreate
	service.UpdateHandler = service.defaultUpdate
	service.DeleteHandler = service.defaultDelete

	return service
}

// Implement HybridServiceInitializer interface
func (h *EndorHybridService[T]) InitializeHybridService(ctx context.Context) error {
	if h.initialized {
		return nil // Already initialized
	}

	err := h.Initialize(ctx)
	if err != nil {
		return err
	}

	// Build the service after initialization
	h.BaseService = h.buildService()
	h.initialized = true
	return nil
}

func (h *EndorHybridService[T]) IsInitialized() bool {
	return h.initialized
}

// Initialize loads the additional attributes from MongoDB and sets up the service
func (h *EndorHybridService[T]) Initialize(ctx context.Context) error {
	// Get the base schema from the static type T
	var zeroT T
	baseSchema := NewSchemaByType(reflect.TypeOf(zeroT))

	// Try to get additional attributes from MongoDB
	resource, err := h.getResourceFromMongoDB(ctx, h.resource)
	var additionalSchema *RootSchema

	if err != nil {
		// If resource doesn't exist in MongoDB, create empty additional schema
		additionalSchema = &RootSchema{
			Schema: Schema{
				Type:       ObjectType,
				Properties: &map[string]Schema{},
			},
			Definitions: map[string]Schema{},
		}
	} else {
		// Parse additional attributes from MongoDB
		additionalSchema, err = resource.UnmarshalAdditionalAttributes()
		if err != nil {
			return fmt.Errorf("failed to parse additional attributes for resource %s: %w", h.resource, err)
		}
	}

	// Merge base schema with additional schema
	h.rootSchema = h.mergeSchemas(baseSchema, additionalSchema)

	// Create repository with auto-generated ID
	autogenerateID := true
	h.repository = NewResourceInstanceRepository[T](h.resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return nil
}

// buildService creates the final EndorService (internal method)
func (h *EndorHybridService[T]) buildService() EndorService {
	if h.rootSchema == nil {
		panic("EndorHybridService must be initialized before building. Call InitializeHybridService() first.")
	}

	return EndorService{
		Resource:    h.resource,
		Description: h.description,
		Methods: map[string]EndorServiceAction{
			"schema": NewAction(
				h.SchemaHandler,
				fmt.Sprintf("Get the schema of the %s (%s)", h.resource, h.description),
			),
			"list": NewAction(
				h.ListHandler,
				fmt.Sprintf("Search for available list of %s (%s)", h.resource, h.description),
			),
			"create": NewConfigurableAction(
				EndorServiceActionOptions{
					Description:     fmt.Sprintf("Create the instance of %s (%s)", h.resource, h.description),
					Public:          false,
					ValidatePayload: true,
					InputSchema: &RootSchema{
						Schema: Schema{
							Reference: fmt.Sprintf("#/$defs/CreateDTO_%s", h.resource),
						},
						Definitions: func() map[string]Schema {
							defs := map[string]Schema{
								fmt.Sprintf("CreateDTO_%s", h.resource): {
									Type: ObjectType,
									Properties: &map[string]Schema{
										"data": h.rootSchema.Schema,
									},
								},
							}
							for k, v := range h.rootSchema.Definitions {
								defs[k] = v
							}
							return defs
						}(),
					},
				},
				h.CreateHandler,
			),
			"instance": NewAction(
				h.InstanceHandler,
				fmt.Sprintf("Get the instance of %s (%s)", h.resource, h.description),
			),
			"update": NewConfigurableAction(
				EndorServiceActionOptions{
					Description:     fmt.Sprintf("Update the existing instance of %s (%s)", h.resource, h.description),
					Public:          false,
					ValidatePayload: true,
					InputSchema: &RootSchema{
						Schema: Schema{
							Reference: fmt.Sprintf("#/$defs/UpdateByIdDTO_%s", h.resource),
						},
						Definitions: func() map[string]Schema {
							defs := map[string]Schema{
								fmt.Sprintf("UpdateByIdDTO_%s", h.resource): {
									Type: ObjectType,
									Properties: &map[string]Schema{
										"id": {
											Type: StringType,
										},
										"data": h.rootSchema.Schema,
									},
								},
							}
							for k, v := range h.rootSchema.Definitions {
								defs[k] = v
							}
							return defs
						}(),
					},
				},
				h.UpdateHandler,
			),
			"delete": NewAction(
				h.DeleteHandler,
				fmt.Sprintf("Delete the existing instance of %s (%s)", h.resource, h.description),
			),
		},
	}
}

// GetEndorService returns the underlying EndorService (for SDK use)
func (h *EndorHybridService[T]) GetEndorService() EndorService {
	if !h.initialized {
		panic("EndorHybridService must be initialized before use. The SDK should handle this automatically.")
	}
	return h.BaseService
}

// Default CRUD handlers - developers can override these

func (h *EndorHybridService[T]) defaultSchema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(h.rootSchema).Build(), nil
}

func (h *EndorHybridService[T]) defaultInstance(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error) {
	instance, err := h.repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[T]]().AddData(&instance).AddSchema(h.rootSchema).Build(), nil
}

func (h *EndorHybridService[T]) defaultList(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error) {
	list, err := h.repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[T]]().AddData(&list).AddSchema(h.rootSchema).Build(), nil
}

func (h *EndorHybridService[T]) defaultCreate(c *EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
	created, err := h.repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[T]]().AddData(created).AddSchema(h.rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build(), nil
}

func (h *EndorHybridService[T]) defaultUpdate(c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
	updated, err := h.repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[T]]().AddData(updated).AddSchema(h.rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.resource))).Build(), nil
}

func (h *EndorHybridService[T]) defaultDelete(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
	err := h.repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.resource))).Build(), nil
}

// Utility methods for developers

// SetSchemaHandler allows developers to customize the schema endpoint
func (h *EndorHybridService[T]) SetSchemaHandler(handler func(*EndorContext[NoPayload]) (*Response[any], error)) {
	h.SchemaHandler = handler
}

// SetInstanceHandler allows developers to customize the instance endpoint
func (h *EndorHybridService[T]) SetInstanceHandler(handler func(*EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error)) {
	h.InstanceHandler = handler
}

// SetListHandler allows developers to customize the list endpoint
func (h *EndorHybridService[T]) SetListHandler(handler func(*EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error)) {
	h.ListHandler = handler
}

// SetCreateHandler allows developers to customize the create endpoint
func (h *EndorHybridService[T]) SetCreateHandler(handler func(*EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error)) {
	h.CreateHandler = handler
}

// SetUpdateHandler allows developers to customize the update endpoint
func (h *EndorHybridService[T]) SetUpdateHandler(handler func(*EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error)) {
	h.UpdateHandler = handler
}

// SetDeleteHandler allows developers to customize the delete endpoint
func (h *EndorHybridService[T]) SetDeleteHandler(handler func(*EndorContext[ReadInstanceDTO]) (*Response[any], error)) {
	h.DeleteHandler = handler
}

// AddCustomAction allows developers to add custom actions beyond CRUD
func (h *EndorHybridService[T]) AddCustomAction(name string, action EndorServiceAction) {
	if h.BaseService.Methods == nil {
		h.BaseService.Methods = make(map[string]EndorServiceAction)
	}
	h.BaseService.Methods[name] = action
}

// GetRepository provides access to the underlying repository for advanced operations
func (h *EndorHybridService[T]) GetRepository() *ResourceInstanceRepository[T] {
	return h.repository
}

// GetRootSchema provides access to the merged schema (base + additional attributes)
func (h *EndorHybridService[T]) GetRootSchema() *RootSchema {
	return h.rootSchema
}

// mergeSchemas combines the base static schema with additional dynamic schema
func (h *EndorHybridService[T]) mergeSchemas(baseSchema *RootSchema, additionalSchema *RootSchema) *RootSchema {
	// Start with a copy of the base schema
	mergedSchema := &RootSchema{
		Schema: Schema{
			Type:       ObjectType,
			Properties: &map[string]Schema{},
		},
		Definitions: make(map[string]Schema),
	}

	// Copy base properties
	if baseSchema != nil && baseSchema.Schema.Properties != nil {
		for k, v := range *baseSchema.Schema.Properties {
			(*mergedSchema.Schema.Properties)[k] = v
		}
	}

	// Copy base definitions
	if baseSchema != nil && baseSchema.Definitions != nil {
		for k, v := range baseSchema.Definitions {
			mergedSchema.Definitions[k] = v
		}
	}

	// Merge additional properties (these override base properties if same key)
	if additionalSchema != nil && additionalSchema.Schema.Properties != nil {
		for k, v := range *additionalSchema.Schema.Properties {
			(*mergedSchema.Schema.Properties)[k] = v
		}
	}

	// Merge additional definitions
	if additionalSchema != nil && additionalSchema.Definitions != nil {
		for k, v := range additionalSchema.Definitions {
			mergedSchema.Definitions[k] = v
		}
	}

	return mergedSchema
}

// getResourceFromMongoDB retrieves a resource definition from MongoDB
func (h *EndorHybridService[T]) getResourceFromMongoDB(ctx context.Context, resourceId string) (*Resource, error) {
	if !GetConfig().EndorDynamicResourcesEnabled {
		return nil, fmt.Errorf("dynamic resources are not enabled")
	}

	client, err := GetMongoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB client: %w", err)
	}

	database := client.Database(GetConfig().EndorDynamicResourceDBName)
	collection := database.Collection(COLLECTION_RESOURCES)

	var resource Resource
	filter := map[string]interface{}{"_id": resourceId}
	err = collection.FindOne(ctx, filter).Decode(&resource)
	if err != nil {
		return nil, fmt.Errorf("resource %s not found: %w", resourceId, err)
	}

	return &resource, nil
}

// InitializeHybridServices automatically initializes all hybrid services in a slice
// This function is called by the SDK to handle hybrid services transparently
func InitializeHybridServices(ctx context.Context, services *[]interface{}) ([]EndorService, error) {
	var endorServices []EndorService

	// Process each service
	for _, service := range *services {
		if hybridService, ok := service.(HybridServiceInitializer); ok {
			// This is a hybrid service
			if !hybridService.IsInitialized() {
				err := hybridService.InitializeHybridService(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to initialize hybrid service: %w", err)
				}
			}
			// Get the initialized EndorService
			endorServices = append(endorServices, hybridService.GetEndorService())
		} else if endorService, ok := service.(EndorService); ok {
			// This is a regular EndorService
			endorServices = append(endorServices, endorService)
		} else {
			return nil, fmt.Errorf("invalid service type: %T", service)
		}
	}

	return endorServices, nil
}

// InitializeHybridServicesInPlace initializes hybrid services in place within a slice of EndorService
// This is used by the SDK server initialization
func InitializeHybridServicesInPlace(microserviceId string, services *[]EndorService) error {
	ctx := context.Background()

	// We need to check if any of the services are actually hybrid services
	// stored as interface{} but passed as EndorService
	// For now, we'll add a registry mechanism

	// Check the global hybrid service registry
	if hybridRegistry != nil {
		initializedServices, err := hybridRegistry.InitializeAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize hybrid services: %w", err)
		}
		// Append initialized hybrid services to the existing services
		*services = append(*services, initializedServices...)
	}

	return nil
}

// Global registry for hybrid services
var hybridRegistry *HybridServiceManager

// RegisterGlobalHybridService registers a hybrid service globally for automatic initialization
func RegisterGlobalHybridService[T ResourceInstanceInterface](
	resourceId string,
	description string,
	configurator func(*EndorHybridService[T]) error,
) {
	if hybridRegistry == nil {
		hybridRegistry = NewHybridServiceManager()
	}
	RegisterHybridService(hybridRegistry, resourceId, description, configurator)
}
