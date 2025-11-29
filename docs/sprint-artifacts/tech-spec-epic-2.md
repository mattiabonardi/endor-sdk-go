# Epic Technical Specification: Dependency Injection Architecture

Date: November 28, 2025
Author: BMad
Epic ID: 2
Status: Draft

---

## Overview

Epic 2 implements the foundation of the endor-sdk-go dependency injection architecture, transforming the framework from a tightly-coupled system with hard-coded MongoDB dependencies into a flexible, testable architecture where all services accept their dependencies as interface parameters. This epic focuses on establishing the core dependency injection container, refactoring all major services (EndorService, EndorHybridService) for constructor injection, and updating the repository layer to eliminate global singleton dependencies.

The primary goal is enabling developers to customize any framework component implementation while maintaining the powerful automatic CRUD capabilities that make endor-sdk-go unique. This transformation directly addresses functional requirements FR1, FR2, FR9-FR12 from the PRD and establishes the foundation for service composition in Epic 3.

## Objectives and Scope

### In Scope
- **DI Container Implementation**: Lightweight, type-safe dependency injection container with interface registration and resolution
- **Service Constructor Injection**: Refactor EndorService and EndorHybridService to accept dependencies via constructors
- **Repository Layer Abstraction**: Remove hard-coded MongoDB client access, implement repository pattern with interface injection
- **Framework Initializer Updates**: Enhance EndorInitializer to wire dependencies automatically with validation
- **Backward Compatibility Factories**: Provide convenience constructors using default implementations for smooth adoption

### Out of Scope  
- **Service Composition Patterns**: Embedding services within services (Epic 3)
- **Middleware Implementation**: Cross-cutting concerns pipeline (Epic 3)
- **Advanced Testing Utilities**: Beyond basic interface mocking (Epic 4)
- **Developer Tooling**: CLI tools, performance profiling (Epic 4)
- **Migration Support**: No backward compatibility required per PRD constraints

## System Architecture Alignment

This epic aligns directly with Architecture Decision 1 (Lightweight Custom DI Container) and Decision 2 (Interface Segregation) from the architecture document. The implementation follows the documented pattern of interface-based registration with type-safe resolution, supporting the three core dependency scopes: Singleton (default), Transient, and Scoped.

The refactoring maintains alignment with the existing dual-service pattern (EndorService + EndorHybridService) while enabling the dependency injection capabilities required for the service composition patterns in Epic 3. All changes preserve the automatic schema generation, MongoDB integration, and type safety that define the endor-sdk-go framework's unique value proposition.

The repository abstraction strategy follows the established Repository Pattern while maintaining compatibility with the existing MongoDB-based implementations, ensuring that the powerful category-based specialization and dynamic schema features continue to function seamlessly with injected dependencies.

## Detailed Design

### Services and Modules

| Service/Module | Responsibility | Inputs | Outputs | Owner |
|----------------|----------------|---------|----------|-------|
| **DI Container** (`sdk/di/container.go`) | Type-safe dependency registration and resolution | Interface registrations, dependency requests | Resolved dependencies, validation errors | Core Framework |
| **EndorService DI** (`sdk/core/endor_service.go`) | Constructor injection for static services | Repository, Config, Logger interfaces | Configured EndorService instances | Core Framework |
| **EndorHybridService DI** (`sdk/core/endor_hybrid_service.go`) | Constructor injection for dynamic services | Repository, Config, Schema interfaces | Configured EndorHybridService instances | Core Framework |
| **Repository Abstraction** (`sdk/repository/`) | Interface-based data access layer | Database client, configuration interfaces | Repository implementations | Data Layer |
| **Framework Initializer** (`sdk/endor_initializer.go`) | Automatic dependency wiring and validation | Service configurations, custom overrides | Fully wired service instances | Framework Core |

### Data Models and Contracts

**DI Container Interface:**
```go
type Container interface {
    Register[T any](impl T, scope Scope) error
    RegisterFactory[T any](factory func(Container) T, scope Scope) error
    Resolve[T any]() (T, error)
    Validate() []error
    GetDependencyGraph() map[string][]string
}

type Scope int
const (
    Singleton Scope = iota  // Default - shared instance
    Transient              // New instance each time
    Scoped                 // Bound to request/context scope
)
```

