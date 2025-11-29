# Epic Technical Specification: Service Composition Framework

Date: 2025-11-29
Author: BMad
Epic ID: 3
Status: Draft

---

## Overview

Epic 3 establishes the **Service Composition Framework** that enables developers to build complex service hierarchies using zero-boilerplate dependency injection patterns. This epic builds upon the interface foundation (Epic 1) and dependency injection architecture (Epic 2) to provide powerful composition capabilities while maintaining the dual-service pattern's flexibility.

The Service Composition Framework introduces middleware pipelines, service embedding patterns, shared dependency management, and advanced composition utilities that transform endor-sdk-go from individual service creation to complex service orchestration. This capability is unique in the Go ecosystem and positions the framework as a powerful tool for building scalable, maintainable microservice architectures.

## Objectives and Scope

**In Scope:**
- Middleware pipeline system for cross-cutting concerns (logging, metrics, authentication)
- EndorService embedding within EndorHybridService with method delegation
- Shared dependency management with singleton, scoped, and transient lifecycles
- Service composition utilities: ServiceChain, ServiceProxy, ServiceBranch, ServiceMerger
- Complete service lifecycle management with dependency-aware startup/shutdown ordering
- Performance-optimized composition patterns with zero-copy where possible

**Out of Scope:**
- Distributed service orchestration (single-process composition only)
- Event-driven service communication (synchronous composition focus)
- Service mesh integration (API gateway integration remains in core framework)
- Dynamic service discovery (services are composed at application startup)
- Backward compatibility with existing non-DI service implementations

## System Architecture Alignment

The Service Composition Framework aligns with the architectural decisions defined in the Architecture document:

**Decorator Pattern with Middleware Pipeline:** Implements Decision 3's middleware approach using chain-of-responsibility pattern for cross-cutting concerns while maintaining Go's composition principles.

**Lightweight Custom DI Container:** Leverages Epic 2's DI container to provide shared dependency management with singleton, scoped, and transient lifecycles as specified in Decision 1.

**Interface Segregation with Smart Composition:** Uses Epic 1's granular interfaces to enable type-safe service composition while maintaining the ability to compose larger interfaces from smaller ones per Decision 2.

**Service Lifecycle Management:** Implements dependency-aware ordering that respects the DI container's dependency graph, ensuring proper resource allocation and cleanup across complex service hierarchies.

This epic transforms the endor-sdk-go framework from a service creation tool into a service composition platform, enabling the advanced patterns described in FR17-FR20 while maintaining all performance and type safety characteristics.

## Detailed Design

### Services and Modules

| Service/Module | Responsibility | Inputs | Outputs | Owner |
|----------------|----------------|---------|---------|-------|
| **MiddlewareInterface** | Cross-cutting concern execution | HTTP context, next handler | Modified context, response | Story 3.1 |
| **ServiceEmbedding** | EndorService delegation within EndorHybridService | Service interface, prefix | Delegated method calls | Story 3.2 |
| **DependencyManager** | Shared singleton/scoped dependency lifecycle | Interface type, scope config | Managed dependency instances | Story 3.3 |
| **CompositionUtilities** | Helper functions for service patterns | Service interfaces, composition config | ServiceChain, ServiceProxy, ServiceBranch, ServiceMerger | Story 3.4 |
| **LifecycleManager** | Dependency-aware service startup/shutdown | Service dependency graph | Ordered start/stop operations | Story 3.5 |

### Data Models and Contracts

```go
// Middleware execution contract
type MiddlewareInterface interface {
    Before(ctx *EndorContext) error
    After(ctx *EndorContext, response *Response) error
}

// Service embedding contract
type EmbeddableService interface {
    EmbedService(prefix string, service EndorServiceInterface) error
    GetEmbeddedServices() map[string]EndorServiceInterface
}

// Dependency lifecycle scoping
type DependencyScope int
const (
    Singleton DependencyScope = iota  // Shared across all services
    Scoped                           // Shared within composition hierarchy
    Transient                        // New instance per resolution
)

// Service composition patterns
type ServiceComposition struct {
    Pattern CompositionPattern    // Chain, Proxy, Branch, Merger
    Services []ServiceInterface   // Services in composition
    Config   CompositionConfig   // Pattern-specific configuration
}

// Lifecycle state management
type ServiceState int
const (
    Created ServiceState = iota
    Starting
    Running
    Stopping
    Stopped
    Failed
)
```

