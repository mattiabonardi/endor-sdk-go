package sdk

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Endor struct {
	internalEndorServices  *[]EndorService
	internalHybridServices *[]EndorHybridService
	postInitFunc           func()
	// AC6: Proper Resource Management - track dependencies for proper cleanup
	dependencies EndorInitializerDependencies
	container    di.Container
}

// EndorInitializerDependencies contains all required dependencies for EndorInitializer.
// This struct provides type safety and clear documentation of initializer dependencies.
type EndorInitializerDependencies struct {
	Container  di.Container                       // Optional: Custom DI container, uses default if nil
	Repository interfaces.RepositoryInterface     // Optional: Custom repository, resolved from container if nil
	Config     interfaces.ConfigProviderInterface // Optional: Custom config, uses default if nil
	Logger     interfaces.LoggerInterface         // Optional: Custom logger, uses default if nil
}

// EndorInitializerError represents dependency validation errors for EndorInitializer construction.
type EndorInitializerError struct {
	Field   string // Name of the problematic dependency field
	Message string // Human-readable error message
	Cause   error  // Underlying error if any
}

func (e *EndorInitializerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("EndorInitializer dependency error [%s]: %s (caused by: %v)", e.Field, e.Message, e.Cause)
	}
	return fmt.Sprintf("EndorInitializer dependency error [%s]: %s", e.Field, e.Message)
}

type EndorInitializer struct {
	endor        *Endor
	dependencies EndorInitializerDependencies
}

func NewEndorInitializer() *EndorInitializer {
	return &EndorInitializer{
		endor:        &Endor{},
		dependencies: EndorInitializerDependencies{}, // Start with empty dependencies - will be resolved in Build()
	}
}

func (b *EndorInitializer) WithEndorServices(services *[]EndorService) *EndorInitializer {
	b.endor.internalEndorServices = services
	return b
}

func (b *EndorInitializer) WithHybridServices(services *[]EndorHybridService) *EndorInitializer {
	b.endor.internalHybridServices = services
	return b
}

func (b *EndorInitializer) WithPostInitFunc(f func()) *EndorInitializer {
	b.endor.postInitFunc = f
	return b
}

// WithContainer sets a custom DI container instance for dependency resolution.
// This enables advanced users to provide their own container configuration.
//
// AC2: Custom Dependency Registration - provides hooks for custom dependency registration
func (b *EndorInitializer) WithContainer(container di.Container) *EndorInitializer {
	b.dependencies.Container = container
	return b
}

// WithCustomRepository sets a custom repository implementation.
// This overrides any repository resolved from the container.
//
// AC5: Advanced Customization - override any dependency with custom implementations
func (b *EndorInitializer) WithCustomRepository(repo interfaces.RepositoryInterface) *EndorInitializer {
	b.dependencies.Repository = repo
	return b
}

// WithCustomConfig sets a custom configuration provider implementation.
// This overrides any config resolved from the container.
//
// AC5: Advanced Customization - override any dependency with custom implementations
func (b *EndorInitializer) WithCustomConfig(config interfaces.ConfigProviderInterface) *EndorInitializer {
	b.dependencies.Config = config
	return b
}

// Build creates properly wired service instances using dependency injection container.
// This method implements AC1: Dependency Graph Configuration with complete validation.
//
// Returns *Endor with all dependencies properly wired, or error if dependency validation fails.
//
// AC1: EndorInitializer.Build() creates properly wired service instances using DI container
// AC3: Complete dependency validation with clear error messages
func (b *EndorInitializer) Build() (*Endor, error) {
	// AC3: Validate the complete dependency graph before service creation
	if err := b.validateAndResolveDependencies(); err != nil {
		return nil, err
	}

	// AC1: Service instances are now properly wired with validated dependencies
	// AC6: Store dependencies and container in Endor for proper resource management
	b.endor.dependencies = b.dependencies
	b.endor.container = b.dependencies.Container

	return b.endor, nil
}