**Repository Interfaces:**
```go
type RepositoryInterface interface {
    Create(ctx context.Context, resource any) error
    Read(ctx context.Context, id string, result any) error
    Update(ctx context.Context, resource any) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter map[string]any, results any) error
}

type DatabaseClientInterface interface {
    Collection(name string) CollectionInterface
    Database(name string) DatabaseInterface
    Close(ctx context.Context) error
    Ping(ctx context.Context) error
}
```

**Service Constructor Signatures:**
```go
// EndorService with dependency injection
type EndorServiceDependencies struct {
    Repository RepositoryInterface
    Config     ConfigProviderInterface
    Logger     LoggerInterface
    Context    EndorContextInterface
}

func NewEndorServiceWithDeps(deps EndorServiceDependencies) EndorService

// EndorHybridService with dependency injection  
type EndorHybridServiceDependencies struct {
    Repository    RepositoryInterface
    Config        ConfigProviderInterface
    SchemaManager SchemaManagerInterface
    Logger        LoggerInterface
}

func NewEndorHybridServiceWithDeps(deps EndorHybridServiceDependencies) EndorHybridService
```

### APIs and Interfaces

**Core Dependency Injection APIs:**

| Method | Path | Request/Response | Purpose |
|--------|------|------------------|---------|
| `container.Register[T]()` | Programmatic | `Register[RepositoryInterface](mongoRepo, Singleton)` | Register concrete implementation for interface |
| `container.RegisterFactory[T]()` | Programmatic | `RegisterFactory[Repository](factoryFunc, Transient)` | Register factory function for complex initialization |
| `container.Resolve[T]()` | Programmatic | `Resolve[RepositoryInterface]() -> (impl, error)` | Resolve dependency by interface type |
| `container.Validate()` | Programmatic | `Validate() -> []error` | Validate dependency graph completeness |

**Service Creation APIs:**

| Method | Purpose | Dependencies | Example |
|--------|---------|--------------|---------|
| `NewEndorServiceWithDeps()` | Create service with injected dependencies | Repository, Config, Logger, Context | Unit testing with mocks |
| `NewEndorService()` | Convenience constructor with defaults | None (uses default implementations) | Production usage |
| `NewEndorHybridServiceWithDeps()` | Create hybrid service with dependencies | Repository, Config, Schema, Logger | Custom repository implementations |
| `NewEndorHybridService()` | Convenience constructor | None (uses default implementations) | Standard usage patterns |

### Workflows and Sequencing

**Dependency Injection Workflow:**

1. **Container Initialization**
   ```
   container := di.NewContainer()
   container.Register[RepositoryInterface](mongoRepo, di.Singleton)
   container.Register[ConfigProviderInterface](envConfig, di.Singleton)
   container.Register[LoggerInterface](structuredLogger, di.Singleton)
   ```

2. **Service Creation with Dependencies**
   ```
   repo := container.Resolve[RepositoryInterface]()
   config := container.Resolve[ConfigProviderInterface]()
   logger := container.Resolve[LoggerInterface]()
   
   service := NewEndorServiceWithDeps(EndorServiceDependencies{
       Repository: repo,
       Config:     config, 
       Logger:     logger,
   })
   ```

3. **Framework Initialization Sequence**
   ```
   initializer := NewEndorInitializer()
   initializer.WithContainer(container)
   initializer.WithServices(serviceConfigs)
   
   // Automatic dependency validation and wiring
   framework := initializer.Build()  // Validates dependency graph
   framework.Init("service-name")    // Starts all services
   ```

**Repository Layer Refactoring Sequence:**

1. **Replace Global MongoDB Access**
   ```
   // Before: Hard-coded global access
   client := GetMongoClient()
   
   // After: Injected client interface
   type MongoRepository struct {
       client DatabaseClientInterface
   }
   ```

2. **Constructor Injection Pattern**
   ```
   func NewMongoRepositoryWithClient(client DatabaseClientInterface, config Config) RepositoryInterface {
       return &MongoRepository{
           client: client,
           config: config,
       }
   }
   ```

## Non-Functional Requirements

### Performance

- **Dependency Resolution Latency**: < 1μs per interface resolution (measured via benchmarks)
- **Service Creation Overhead**: < 10μs additional latency compared to current direct instantiation
- **Memory Footprint**: DI container uses < 1MB RAM for typical service configurations (50-100 dependencies)
- **Container Validation Time**: Complete dependency graph validation < 5ms during application startup
- **Zero Runtime Reflection**: All dependency resolution uses compile-time type information, no reflection overhead

