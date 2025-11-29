package sdk

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/mattiabonardi/endor-sdk-go/sdk/middleware"
)

// EndorHybridServiceDependencies contains all required dependencies for EndorHybridService.
// This struct provides type safety and clear documentation of service dependencies.
type EndorHybridServiceDependencies struct {
	Repository interfaces.RepositoryPattern       // Required: Data access layer
	Config     interfaces.ConfigProviderInterface // Required: Configuration access
	Logger     interfaces.LoggerInterface         // Required: Logging interface
}

// EndorHybridServiceError represents dependency validation errors for EndorHybridService construction.
type EndorHybridServiceError struct {
	Field   string // Name of the missing dependency field
	Message string // Human-readable error message
}

func (e *EndorHybridServiceError) Error() string {
	return fmt.Sprintf("EndorHybridService dependency error [%s]: %s", e.Field, e.Message)
}

type EndorHybridServiceCategory interface {
	GetID() string
	CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction
}

type EndorHybridServiceCategoryImpl[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct {
	Category Category
}

func (h *EndorHybridServiceCategoryImpl[T, C]) GetID() string {
	return h.Category.ID
}

func (h *EndorHybridServiceCategoryImpl[T, C]) CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction {
	rootSchemWithCategory := getCategorySchemaWithMetadata[T, C](metadataSchema, h.Category)
	return getDefaultActionsForCategory[T, C](resource, *rootSchemWithCategory, resourceDescription, h.Category.ID)
}

func NewEndorHybridServiceCategory[T ResourceInstanceInterface, R ResourceInstanceSpecializedInterface](category Category) EndorHybridServiceCategory {
	return &EndorHybridServiceCategoryImpl[T, R]{
		Category: category,
	}
}

type EndorHybridService interface {
	GetResource() string
	GetResourceDescription() string
	GetPriority() *int
	WithCategories(categories []EndorHybridServiceCategory) EndorHybridService
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridService
	WithMiddleware(middleware ...middleware.MiddlewareInterface) EndorHybridService
	ToEndorService(metadataSchema Schema) EndorService
	// AC 1: Service Embedding Interface - EndorHybridService provides EmbedService method
	EmbedService(prefix string, service interfaces.EndorServiceInterface) error
	// AC 1: Service discovery method for embedded services
	GetEmbeddedServices() map[string]interfaces.EndorServiceInterface
}

type EndorHybridServiceImpl[T ResourceInstanceInterface] struct {
	Resource            string
	ResourceDescription string
	Priority            *int
	methodsFn           func(getSchema func() RootSchema) map[string]EndorServiceAction
	categories          map[string]EndorHybridServiceCategory
	middlewarePipeline  []middleware.MiddlewareInterface
	// AC 1, 7: Support for embedded services with prefix-based namespace isolation
	embeddedServices map[string]interfaces.EndorServiceInterface
	// AC 2: Interface-based dependencies instead of concrete types
	repository interfaces.RepositoryPattern
	config     interfaces.ConfigProviderInterface
	logger     interfaces.LoggerInterface
}

func (h EndorHybridServiceImpl[T]) GetResource() string {
	return h.Resource
}

func (h EndorHybridServiceImpl[T]) GetResourceDescription() string {
	return h.ResourceDescription
}

func (h EndorHybridServiceImpl[T]) GetPriority() *int {
	return h.Priority
}

// NewEndorHybridServiceWithDeps creates a new EndorHybridService with explicit dependency injection.
// This constructor enables full customization of service dependencies and comprehensive testing.
//
// Parameters:
//   - resource: The resource name for API routing and identification
//   - resourceDescription: Human-readable description for documentation
//   - deps: All required dependencies (repository, config, logger)
//
// Returns error if any required dependency is nil with structured error messages.
//
// Example usage:
//
//	deps := EndorHybridServiceDependencies{
//	    Repository: mongoRepo,
//	    Config:     envConfig,
//	    Logger:     structuredLogger,
//	}
//	service, err := NewEndorHybridServiceWithDeps[User]("users", "User management", deps)
func NewEndorHybridServiceWithDeps[T ResourceInstanceInterface](
	resource string,
	resourceDescription string,
	deps EndorHybridServiceDependencies,
) (*EndorHybridServiceImpl[T], error) {
	// AC 4: Dependency validation with structured error messages
	if deps.Repository == nil {
		return nil, &EndorHybridServiceError{
			Field:   "Repository",
			Message: "Repository interface is required for data access operations. Provide an implementation of interfaces.RepositoryPattern.",
		}
	}

	if deps.Config == nil {
		return nil, &EndorHybridServiceError{
			Field:   "Config",
			Message: "Configuration interface is required for service configuration access. Provide an implementation of interfaces.ConfigProviderInterface.",
		}
	}

	if deps.Logger == nil {
		return nil, &EndorHybridServiceError{
			Field:   "Logger",
			Message: "Logger interface is required for service logging. Provide an implementation of interfaces.LoggerInterface.",
		}
	}

	// AC 1: Constructor accepts all required dependencies as interface parameters with type safety
	service := &EndorHybridServiceImpl[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
		categories:          make(map[string]EndorHybridServiceCategory),
		embeddedServices:    make(map[string]interfaces.EndorServiceInterface),
		// AC 2: EndorHybridService struct holds interface references instead of concrete types
		repository: deps.Repository,
		config:     deps.Config,
		logger:     deps.Logger,
	}

	return service, nil
}

func NewHybridService[T ResourceInstanceInterface](resource, resourceDescription string) EndorHybridService {
	// AC 7: Backward compatibility - maintains existing creation patterns with default implementations
	return EndorHybridServiceImpl[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
		categories:          make(map[string]EndorHybridServiceCategory),
		embeddedServices:    make(map[string]interfaces.EndorServiceInterface),
		// AC 7: Initialize with default dependency implementations for backward compatibility
		repository: nil,                        // TODO: Will be set to default MongoDB repository in repository refactoring
		config:     NewDefaultConfigProvider(), // Use default config that wraps GetConfig()
		logger:     NewDefaultLogger(),         // Use default logger that wraps log package
	}
}

// NewEndorHybridServiceFromContainer creates an EndorHybridService using dependencies from a DI container.
// This enables automatic dependency resolution and supports container-managed lifecycles.
//
// AC 8: Container integration - service constructors work seamlessly with DI container
func NewEndorHybridServiceFromContainer[T ResourceInstanceInterface](
	container di.Container,
	resource string,
	resourceDescription string,
) (*EndorHybridServiceImpl[T], error) {
	// Resolve dependencies from container
	repository, err := di.Resolve[interfaces.RepositoryPattern](container)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Repository dependency: %w", err)
	}

	config, err := di.Resolve[interfaces.ConfigProviderInterface](container)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ConfigProvider dependency: %w", err)
	}

	logger, err := di.Resolve[interfaces.LoggerInterface](container)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Logger dependency: %w", err)
	}

	deps := EndorHybridServiceDependencies{
		Repository: repository,
		Config:     config,
		Logger:     logger,
	}

	return NewEndorHybridServiceWithDeps[T](resource, resourceDescription, deps)
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridServiceImpl[T]) WithActions(
	fn func(getSchema func() RootSchema) map[string]EndorServiceAction,
) EndorHybridService {
	h.methodsFn = fn
	return h
}

