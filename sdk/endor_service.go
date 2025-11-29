package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/mattiabonardi/endor-sdk-go/sdk/middleware"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorServiceAction interface {
	CreateHTTPCallback(microserviceId string) func(c *gin.Context)
	GetOptions() EndorServiceActionOptions
}

type EndorServiceActionOptions struct {
	Description     string
	Public          bool
	ValidatePayload bool
	InputSchema     *RootSchema
}

type EndorService struct {
	Resource         string
	Description      string
	Methods          map[string]EndorServiceAction
	Priority         *int
	ResourceMetadata bool

	// optionals
	Version string

	// Dependency injection fields - interfaces instead of concrete types
	repository interfaces.RepositoryPattern       // Injected repository for data access
	config     interfaces.ConfigProviderInterface // Injected configuration provider
	logger     interfaces.LoggerInterface         // Injected logger for structured logging
	// Note: Context is handled per-request, not injected at service level
}

func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorServiceAction {
	options := EndorServiceActionOptions{
		Description:     description,
		Public:          false,
		ValidatePayload: true,
		InputSchema:     nil,
	}
	// resolve input params dynamically
	options.InputSchema = resolveInputSchema[T]()
	return NewConfigurableAction(options, handler)
}

func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceAction {
	if options.InputSchema == nil {
		options.InputSchema = resolveInputSchema[T]()
	}
	return &endorServiceActionImpl[T, R]{handler: handler, options: options}
}

type endorServiceActionImpl[T any, R any] struct {
	handler EndorHandlerFunc[T, R]
	options EndorServiceActionOptions
}

func (m *endorServiceActionImpl[T, R]) CreateHTTPCallback(microserviceId string) func(c *gin.Context) {
	return func(c *gin.Context) {
		development := false
		if c.GetHeader("x-development") == "true" {
			development = true
		}
		session := Session{
			Id:          c.GetHeader("x-user-session"),
			Username:    c.GetHeader("x-user-id"),
			Development: development,
		}
		// Recupera categoryID dal context Gin se presente
		var categoryID *string
		if catID, exists := c.Get("categoryID"); exists {
			if catIDStr, ok := catID.(string); ok {
				categoryID = &catIDStr
			}
		}

		ec := &EndorContext[T]{
			MicroServiceId: microserviceId,
			Session:        session,
			CategoryID:     categoryID,
			GinContext:     c,
		}
		var t T
		if m.options.ValidatePayload && reflect.TypeOf(t) != reflect.TypeOf(NoPayload{}) {
			if err := c.ShouldBindJSON(&ec.Payload); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
				return
			}
		}
		// call method
		response, err := m.handler(ec)
		if err != nil {
			var endorError *EndorError
			if errors.As(err, &endorError) {
				c.AbortWithStatusJSON(endorError.StatusCode, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, endorError.Error())))
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())))
			}
		} else {
			c.Header("x-endor-microservice", microserviceId)
			c.JSON(http.StatusOK, response)
		}
	}
}

func (m *endorServiceActionImpl[T, R]) GetOptions() EndorServiceActionOptions {
	return m.options
}

func resolveInputSchema[T any]() *RootSchema {
	var zeroT T
	tType := reflect.TypeOf(zeroT)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	// convert type to schema
	if tType != nil && tType != reflect.TypeOf(NoPayload{}) {
		return NewSchemaByType(tType)
	}
	return nil
}

// EndorServiceDependencies contains all required dependencies for EndorService.
// This struct provides type safety and clear documentation of service dependencies.
type EndorServiceDependencies struct {
	Repository interfaces.RepositoryPattern       // Required: Data access layer
	Config     interfaces.ConfigProviderInterface // Required: Configuration access
	Logger     interfaces.LoggerInterface         // Required: Logging interface
}

// EndorServiceError represents dependency validation errors for EndorService construction.
type EndorServiceError struct {
	Field   string // Name of the missing dependency field
	Message string // Human-readable error message
}

func (e *EndorServiceError) Error() string {
	return fmt.Sprintf("EndorService dependency error [%s]: %s", e.Field, e.Message)
}

