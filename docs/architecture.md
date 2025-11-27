# endor-sdk-go - Refactored Architecture Document

**Author:** BMad
**Date:** November 27, 2025
**Project:** Dependency Injection & Service Composition Refactoring
**Skill Level:** Expert

---

## Executive Summary

This document defines the architectural decisions for refactoring endor-sdk-go into an interface-driven, dependency-injectable framework that enables testability and service composition while preserving the innovative dual-service pattern (EndorService + EndorHybridService).

**Key Architectural Goals:**
- **Testability**: All components mockable through interfaces
- **Composition**: Services can embed other services via dependency injection
- **Flexibility**: Custom implementations injectable for any framework component
- **Performance**: Zero overhead abstractions with compile-time safety

---

## Project Context Understanding

Based on my review of your PRD and epics, I understand the endor-sdk-go project as:

**Core Challenge:** The current framework has tightly coupled dependencies (hard-coded MongoDB access, global singletons) that prevent effective unit testing and service composition patterns.

**Key Aspects:**
- **Dual Service Architecture**: Static EndorService (manual control) + EndorHybridService (automatic CRUD)
- **Type Safety**: Extensive use of Go generics for compile-time validation
- **Dynamic Schema**: Automatic JSON schema generation from Go structs
- **MongoDB Integration**: Automatic persistence with category-based specialization

**Critical NFRs:**
- Maintain current performance characteristics
- Preserve automatic schema generation
- Support both testability and production deployment
- Enable clean service composition patterns

**Epic Structure:** 4 epics with 20 stories covering interface extraction, dependency injection, service composition, and developer experience.

This matches the framework refactoring project focused on enabling testability and composition while preserving the powerful automation features that make endor-sdk-go unique.

---

## Architectural Decisions

Given your expert skill level and Go framework requirements, here are the critical architectural decisions:

### Decision 1: Dependency Injection Pattern

**Choice: Lightweight Custom DI Container**

**Options Considered:**
- Heavy framework (wire/dig) - Too complex, compile-time overhead
- Manual DI (no container) - Too much boilerplate for framework users
- **Custom lightweight container** - Perfect balance for Go framework

**Rationale:** Go frameworks benefit from simple, explicit DI without magic. Custom container gives us:
- Interface-based registration: `container.Register[RepositoryInterface](mongoRepo)`
- Type-safe resolution: `container.Resolve[RepositoryInterface]()`
- Lifecycle management: Singleton (default), Transient, Scoped
- Zero runtime reflection overhead
- Circular dependency detection

**Implementation Pattern:**
```go
type Container interface {
    Register[T any](impl T, scope Scope)
    Resolve[T any]() (T, error)
    Validate() error
}
```

### Decision 2: Interface Granularity

**Choice: Interface Segregation with Smart Composition**

**Options Considered:**
- Monolithic interfaces - Harder to mock, violates ISP
- **Granular interfaces** - Easy mocking, compose as needed
- Per-method interfaces - Too fragmented

**Rationale:** Following Go's interface philosophy of "accept interfaces, return structs":
- `RepositoryInterface` for data access
- `ConfigProviderInterface` for configuration
- `EndorServiceInterface` and `EndorHybridServiceInterface` for services
- Compose larger interfaces from smaller ones when needed

**Interface Strategy:**
```go
// Small, focused interfaces
type RepositoryInterface interface {
    Create(ctx context.Context, resource Resource) error
    Read(ctx context.Context, id string) (Resource, error)
    Update(ctx context.Context, resource Resource) error
    Delete(ctx context.Context, id string) error
}

// Composed when needed
type FullServiceInterface interface {
    RepositoryInterface
    ConfigProviderInterface
}
```

### Decision 3: Service Composition Pattern

**Choice: Decorator Pattern with Middleware Pipeline**

**Options Considered:**
- Direct embedding - Tight coupling
- **Decorator with middleware** - Clean, extensible
- Proxy pattern - Too much indirection for Go