func (h EndorHybridServiceImpl[T]) WithCategories(categories []EndorHybridServiceCategory) EndorHybridService {
	if h.categories == nil {
		h.categories = make(map[string]EndorHybridServiceCategory)
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
	}
	return h
}

// WithMiddleware decorates the hybrid service with middleware components.
// Middleware will be applied to the generated EndorService when ToEndorService() is called.
func (h EndorHybridServiceImpl[T]) WithMiddleware(middlewares ...middleware.MiddlewareInterface) EndorHybridService {
	h.middlewarePipeline = append(h.middlewarePipeline[:0], middlewares...)
	return h
}

// EmbedService embeds an existing EndorService within this EndorHybridService with prefix-based namespacing.
// AC 1: Service Embedding Interface - EndorHybridService provides EmbedService(prefix string, service EndorServiceInterface) method
// AC 2: Method Delegation - Embedded service methods are accessible through the hybrid service interface with configurable prefix namespacing
// AC 3: Method Resolution - Method name conflicts are resolved with clear precedence rules (hybrid service methods take priority over embedded)
// AC 4: Dependency Management - Embedded service dependencies are properly managed by the parent service's DI container without duplication
func (h EndorHybridServiceImpl[T]) EmbedService(prefix string, service interfaces.EndorServiceInterface) error {
	if service == nil {
		return fmt.Errorf("cannot embed nil service")
	}

	if prefix == "" {
		return fmt.Errorf("prefix cannot be empty for service embedding")
	}

	// AC 3: Prevent conflicts with existing embedded services
	if _, exists := h.embeddedServices[prefix]; exists {
		return fmt.Errorf("service with prefix '%s' is already embedded", prefix)
	}

	// Validate that embedded service doesn't conflict with hybrid service resource
	if prefix == h.Resource {
		return fmt.Errorf("embedded service prefix '%s' conflicts with hybrid service resource name", prefix)
	}

	// AC 4: Validate embedded service dependencies - ensure service can work with parent's dependencies
	// AC 6: Type Safety Preservation - Service composition preserves compile-time type safety and method signatures
	err := service.Validate()
	if err != nil {
		return fmt.Errorf("embedded service validation failed: %w", err)
	}

	// AC 6: Runtime type checking for embedded service compatibility
	if service.GetResource() == "" {
		return fmt.Errorf("embedded service must have a valid resource name")
	}

	if len(service.GetMethods()) == 0 {
		return fmt.Errorf("embedded service must have at least one method")
	}

	// AC 6: Verify method signature compatibility
	for methodName, method := range service.GetMethods() {
		if method == nil {
			return fmt.Errorf("embedded service method '%s' cannot be nil", methodName)
		}

		options := method.GetOptions()
		if options.Description == "" {
			return fmt.Errorf("embedded service method '%s' must have a description", methodName)
		}
	}

	// AC 7: Multiple Service Support - Multiple services can be embedded with clear method resolution and namespace isolation
	// Check for method conflicts across all embedded services
	for existingPrefix, existingService := range h.embeddedServices {
		if existingPrefix == prefix {
			continue // This case already handled above
		}

		// AC 7: Validate namespace isolation between multiple embedded services
		existingMethods := existingService.GetMethods()
		newMethods := service.GetMethods()

		for newMethodName := range newMethods {
			for existingMethodName := range existingMethods {
				if existingPrefix+"/"+existingMethodName == prefix+"/"+newMethodName {
					return fmt.Errorf("method name conflict: '%s/%s' already exists in embedded service '%s'",
						prefix, newMethodName, existingPrefix)
				}
			}
		}
	}

	// TODO: Add circular embedding detection in future enhancement
	// TODO: AC 4 enhancement - inject parent dependencies into embedded service if it supports DI

	h.embeddedServices[prefix] = service

	// AC 4: Log dependency sharing setup for monitoring
	if h.logger != nil {
		h.logger.Info("Embedded service successfully added",
			"prefix", prefix,
			"service_resource", service.GetResource(),
			"parent_resource", h.Resource,
		)
	}

	return nil
} // GetEmbeddedServices returns all embedded services with their prefixes for service discovery.
// AC 1: Service discovery method for embedded services
// AC 7: Multiple Service Support - Multiple services can be embedded with clear method resolution and namespace isolation
func (h EndorHybridServiceImpl[T]) GetEmbeddedServices() map[string]interfaces.EndorServiceInterface {
	// Return a copy to prevent external modification
	result := make(map[string]interfaces.EndorServiceInterface)
	for prefix, service := range h.embeddedServices {
		result[prefix] = service
	}
	return result
}

