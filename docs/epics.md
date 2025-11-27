# endor-sdk-go - Epic Breakdown

**Author:** BMad
**Date:** November 27, 2025
**Project Level:** Framework/SDK Architectural Refactoring
**Target Scale:** Developer Tool for Go Microservices

---

## Overview

This document provides the complete epic and story breakdown for endor-sdk-go architectural refactoring, decomposing the requirements from the [PRD](./prd.md) into implementable stories focused on making EndorService and EndorHybridService testable and composable.

**Living Document Notice:** This is the initial version. It will be updated after Architecture workflows add technical details to stories.

**Workflow Mode:** INITIAL CREATION
**Available Context:** PRD requirements extracted and analyzed

---

## Functional Requirements Inventory

**Core Service Framework:**
- FR1: Framework provides interface-based EndorService creation with dependency injection support
- FR2: Framework provides interface-based EndorHybridService creation with automatic CRUD capabilities
- FR3: Developers can compose services within services using dependency injection patterns  
- FR4: All framework components implement well-defined interfaces for easy mocking

**Enhanced Testability:**
- FR5: Developers can unit test services without MongoDB dependencies through interface mocking
- FR6: Framework provides test utility package with pre-built mocks for common interfaces
- FR7: Developers can test service composition patterns in isolation
- FR8: Integration tests can use in-memory or test database implementations

**Dependency Management:**
- FR9: All service constructors accept dependencies as interface parameters
- FR10: Framework eliminates hard-coded singleton dependencies (like MongoDB client)
- FR11: Developers can inject custom implementations for any framework dependency
- FR12: Framework supports both constructor injection and factory patterns

**Backward Compatibility:**
- ~~FR13: Existing EndorService implementations continue to work with minimal changes~~ (REMOVED - No backward compatibility needed)
- ~~FR14: Existing EndorHybridService implementations migrate through deprecation cycle~~ (REMOVED - No backward compatibility needed)  
- ~~FR15: Current API contracts remain stable during transition period~~ (REMOVED - No backward compatibility needed)
- ~~FR16: Migration tools assist developers in updating to new architecture~~ (REMOVED - No backward compatibility needed)

**Service Composition:**
- FR17: EndorHybridService can embed other EndorServices through dependency injection
- FR18: Services can share common dependencies without coupling to implementations
- FR19: Middleware pattern enables cross-cutting concerns (logging, metrics, auth)
- FR20: Service lifecycle management supports proper dependency teardown

**Developer Experience:**
- FR21: Framework provides clear error messages for dependency injection failures
- FR22: Documentation includes complete examples for testing patterns
- FR23: Migration guide covers all common refactoring scenarios
- FR24: Framework maintains all current performance characteristics
- FR25: API documentation generation continues to work seamlessly

**MongoDB Integration:**
- FR26: Database access occurs through repository interfaces that can be mocked
- FR27: Schema generation and validation remain automatic for hybrid services
- FR28: Developers can provide custom repository implementations for testing
- FR29: Connection management becomes dependency-injectable

---

## Epic Structure and Value Delivery

**Epic 1: Interface Foundation & Testability**
**Goal:** Framework components become mockable and testable
**User Value:** Developers can write unit tests without external dependencies

**Epic 2: Dependency Injection Architecture**  
**Goal:** All services use constructor injection for dependencies
**User Value:** Developers can compose services cleanly and customize implementations

**Epic 3: Service Composition Framework**
**Goal:** Services can embed other services with zero boilerplate
**User Value:** Developers can build complex service hierarchies using simple patterns

**Epic 4: Developer Experience & Tooling**
**Goal:** Excellent developer experience with clear documentation and debugging tools
**User Value:** Developers can quickly adopt, debug, and optimize their use of the new architecture

---

## FR Coverage Map

**Epic 1 (Interface Foundation & Testability):**
- FR4: All framework components implement interfaces for mocking
- FR5: Unit testing without MongoDB dependencies
- FR6: Test utility package with pre-built mocks
- FR7: Test service composition patterns in isolation
- FR8: Integration tests with in-memory implementations