**Rationale:** Go's composition over inheritance maps perfectly to decorator pattern:
- `EndorHybridService` can wrap `EndorService` instances
- Middleware pipeline for cross-cutting concerns
- Method delegation with clear precedence rules
- Type-safe composition

**Composition Strategy:**
```go
type EndorHybridService interface {
    EmbedService(prefix string, service EndorServiceInterface)
    WithMiddleware(middleware ...MiddlewareInterface) EndorHybridService
    ToEndorService(schema Schema) EndorServiceInterface
}
```

**Impact:** This pattern enables the powerful service composition capabilities required by FR17-FR19 while maintaining clean separation of concerns and testability.

### Decision 4: Error Handling and Logging Strategy

**Context:** With dependency injection and service composition, we need consistent error handling and logging patterns that work across service boundaries and maintain observability.

**Decision:** Implement context-aware error handling with structured logging through dependency injection.

**Alternatives Considered:**
1. Global error handling with singleton logger
2. Error interfaces with embedded context
3. Context-aware errors with injected logger (CHOSEN)

**Rationale:** 
- Service composition requires error context to flow through service hierarchies
- Dependency injection enables different logging strategies per environment
- Structured logging with context enables better observability and debugging
- Go's error wrapping patterns work well with service composition

**Implementation Pattern:**
```go
// Error handling interface
type ErrorHandler interface {
    HandleError(ctx context.Context, err error) error
    WrapError(err error, context string) error
}

// Logger interface for dependency injection
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, err error, fields ...Field)
}

// Service with injected error handling
type EndorService struct {
    errorHandler ErrorHandler
    logger       Logger
}
```

**Impact:** This approach ensures consistent error handling across service compositions while maintaining testability through interface-based logging.

### Decision 5: Testing Strategy and Test Organization

**Context:** The new architecture must support comprehensive testing at unit, integration, and composition levels while maintaining fast test execution.

**Decision:** Implement layered testing strategy with clear separation between unit tests (mocked dependencies) and integration tests (real dependencies).

**Alternatives Considered:**
1. Mixed testing approach with conditional mocking
2. Pure integration testing with test databases
3. Layered testing with build tags (CHOSEN)

**Rationale:**
- Unit tests must run fast without external dependencies (FR5)
- Integration tests validate real database interactions
- Service composition testing requires both levels
- Build tags enable selective test execution in CI/CD

**Implementation Pattern:**
```go
// Unit tests (fast, mocked dependencies)
//go:build unit

func TestEndorService_GetResource_Unit(t *testing.T) {
    mockRepo := testutils.NewMockRepository()
    mockRepo.On("FindByID", "123").Return(expectedResource, nil)
    
    service := NewEndorServiceWithDeps(mockRepo, mockConfig, mockLogger)
    result, err := service.GetResource("123")
    
    assert.NoError(t, err)
    assert.Equal(t, expectedResource, result)
}

// Integration tests (slower, real dependencies)
//go:build integration

func TestEndorService_GetResource_Integration(t *testing.T) {
    testDB := testutils.SetupTestDatabase(t)
    defer testDB.Cleanup()
    
    service := NewEndorServiceWithDeps(testDB.Repository(), testConfig, testLogger)
    // Test with real database...
}
```

**Impact:** This strategy enables fast unit test execution while ensuring integration validation, supporting the comprehensive testing requirements from FR5-FR8.

### Decision 6: Project Structure and Package Organization

**Context:** The new architecture with interfaces, implementations, and testing utilities requires clear project organization that supports both framework development and user consumption.

**Decision:** Implement domain-driven package structure with clear separation between interfaces, implementations, and utilities.

**Alternatives Considered:**
1. Flat package structure with all code in `sdk/`
2. Layer-based structure (interfaces/, implementations/, utils/)
3. Domain-driven structure with feature packages (CHOSEN)

**Rationale:**
- Domain-driven structure aligns with Go conventions
- Clear separation enables selective imports
- Testing utilities are separate from core framework
- Extension points are clearly defined and discoverable