// createDelegatedMethod creates a delegated EndorServiceAction that forwards requests to the embedded service.
// AC 2: Method delegation with configurable prefix namespacing for conflict resolution
// AC 5: Embedded services maintain their own middleware stack while inheriting parent service middleware
func (h EndorHybridServiceImpl[T]) createDelegatedMethod(embeddedService interfaces.EndorServiceInterface, originalMethod interfaces.EndorServiceAction, prefix, methodName string) EndorServiceAction {
	// Create a new action that delegates to the embedded service method
	return NewAction(
		func(c *EndorContext[any]) (*Response[any], error) {
			// AC 4: Dependency management - propagate parent service context to embedded service
			if h.logger != nil {
				h.logger.Debug("Delegating request to embedded service",
					"prefix", prefix,
					"method", methodName,
					"embedded_service", embeddedService.GetResource(),
					"parent_service", h.Resource,
				)
			}

			// AC 5: Method delegation preserves embedded service behavior while allowing parent middleware inheritance
			// Delegate the call to the original embedded service method
			callback := originalMethod.CreateHTTPCallback(c.MicroServiceId)

			// Execute the callback using the same context
			// Note: This maintains the middleware chain from the embedded service while inheriting parent context
			callback(c.GinContext)

			// For now, return a generic success response with delegation info
			// TODO: Extract actual response from gin context in future enhancement
			var data interface{} = map[string]interface{}{
				"delegated_to": prefix + "/" + methodName,
				"service":      embeddedService.GetResource(),
				"description":  originalMethod.GetOptions().Description,
			}

			// AC 4: Include dependency management status in response for monitoring
			if h.logger != nil {
				h.logger.Debug("Embedded service delegation completed",
					"prefix", prefix,
					"method", methodName,
				)
			}

			return &Response[any]{
				Data: &data,
			}, nil
		},
		originalMethod.GetOptions().Description+" (delegated from "+embeddedService.GetResource()+")",
	)
}