**Epic 2 (Dependency Injection Architecture):**
- FR1: Interface-based EndorService creation with DI
- FR2: Interface-based EndorHybridService creation with DI
- FR9: Service constructors accept interface parameters
- FR10: Eliminate hard-coded singleton dependencies
- FR11: Inject custom implementations for any dependency
- FR12: Support constructor injection and factory patterns

**Epic 3 (Service Composition Framework):**
- FR3: Compose services within services using DI patterns
- FR17: EndorHybridService embeds EndorServices through DI
- FR18: Share dependencies without implementation coupling
- FR19: Middleware pattern for cross-cutting concerns
- FR20: Proper service lifecycle management

**Epic 4: Developer Experience & Tooling**
- FR21: Clear error messages for DI failures → Story 4.2
- FR22: Complete testing pattern documentation → Story 4.1, 4.3
- FR23: Comprehensive developer guide and examples → Story 4.1
- FR24: Framework maintains excellent performance characteristics → Story 4.5
- FR25: API documentation generation works seamlessly → Story 4.1
- FR26: Database access through repository interfaces → Story 2.4
- FR27: Schema generation remains automatic → Story 2.3
- FR28: Custom repository implementations for testing → Story 2.4, 4.3
- FR29: Connection management becomes dependency-injectable → Story 2.4
- FR30: Preserve and enhance generic type safety → Story 2.2, 2.3
- FR31: JSON schema generation with DI services → Story 2.3
- FR32: Automatic input validation with extension points → Story 2.2, 2.3
- FR33: Maintain current error handling patterns → Story 2.2, 2.3, 4.2

---

## Epic 1: Interface Foundation & Testability

**Goal:** Framework components become mockable and testable, enabling developers to write unit tests without external dependencies.

### Story 1.1: Extract Core Service Interfaces

As a framework developer,
I want to extract interfaces for EndorService and EndorHybridService,
So that developers can mock these components in their tests.

**Acceptance Criteria:**

**Given** the current concrete EndorService and EndorHybridService implementations
**When** I extract their public methods into Go interfaces
**Then** EndorServiceInterface and EndorHybridServiceInterface are defined

**And** all current method signatures are preserved in the interfaces
**And** concrete implementations satisfy their respective interfaces
**And** interfaces include comprehensive GoDoc documentation with examples

**Prerequisites:** None (first story)

**Technical Notes:** 
- Create `sdk/interfaces/` package for all framework interfaces
- EndorServiceInterface should include: GetResource(), GetDescription(), GetMethods(), etc.
- EndorHybridServiceInterface should include: WithCategories(), WithActions(), ToEndorService()
- Use Go's implicit interface satisfaction (no explicit implements)
- Add interface compliance tests: `var _ EndorServiceInterface = (*EndorService)(nil)`

### Story 1.2: Extract Repository Interfaces

As a developer using the framework,
I want repository operations to be interface-based,
So that I can mock database interactions in my unit tests.

**Acceptance Criteria:**

**Given** the current MongoDB-coupled repository implementations
**When** I extract repository interfaces for all data access patterns
**Then** RepositoryInterface, ResourceRepositoryInterface, and MongoRepositoryInterface are defined

**And** all current repository methods are abstracted into interfaces
**And** MongoDB implementations satisfy the repository interfaces  
**And** interfaces are generic-friendly for type safety
**And** interface methods return domain errors (not database-specific errors)

**Prerequisites:** Story 1.1 (interface foundation established)

**Technical Notes:**
- Focus on EndorServiceRepository and related data access patterns
- Abstract away MongoDB-specific types (use domain models in interface signatures)
- Include CRUD operations: Create, Read, Update, Delete, List, Query
- Support for both static and dynamic resource patterns
- Interface should support both synchronous and future asynchronous patterns

### Story 1.3: Extract Configuration and Context Interfaces  