### Security

- **Interface Validation**: Container prevents registration of non-interface types as dependency keys
- **Singleton Thread Safety**: All singleton dependencies are protected against concurrent access races
- **Dependency Isolation**: Services cannot access dependencies not explicitly injected, preventing privilege escalation
- **Container Immutability**: Once container validation passes, dependency graph cannot be modified during runtime
- **Type Safety Enforcement**: Generic constraints prevent runtime type casting errors and ensure compile-time compatibility

### Reliability/Availability

- **Circular Dependency Detection**: Container validation detects and rejects circular dependencies before service startup
- **Graceful Degradation**: Missing optional dependencies result in warnings, not service failures
- **Dependency Health Monitoring**: Failed dependency creation logs structured errors with resolution guidance
- **Startup Validation**: Complete dependency graph validation prevents runtime dependency resolution failures
- **Error Recovery**: Dependency creation failures include retry mechanisms for transient database connection issues

### Observability

- **Dependency Resolution Metrics**: Prometheus metrics tracking resolution time per interface type and dependency scope
- **Container Health Endpoint**: `/health/dependencies` endpoint showing dependency graph status and resolution statistics
- **Structured Logging**: All dependency operations log with correlation IDs, dependency chains, and resolution context
- **Debug Tracing**: Debug mode provides detailed dependency resolution traces showing registration/resolution flow
- **Dependency Graph Visualization**: Container introspection APIs enable runtime dependency graph visualization for debugging

## Dependencies and Integrations

### External Dependencies

| Dependency | Version | Purpose | DI Integration Point |
|------------|---------|---------|----------------------|
| **go.mongodb.org/mongo-driver** | v1.17.3 | MongoDB client and operations | DatabaseClientInterface implementation |
| **github.com/gin-gonic/gin** | v1.10.0 | HTTP routing and middleware | HTTP handler integration (unchanged) |
| **github.com/prometheus/client_golang** | v1.21.0 | Metrics collection | MetricsInterface implementation |
| **github.com/joho/godotenv** | v1.5.1 | Environment configuration | ConfigProviderInterface implementation |
| **github.com/stretchr/testify** | v1.11.1 | Testing utilities and mocks | TestUtils package foundation |

### Internal Integration Points

**DI Container → Services Integration:**
- Container provides type-safe resolution for all service dependencies
- Services register their interface implementations during initialization
- Framework initializer orchestrates dependency wiring and validation

**Repository → Database Client Integration:**
- Repository implementations accept DatabaseClientInterface instead of concrete MongoDB client
- Client interface abstracts MongoDB operations for testing and alternative implementations
- Connection lifecycle management becomes container-managed dependency

**Service → Repository Integration:**  
- EndorService and EndorHybridService receive repository dependencies via constructors
- Repository interface abstracts CRUD operations maintaining existing method signatures
- Category-based specialization continues working through injected repository interface

**Configuration → Environment Integration:**
- ConfigProviderInterface abstracts environment variable access and file-based configuration
- Multiple configuration sources can be registered (environment, files, remote config)
- Services receive configuration through injected interface, enabling test configurations

### Backward Compatibility Strategy

**Convenience Constructors:**
- `NewEndorService()` maintains existing creation patterns using default dependencies
- `NewEndorHybridService()` preserves current usage for standard scenarios
- Factory functions handle dependency wiring automatically for simple use cases

**Default Implementation Registry:**
- Framework provides standard implementations (MongoRepository, EnvConfig, StructuredLogger)
- Default registrations occur automatically during framework initialization
- Advanced users can override any dependency while maintaining interface compatibility

## Acceptance Criteria (Authoritative)

### AC-2.1: Dependency Injection Container Implementation
**Given** the need for type-safe dependency management
**When** I implement the DI container with interface registration and resolution
**Then** the container provides type-safe dependency registration and resolution
**And** container supports Singleton, Transient, and Scoped dependency lifecycles
**And** circular dependency detection prevents invalid configurations
**And** container validation occurs during application startup with clear error messages
**And** dependency resolution time is < 1μs per interface resolution

