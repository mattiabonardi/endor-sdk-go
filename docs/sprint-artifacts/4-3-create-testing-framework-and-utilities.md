# Story 4.3: Create Testing Framework and Utilities

Status: ready-for-dev

## Story

As a **Go developer building services with the endor-sdk-go framework**,
I want **comprehensive testing utilities and patterns for dependency-injected and composed services**,
so that **I can write effective tests for complex service hierarchies and composition scenarios with confidence**.

## Acceptance Criteria

1. **Test Builder Framework**: Testing framework provides builder patterns for creating complex test scenarios:
   - `NewTestServiceHierarchy()` creates realistic multi-service test setups
   - `WithMockedDependencies()` allows selective mocking of specific dependencies
   - `WithTestData()` provides realistic data fixtures for comprehensive testing
   - Builders support both simple and complex dependency injection scenarios

2. **Composition Testing Utilities**: Framework includes specialized utilities for testing service composition:
   - `AssertDependencyInjected()` validates that dependencies are correctly wired
   - `AssertCompositionValid()` verifies service hierarchies are properly constructed
   - `TestServiceLifecycle()` validates initialization and cleanup patterns
   - Support for testing middleware chains and cross-cutting concerns

3. **Performance Testing Integration**: Testing utilities include performance measurement capabilities:
   - Benchmarking utilities measure composition overhead
   - Memory usage tracking for dependency injection patterns
   - Latency measurement for service interaction patterns
   - Performance regression detection for complex service hierarchies

4. **Integration Testing Support**: Framework provides realistic multi-service testing scenarios:
   - Test containers for realistic database interactions
   - Multi-service orchestration for end-to-end testing
   - Network simulation for distributed service testing
   - Environment configuration management for test scenarios

5. **Test Data Management**: Comprehensive test fixture system for realistic testing:
   - Schema-compliant test data generators
   - Relationship-aware test data creation (services, resources, categories)
   - Test data cleanup and isolation between test runs
   - Export/import utilities for shared test scenarios

## Tasks / Subtasks

- [ ] Task 1: Implement Test Builder Framework (AC: 1)
  - [ ] Create `TestServiceBuilder` with fluent API for service hierarchy creation
  - [ ] Implement dependency mocking selectors (`WithMockedRepo()`, `WithMockedConfig()`)
  - [ ] Add realistic test data builders for common service patterns
  - [ ] Create test scenario templates for single service, composed services, and complex hierarchies

- [ ] Task 2: Build Composition Testing Utilities (AC: 2)
  - [ ] Implement dependency injection assertion helpers
  - [ ] Create service composition validation utilities
  - [ ] Add lifecycle testing utilities (initialization, cleanup, error scenarios)
  - [ ] Build middleware chain testing helpers

- [ ] Task 3: Add Performance Testing Capabilities (AC: 3)
  - [ ] Create benchmarking utilities for dependency injection overhead
  - [ ] Implement memory usage tracking for service composition patterns
  - [ ] Add latency measurement utilities for service interactions
  - [ ] Build performance regression testing framework

- [ ] Task 4: Implement Integration Testing Support (AC: 4)
  - [ ] Add test container utilities for realistic database testing
  - [ ] Create multi-service test orchestration framework
  - [ ] Implement environment configuration management for testing
  - [ ] Add network simulation utilities for distributed testing scenarios

- [ ] Task 5: Create Test Data Management System (AC: 5)
  - [ ] Build schema-compliant test data generators
  - [ ] Implement relationship-aware test data creation utilities
  - [ ] Add test data cleanup and isolation mechanisms
  - [ ] Create export/import utilities for shared test scenarios

- [ ] Task 6: Testing and Documentation (All ACs)
  - [ ] Write comprehensive test suite for all testing utilities
  - [ ] Create usage examples for each testing pattern
  - [ ] Add performance benchmarks for testing framework overhead
  - [ ] Document best practices for testing complex service hierarchies

## Dev Notes

This story extends the foundation built in Epic 1 (interface-based testing) and Epic 2 (dependency injection) to create a comprehensive testing framework specifically designed for the unique challenges of testing dependency-injected and composed services.

### Key Architecture Considerations

- **Testing Framework Integration**: Built on top of the existing `sdk/testutils` package established in Story 1.4
- **Performance Consciousness**: Testing utilities should have minimal overhead and not impact production builds
- **Realistic Test Scenarios**: Focus on real-world service composition patterns that developers actually encounter
- **Integration with DI Container**: Testing utilities should work seamlessly with the dependency injection container from Epic 2

### Learnings from Previous Story