As a developer building services with the framework,
I want configuration and context management to be interface-based,
So that I can provide custom configurations and mock contexts in tests.

**Acceptance Criteria:**

**Given** the current hard-coded configuration and context handling
**When** I extract interfaces for configuration providers and context management
**Then** ConfigProviderInterface and EndorContextInterface are defined

**And** current configuration loading logic is abstracted behind interfaces
**And** EndorContext operations become mockable for testing
**And** interface includes both environment-based and custom configuration sources
**And** context propagation patterns are preserved

**Prerequisites:** Story 1.2 (core repository interfaces exist)

**Technical Notes:**
- ConfigProviderInterface should abstract environment variables, file-based config, etc.
- EndorContextInterface should handle request context, user sessions, metadata
- Support for both production (env-based) and test (in-memory) implementations
- Include context cancellation and timeout patterns
- Add validation for configuration completeness and correctness

### Story 1.4: Create Test Utility Package with Mocks

As a developer using the framework,
I want pre-built mock implementations of framework interfaces,
So that I can quickly write unit tests without creating mocks manually.

**Acceptance Criteria:**

**Given** the framework interfaces from previous stories  
**When** I create a comprehensive test utilities package
**Then** `sdk/testutils` package provides ready-to-use mocks

**And** MockEndorService, MockEndorHybridService, MockRepository implementations exist
**And** mocks support behavior configuration (return values, errors, call tracking)
**And** helper functions exist for common test scenarios (setup service, create test data, etc.)
**And** examples demonstrate testing patterns for each service type
**And** performance mocks simulate realistic latency and load

**Prerequisites:** Stories 1.1, 1.2, 1.3 (all core interfaces defined)

**Technical Notes:**
- Use testify/mock or similar for behavior verification
- Include builders for test data creation (test services, test schemas, etc.)
- Provide both strict mocks (fail on unexpected calls) and lenient mocks
- Include helper functions: `NewTestEndorService()`, `MockWithBehavior()`, etc.
- Add examples in GoDoc showing full test scenarios

### Story 1.5: Refactor Existing Tests to Use Interface Mocks

As a framework maintainer,
I want existing tests updated to use the new interface-based mocking approach,
So that I demonstrate best practices and validate the new testing capabilities.

**Acceptance Criteria:**

**Given** existing test files in `/test` directory and new mock utilities  
**When** I refactor tests to use interface-based mocks instead of concrete dependencies
**Then** all tests in `test/` directory use the new mocking approach

**And** test execution time improves (no MongoDB dependencies in unit tests)
**And** tests become more focused on business logic rather than infrastructure  
**And** test coverage increases due to ability to test error scenarios easily
**And** examples exist for testing service composition patterns

**Prerequisites:** Story 1.4 (test utilities with mocks available)

**Technical Notes:**
- Focus on `endor_hybrid_service_test.go` and `endor_service_test.go` as primary examples
- Separate unit tests (mocked dependencies) from integration tests (real dependencies)
- Add test categories: `//go:build unit` and `//go:build integration`
- Create examples of testing error scenarios that were difficult before
- Document testing patterns in test files for developer reference

---

## Epic 2: Dependency Injection Architecture

**Goal:** All services use constructor injection for dependencies, eliminating hard-coded singletons and enabling clean service composition.

### Story 2.1: Implement Dependency Injection Container

As a framework developer,
I want a lightweight dependency injection container,
So that services can declare and receive their dependencies automatically.

**Acceptance Criteria:**

**Given** the need for dependency injection throughout the framework
**When** I implement a DI container with registration and resolution capabilities
**Then** `sdk/di` package provides Container interface and implementation

**And** container supports interface-based dependency registration
**And** container resolves dependencies with proper lifecycle management (singleton, transient)
**And** circular dependency detection prevents infinite loops
**And** container provides clear error messages for missing or misconfigured dependencies
**And** container supports both constructor injection and factory patterns

**Prerequisites:** Story 1.5 (interface foundation complete)