### AC-2.2: EndorService Constructor Injection
**Given** the current EndorService with hard-coded dependencies
**When** I refactor EndorService to accept dependencies via constructor injection
**Then** `NewEndorServiceWithDeps()` accepts Repository, Config, Logger, and Context interfaces as parameters
**And** all internal EndorService operations use injected dependencies instead of globals
**And** `NewEndorService()` convenience constructor provides default implementations for backward compatibility
**And** dependency validation ensures all required dependencies are provided
**And** existing EndorService method signatures and behavior remain unchanged

### AC-2.3: EndorHybridService Constructor Injection  
**Given** the current EndorHybridService with tightly coupled dependencies
**When** I refactor EndorHybridService to support dependency injection
**Then** `NewEndorHybridServiceWithDeps()` accepts Repository, Config, SchemaManager, and Logger interfaces
**And** `ToEndorService()` method passes injected dependencies to generated EndorService instances
**And** category operations and automatic CRUD functionality use injected repository interface
**And** schema generation and validation work with injected SchemaManager dependency
**And** automatic action generation uses injected dependencies consistently

### AC-2.4: Repository Layer Dependency Injection
**Given** the current repository layer with hard-coded MongoDB client access
**When** I refactor repositories to accept database client via constructor injection
**Then** `NewRepositoryWithClient()` accepts DatabaseClientInterface parameter instead of using global GetMongoClient()
**And** all CRUD operations use injected client interface instead of direct MongoDB calls
**And** repository implementations support both MongoDB and alternative database clients through interface abstraction
**And** transaction support and connection pooling work with injected client dependencies
**And** existing repository functionality and performance characteristics are preserved

### AC-2.5: Framework Initializer Dependency Wiring
**Given** the new dependency injection patterns across all framework components
**When** I enhance EndorInitializer to handle automatic dependency wiring
**Then** `EndorInitializer.Build()` creates properly wired service instances with validated dependency graphs
**And** initializer provides `WithContainer()`, `WithCustomRepository()`, and `WithCustomConfig()` methods for dependency override
**And** dependency graph validation occurs during Build() with clear error messages for missing or misconfigured dependencies
**And** initializer supports both simple default configurations and advanced custom dependency scenarios
**And** proper dependency cleanup occurs during framework shutdown

## Traceability Mapping

| Acceptance Criteria | Epic 2 Stories | Framework Components | Test Coverage |
|-------------------|----------------|---------------------|---------------|
| **AC-2.1: DI Container** | Story 2.1 | `sdk/di/container.go`, Container interface | Unit tests for registration/resolution, circular dependency detection, performance benchmarks |
| **AC-2.2: EndorService DI** | Story 2.2 | `sdk/core/endor_service.go`, EndorServiceDependencies struct | Unit tests with mocked dependencies, integration tests with real dependencies |
| **AC-2.3: EndorHybridService DI** | Story 2.3 | `sdk/core/endor_hybrid_service.go`, EndorHybridServiceDependencies | Tests for ToEndorService() dependency passing, category operations with injected repo |
| **AC-2.4: Repository DI** | Story 2.4 | `sdk/repository/`, RepositoryInterface, DatabaseClientInterface | Repository tests with mock client, MongoDB implementation tests |
| **AC-2.5: Framework Initializer** | Story 2.5 | `sdk/endor_initializer.go`, dependency wiring logic | End-to-end initialization tests, dependency override validation |

### Functional Requirements Traceability

| FR | Description | Epic 2 Implementation | Validation Method |
|----|-------------|----------------------|-------------------|
| **FR1** | Interface-based EndorService with DI | Story 2.2: Constructor injection | Unit tests with interface mocks |
| **FR2** | Interface-based EndorHybridService with DI | Story 2.3: Constructor injection | Unit tests with interface mocks |
| **FR9** | Service constructors accept interface parameters | Stories 2.2, 2.3, 2.4 | Constructor signature validation |
| **FR10** | Eliminate hard-coded singleton dependencies | Stories 2.1, 2.4, 2.5 | Dependency graph analysis |
| **FR11** | Inject custom implementations | Stories 2.1, 2.5 | Custom implementation tests |
| **FR12** | Support constructor injection and factory patterns | Stories 2.1, 2.5 | Both pattern usage tests |
| **FR26** | Database access through repository interfaces | Story 2.4 | Repository interface compliance tests |
| **FR27** | Schema generation remains automatic | Story 2.3 | Schema generation with DI tests |
| **FR28** | Custom repository implementations | Story 2.4 | In-memory repository tests |
| **FR29** | Connection management dependency-injectable | Story 2.4 | Custom client injection tests |