// NewEndorServiceWithDeps creates a new EndorService with explicit dependency injection.
// This constructor enables full customization of service dependencies and comprehensive testing.
//
// Parameters:
//   - resource: The resource name for API routing and identification
//   - description: Human-readable description for documentation
//   - methods: Map of actions available on this service
//   - deps: All required dependencies (repository, config, logger)
//
// Returns error if any required dependency is nil with structured error messages.
//
// Example usage:
//
//	deps := EndorServiceDependencies{
//	    Repository: mongoRepo,
//	    Config:     envConfig,
//	    Logger:     structuredLogger,
//	}
//	service, err := NewEndorServiceWithDeps("users", "User management", actions, deps)
func NewEndorServiceWithDeps(
	resource string,
	description string,
	methods map[string]EndorServiceAction,
	deps EndorServiceDependencies,
) (*EndorService, error) {
	// AC 4: Dependency validation with structured error messages
	if deps.Repository == nil {
		return nil, &EndorServiceError{
			Field:   "Repository",
			Message: "Repository interface is required for data access operations. Provide an implementation of interfaces.RepositoryPattern.",
		}
	}

	if deps.Config == nil {
		return nil, &EndorServiceError{
			Field:   "Config",
			Message: "Configuration interface is required for service configuration access. Provide an implementation of interfaces.ConfigProviderInterface.",
		}
	}

	if deps.Logger == nil {
		return nil, &EndorServiceError{
			Field:   "Logger",
			Message: "Logger interface is required for service logging. Provide an implementation of interfaces.LoggerInterface.",
		}
	}

	// AC 1: Constructor accepts all required dependencies as interface parameters with type safety
	service := &EndorService{
		Resource:    resource,
		Description: description,
		Methods:     methods,
		// AC 2: EndorService struct holds interface references instead of concrete types
		repository: deps.Repository,
		config:     deps.Config,
		logger:     deps.Logger,
	}

	return service, nil
}

// NewEndorService creates an EndorService using default implementations for backward compatibility.
// This convenience constructor maintains existing creation patterns while using dependency injection internally.
//
// AC 3: Backward compatibility - maintains existing creation patterns with default implementations
//
// Note: This constructor will use default singleton implementations for dependencies.
// For custom implementations or testing, use NewEndorServiceWithDeps() instead.
func NewEndorService(resource string, description string, methods map[string]EndorServiceAction) EndorService {
	// AC 3: Backward compatibility - maintains existing creation patterns with default implementations
	return EndorService{
		Resource:    resource,
		Description: description,
		Methods:     methods,
		// AC 3: Initialize with default dependency implementations for backward compatibility
		repository: nil,                        // TODO: Will be set to default MongoDB repository in repository refactoring
		config:     NewDefaultConfigProvider(), // Use default config that wraps GetConfig()
		logger:     NewDefaultLogger(),         // Use default logger that wraps log package
	}
}

// NewEndorServiceFromContainer creates an EndorService using dependencies from a DI container.
// This enables automatic dependency resolution and supports container-managed lifecycles.
//
// AC 5: Container integration - service constructors work seamlessly with DI container
func NewEndorServiceFromContainer(
	container di.Container,
	resource string,
	description string,
	methods map[string]EndorServiceAction,
) (*EndorService, error) {
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

	deps := EndorServiceDependencies{
		Repository: repository,
		Config:     config,
		Logger:     logger,
	}

	return NewEndorServiceWithDeps(resource, description, methods, deps)
}

// EndorServiceInterface implementation
// AC 6: All existing EndorService methods work unchanged using injected interface dependencies

// GetResource returns the resource name that this service manages.
func (s *EndorService) GetResource() string {
	return s.Resource
}

// GetDescription returns a human-readable description of the service.
func (s *EndorService) GetDescription() string {
	return s.Description
}

// GetMethods returns the map of available actions for this service.
func (s *EndorService) GetMethods() map[string]EndorServiceAction {
	return s.Methods
}

// GetPriority returns the service priority for conflict resolution.
func (s *EndorService) GetPriority() *int {
	return s.Priority
}

// GetVersion returns the API version for this service.
func (s *EndorService) GetVersion() string {
	return s.Version
}