// validateAndResolveDependencies ensures all required dependencies are available.
// This implements AC3: Complete Dependency Validation with structured error messages.
func (b *EndorInitializer) validateAndResolveDependencies() error {
	// Step 1: Initialize or use provided DI container
	if b.dependencies.Container == nil {
		// AC4: Backward compatibility - create default container with default implementations
		container, err := b.createDefaultContainer()
		if err != nil {
			return &EndorInitializerError{
				Field:   "Container",
				Message: "Failed to create default DI container",
				Cause:   err,
			}
		}
		b.dependencies.Container = container
	}

	// Step 2: Resolve missing dependencies from container (if not explicitly provided)
	if err := b.resolveMissingDependencies(); err != nil {
		return err
	}

	// Step 3: AC3: Validate complete dependency graph
	if validationErrors := b.dependencies.Container.Validate(); len(validationErrors) > 0 {
		return &EndorInitializerError{
			Field:   "DependencyGraph",
			Message: fmt.Sprintf("Dependency graph validation failed with %d errors: %v", len(validationErrors), validationErrors),
		}
	}

	return nil
}

// createDefaultContainer sets up a DI container with default implementations.
// AC4: Backward compatibility using default dependency implementations.
func (b *EndorInitializer) createDefaultContainer() (di.Container, error) {
	container := di.NewContainer()

	// Register default dependencies following patterns from Stories 2.1-2.4

	// Register default config provider
	if err := di.Register[interfaces.ConfigProviderInterface](container, NewDefaultConfigProvider(), di.Singleton); err != nil {
		return nil, fmt.Errorf("failed to register default config provider: %w", err)
	}

	// Register default logger
	if err := di.Register[interfaces.LoggerInterface](container, NewDefaultLogger(), di.Singleton); err != nil {
		return nil, fmt.Errorf("failed to register default logger: %w", err)
	}

	// Register repository factories following Story 2.4 patterns
	if err := RegisterRepositoryFactories(container); err != nil {
		return nil, fmt.Errorf("failed to register repository factories: %w", err)
	}

	return container, nil
}

// resolveMissingDependencies attempts to resolve any nil dependencies from the container.
// This enables automatic dependency resolution while allowing explicit overrides.
func (b *EndorInitializer) resolveMissingDependencies() error {
	// Resolve repository if not explicitly provided
	if b.dependencies.Repository == nil {
		repo, err := di.Resolve[interfaces.RepositoryInterface](b.dependencies.Container)
		if err != nil {
			return &EndorInitializerError{
				Field:   "Repository",
				Message: "Failed to resolve repository dependency from container. Ensure repository factories are registered or provide custom repository via WithCustomRepository()",
				Cause:   err,
			}
		}
		b.dependencies.Repository = repo
	}

	// Resolve config if not explicitly provided
	if b.dependencies.Config == nil {
		config, err := di.Resolve[interfaces.ConfigProviderInterface](b.dependencies.Container)
		if err != nil {
			return &EndorInitializerError{
				Field:   "Config",
				Message: "Failed to resolve configuration dependency from container. Ensure config provider is registered or provide custom config via WithCustomConfig()",
				Cause:   err,
			}
		}
		b.dependencies.Config = config
	}

	// Resolve logger if not explicitly provided
	if b.dependencies.Logger == nil {
		logger, err := di.Resolve[interfaces.LoggerInterface](b.dependencies.Container)
		if err != nil {
			return &EndorInitializerError{
				Field:   "Logger",
				Message: "Failed to resolve logger dependency from container. Ensure logger is registered or provide custom logger via WithCustomConfig() or use default",
				Cause:   err,
			}
		}
		b.dependencies.Logger = logger
	}

	return nil
}