**Implementation Pattern:**
```
sdk/
├── interfaces/           # All framework interfaces
│   ├── service.go       # EndorServiceInterface
│   ├── repository.go    # Repository interfaces
│   └── config.go        # Configuration interfaces
├── core/                # Core implementations
│   ├── endor_service.go
│   ├── endor_hybrid_service.go
│   └── di_container.go
├── repository/          # Repository implementations
│   ├── mongo_repository.go
│   └── memory_repository.go
├── middleware/          # Middleware implementations
│   ├── logging.go
│   ├── metrics.go
│   └── auth.go
├── testutils/          # Testing utilities and mocks
│   ├── mocks.go
│   ├── builders.go
│   └── fixtures.go
└── examples/           # Usage examples
    ├── simple_service/
    ├── composed_service/
    └── testing_patterns/
```

**Impact:** This structure provides clear boundaries between interfaces and implementations, making the framework easy to understand and extend while supporting clean dependency injection patterns.

### Decision 7: Implementation Consistency and Agent Coordination

**Context:** With 20 stories across 4 epics, multiple AI agents will implement different parts of the framework. We need consistency rules to ensure cohesive implementation.

**Decision:** Implement strict naming conventions, interface patterns, and code organization rules that all implementing agents must follow.

**Alternatives Considered:**
1. Loose guidelines with post-implementation cleanup
2. Code generation with templates
3. Strict conventions with validation tools (CHOSEN)

**Rationale:**
- Multiple agents need clear rules to maintain consistency
- Go has strong conventions that should be followed
- Interface-driven design requires consistent patterns
- Testing utilities must work seamlessly across all implementations

**Implementation Standards:**
```go
// Naming conventions
type ServiceInterface interface {}     // Interface suffix
type ServiceImpl struct {}            // Impl suffix for implementations
func NewServiceWithDeps() ServiceInterface {} // Constructor naming

// Error handling pattern
func (s *ServiceImpl) Method() (Result, error) {
    if err := s.validate(); err != nil {
        return nil, s.errorHandler.WrapError(err, "ServiceImpl.Method")
    }
    // implementation...
}

// Dependency injection pattern
type ServiceImpl struct {
    repo   RepositoryInterface   // Interface dependencies
    config ConfigInterface
    logger LoggerInterface
}

// Testing pattern
func TestService_Method(t *testing.T) {
    // Given
    mockDep := testutils.NewMockRepository()
    service := NewServiceWithDeps(mockDep, testConfig, testLogger)
    
    // When
    result, err := service.Method()
    
    // Then
    assert.NoError(t, err)
}
```

**Validation Rules:**
- All public interfaces must have comprehensive GoDoc
- All implementations must have corresponding unit tests
- Error handling must use context-aware patterns
- Dependency injection must use interface parameters only
- Test utilities must support both behavior verification and state testing

**Impact:** These standards ensure that all 20 implementation stories produce cohesive, maintainable code that works together seamlessly.

---

## Cross-Cutting Concerns

### Performance Requirements

**Memory Management:**
- Dependency container uses interface references (pointer-sized)
- Service composition through delegation (no object copying)
- Repository pooling for database connections
- Lazy initialization of expensive dependencies

**Latency Targets:**
- Dependency resolution: < 1μs per service
- Service composition: < 10μs additional overhead
- Middleware pipeline: < 5μs per middleware
- Total request overhead: < 50μs compared to current implementation

### Security Architecture

**Dependency Injection Security:**
- Interface validation prevents injection of malicious implementations
- Singleton dependencies are thread-safe by design
- Service isolation through interface boundaries
- Audit logging for dependency resolution in production

**Service Composition Security:**
- Embedded services inherit parent authentication context
- Middleware pipeline enforces security policies consistently
- Service boundaries prevent privilege escalation
- Context propagation maintains security tokens

### Monitoring and Observability

**Dependency Injection Metrics:**
- Container resolution time per interface
- Dependency graph health status
- Service composition depth metrics
- Memory usage tracking for injected dependencies

**Service Composition Tracing:**
- Request flow through composed services
- Middleware execution timing
- Embedded service call tracking
- Error propagation through service hierarchies