### Architecture Decision Mapping

| Architecture Decision | Epic 2 Implementation | Validation |
|----------------------|----------------------|-----------|
| **Decision 1: Lightweight DI Container** | Story 2.1: Custom container implementation | Performance benchmarks < 1μs resolution |
| **Decision 2: Interface Segregation** | Stories 2.2, 2.3, 2.4: Focused interfaces | Interface compliance tests |
| **Decision 6: Package Organization** | All stories: Clear separation interfaces/implementations | Package structure validation |
| **Decision 7: Implementation Consistency** | All stories: Consistent naming and patterns | Code style validation |

## Risks, Assumptions, Open Questions

### **RISK**: DI Container Performance Overhead
**Mitigation**: Implement zero-allocation resolution using compile-time generics and singleton caching. Benchmark against current performance baselines during implementation.
**Action**: Story 2.1 includes comprehensive performance testing to validate < 1μs resolution requirement.

### **RISK**: Breaking Changes During Refactoring  
**Mitigation**: Maintain convenience constructors (`NewEndorService()`, `NewEndorHybridService()`) that preserve existing usage patterns.
**Action**: Each story includes backward compatibility validation tests to ensure existing code continues working.

### **ASSUMPTION**: MongoDB Client Interface Abstraction
**Validation Required**: Verify that MongoDB driver operations can be cleanly abstracted without losing functionality.
**Action**: Story 2.4 includes proof-of-concept implementation to validate interface design before full implementation.

### **ASSUMPTION**: Generic Type Safety Preservation
**Validation Required**: Confirm that DI container generics work correctly with existing EndorService/EndorHybridService generics.
**Action**: Story 2.1 includes type safety validation tests with complex generic scenarios.

### **QUESTION**: Dependency Scope Strategy for Services
**Current Thinking**: Default to Singleton for services, Scoped for request-bound dependencies, Transient for stateless utilities.
**Decision Needed**: Confirm scope strategy aligns with framework usage patterns and performance requirements.

### **QUESTION**: Error Handling Strategy for Missing Dependencies
**Options**: (1) Panic on missing required dependencies, (2) Return errors from constructors, (3) Provide default implementations
**Decision**: Use panic for required dependencies (fail-fast), warnings for optional dependencies with defaults.

### **ASSUMPTION**: Interface Backward Compatibility
**Validation Required**: New interfaces must not break existing method signatures or behavior contracts.
**Action**: Each story includes interface compliance tests comparing old vs new method signatures and behavior.

## Test Strategy Summary

### **Unit Testing Strategy**
- **DI Container**: Test registration, resolution, validation, and error scenarios with comprehensive edge case coverage
- **Service Constructors**: Test dependency injection with mocked interfaces, validate proper dependency usage
- **Repository Layer**: Test interface implementations with mock database clients, validate operation correctness
- **Integration Validation**: Test end-to-end dependency wiring through framework initializer

### **Performance Testing Approach**
- **Benchmark DI Resolution**: Measure dependency resolution time across different dependency graph sizes
- **Memory Usage Validation**: Verify DI container memory footprint remains under 1MB for typical configurations
- **Regression Testing**: Compare performance against current implementation baselines for all operations

### **Compatibility Testing Strategy**  
- **Interface Compliance**: Validate new interfaces match existing method signatures exactly
- **Behavior Preservation**: Test that dependency injection preserves all existing service behaviors
- **Convenience Constructor Testing**: Ensure `NewEndorService()` and `NewEndorHybridService()` maintain current usage patterns

### **Integration Testing Coverage**
- **Full Stack Wiring**: Test complete dependency injection from framework initializer through service operations  
- **Database Integration**: Validate MongoDB operations work correctly through repository interface abstraction
- **Configuration Integration**: Test environment-based and custom configuration scenarios with injected providers

### **Error Scenario Validation**
- **Missing Dependencies**: Test error messages and failure modes for incomplete dependency configurations
- **Circular Dependencies**: Validate detection and clear error reporting for invalid dependency graphs
- **Type Safety**: Test compile-time and runtime type safety enforcement across all dependency scenarios