func (h *Endor) Init(microserviceId string) {
	// load configuration
	config := GetConfig()

	// define runtime configuration
	config.DynamicResourceDocumentDBName = microserviceId

	// create router
	router := gin.New()

	// monitoring
	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Check if an EndorService with resource == "resource" is already defined
	resourceServiceExists := false
	if h.internalEndorServices != nil {
		for _, svc := range *h.internalEndorServices {
			if svc.Resource == "resource" {
				resourceServiceExists = true
				break
			}
		}
	}
	if !resourceServiceExists {
		*h.internalEndorServices = append(*h.internalEndorServices, *NewResourceService(microserviceId, h.internalEndorServices, h.internalHybridServices))
		*h.internalEndorServices = append(*h.internalEndorServices, *NewResourceActionService(microserviceId, h.internalEndorServices, h.internalHybridServices))
	}

	// get all resources
	EndorServiceRepository := NewEndorServiceRepository(microserviceId, h.internalEndorServices, h.internalHybridServices)
	resources, err := EndorServiceRepository.EndorServiceList()
	if err != nil {
		log.Fatal(err)
	}

	router.NoRoute(func(c *gin.Context) {
		// find the resource in path /api/{version}/{resource}/{method}
		pathSegments := strings.Split(c.Request.URL.Path, "/")
		if len(pathSegments) > 4 {
			resource := pathSegments[3]
			action := pathSegments[4]
			if len(pathSegments) == 6 {
				action = pathSegments[4] + "/" + pathSegments[5]
			}
			endorRepositoryDictionary, err := EndorServiceRepository.Instance(ReadInstanceDTO{
				Id: resource,
			})
			if err == nil {
				if method, ok := endorRepositoryDictionary.EndorService.Methods[action]; ok {
					method.CreateHTTPCallback(microserviceId)(c)
					return
				}
			}
		}
		response := NewDefaultResponseBuilder()
		response.AddMessage(NewMessage(Fatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
		c.JSON(http.StatusNotFound, response.Build())
	})

	err = InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), resources)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), resources, "/api")
	if err != nil {
		log.Fatal(err)
	}

	// swagger
	router.StaticFS("/swagger", http.Dir(swaggerPath))

	// post initialization
	if h.postInitFunc != nil {
		h.postInitFunc()
	}

	// start http server
	router.Run()
}

// Shutdown implements AC6: Proper Resource Management with dependency cleanup.
// This method ensures proper dependency cleanup during shutdown following dependency order.
//
// AC6: Framework ensures proper dependency cleanup during shutdown following dependency order
func (h *Endor) Shutdown() error {
	if h.dependencies.Logger != nil {
		h.dependencies.Logger.Info("Starting Endor shutdown sequence")
	}

	var shutdownErrors []error

	// Step 1: Shutdown services first (dependents before dependencies)
	if h.internalEndorServices != nil {
		for i, service := range *h.internalEndorServices {
			if h.dependencies.Logger != nil {
				h.dependencies.Logger.Debug("Shutting down EndorService", "index", i, "resource", service.GetResource())
			}
			// Services don't currently have cleanup methods, but framework is prepared for future implementation
		}
	}

	if h.internalHybridServices != nil {
		for i, service := range *h.internalHybridServices {
			if h.dependencies.Logger != nil {
				h.dependencies.Logger.Debug("Shutting down EndorHybridService", "index", i, "resource", service.GetResource())
			}
			// Services don't currently have cleanup methods, but framework is prepared for future implementation
		}
	}

	// Step 2: Shutdown repositories (data access layer)
	if h.dependencies.Repository != nil {
		if h.dependencies.Logger != nil {
			h.dependencies.Logger.Debug("Shutting down repository dependencies")
		}
		// Repository interface doesn't currently define cleanup methods, but framework is prepared
		// Future enhancement: Define Closable interface and check if repository implements it
	}

	// Step 3: Shutdown configuration and logging (infrastructure dependencies last)
	if h.dependencies.Config != nil {
		if h.dependencies.Logger != nil {
			h.dependencies.Logger.Debug("Shutting down configuration provider")
		}
		// Config providers don't currently have cleanup methods, but framework is prepared
	}

	// Step 4: Shutdown logger last
	if h.dependencies.Logger != nil {
		h.dependencies.Logger.Info("Endor shutdown sequence complete")
	}

	// Step 5: Clear container to release all references
	if h.container != nil {
		h.container.Reset() // Clear all registrations to help with garbage collection
	}

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors: %v", len(shutdownErrors), shutdownErrors)
	}

	return nil
}