---

## Technical Implementation Mapping

### Epic 1: Interface Foundation → Technical Components

**Story 1.1 (Extract Core Service Interfaces):**
- Package: `sdk/interfaces/service.go`
- Key Interfaces: `EndorServiceInterface`, `EndorHybridServiceInterface`
- Implementation: Extract existing public methods, add GoDoc, create compliance tests

**Story 1.2 (Extract Repository Interfaces):**
- Package: `sdk/interfaces/repository.go`
- Key Interfaces: `RepositoryInterface[T]`, `ResourceRepositoryInterface[T]`
- Implementation: Abstract MongoDB operations, use generics for type safety

**Story 1.3 (Extract Configuration Interfaces):**
- Package: `sdk/interfaces/config.go`
- Key Interfaces: `ConfigProviderInterface`, `EndorContextInterface`
- Implementation: Abstract environment loading, context management

**Story 1.4 (Create Test Utility Package):**
- Package: `sdk/testutils/mocks.go`
- Components: Mock implementations using testify/mock
- Implementation: Behavior-configurable mocks, test data builders

**Story 1.5 (Refactor Existing Tests):**
- Files: `test/*_test.go`
- Changes: Replace concrete dependencies with interface mocks
- Implementation: Build tags for unit vs integration tests

### Epic 2: Dependency Injection → Technical Architecture

**Story 2.1 (DI Container Implementation):**
- Package: `sdk/core/di_container.go`
- Components: `Container` interface, registration/resolution logic
- Implementation: Type-safe generics, lifecycle management, validation

**Story 2.2 (EndorService Constructor Injection):**
- File: `sdk/core/endor_service.go`
- Changes: Add `NewEndorServiceWithDeps()` constructor
- Implementation: Interface parameters, backward-compatible factory

**Story 2.3 (EndorHybridService Constructor Injection):**
- File: `sdk/core/endor_hybrid_service.go`
- Changes: Add dependency injection to `ToEndorService()`
- Implementation: Pass injected dependencies to generated services

**Story 2.4 (Repository Layer Refactoring):**
- Package: `sdk/repository/`
- Changes: Constructor injection for database clients
- Implementation: Remove GetMongoClient() calls, inject client interfaces

**Story 2.5 (Framework Initializer Updates):**
- File: `sdk/endor_initializer.go`
- Changes: Wire dependencies automatically in `Build()`
- Implementation: Container configuration, dependency graph validation

### Epic 3: Service Composition → Architecture Patterns

**Story 3.1 (Middleware Pipeline):**
- Package: `sdk/middleware/`
- Components: `MiddlewareInterface`, built-in middleware (auth, logging, metrics)
- Implementation: Chain-of-responsibility pattern, context propagation

**Story 3.2 (Service Embedding):**
- Enhancement: `EndorHybridService.EmbedService()`
- Components: Method delegation, namespace resolution
- Implementation: Reflection-based method forwarding, precedence rules

**Story 3.3 (Shared Dependency Management):**
- Enhancement: DI container scoping (Singleton, Transient, Scoped)
- Components: Lifecycle management, concurrent access safety
- Implementation: Thread-safe dependency sharing, health monitoring

**Story 3.4 (Composition Utilities):**
- Package: `sdk/composition/`
- Components: `ServiceChain`, `ServiceProxy`, `ServiceBranch`, `ServiceMerger`
- Implementation: Type-safe composition patterns, error propagation

**Story 3.5 (Service Lifecycle Management):**
- Enhancement: `ServiceLifecycleInterface`
- Components: Start/stop ordering, health aggregation
- Implementation: Dependency-aware lifecycle, graceful degradation

### Epic 4: Developer Experience → Tooling

**Story 4.1 (Comprehensive Documentation):**
- Files: `docs/developer-guide.md`, API documentation generation
- Content: Before/after examples, testing patterns, tutorials
- Implementation: Runnable examples, CI/CD validation

**Story 4.2 (Enhanced Error Messages):**
- Enhancement: Structured error types, debug tracing
- Components: `DependencyError`, `CompositionError`, debugging tools
- Implementation: Context-aware errors, actionable error messages