### APIs and Interfaces

```go
// Middleware Pipeline API
type MiddlewarePipeline interface {
    AddMiddleware(middleware MiddlewareInterface) MiddlewarePipeline
    Execute(ctx *EndorContext, handler EndorHandlerFunc) (*Response, error)
}

// Service Embedding API
func (h *EndorHybridService) EmbedService(prefix string, service EndorServiceInterface) error
func (h *EndorHybridService) GetEmbeddedService(prefix string) (EndorServiceInterface, error)

// Shared Dependency Management API
func (c *Container) RegisterScoped[T any](impl T, scope DependencyScope) error
func (c *Container) ResolveShared[T any]() (T, error)

// Composition Utilities API
func ServiceChain(services ...EndorServiceInterface) ChainedService
func ServiceProxy(target EndorServiceInterface, interceptor ProxyInterceptor) ProxiedService
func ServiceBranch(router BranchRouter, services map[string]EndorServiceInterface) BranchedService
func ServiceMerger(services []EndorServiceInterface, merger ResultMerger) MergedService

// Lifecycle Management API
type ServiceLifecycle interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck() error
    GetState() ServiceState
}
```

### Workflows and Sequencing

**1. Middleware Pipeline Execution:**
```
Request → Auth Middleware → Logging Middleware → Metrics Middleware → Service Handler → Response
         ↓ Before()       ↓ Before()          ↓ Before()
         ↑ After()        ↑ After()           ↑ After()
```

**2. Service Embedding Flow:**
```
HybridService.HandleRequest() 
  → Check embedded services for method match
  → Delegate to embedded service with prefix context
  → Merge results with hybrid service response
  → Return unified response
```

**3. Shared Dependency Resolution:**
```
Service A requests Repository → Container checks scope
  → Singleton exists? Return existing instance
  → Scoped exists in hierarchy? Return scoped instance  
  → Create new instance, store per scope, return
```

**4. Service Lifecycle Startup:**
```
1. Build dependency graph from service registrations
2. Topological sort for dependency order
3. Start services in dependency order (dependencies first)
4. Validate all services reached Running state
5. Begin health monitoring loop
```

## Non-Functional Requirements

### Performance

- **Middleware Pipeline Overhead:** < 10μs per middleware execution (target: 5μs average)
- **Service Embedding Latency:** < 15μs additional overhead for delegated method calls
- **Dependency Resolution:** < 1μs for singleton dependencies, < 5μs for scoped dependencies
- **Service Composition:** ServiceChain < 20μs, ServiceProxy < 10μs, ServiceBranch < 25μs per routing decision
- **Lifecycle Operations:** Service startup < 100ms for dependency graphs up to 50 services
- **Memory Usage:** Composition overhead < 512 bytes per embedded service, < 1KB per middleware
- **Concurrent Access:** Shared dependencies must support 1000+ concurrent resolutions with zero contention

### Security

- **Middleware Security:** Authentication middleware must validate session tokens and propagate security context through pipeline
- **Service Isolation:** Embedded services cannot access parent service's private dependencies without explicit injection
- **Dependency Injection Security:** Container prevents injection of unauthorized implementations through interface validation
- **Context Propagation:** Security tokens and user context must flow correctly through service composition hierarchies
- **Audit Logging:** All service embedding operations and middleware executions must be logged with request correlation IDs
- **Error Information:** Error messages must not leak sensitive dependency configuration or service topology details

### Reliability/Availability