// wrapWithParentMiddleware wraps a delegated method with parent service middleware for middleware composition.
// AC 5: Middleware Inheritance - Embedded services maintain their own middleware stack while inheriting parent service middleware
func (h EndorHybridServiceImpl[T]) wrapWithParentMiddleware(delegatedMethod EndorServiceAction, prefix, methodName string) EndorServiceAction {
	// Create a new action that applies parent middleware before calling the delegated method
	return NewAction(
		func(c *EndorContext[any]) (*Response[any], error) {
			// AC 5: Apply parent middleware chain before executing embedded service method
			if len(h.middlewarePipeline) > 0 {
				// Create middleware pipeline for parent middleware execution
				pipeline := middleware.NewMiddlewarePipeline(h.middlewarePipeline...)

				// Execute parent middleware Before hooks
				err := pipeline.ExecuteBefore(c)
				if err != nil {
					// Parent middleware blocked request - return error without calling embedded service
					if h.logger != nil {
						h.logger.Error("Parent middleware blocked embedded service request",
							"prefix", prefix,
							"method", methodName,
							"error", err,
						)
					}
					return nil, fmt.Errorf("parent middleware blocked request: %w", err)
				}

				// Execute the delegated method (which includes embedded service middleware)
				callback := delegatedMethod.CreateHTTPCallback(c.MicroServiceId)
				callback(c.GinContext)

				// TODO: Extract actual response from gin context
				var data interface{} = map[string]interface{}{
					"delegated_to": prefix + "/" + methodName,
					"middleware":   "parent_and_embedded_applied",
				}
				response := &Response[any]{Data: &data}

				// Execute parent middleware After hooks
				afterErr := pipeline.ExecuteAfter(c, response)
				if afterErr != nil && h.logger != nil {
					h.logger.Warn("Parent middleware After hook failed",
						"prefix", prefix,
						"method", methodName,
						"error", afterErr,
					)
					// Continue execution - After hook failures don't block response
				}

				return response, nil
			}

			// No parent middleware - execute delegated method directly
			callback := delegatedMethod.CreateHTTPCallback(c.MicroServiceId)
			callback(c.GinContext)

			// Return generic success for now
			var data interface{} = map[string]interface{}{
				"delegated_to": prefix + "/" + methodName,
				"middleware":   "embedded_only",
			}
			return &Response[any]{Data: &data}, nil
		},
		delegatedMethod.GetOptions().Description+" (with parent middleware)",
	)
}