**Technical Notes:**
- Keep DI container lightweight (avoid heavy frameworks like wire/dig if possible)
- Support registration: `container.Register[Interface](implementation)`
- Support resolution: `container.Resolve[Interface]()`
- Include dependency scopes: Singleton (default), Transient, Scoped
- Add container validation to detect configuration issues early
- Support optional dependencies with fallback to defaults

### Story 2.2: Refactor EndorService for Constructor Injection

As a developer using the framework,
I want EndorService creation to accept dependencies via constructor injection,
So that I can customize implementations and mock dependencies easily.

**Acceptance Criteria:**

**Given** the current EndorService with hard-coded dependencies
**When** I refactor constructors to accept dependencies as parameters
**Then** NewEndorService() accepts all required dependencies as interface parameters

**And** all internal dependencies (repositories, config, context) are injected
**And** existing EndorService creation patterns remain backward compatible
**And** factory functions provide sensible defaults for production use
**And** validation ensures all required dependencies are provided
**And** clear error messages indicate missing dependencies

**Prerequisites:** Story 2.1 (DI container available)

**Technical Notes:**
- Create NewEndorServiceWithDeps(repo RepositoryInterface, config ConfigInterface, ...) 
- Maintain NewEndorService() as convenience function using default implementations
- Update EndorService struct to hold interface references, not concrete types
- Add dependency validation in constructor (panic on nil required dependencies)
- Support optional dependencies with reasonable defaults
- Update all internal EndorService methods to use injected dependencies

### Story 2.3: Refactor EndorHybridService for Constructor Injection

As a developer using the framework,
I want EndorHybridService creation to support dependency injection,
So that I can provide custom repositories and configurations for different environments.

**Acceptance Criteria:**

**Given** the current EndorHybridService with tightly coupled dependencies
**When** I refactor creation to use dependency injection patterns
**Then** NewHybridServiceWithDeps() accepts all dependencies as interface parameters

**And** ToEndorService() method uses injected dependencies instead of globals
**And** WithCategories() and WithActions() work with injected dependencies
**And** automatic CRUD operations use injected repository interfaces
**And** schema generation works with injected configuration
**And** backward compatibility is maintained for existing creation patterns

**Prerequisites:** Story 2.2 (EndorService DI patterns established)

**Technical Notes:**
- Update EndorHybridServiceImpl to hold interface references
- Refactor ToEndorService() to pass dependencies to created EndorService
- Ensure category operations use injected repository
- Update automatic action generation to use injected dependencies
- Support dependency inheritance: hybrid service passes dependencies to generated EndorService
- Add validation for required dependencies

### Story 2.4: Refactor Repository Layer for Dependency Injection

As a developer using the framework,
I want repository creation to use dependency injection,
So that I can provide custom database connections and configurations.

**Acceptance Criteria:**

**Given** the current repository layer with hard-coded MongoDB client access
**When** I refactor repositories to accept dependencies via constructors
**Then** NewEndorServiceRepository() accepts database client and config interfaces

**And** MongoDB-specific implementation uses injected client instead of global GetMongoClient()
**And** repository interfaces support both MongoDB and alternative implementations
**And** connection lifecycle management is handled by injected dependencies
**And** repository creation validates all required dependencies
**And** existing repository functionality is preserved

**Prerequisites:** Story 2.3 (service layer DI complete)

**Technical Notes:**
- Abstract database client access: `NewRepositoryWithClient(client DatabaseClientInterface)`
- Remove direct calls to GetMongoClient() from repository implementations
- Support repository chaining (repository can depend on other repositories)
- Add repository factory patterns for common configurations
- Ensure transaction support works with injected clients
- Update all CRUD operations to use injected client

### Story 2.5: Update Framework Initializer for Dependency Injection

As a developer using the framework,
I want the EndorInitializer to wire dependencies automatically,
So that I can configure the entire service graph in one place.

**Acceptance Criteria:**

