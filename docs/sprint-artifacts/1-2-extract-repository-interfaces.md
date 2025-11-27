# Story 1.2: extract-repository-interfaces

Status: ready-for-dev

## Story

As a developer using the framework,
I want repository operations to be interface-based,
So that I can mock database interactions in my unit tests.

## Acceptance Criteria

1. **Interface Definition** - RepositoryInterface, ResourceRepositoryInterface, and MongoRepositoryInterface are defined in `sdk/interfaces/repository.go`
2. **Method Abstraction** - All current repository methods are abstracted into interfaces with domain-focused signatures  
3. **Implementation Compliance** - MongoDB repository implementations satisfy their respective interfaces
4. **Generic Type Safety** - Repository interfaces support generic types for compile-time type safety
5. **Domain Error Handling** - Interface methods return domain errors rather than database-specific errors

## Tasks / Subtasks

- [ ] Task 1: Analyze existing repository implementations and extract interface contracts (AC: 1, 2)
  - [ ] Review `sdk/endor_resource_repository.go` and related repository files
  - [ ] Identify all public methods used by EndorService and EndorHybridService
  - [ ] Extract common CRUD patterns: Create, Read, Update, Delete, List, Query operations
  - [ ] Design generic-friendly interfaces with type parameters for resource types
- [ ] Task 2: Create repository interface definitions (AC: 1, 4, 5)
  - [ ] Create `sdk/interfaces/repository.go` file
  - [ ] Define RepositoryInterface with core CRUD operations
  - [ ] Define ResourceRepositoryInterface with endor-specific resource operations
  - [ ] Define MongoRepositoryInterface with MongoDB-specific optimizations
  - [ ] Add comprehensive GoDoc documentation with usage examples
- [ ] Task 3: Implement interface compliance and testing (AC: 3)
  - [ ] Add interface compliance tests to `sdk/interfaces_test.go`
  - [ ] Create unit tests demonstrating repository interface usage patterns
  - [ ] Verify existing repository implementations satisfy new interfaces
  - [ ] Test generic type safety with concrete resource types
- [ ] Task 4: Domain error abstraction (AC: 5)
  - [ ] Review current error handling in repository implementations
  - [ ] Define domain error types that abstract away MongoDB specifics
  - [ ] Update interface signatures to use domain errors
  - [ ] Ensure error semantics remain clear for framework users

## Dev Notes

**Architectural Context:**
- Builds on Story 1.1's interface foundation for EndorService and EndorHybridService
- Follows Interface Segregation with Smart Composition pattern from architecture document
- Implements "accept interfaces, return structs" Go philosophy for repository layer
- Enables Epic 1's goal of testability by making database interactions mockable

**Framework Requirements:**
- Repository interfaces must support both static and dynamic resource patterns used by EndorService and EndorHybridService
- Generic type safety preserves compile-time validation of resource types
- Domain error abstraction prevents tests from depending on MongoDB-specific error types
- Interface design enables future alternative database implementations

**Testing Strategy:**
- Interface compliance tests ensure repository implementations satisfy contracts
- Unit tests demonstrate mocking patterns for database-dependent business logic
- Generic type safety tests validate compile-time type checking
- Error handling tests verify domain error abstraction works correctly

### Learnings from Previous Story

**From Story 1-1-extract-core-service-interfaces (Status: review)**

- **Interface Package Created**: `sdk/interfaces/service.go` established - extend with `repository.go`
- **Testing Patterns**: Interface compliance tests using `var _ Interface = (*Implementation)(nil)` - reuse this pattern
- **Documentation Standards**: Comprehensive GoDoc with usage examples - follow same standard
- **No Breaking Changes**: All existing APIs preserved during interface extraction - maintain this approach
- **Architecture Foundation**: Interface extraction foundation established for dependency injection patterns

[Source: docs/sprint-artifacts/1-1-extract-core-service-interfaces.md#Dev-Agent-Record]

### Project Structure Notes

**Interface Package Extension:**
- Extend existing `sdk/interfaces/` package with `repository.go`
- Follow same patterns established in `sdk/interfaces/service.go`
- Add repository interface tests to existing `sdk/interfaces_test.go`

**Repository Implementation Locations:**
- Current implementations: `sdk/endor_resource_repository.go`, `sdk/mongo_resource_instance_repository.go`, `sdk/mongo_static_resource_instance_repository.go`
- Interface compliance will be verified against existing implementations
- No structural changes to existing repository files required

**Architecture Alignment:**
- Supports Epic 1 goal: "Framework components become mockable and testable"
- Establishes repository abstraction needed for Epic 2's dependency injection
- Enables Epic 1 Story 1.4's test utility package to provide repository mocks

### References

- [Source: docs/epics.md#Story 1.2: Extract Repository Interfaces]
- [Source: docs/architecture.md#Decision 2: Interface Granularity]
- [Source: docs/prd.md#Enhanced Testability FR5, FR26, FR28]
- [Source: docs/sprint-artifacts/1-1-extract-core-service-interfaces.md#Completion Notes List]

## Dev Agent Record

### Context Reference

- [docs/sprint-artifacts/1-2-extract-repository-interfaces.context.xml](./1-2-extract-repository-interfaces.context.xml) - Complete story context with documentation artifacts, code analysis, interfaces, constraints, and testing guidance

### Agent Model Used

<!-- Agent model will be filled when story is implemented -->

### Debug Log References

### Completion Notes List

### File List