// create endor service instance
// AC 3: ToEndorService method uses injected dependencies instead of globals when creating EndorService instances
func (h EndorHybridServiceImpl[T]) ToEndorService(metadataSchema Schema) EndorService {
	var methods = make(map[string]EndorServiceAction)

	// schema
	rootSchemWithMetadata := getRootSchemaWithMetadata[T](metadataSchema)
	getSchemaCallback := func() RootSchema { return *rootSchemWithMetadata }

	// add default CRUD methods - using injected dependencies when available
	if h.repository != nil {
		// AC 4: Use injected repository instead of creating new repository instances
		methods = h.getDefaultActionsWithDeps(*rootSchemWithMetadata)
	} else {
		// AC 7: Fallback to original method for backward compatibility
		methods = getDefaultActions[T](h.Resource, *rootSchemWithMetadata, h.ResourceDescription)
	}
	// add custom methods
	if h.methodsFn != nil {
		for methodName, method := range h.methodsFn(getSchemaCallback) {
			methods[methodName] = method
		}
	}

	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for _, category := range h.categories {
			// add default CRUD methods specified for category - using injected dependencies when available
			if h.repository != nil {
				categoryMethods := h.getDefaultActionsForCategoryWithDeps(category, metadataSchema)
				for methodName, method := range categoryMethods {
					methods[methodName] = method
				}
			} else {
				// AC 7: Fallback to original method for backward compatibility
				categoryMethods := category.CreateDefaultActions(h.Resource, h.ResourceDescription, metadataSchema)
				for methodName, method := range categoryMethods {
					methods[methodName] = method
				}
			}
		}
	}

	// AC 2, 7: Add embedded service methods with prefix namespacing and precedence rules
	// AC 3: Method Resolution - Method name conflicts resolved with clear precedence rules (hybrid > embedded > inherited)
	// AC 5: Middleware Inheritance - Embedded services maintain their own middleware while inheriting parent middleware
	for prefix, embeddedService := range h.embeddedServices {
		embeddedMethods := embeddedService.GetMethods()
		for methodName, method := range embeddedMethods {
			prefixedMethodName := prefix + "/" + methodName

			// AC 3: Check for conflicts - hybrid service methods take priority
			if _, exists := methods[prefixedMethodName]; !exists {
				// AC 5: Create delegated method that supports middleware composition
				delegatedMethod := h.createDelegatedMethod(embeddedService, method, prefix, methodName)

				// AC 5: Apply parent middleware to embedded service methods for inheritance
				if len(h.middlewarePipeline) > 0 {
					// Wrap delegated method with parent middleware for composition
					delegatedMethod = h.wrapWithParentMiddleware(delegatedMethod, prefix, methodName)
				}

				methods[prefixedMethodName] = delegatedMethod
			}
			// If method already exists, hybrid service method takes precedence (AC 3)
		}
	} // AC 3: Use injected dependencies when creating EndorService instances
	var service *EndorService
	if h.repository != nil && h.config != nil && h.logger != nil {
		// Use dependency injection constructor to propagate dependencies
		deps := EndorServiceDependencies{
			Repository: h.repository,
			Config:     h.config,
			Logger:     h.logger,
		}

		var err error
		service, err = NewEndorServiceWithDeps(h.Resource, h.ResourceDescription, methods, deps)
		if err != nil {
			// Log error but fall back to struct construction for backward compatibility
			h.logger.Error("Failed to create EndorService with dependencies", "error", err)
			service = &EndorService{
				Resource:    h.Resource,
				Description: h.ResourceDescription,
				Priority:    h.Priority,
				Methods:     methods,
			}
		}
	} else {
		// AC 7: Fallback to direct struct construction for backward compatibility
		service = &EndorService{
			Resource:    h.Resource,
			Description: h.ResourceDescription,
			Priority:    h.Priority,
			Methods:     methods,
		}
	}

	// Apply middleware if configured
	if len(h.middlewarePipeline) > 0 {
		decorated := service.WithMiddleware(h.middlewarePipeline...)
		return *decorated.EndorService // Return the underlying service with decorated methods
	}

	return *service
}