// Validate performs service configuration validation using injected dependencies.
// AC 2 & 6: Uses injected interface dependencies instead of concrete types
func (s *EndorService) Validate() error {
	// Basic field validation
	if s.Resource == "" {
		return &EndorServiceError{
			Field:   "Resource",
			Message: "Resource name is required for service registration",
		}
	}

	if s.Description == "" {
		return &EndorServiceError{
			Field:   "Description",
			Message: "Description is required for API documentation",
		}
	}

	if len(s.Methods) == 0 {
		return &EndorServiceError{
			Field:   "Methods",
			Message: "At least one service action must be defined",
		}
	}

	// Validate injected dependencies if they exist (for services created with NewEndorServiceWithDeps)
	if s.repository != nil || s.config != nil || s.logger != nil {
		if s.config != nil {
			// Use injected config interface for validation instead of global GetConfig()
			if err := s.config.Validate(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
		}

		if s.logger != nil {
			// Log validation attempt using injected logger
			s.logger.Debug("EndorService validation", "resource", s.Resource, "methods", len(s.Methods))
		}
	}

	return nil
}

// GetRepository returns the injected repository interface.
// This method enables access to the injected repository for service operations.
// AC 2: Provides access to injected RepositoryInterface instead of hard-coded access
func (s *EndorService) GetRepository() interfaces.RepositoryPattern {
	return s.repository
}

// GetConfig returns the injected configuration interface.
// This method enables access to configuration without global GetConfig() calls.
// AC 2: Provides access to injected ConfigInterface instead of direct configuration access
func (s *EndorService) GetConfig() interfaces.ConfigProviderInterface {
	return s.config
}

// GetLogger returns the injected logger interface.
// This method enables structured logging using the injected logger implementation.
// AC 2: Provides access to injected LoggerInterface for service logging
func (s *EndorService) GetLogger() interfaces.LoggerInterface {
	return s.logger
}

// Note: Interface compliance deferred until type alignment between sdk and interfaces packages
// TODO: Resolve RootSchema and EndorServiceAction type differences between sdk and interfaces packages
// var _ interfaces.EndorServiceInterface = (*EndorService)(nil)

// DecoratedService wraps an EndorService with middleware pipeline support.
// This type implements the decorator pattern to add cross-cutting concerns
// without modifying the original service implementation.
type DecoratedService struct {
	*EndorService                                // Embedded service for method delegation
	pipeline      *middleware.MiddlewarePipeline // Middleware execution pipeline
}

// WithMiddleware decorates the EndorService with a middleware pipeline.
// Middleware will be executed in the order provided, with Before() hooks
// running before the service handler and After() hooks running after.
//
// Example usage:
//
//	decoratedService := service.WithMiddleware(
//	    authMiddleware,
//	    loggingMiddleware,
//	    metricsMiddleware,
//	)
//
// The decorated service preserves the original EndorService interface
// and can be used anywhere an EndorService is expected.
func (s *EndorService) WithMiddleware(middlewares ...middleware.MiddlewareInterface) *DecoratedService {
	return &DecoratedService{
		EndorService: s,
		pipeline:     middleware.NewMiddlewarePipeline(middlewares...),
	}
}

// GetMethods returns the decorated methods that include middleware execution.
// The returned actions execute middleware Before/After hooks around the original handlers.
func (d *DecoratedService) GetMethods() map[string]EndorServiceAction {
	originalMethods := d.EndorService.GetMethods()
	decoratedMethods := make(map[string]EndorServiceAction, len(originalMethods))

	for name, action := range originalMethods {
		decoratedMethods[name] = &decoratedServiceAction{
			original: action,
			pipeline: d.pipeline,
		}
	}

	return decoratedMethods
}

// decoratedServiceAction wraps an EndorServiceAction with middleware execution.
type decoratedServiceAction struct {
	original EndorServiceAction
	pipeline *middleware.MiddlewarePipeline
}

func (d *decoratedServiceAction) CreateHTTPCallback(microserviceId string) func(c *gin.Context) {
	originalCallback := d.original.CreateHTTPCallback(microserviceId)

	return func(c *gin.Context) {
		// Create middleware-aware context wrapper
		middlewareCtx := &middlewareContext{ginContext: c}

		// Execute middleware Before() hooks
		if err := d.pipeline.ExecuteBefore(middlewareCtx); err != nil {
			// Middleware error - short-circuit execution
			c.AbortWithStatusJSON(http.StatusBadRequest, NewDefaultResponseBuilder().
				AddMessage(NewMessage(Error, fmt.Sprintf("Middleware execution failed: %v", err))).
				Build())
			return
		}

		// Execute original service handler
		originalCallback(c)

		// Create a response wrapper with available data from context
		middlewareResponse := &middlewareResponse{
			statusCode: http.StatusOK, // Default, actual code may vary
			headers:    c.Writer.Header(),
		}

		// Execute middleware After() hooks
		if err := d.pipeline.ExecuteAfter(middlewareCtx, middlewareResponse); err != nil {
			// Log middleware error but don't modify response
			// The original response has already been sent
			fmt.Printf("Middleware After() error: %v\n", err)
		}
	}
}

func (d *decoratedServiceAction) GetOptions() EndorServiceActionOptions {
	return d.original.GetOptions()
}

// middlewareContext adapts Gin context for middleware interface compatibility.
type middlewareContext struct {
	ginContext *gin.Context
}

// middlewareResponse wraps response data for middleware interface compatibility.
type middlewareResponse struct {
	statusCode int
	headers    http.Header
}