- **Circuit Breaking:** Service composition must include circuit breaker patterns for embedded service failures
- **Graceful Degradation:** When embedded services fail, parent services should continue operating with reduced functionality
- **Health Aggregation:** Composite service health checks must accurately reflect the health of all embedded services
- **Recovery Patterns:** Failed services in compositions should support automatic restart with exponential backoff
- **Dependency Availability:** 99.9% availability target for shared singleton dependencies across service restarts
- **State Consistency:** Service lifecycle operations must be atomic - no partial startup/shutdown states

### Observability

- **Middleware Metrics:** Execution time, success rate, and error count per middleware type with Prometheus integration
- **Composition Tracing:** Request flow through service compositions with OpenTelemetry-compatible trace spans
- **Dependency Monitoring:** Health status, resolution count, and memory usage for all managed dependencies
- **Lifecycle Events:** Structured logging for service state transitions with correlation to dependency graph changes
- **Performance Monitoring:** Real-time dashboards for composition overhead, embedding performance, and shared resource usage
- **Error Tracking:** Detailed error attribution through service composition hierarchies with actionable debugging information

## Dependencies and Integrations

**Internal Framework Dependencies:**
- **Epic 1 Interfaces:** MiddlewareInterface, EndorServiceInterface, EndorHybridServiceInterface, RepositoryInterface
- **Epic 2 DI Container:** Container registration/resolution APIs, lifecycle management, dependency graph validation
- **Core EndorService:** Method delegation patterns, handler function signatures, context propagation
- **Core EndorHybridService:** ToEndorService() integration, category handling, schema generation preservation

**External Dependencies:**
- **Go Standard Library:** context package for cancellation, sync package for concurrent access patterns
- **Gin Framework:** HTTP middleware integration, context extension for service composition
- **Prometheus:** Metrics collection for middleware execution, composition performance, dependency health
- **MongoDB Driver:** Repository interface implementations, connection pooling through dependency injection
- **testify/mock:** Mock implementations for all composition interfaces in testutils package

**Integration Constraints:**
- Middleware must be compatible with existing Gin middleware chain
- Service composition must preserve existing EndorService API contracts
- Shared dependencies must work with existing MongoDB connection management
- Lifecycle management must integrate with existing health check endpoints (/readyz, /livez)

## Acceptance Criteria (Authoritative)

1. **Middleware Pipeline (Story 3.1):** Services can be wrapped with multiple middleware components in configurable order
2. **Cross-Cutting Concerns:** Built-in middleware for authentication, logging, and metrics are provided and functional
3. **Service Embedding (Story 3.2):** EndorHybridService can embed EndorService instances with method delegation
4. **Method Resolution:** Embedded service methods are accessible with clear precedence rules for conflicts
5. **Shared Dependencies (Story 3.3):** Multiple services share singleton dependencies efficiently without resource duplication
6. **Dependency Scoping:** Container supports Singleton, Scoped, and Transient dependency lifecycles
7. **Composition Utilities (Story 3.4):** ServiceChain, ServiceProxy, ServiceBranch, and ServiceMerger utilities are implemented and tested
8. **Type Safety:** All composition patterns maintain compile-time type safety and interface contracts
9. **Lifecycle Management (Story 3.5):** Services start and stop in correct dependency order with health aggregation
10. **Performance Requirements:** All composition overhead targets are met or exceeded
11. **Error Handling:** Errors propagate correctly through service composition hierarchies
12. **Testing Support:** All composition patterns can be unit tested with mocked dependencies

## Traceability Mapping