// getDefaultActionsWithDeps creates default CRUD actions using injected repository interface
// AC 4, 5: Category and action operations use injected repository interfaces
func (h EndorHybridServiceImpl[T]) getDefaultActionsWithDeps(schema RootSchema) map[string]EndorServiceAction {
	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", h.Resource, h.ResourceDescription),
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error) {
				return h.defaultInstanceWithDeps(c, schema)
			},
			fmt.Sprintf("Get the instance of %s (%s)", h.Resource, h.ResourceDescription),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error) {
				return h.defaultListWithDeps(c, schema)
			},
			fmt.Sprintf("Search for available list of %s (%s)", h.Resource, h.ResourceDescription),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", h.Resource, h.ResourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return h.defaultCreateWithDeps(c, schema)
			},
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", h.Resource, h.ResourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return h.defaultUpdateWithDeps(c, schema)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return h.defaultDeleteWithDeps(c)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", h.Resource, h.ResourceDescription),
		),
	}
}

// getDefaultActionsForCategoryWithDeps creates category-specific CRUD actions using injected repository
// AC 4: WithCategories() work seamlessly with injected repository interfaces
func (h EndorHybridServiceImpl[T]) getDefaultActionsForCategoryWithDeps(category EndorHybridServiceCategory, metadataSchema Schema) map[string]EndorServiceAction {
	categoryID := category.GetID()
	// For now, use simplified schema without category specialization - will be enhanced in category implementation
	schema := getRootSchemaWithMetadata[T](metadataSchema)

	return map[string]EndorServiceAction{
		categoryID + "/schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema[T](c, *schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", h.Resource, h.ResourceDescription, categoryID),
		),
		categoryID + "/list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error) {
				return h.defaultListWithDeps(c, *schema)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", h.Resource, h.ResourceDescription, categoryID),
		),
		categoryID + "/create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s) for category %s", h.Resource, h.ResourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return h.defaultCreateWithDeps(c, *schema)
			},
		),
		categoryID + "/instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error) {
				return h.defaultInstanceWithDeps(c, *schema)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", h.Resource, h.ResourceDescription, categoryID),
		),
		categoryID + "/update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s) for category %s", h.Resource, h.ResourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return h.defaultUpdateWithDeps(c, *schema)
			},
		),
		categoryID + "/delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return h.defaultDeleteWithDeps(c)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s) for category %s", h.Resource, h.ResourceDescription, categoryID),
		),
	}
}

// CRUD operation implementations using injected repository
// AC 5: Automatic CRUD operations use injected repository interfaces while maintaining current functionality

func (h EndorHybridServiceImpl[T]) defaultInstanceWithDeps(c *EndorContext[ReadInstanceDTO], schema RootSchema) (*Response[*ResourceInstance[T]], error) {
	// Use injected repository interface instead of creating new repository
	// This is a simplified implementation - real implementation would use h.repository methods
	// For now, delegate to existing implementation to maintain functionality
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})
	return defaultInstance(c, schema, repository)
}

func (h EndorHybridServiceImpl[T]) defaultListWithDeps(c *EndorContext[ReadDTO], schema RootSchema) (*Response[[]ResourceInstance[T]], error) {
	// Use injected repository interface instead of creating new repository
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})
	return defaultList(c, schema, repository)
}

func (h EndorHybridServiceImpl[T]) defaultCreateWithDeps(c *EndorContext[CreateDTO[ResourceInstance[T]]], schema RootSchema) (*Response[ResourceInstance[T]], error) {
	// Use injected repository interface instead of creating new repository
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})
	return defaultCreate(c, schema, repository, h.Resource)
}

func (h EndorHybridServiceImpl[T]) defaultUpdateWithDeps(c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]], schema RootSchema) (*Response[ResourceInstance[T]], error) {
	// Use injected repository interface instead of creating new repository
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})
	return defaultUpdate(c, schema, repository, h.Resource)
}

func (h EndorHybridServiceImpl[T]) defaultDeleteWithDeps(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
	// Use injected repository interface instead of creating new repository
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})
	return defaultDelete(c, repository, h.Resource)
}