**Given** the new dependency injection patterns in services and repositories
**When** I update EndorInitializer to handle dependency wiring
**Then** EndorInitializer.Build() creates properly wired service instances

**And** initializer provides hooks for custom dependency registration
**And** initializer validates the complete dependency graph before starting
**And** clear error messages indicate dependency configuration problems
**And** backward compatibility is maintained for simple use cases
**And** advanced users can override any dependency with custom implementations

**Prerequisites:** Story 2.4 (full dependency chain supports injection)

**Technical Notes:**
- Add WithContainer() method to provide custom DI container
- Add WithCustomRepository(), WithCustomConfig() methods for overrides
- Implement dependency graph validation during Build()
- Support both simple (defaults) and advanced (custom) initialization patterns
- Ensure proper dependency cleanup during shutdown
- Add container introspection for debugging dependency issues

---

## Epic 3: Service Composition Framework

**Goal:** Services can embed other services with zero boilerplate, enabling complex service hierarchies using simple dependency injection patterns.

### Story 3.1: Implement Service Middleware Pipeline

As a developer using the framework,
I want to add cross-cutting concerns (logging, metrics, auth) to services declaratively,
So that I can enhance services without modifying their core logic.

**Acceptance Criteria:**

**Given** services created with dependency injection
**When** I implement a middleware pipeline system
**Then** services can be wrapped with middleware components

**And** middleware can be chained in configurable order
**And** middleware interfaces are simple and composable
**And** common middleware (logging, metrics, auth) are provided out-of-the-box
**And** custom middleware can be implemented easily
**And** middleware execution preserves service interface contracts
**And** middleware supports both synchronous and asynchronous patterns

**Prerequisites:** Story 2.5 (dependency injection framework complete)

**Technical Notes:**
- Create MiddlewareInterface with Before/After hooks
- Implement middleware chaining: `WithMiddleware(auth, logging, metrics)`
- Provide built-in middleware: AuthMiddleware, LoggingMiddleware, MetricsMiddleware
- Ensure middleware can access and modify request/response context
- Support middleware short-circuiting (early returns for auth failures, etc.)
- Add middleware performance monitoring (execution time per middleware)

### Story 3.2: Enable EndorService Embedding in EndorHybridService

As a developer building complex services,
I want to embed existing EndorServices within my EndorHybridService,
So that I can reuse business logic without code duplication.

**Acceptance Criteria:**

**Given** dependency-injected EndorService and EndorHybridService components
**When** I embed an EndorService within an EndorHybridService
**Then** the hybrid service can delegate specific actions to the embedded service

**And** embedded service methods are accessible through the hybrid service interface
**And** method name conflicts are resolved with clear precedence rules
**And** embedded service dependencies are properly managed by the parent service
**And** embedded services maintain their own middleware and configuration
**And** composition preserves type safety and method signatures

**Prerequisites:** Story 3.1 (middleware pipeline available for embedded services)

**Technical Notes:**
- Add EmbedService() method to EndorHybridService: `hybrid.EmbedService("prefix", endorService)`
- Implement method delegation with optional prefix namespacing
- Handle dependency sharing between parent and embedded services
- Preserve embedded service middleware stack
- Add validation to prevent circular service embedding
- Support multiple embedded services with clear method resolution

### Story 3.3: Implement Shared Dependency Management

As a developer composing multiple services,
I want services to share common dependencies efficiently,
So that I avoid duplicate resource allocation and ensure consistency.

**Acceptance Criteria:**

**Given** multiple services requiring the same dependencies (database, config, etc.)
**When** I configure shared dependency management
**Then** services automatically share singleton dependencies

**And** dependency lifecycle is managed centrally (creation, cleanup)
**And** shared dependencies support concurrent access safely
**And** dependency updates propagate to all dependent services
**And** memory usage is optimized through dependency sharing
**And** dependency health monitoring affects all dependent services

**Prerequisites:** Story 3.2 (service embedding patterns established)