**Story 4.3 (Testing Framework):**
- Enhancement: `sdk/testutils/` advanced patterns
- Components: Test builders, composition testing, performance utilities
- Implementation: Realistic test scenarios, benchmark utilities

**Story 4.4 (Development Tools):**
- Tool: `endor-cli` with service generation, validation
- Components: Code generators, dependency validators, profiling
- Implementation: CLI tool with multiple subcommands, IDE integration

**Story 4.5 (Performance Validation):**
- Enhancement: Comprehensive benchmark suite
- Components: Performance tests, regression detection, optimization guide
- Implementation: Continuous benchmarking, performance dashboard

---

## Integration Points and Dependencies

### Story Dependencies (Implementation Order)

**Phase 1: Foundation (Epic 1)**
1.1 → 1.2 → 1.3 → 1.4 → 1.5

**Phase 2: Core Architecture (Epic 2)**  
2.1 → 2.2 → 2.3 → 2.4 → 2.5

**Phase 3: Advanced Features (Epic 3)**
3.1 → 3.2 → 3.3 → 3.4 → 3.5

**Phase 4: Developer Experience (Epic 4)**
4.1 → 4.2 → 4.3 → 4.4 → 4.5

### Critical Integration Points

**Interface Compatibility:** Stories 1.1-1.3 must produce interfaces that work seamlessly with Stories 2.2-2.4
**DI Container:** Story 2.1 is foundation for all subsequent dependency injection
**Service Composition:** Stories 3.2-3.3 depend on solid DI foundation from Epic 2
**Testing Infrastructure:** Story 1.4 enables all subsequent testing, enhanced by Story 4.3

### Validation Checkpoints

**After Epic 1:** All existing functionality works with new interfaces
**After Epic 2:** Services can be created with custom dependencies  
**After Epic 3:** Complex service compositions work correctly
**After Epic 4:** Complete developer experience with documentation and tools

---

## Success Metrics

### Technical Metrics

**Performance:**
- < 50μs additional overhead per request
- < 1μs dependency resolution time
- Memory usage increase < 5% of current baseline
- Zero performance regression in core business logic

**Testability:**
- 100% of framework components have interface abstractions
- Unit test execution time < 1s (without database dependencies)
- Integration test coverage > 90% of component interactions
- Test code coverage > 85% across all framework components

**Maintainability:**
- All public APIs have comprehensive GoDoc documentation
- Zero circular dependencies in final architecture
- Consistent error handling patterns across all components
- Clear separation between interfaces and implementations

### Developer Experience Metrics

**Adoption:**
- Migration guide enables 90%+ of existing services to adopt new patterns
- New service creation time reduced by 50% with DI patterns
- Debugging time for dependency issues reduced by 75%
- Developer onboarding time reduced by 40%

**Quality:**
- Service composition reduces code duplication by 60%+
- Custom dependency injection reduces coupling by 80%+
- Interface-based testing increases test scenario coverage by 200%+
- Clear error messages resolve 90% of configuration issues without support

---

## Migration Strategy

### Phase 1: Interface Foundation (Week 1)
- Extract all core interfaces
- Create comprehensive test utilities
- Validate existing functionality works unchanged

### Phase 2: Dependency Injection (Week 2)
- Implement DI container and core patterns
- Refactor services for constructor injection
- Maintain backward compatibility with factory functions

### Phase 3: Service Composition (Week 3)  
- Enable service embedding and middleware
- Implement advanced composition patterns
- Validate performance characteristics

### Phase 4: Developer Experience (Week 4)
- Complete documentation and tooling
- Validate migration paths work correctly
- Establish performance baselines and monitoring

### Rollout Strategy

**Validation:** Each phase includes comprehensive testing and validation
**Backward Compatibility:** Not required - clean architecture implementation  
**Training:** Developer guide and examples available from Phase 4
**Support:** Architecture decisions document guides consistent implementation

---

## Conclusion