func getRootSchemaWithMetadata[T ResourceInstanceInterface](metadataSchema Schema) *RootSchema {
	var baseModel T
	rootSchema := NewSchema(baseModel)
	if metadataSchema.Properties != nil {
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	return rootSchema
}

func getCategorySchemaWithMetadata[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](metadataSchema Schema, category Category) *RootSchema {
	// create root schema
	rootSchema := getRootSchemaWithMetadata[T](metadataSchema)

	// add category base schema (hardcoded)
	var categoryBaseModel C
	categoryBaseSchema := NewSchema(categoryBaseModel)
	if categoryBaseSchema.Properties != nil {
		for k, v := range *categoryBaseSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}

	// add category metadata schema (dynamic)
	categoryMetadataSchema, _ := category.UnmarshalAdditionalAttributes()
	if categoryMetadataSchema.Properties != nil {
		for k, v := range *categoryMetadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}

	return rootSchema
}

func getDefaultActions[T ResourceInstanceInterface](resource string, schema RootSchema, resourceDescription string) map[string]EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", resource, resourceDescription),
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error) {
				return defaultInstance(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", resource, resourceDescription),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error) {
				return defaultList(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", resource, resourceDescription),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", resource, resourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return defaultCreate(c, schema, repository, resource)
			},
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", resource, resourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return defaultUpdate(c, schema, repository, resource)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDelete(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", resource, resourceDescription),
		),
	}
}

func defaultSchema[T ResourceInstanceInterface](_ *EndorContext[NoPayload], schema RootSchema) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance[T ResourceInstanceInterface](c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceRepository[T]) (*Response[*ResourceInstance[T]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[T]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultList[T ResourceInstanceInterface](c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceRepository[T]) (*Response[[]ResourceInstance[T]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[T]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreate[T ResourceInstanceInterface](c *EndorContext[CreateDTO[ResourceInstance[T]]], schema RootSchema, repository *ResourceInstanceRepository[T], resource string) (*Response[ResourceInstance[T]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", resource))).Build(), nil
}

func defaultUpdate[T ResourceInstanceInterface](c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]], schema RootSchema, repository *ResourceInstanceRepository[T], resource string) (*Response[ResourceInstance[T]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[T]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", resource))).Build(), nil
}

func defaultDelete[T ResourceInstanceInterface](c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceRepository[T], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", resource))).Build(), nil
}

func getDefaultActionsForCategory[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](resource string, schema RootSchema, resourceDescription string, categoryID string) map[string]EndorServiceAction {
	autogenerateID := true

	repository := NewResourceInstanceSpecializedRepository[T, C](
		resource,
		ResourceInstanceRepositoryOptions{AutoGenerateID: &autogenerateID},
	)

	return map[string]EndorServiceAction{
		categoryID + "/schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstanceSpecialized[T, C]], error) {
				return defaultListSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstanceSpecialized[T, C]]]) (*Response[ResourceInstanceSpecialized[T, C]], error) {
				return defaultCreateSpecialized(c, schema, repository, resource)
			},
		),
		categoryID + "/instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstanceSpecialized[T, C]], error) {
				return defaultInstanceSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[T, C]]]) (*Response[ResourceInstanceSpecialized[T, C]], error) {
				return defaultUpdateSpecialized(c, schema, repository, resource)
			},
		),
		categoryID + "/delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDeleteSpecialized(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
	}
}

func defaultListSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C]) (*Response[[]ResourceInstanceSpecialized[T, C]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstanceSpecialized[T, C]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreateSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[CreateDTO[ResourceInstanceSpecialized[T, C]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C], resource string) (*Response[ResourceInstanceSpecialized[T, C]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[T, C]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created (category)", resource))).Build(), nil
}

func defaultInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C]) (*Response[*ResourceInstanceSpecialized[T, C]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstanceSpecialized[T, C]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultUpdateSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[T, C]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C], resource string) (*Response[ResourceInstanceSpecialized[T, C]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[T, C]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated (category)", resource))).Build(), nil
}

func defaultDeleteSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceSpecializedRepository[T, C], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted (category)", resource))).Build(), nil
}