**Technical Notes:**
- Implement dependency scoping in DI container: Singleton, Scoped, Transient
- Add dependency health checking and automatic failover
- Implement dependency update notifications for dependent services
- Support dependency proxying for thread safety
- Add dependency usage tracking and monitoring
- Ensure proper cleanup order during shutdown (dependents before dependencies)

### Story 3.4: Create Service Composition Utilities

As a developer building service hierarchies,
I want utility functions for common service composition patterns,
So that I can build complex service graphs quickly and reliably.

**Acceptance Criteria:**

**Given** the service embedding and dependency sharing capabilities
**When** I use service composition utilities
**Then** common composition patterns are available as helper functions

**And** utilities include: ServiceChain, ServiceProxy, ServiceBranch, ServiceMerger
**And** composition utilities preserve type safety and interface contracts
**And** error handling flows correctly through composed service chains
**And** performance characteristics are documented for each composition pattern
**And** utilities support both static (compile-time) and dynamic (runtime) composition

**Prerequisites:** Story 3.3 (shared dependency management available)

**Technical Notes:**
- ServiceChain: Sequential execution through multiple services
- ServiceProxy: Transparent forwarding with potential transformation
- ServiceBranch: Conditional routing to different services
- ServiceMerger: Combine results from multiple services
- Add composition validation (type compatibility, circular dependency detection)
- Support async composition patterns for I/O-bound operations

### Story 3.5: Implement Service Lifecycle Management

As a framework user,
I want composed services to have proper lifecycle management,
So that resources are allocated and cleaned up correctly in service hierarchies.

**Acceptance Criteria:**

**Given** complex service compositions with embedded services and shared dependencies
**When** I start and stop the service hierarchy
**Then** lifecycle events (start, stop, health check) propagate correctly through the hierarchy

**And** services start in correct dependency order (dependencies before dependents)
**And** services stop in reverse dependency order (dependents before dependencies)
**And** health checks aggregate status from all composed services
**And** partial failures are handled gracefully (continue with available services)
**And** lifecycle hooks allow custom initialization and cleanup logic

**Prerequisites:** Story 3.4 (composition utilities available)

**Technical Notes:**
- Add ServiceLifecycleInterface with Start(), Stop(), HealthCheck() methods
- Implement dependency-aware startup/shutdown ordering
- Support graceful degradation when embedded services fail
- Add lifecycle event broadcasting for monitoring and logging
- Implement service recovery patterns for transient failures
- Support hot-swapping of services in composition hierarchy (advanced feature)

---

## Epic 4: Developer Experience & Tooling

**Goal:** Excellent developer experience with clear documentation and debugging tools, enabling developers to quickly adopt and optimize the new architecture.

### Story 4.1: Create Comprehensive Developer Documentation

As a developer learning the new architecture,
I want detailed documentation and examples for all framework patterns,
So that I can understand and implement dependency injection and service composition effectively.

**Acceptance Criteria:**

**Given** the complete new architecture with DI and service composition
**When** I create comprehensive developer documentation
**Then** complete developer guide covers all framework capabilities

**And** before/after code examples show dependency injection patterns clearly
**And** documentation covers testing strategies with mocked dependencies
**And** step-by-step tutorials guide developers through common scenarios
**And** API documentation is automatically generated and always up-to-date
**And** code examples are runnable and tested as part of CI/CD

**Prerequisites:** Story 3.5 (new architecture fully implemented)

**Technical Notes:**
- Create `/docs/developer-guide.md` with comprehensive examples
- Include testing patterns: unit tests, integration tests, composition testing
- Document performance characteristics and optimization tips
- Add tutorial series: "Building Your First Service", "Advanced Composition", etc.
- Ensure all code examples compile and pass tests
- Include common patterns library with copy-paste examples

### Story 4.2: Implement Enhanced Error Messages and Debugging

As a developer working with dependency injection and service composition,
I want clear error messages when something goes wrong,
So that I can quickly identify and fix configuration issues.