| Acceptance Criteria | Spec Section | Component/API | Test Strategy |
|---------------------|--------------|---------------|---------------|
| AC1: Middleware Pipeline | APIs - MiddlewarePipeline | MiddlewareInterface.AddMiddleware() | Unit test with mock middleware chain |
| AC2: Cross-Cutting Concerns | Services - MiddlewareInterface | AuthMiddleware, LoggingMiddleware, MetricsMiddleware | Integration test with real auth tokens |
| AC3: Service Embedding | APIs - EmbedService() | EndorHybridService.EmbedService() | Unit test with mock EndorService |
| AC4: Method Resolution | Workflows - Service Embedding Flow | Method delegation logic | Unit test with conflicting method names |
| AC5: Shared Dependencies | APIs - RegisterScoped() | Container.ResolveShared() | Unit test with multiple service instances |
| AC6: Dependency Scoping | Data Models - DependencyScope | Singleton/Scoped/Transient logic | Unit test with lifecycle verification |
| AC7: Composition Utilities | APIs - ServiceChain/Proxy/Branch/Merger | CompositionUtilities package | Unit test each utility pattern |
| AC8: Type Safety | All APIs | Interface constraint validation | Compile-time verification tests |
| AC9: Lifecycle Management | APIs - ServiceLifecycle | Start/Stop dependency ordering | Integration test with complex dependency graph |
| AC10: Performance | NFRs - Performance | Middleware/embedding overhead | Benchmark tests with performance regression detection |
| AC11: Error Handling | Workflows - All flows | Error propagation patterns | Unit test with simulated failures |
| AC12: Testing Support | All components | testutils mock implementations | Meta-test: testing the testing utilities |

## Risks, Assumptions, Open Questions

**RISKS:**
- **R1:** Middleware execution overhead could accumulate significantly in deep middleware chains → Mitigation: Performance benchmarks and middleware count limits
- **R2:** Service embedding complexity could introduce circular dependency issues → Mitigation: Static analysis validation in DI container
- **R3:** Shared dependency management might create memory leaks in long-running services → Mitigation: Explicit cleanup protocols and memory monitoring
- **R4:** Complex service composition could make debugging difficult → Mitigation: Enhanced error messages with composition trace information

**ASSUMPTIONS:**
- **A1:** Developers will follow dependency injection patterns rather than global singleton access
- **A2:** Service composition will primarily be configured at startup, not dynamically at runtime
- **A3:** Performance overhead of composition abstractions is acceptable for the flexibility gained
- **A4:** Existing EndorService implementations can be refactored to support embedding without breaking changes

**OPEN QUESTIONS:**
- **Q1:** Should middleware support async execution patterns for I/O-bound operations? → **Next Step:** Research async middleware patterns in Go
- **Q2:** What is the maximum reasonable depth for service composition hierarchies? → **Next Step:** Performance testing with deep compositions
- **Q3:** How should we handle version compatibility when embedding services with different dependency versions? → **Next Step:** Dependency version conflict resolution strategy
- **Q4:** Should composition utilities support runtime modification of service graphs? → **Next Step:** Evaluate use cases for dynamic composition

## Test Strategy Summary

**Unit Testing Strategy:**
- **Middleware Testing:** Each middleware component tested in isolation with mock contexts and handlers
- **Service Embedding:** EndorService embedding tested with mock services to verify method delegation
- **Dependency Management:** DI container scoping tested with mock dependencies across different lifecycle scenarios
- **Composition Utilities:** Each utility pattern tested independently with mock services and known inputs/outputs
- **Lifecycle Management:** Service startup/shutdown ordering tested with mock dependency graphs

**Integration Testing Strategy:**
- **End-to-End Composition:** Full service hierarchies tested with real HTTP requests through middleware pipeline
- **Database Integration:** Shared repository dependencies tested with real MongoDB connections
- **Performance Integration:** Composition overhead validated under realistic load conditions
- **Health Check Integration:** Composite service health aggregation tested with partial service failures

**Test Framework:**
- **Build Tags:** `//go:build unit` for fast isolated tests, `//go:build integration` for slower database tests
- **Test Utilities:** Enhanced testutils package with composition test builders and middleware mocks
- **Benchmarking:** Continuous performance regression testing for all composition patterns
- **Coverage Target:** 90%+ code coverage across all composition components

**Test Scenarios:**
- Happy path: Successful service composition with all dependencies healthy
- Error handling: Service failures, middleware errors, dependency unavailability
- Edge cases: Circular dependencies, deep composition hierarchies, resource exhaustion
- Performance: High-load scenarios, memory leak detection, latency validation