**From Story 4-2 (Enhanced Error Messages and Debugging)**

- **New Error System Available**: `sdk/errors/` package provides comprehensive error classification that testing framework should integrate with for error scenario testing
- **Debug Mode Integration**: Testing framework should leverage the debug mode capabilities for test troubleshooting
- **Performance Metrics**: Can use the metrics collection system to measure testing framework overhead
- **Error Scenario Testing**: Use `sdk/testutils/error_scenarios.go` for testing error conditions in composed services

**Testing Setup Integration**:
- Reuse error classification system for testing error scenarios
- Integrate with debug logging for test troubleshooting
- Use validation framework for test configuration validation
- Build on contextual error reporting for better test failure messages

[Source: stories/4-2-implement-enhanced-error-messages-and-debugging.md#Dev-Agent-Record]

### Integration with Framework Architecture

**From Epic 1 (Interface Foundation)**:
- Build upon `MockEndorService`, `MockEndorHybridService`, `MockRepository` implementations
- Extend existing test utilities with composition-specific patterns
- Use established interface mocking patterns as foundation

**From Epic 2 (Dependency Injection)**:
- Testing framework must work with DI container for realistic test scenarios
- Support testing of constructor injection patterns
- Validate dependency resolution in test environments

**From Epic 3 (Service Composition)**:
- Focus on testing service embedding patterns
- Test middleware pipeline functionality
- Validate shared dependency management in test scenarios

### Project Structure Notes

**Testing Framework Organization**:
```
sdk/
├── testutils/
│   ├── builders/
│   │   ├── service_builder.go      # TestServiceBuilder implementation
│   │   ├── hierarchy_builder.go    # Complex service hierarchy builders
│   │   └── data_builder.go         # Test data creation utilities
│   ├── assertions/
│   │   ├── composition_assertions.go   # Composition validation utilities
│   │   ├── di_assertions.go           # Dependency injection validation
│   │   └── lifecycle_assertions.go    # Service lifecycle validation
│   ├── performance/
│   │   ├── benchmarks.go           # Performance testing utilities
│   │   ├── memory_tracker.go       # Memory usage measurement
│   │   └── regression_detector.go  # Performance regression detection
│   ├── integration/
│   │   ├── test_containers.go      # Database container utilities
│   │   ├── orchestration.go        # Multi-service test orchestration
│   │   └── environment.go          # Test environment management
│   └── fixtures/
│       ├── data_generators.go      # Schema-compliant test data
│       ├── cleanup.go              # Test isolation utilities
│       └── export_import.go        # Test scenario sharing
```

**Testing Patterns**:
```
test/
├── composition/
│   ├── service_embedding_test.go   # Test service composition patterns
│   ├── middleware_chains_test.go   # Test middleware integration
│   └── complex_hierarchies_test.go # Test deep service hierarchies
├── performance/
│   ├── di_overhead_test.go         # DI performance benchmarks
│   ├── composition_latency_test.go # Service composition performance
│   └── memory_usage_test.go        # Memory usage validation
└── integration/
    ├── multi_service_test.go       # End-to-end multi-service tests
    ├── database_integration_test.go # Database interaction testing
    └── error_scenarios_test.go     # Comprehensive error testing
```

### Technical Implementation Notes

**Framework Integration Points**:
- Extends Epic 1's mock utilities with composition-specific patterns
- Uses Epic 2's DI container for realistic dependency injection testing
- Leverages Epic 3's service composition patterns for test scenarios
- Integrates with Story 4.2's error system for comprehensive error testing

**Performance Considerations**:
- Testing framework should not impact production builds (build tags: `//go:build testing`)
- Benchmarking utilities should have minimal overhead
- Memory tracking should be accurate without affecting test performance
- Support both unit testing (mocked dependencies) and integration testing (real dependencies)

### References

- [Source: docs/prd.md#Enhanced-Testability] - FR5-8: Unit testing without dependencies, test utilities, composition testing
- [Source: docs/architecture.md#Testing-Strategy] - Comprehensive testing approach for dependency-injected services
- [Source: docs/epics.md#Story-4-3] - Testing framework requirements and acceptance criteria
- [Source: stories/1-4-create-test-utility-package-with-mocks.md] - Foundation testing utilities to extend
- [Source: stories/4-2-implement-enhanced-error-messages-and-debugging.md] - Error system integration points

## Dev Agent Record

### Context Reference

- [Story Context XML](./4-3-create-testing-framework-and-utilities.context.xml)

### Agent Model Used

Claude-3-5-Sonnet

### Debug Log References

### Completion Notes List

### File List