**Acceptance Criteria:**

**Given** the new dependency injection and service composition features
**When** configuration errors or dependency issues occur
**Then** error messages clearly identify the problem and suggest solutions

**And** dependency injection failures show complete dependency chain with clear resolution path
**And** service composition errors identify which service in the hierarchy failed
**And** runtime diagnostics help debug performance and lifecycle issues
**And** debug mode provides detailed tracing of dependency resolution
**And** error messages include code examples showing correct usage

**Prerequisites:** Story 4.1 (documentation available to reference in error messages)

**Technical Notes:**
- Implement structured error types with context: `DependencyError`, `CompositionError`, etc.
- Add dependency graph visualization for debugging: `endor debug dependencies`
- Include stack traces with service composition context
- Add runtime profiling hooks for performance debugging
- Implement health check endpoints that show service composition status
- Create error message templates with actionable suggestions and code examples

### Story 4.3: Create Testing Framework and Utilities

As a developer building services with the framework,
I want comprehensive testing utilities and patterns,
So that I can write effective tests for dependency-injected and composed services.

**Acceptance Criteria:**

**Given** the interface-based architecture and test utilities from Epic 1
**When** I enhance testing capabilities with advanced patterns
**Then** testing framework supports complex composition testing scenarios

**And** test builders make it easy to create test service hierarchies
**And** assertion helpers validate dependency injection configurations
**And** performance testing utilities measure composition overhead
**And** integration testing supports realistic multi-service scenarios
**And** test fixtures provide realistic data for common testing scenarios

**Prerequisites:** Story 4.2 (debugging tools available for test troubleshooting)

**Technical Notes:**
- Extend `sdk/testutils` with composition testing utilities
- Add test builders: `NewTestServiceHierarchy()`, `WithMockedDependencies()`
- Create assertion helpers: `AssertDependencyInjected()`, `AssertCompositionValid()`
- Add benchmarking utilities for performance testing
- Support test scenarios: single service, composed services, middleware chains
- Include test data generators for realistic testing scenarios

### Story 4.4: Implement Development Tools and CLI

As a developer working with the framework,
I want development tools that help me build and debug services efficiently,
So that I can be productive and avoid common mistakes.

**Acceptance Criteria:**

**Given** the complete framework with documentation and testing utilities
**When** I create development tools for common tasks
**Then** CLI tools assist with service creation and validation

**And** service generators create boilerplate for common service patterns
**And** dependency graph validator detects configuration issues early
**And** performance profiler identifies bottlenecks in service composition
**And** code generators create test scaffolding automatically
**And** IDE integration provides autocomplete and error detection

**Prerequisites:** Story 4.3 (testing framework provides foundation for validation tools)

**Technical Notes:**
- Create `endor-cli` tool with subcommands: generate, validate, profile
- Service generator: `endor generate service --type=hybrid --name=UserService`
- Dependency validator: `endor validate dependencies --config=services.yaml`
- Performance profiler: `endor profile --duration=30s --output=report.html`
- Add IDE extensions for popular editors (VS Code, GoLand)
- Support project templates for quick setup

### Story 4.5: Validate Performance and Create Benchmarks

As a framework maintainer and user,
I want to ensure the new architecture delivers excellent performance,
So that dependency injection and composition don't impact application speed.

**Acceptance Criteria:**

**Given** the complete new architecture with DI and service composition
**When** I run comprehensive performance benchmarks
**Then** performance characteristics are documented and optimized

**And** dependency injection overhead is minimal in production scenarios
**And** service composition adds negligible latency to request processing
**And** memory usage is optimized through efficient dependency sharing
**And** benchmark suite covers realistic usage patterns and load scenarios
**And** performance regression tests prevent future performance degradation
**And** optimization guide helps developers write performant services

**Prerequisites:** Story 4.4 (profiling tools available for performance analysis)