This architecture enables the transformation of endor-sdk-go from a tightly-coupled framework to a flexible, testable, composable system while preserving all innovative features that make it unique. The interface-driven design with lightweight dependency injection provides the foundation for powerful service composition patterns that will significantly enhance developer productivity and code quality.

The 7 architectural decisions create a cohesive technical foundation that addresses all 29 functional requirements while establishing clear implementation patterns for the 20 stories across 4 epics. Most importantly, this architecture solves the core testability problems while enabling the service composition capabilities that will make endor-sdk-go a truly powerful framework for building scalable Go microservices.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Endor SDK Framework                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
│  │  EndorService   │  │ EndorHybridSvc   │  │  API Gateway    │ │
│  │   (Static)      │  │   (Dynamic)      │  │  Integration    │ │
│  └─────────────────┘  └──────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
│  │   Context &     │  │  Schema System   │  │   Middleware    │ │
│  │   Actions       │  │   (Reflection)   │  │   Pipeline      │ │
│  └─────────────────┘  └──────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
│  │  Repository     │  │   MongoDB        │  │   Error         │ │
│  │   Pattern       │  │   Integration    │  │   Handling      │ │
│  └─────────────────┘  └──────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                     Gin HTTP Framework                         │
└─────────────────────────────────────────────────────────────────┘
```

## Technology Stack

### Core Technologies

| Component | Technology | Version | Purpose |
|-----------|------------|---------|---------|
| **Runtime** | Go | 1.21.4 | Primary development language |
| **HTTP Framework** | Gin | v1.10.0 | REST API routing and middleware |
| **Database** | MongoDB | v1.17.3 | Document storage with dynamic schemas |
| **Monitoring** | Prometheus | v1.21.0 | Metrics collection and monitoring |
| **Configuration** | godotenv | v1.5.1 | Environment-based configuration |
| **Documentation** | Swagger/OpenAPI | 3.1 | Automatic API documentation |

### Supporting Libraries

| Library | Purpose |
|---------|---------|
| `golang/snappy` | Data compression |
| `klauspost/compress` | Enhanced compression algorithms |
| `go-playground/validator` | Input validation |
| `goccy/go-json` | High-performance JSON processing |

## Architecture Patterns

### 1. Service Framework Pattern

The SDK implements a dual-service architecture:

#### Static Services (EndorService)
```go
type EndorService struct {
    Resource         string
    Description      string
    Methods          map[string]EndorServiceAction
    Priority         *int
    ResourceMetadata bool
    Version          string
}
```

**Characteristics:**
- Predefined endpoint structure
- Manual action registration
- Custom business logic focus
- Full control over HTTP handling

#### Dynamic Services (EndorHybridService)
```go
type EndorHybridService interface {
    GetResource() string
    GetResourceDescription() string
    GetPriority() *int
    WithCategories(categories []EndorHybridServiceCategory) EndorHybridService
    WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridService
    ToEndorService(metadataSchema Schema) EndorService
}
```

**Characteristics:**
- Automatic CRUD operations
- Schema-driven development
- MongoDB integration
- Category-based specialization

### 2. Repository Pattern

Provides clean separation between business logic and data access:

```go
type ResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface {
    Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error)
    List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error)
    Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
    Delete(ctx context.Context, dto ReadInstanceDTO) error
    Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
}
```

### 3. Generic Type Safety

Leverages Go generics for compile-time type safety:

```go
type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)
```

### 4. Schema-Driven Architecture

Dynamic schema generation using reflection:

```go
func resolveInputSchema[T any]() *RootSchema {
    var zeroValue T
    return convertStructToJsonSchema(reflect.TypeOf(zeroValue))
}
```

## Component Architecture

### Request Processing Pipeline

```
HTTP Request → Authentication → Validation → Context Creation → Handler → Response Serialization
```

1. **Authentication**: Header-based session extraction
2. **Validation**: Automatic JSON payload validation against schema
3. **Context Creation**: Type-safe Endor context with session and payload
4. **Handler Execution**: Business logic execution
5. **Response Serialization**: Standardized response format

### Data Layer Design

#### Core Data Models

```go
type ResourceInstance[T ResourceInstanceInterface] struct {
    This     T              `bson:",inline"`
    Metadata map[string]any `bson:"metadata,omitempty"`
}
```

#### Specialized Resources (Categories)
```go
type ResourceInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct {
    This         T                      `json:",inline" bson:"this"`
    CategoryThis C                      `json:",inline" bson:"categoryThis"`
    Metadata     map[string]interface{} `json:",inline" bson:"metadata,omitempty"`
}
```

### MongoDB Integration Strategy

- **Connection Management**: Singleton pattern with automatic reconnection
- **Collection Strategy**: One collection per resource type
- **ID Management**: Configurable auto-generation (ObjectID or custom)
- **Schema Flexibility**: Dynamic metadata fields alongside typed data

## API Design Principles

### RESTful Conventions

Standard endpoint patterns for hybrid services:
- `GET /{resource}/schema` - Schema definition
- `GET /{resource}/list` - Resource listing
- `GET /{resource}/instance/{id}` - Single resource
- `POST /{resource}/instance` - Create resource
- `PUT /{resource}/instance/{id}` - Update resource
- `DELETE /{resource}/instance/{id}` - Delete resource

### Category-Based Endpoints

For specialized resources:
- `GET /{resource}/categories/{categoryId}/instance/{id}`
- `POST /{resource}/categories/{categoryId}/instance`
- `PUT /{resource}/categories/{categoryId}/instance/{id}`

### Response Format Standardization

```go
type Response[T any] struct {
    Messages []Message   `json:"messages"`
    Data     *T          `json:"data"`
    Schema   *RootSchema `json:"schema"`
}
```

## Configuration Architecture

### Server Configuration
```go
type ServerConfig struct {
    ServerPort                    string
    DocumentDBUri                 string
    HybridResourcesEnabled        bool
    DynamicResourcesEnabled       bool
    DynamicResourceDocumentDBName string
}
```

### Environment-Based Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | 8080 | HTTP server port |
| `DOCUMENT_DB_URI` | mongodb://localhost:27017 | MongoDB connection |
| `HYBRID_RESOURCES_ENABLED` | false | Enable hybrid service features |
| `DYNAMIC_RESOURCES_ENABLED` | false | Enable runtime resource creation |

## Error Handling Architecture

### Custom Error Types
```go
type EndorError struct {
    StatusCode  int
    InternalErr error
}
```

### Error Response Pattern
Errors are wrapped in standardized response format with severity levels:
- **Info**: Informational messages
- **Warning**: Non-blocking issues
- **Error**: Recoverable errors
- **Fatal**: Unrecoverable errors

## Deployment Architecture

### Built-in Observability
- **Health Checks**: `/readyz`, `/livez` endpoints
- **Metrics**: Prometheus `/metrics` endpoint
- **Documentation**: Auto-generated Swagger UI at `/swagger`

### API Gateway Integration
- Automatic Traefik configuration generation
- Service discovery and load balancing
- Authentication middleware integration

## Security Architecture

### Authentication Strategy
- Header-based session management
- Configurable authentication middleware
- Session context propagation through request pipeline

### Authorization Patterns
- Action-level access control
- Resource-based permissions
- Category-specific authorization (for specialized resources)

## Development Architecture

### Builder Pattern for Initialization
```go
sdk.NewEndorInitializer().
    WithEndorServices(&[]sdk.EndorService{...}).
    WithHybridServices(&[]sdk.EndorHybridService{...}).
    Build().
    Init("microservice-name")
```

### Testing Strategy
- Interface-based design for easy mocking
- Separate test services for validation
- Integration testing through full HTTP stack

## Scalability Considerations

### Horizontal Scaling
- Stateless service design
- Shared MongoDB backend
- API gateway load balancing

### Performance Optimizations
- Connection pooling for MongoDB
- JSON processing optimizations
- Prometheus metrics for monitoring

---

*This architecture enables rapid development of consistent, scalable microservices while maintaining flexibility for custom business requirements.*