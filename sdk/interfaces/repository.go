// Package interfaces provides repository interfaces for endor-sdk-go framework.
// These interfaces enable dependency injection, testing with mocks, and service composition
// while maintaining the framework's type-safe repository patterns.
//
// This package contains the extracted repository interfaces that establish contracts
// for dependency injection and testing without requiring changes to existing implementations.
//
// Phase 1: Domain Error Interfaces and Base Contracts
// This initial implementation focuses on domain error abstraction and interface foundations
// to enable testing patterns while avoiding import cycles with existing implementations.
package interfaces

import "context"

// Base interface contracts that define repository behavior patterns

// ResourceInstanceInterface defines the contract that all resource instances must implement.
// This matches the existing interface definition in the framework.
type ResourceInstanceInterface interface {
	GetID() *string
	SetID(id string)
}

// ResourceInstanceSpecializedInterface defines the contract for category-specialized resources.
// This matches the existing interface definition in the framework.
type ResourceInstanceSpecializedInterface interface {
	GetCategoryType() *string
	SetCategoryType(categoryType string)
}

// Domain Error Interfaces - These provide domain error abstraction (Acceptance Criteria 5)
// These abstract away database-specific error details and provide consistent error handling

// RepositoryError defines the base interface for all repository-related errors.
// This enables consistent error handling and testing patterns across implementations.
//
// This interface abstracts MongoDB-specific errors (like mongo.ErrNoDocuments) into
// domain-appropriate error types that can be easily tested and handled.
type RepositoryError interface {
	error
	// ErrorCode returns a domain-specific error code for programmatic handling
	ErrorCode() string
	// ErrorMessage returns a human-readable error message
	ErrorMessage() string
}

// NotFoundError indicates that a requested resource was not found.
// This abstracts database-specific "no documents" errors into domain language.
//
// Instead of mongo.ErrNoDocuments, repositories should return implementations
// of this interface to enable proper testing and error handling.
type NotFoundError interface {
	RepositoryError
	// ResourceType returns the type of resource that was not found (e.g., "User", "Product")
	ResourceType() string
	// ResourceID returns the ID that was not found
	ResourceID() string
}

// ValidationError indicates that data validation failed.
// This abstracts database constraint violations and input validation failures.
//
// Instead of MongoDB validation errors or constraint violations,
// repositories should return implementations of this interface.
type ValidationError interface {
	RepositoryError
	// ValidationDetails returns specific validation failure information
	// Key is the field name, value is the validation error message
	ValidationDetails() map[string]string
}

// ConflictError indicates that a create or update operation failed due to conflicts.
// This abstracts database unique constraint violations and concurrent modification issues.
//
// Instead of MongoDB duplicate key errors or optimistic concurrency failures,
// repositories should return implementations of this interface.
type ConflictError interface {
	RepositoryError
	// ConflictType returns the type of conflict (e.g., "duplicate_id", "concurrent_modification")
	ConflictType() string
}

// InternalError indicates an internal repository error that should not be exposed to clients.
// This abstracts database connection errors, query failures, and other infrastructure issues.
//
// Instead of MongoDB connection errors or query syntax errors,
// repositories should return implementations of this interface.
type InternalError interface {
	RepositoryError
	// ShouldRetry indicates whether the operation can be safely retried
	ShouldRetry() bool
}

// Repository Interface Patterns - Phase 1
// These define the high-level contracts that repositories should implement
// Specific method signatures will be resolved through existing implementations

// RepositoryPattern defines the general contract for CRUD repository operations.
// This interface establishes the pattern that all repositories should follow
// without defining specific method signatures that would cause import cycles.
//
// Concrete implementations in the sdk package implement the actual method signatures
// while following this pattern for testability and dependency injection.
type RepositoryPattern interface {
	// Pattern: Instance retrieval with domain error handling
	// Implementations should return NotFoundError for missing resources

	// Pattern: List operations with filtering support
	// Implementations should return ValidationError for malformed filters

	// Pattern: Create operations with conflict detection
	// Implementations should return ConflictError for duplicate resources

	// Pattern: Delete operations with proper error handling
	// Implementations should return NotFoundError for missing resources

	// Pattern: Update operations with validation and conflict detection
	// Implementations should return NotFoundError, ValidationError, or ConflictError as appropriate
}