**Technical Notes:**
- Create comprehensive benchmark suite covering all framework features
- Test scenarios: simple services, complex composition, high concurrency
- Measure: request latency, memory usage, CPU usage, dependency resolution time
- Add continuous benchmarking to CI/CD pipeline
- Create performance dashboard showing trends over time
- Include load testing with realistic service hierarchies
- Document performance best practices and optimization strategies

---

## FR Coverage Matrix

**Epic 1: Interface Foundation & Testability**
- FR4: All framework components implement interfaces for mocking → Story 1.1, 1.2, 1.3
- FR5: Unit testing without MongoDB dependencies → Story 1.2, 1.4, 1.5
- FR6: Test utility package with pre-built mocks → Story 1.4
- FR7: Test service composition patterns in isolation → Story 1.4, 1.5
- FR8: Integration tests with in-memory implementations → Story 1.5

**Epic 2: Dependency Injection Architecture**
- FR1: Interface-based EndorService creation with DI → Story 2.2
- FR2: Interface-based EndorHybridService creation with DI → Story 2.3
- FR9: Service constructors accept interface parameters → Story 2.2, 2.3, 2.4
- FR10: Eliminate hard-coded singleton dependencies → Story 2.1, 2.4, 2.5
- FR11: Inject custom implementations for any dependency → Story 2.1, 2.5
- FR12: Support constructor injection and factory patterns → Story 2.1, 2.5

**Epic 3: Service Composition Framework**
- FR3: Compose services within services using DI patterns → Story 3.2, 3.4
- FR17: EndorHybridService embeds EndorServices through DI → Story 3.2
- FR18: Share dependencies without implementation coupling → Story 3.3
- FR19: Middleware pattern for cross-cutting concerns → Story 3.1
- FR20: Proper service lifecycle management → Story 3.5

**Epic 4: Developer Experience & Migration**
- FR13: Existing EndorService implementations work with minimal changes → Story 4.1
- FR14: EndorHybridService migration through deprecation cycle → Story 4.1
- FR15: Current API contracts remain stable during transition → Story 4.1
- FR16: Migration tools assist with architecture updates → Story 4.2
- FR21: Clear error messages for DI failures → Story 4.4
- FR22: Complete testing pattern documentation → Story 4.3
- FR23: Migration guide for common scenarios → Story 4.3
- FR24: Maintain current performance characteristics → Story 4.5
- FR25: API documentation continues working seamlessly → Story 4.1
- FR26: Database access through repository interfaces → Story 2.4
- FR27: Schema generation remains automatic → Story 2.3
- FR28: Custom repository implementations for testing → Story 2.4
- FR29: Connection management becomes dependency-injectable → Story 2.4
- FR30: Preserve and enhance generic type safety → Story 2.2, 2.3
- FR31: JSON schema generation with DI services → Story 2.3
- FR32: Automatic input validation with extension points → Story 2.2, 2.3
- FR33: Maintain current error handling patterns → Story 2.2, 2.3, 4.4

---

## Summary

**✅ Epic Breakdown Complete - endor-sdk-go Architectural Refactoring**

**4 Epics, 20 Stories, 29 Functional Requirements Covered**

**Epic Progression:**
1. **Interface Foundation** → Developers can write tests
2. **Dependency Injection** → Developers can customize and compose
3. **Service Composition** → Developers can build complex hierarchies  
4. **Developer Experience** → Developers can adopt and optimize effectively

**Value Delivery Strategy:**
- Each epic delivers immediate user value
- No technical layer epics - all user-focused capabilities
- Sequential execution enables smooth adoption
- No backward compatibility constraints - clean architecture implementation

**Context Incorporated:**
- ✅ PRD requirements (29 of 33 FRs covered - removed 4 migration FRs)
- ℹ️ Focused epic structure without migration complexity
- **Next:** Run Architecture workflow for technical implementation details
- **Note:** Epics will be enhanced with architectural context later

**Breaking Changes Accepted:**
Since the framework is not yet in production use, we can implement the new architecture without backward compatibility concerns, resulting in a cleaner, more maintainable design.