// TestableRepository defines the interface for repository implementations that support testing.
// This interface enables dependency injection and mock creation for unit testing.
//
// Repositories implementing this pattern can be easily mocked and tested without
// requiring actual database connections or external dependencies.
type TestableRepository interface {
	RepositoryPattern
	// Pattern: Validation support for testing
	// Implementations should provide validation methods that can be tested independently

	// Pattern: Error simulation for testing
	// Implementations should allow error injection for testing error handling paths
}

// Configuration interfaces for repository initialization and testing

// RepositoryOptions defines configuration options that repositories should support.
// This enables consistent configuration patterns across different repository implementations.
type RepositoryOptions interface {
	// GetAutoGenerateID returns whether IDs should be auto-generated
	GetAutoGenerateID() bool
	// SetAutoGenerateID configures ID generation behavior
	SetAutoGenerateID(auto bool)
}

// MockableRepository defines the interface for repositories that can be easily mocked.
// This enables comprehensive unit testing without database dependencies.
type MockableRepository interface {
	TestableRepository
	// Pattern: Mock behavior configuration
	// Implementations should support configurable responses for testing scenarios

	// Pattern: Call verification
	// Implementations should enable verification of method calls and arguments

	// Pattern: Error injection
	// Implementations should allow injection of specific error types for testing
}

// Documentation for interface usage patterns:
//
// Basic Repository Pattern:
//   - Implement domain error interfaces instead of exposing database errors
//   - Follow consistent CRUD operation patterns
//   - Support dependency injection through interface types
//
// Testing Pattern:
//   - Create mock implementations that satisfy repository interfaces
//   - Use domain error interfaces for predictable error testing
//   - Inject repository dependencies through interface types
//
// Error Handling Pattern:
//   - Return NotFoundError instead of mongo.ErrNoDocuments
//   - Return ValidationError instead of database constraint violations
//   - Return ConflictError instead of duplicate key errors
//   - Return InternalError for infrastructure failures
//
// This approach enables:
//   1. Interface-based dependency injection
//   2. Comprehensive unit testing with mocks
//   3. Domain-appropriate error handling
//   4. Consistent patterns across repository implementations
//   5. Testing without external database dependencies

// Core repository interface for dependency injection and testing

// RepositoryInterface defines the core contract for all repository implementations.
// This interface enables dependency injection, testing with mocks, and service composition
// while maintaining the framework's automatic CRUD capabilities.
//
// Acceptance Criteria 2: Repository implementations satisfy RepositoryInterface from
// interfaces package and can be mocked in tests.
type RepositoryInterface interface {
	// Create inserts a new resource into the repository
	Create(ctx context.Context, resource any) error
	// Read retrieves a resource by ID and populates the result
	Read(ctx context.Context, id string, result any) error
	// Update modifies an existing resource in the repository
	Update(ctx context.Context, resource any) error
	// Delete removes a resource from the repository by ID
	Delete(ctx context.Context, id string) error
	// List retrieves multiple resources matching the filter criteria
	List(ctx context.Context, filter map[string]any, results any) error
}

// Repository factory function types for dependency injection patterns

// RepositoryFactoryFunc defines the signature for repository factory functions.
// This enables both direct construction and container-based resolution patterns.
//
// Acceptance Criteria 5: Repository Factory Patterns support both NewRepositoryWithClient()
// direct construction and NewRepositoryFromContainer() for DI container resolution.
type RepositoryFactoryFunc func(deps RepositoryDependencies) (RepositoryInterface, error)

// RepositoryContainerFactoryFunc defines the signature for container-based repository factories.
// This enables automatic dependency resolution from DI containers.
//
// Acceptance Criteria 8: Container integration enables repository resolution by interface type.
type RepositoryContainerFactoryFunc func(container DIContainerInterface) (RepositoryInterface, error)

// DIContainerInterface defines the minimum interface required for dependency resolution.
// This is a subset of the full DI container interface to avoid import cycles.
type DIContainerInterface interface {
	// ResolveType resolves a dependency by interface type
	ResolveType(interfaceType interface{}) (interface{}, error)
}
