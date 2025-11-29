// Package sdk provides repository factory functions for dependency injection.
// These factories enable both direct construction and container-based resolution
// patterns for repository instances.
package sdk

import (
	"context"
	"reflect"

	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// RegisterRepositoryFactories registers repository factory functions with the DI container.
// This enables automatic repository resolution and shared instance management.
//
// Acceptance Criteria 8: Container Integration - repository implementations register
// with DI container and can be resolved by interface type.
func RegisterRepositoryFactories(container di.Container) error {
	// Register database client factory
	err := di.RegisterFactory[interfaces.DatabaseClientInterface](container, func(c di.Container) (interfaces.DatabaseClientInterface, error) {
		return DefaultDatabaseClient()
	}, di.Singleton)
	if err != nil {
		return err
	}

	// Register repository factory using the factory pattern
	err = di.RegisterFactory[interfaces.RepositoryInterface](container, func(c di.Container) (interfaces.RepositoryInterface, error) {
		return NewRepositoryFromContainer(c)
	}, di.Singleton)
	if err != nil {
		return err
	}

	return nil
}

// NewRepositoryFromContainer creates a repository by resolving dependencies from DI container.
// This is the primary container integration factory following the established pattern.
//
// Acceptance Criteria 5: Support NewRepositoryFromContainer() for DI container resolution.
func NewRepositoryFromContainer(container di.Container) (interfaces.RepositoryInterface, error) {
	// Resolve database client
	dbClientType := reflect.TypeOf((*interfaces.DatabaseClientInterface)(nil)).Elem()
	dbClientInterface, err := container.ResolveType(dbClientType)
	if err != nil {
		return nil, err
	}

	dbClient, ok := dbClientInterface.(interfaces.DatabaseClientInterface)
	if !ok {
		return nil, interfaces.NewDatabaseRepositoryError(
			interfaces.RepositoryErrorCodeInvalidDependencies,
			"Resolved database client is not of correct type",
			nil,
		)
	}

	// Resolve config provider
	configType := reflect.TypeOf((*interfaces.ConfigProviderInterface)(nil)).Elem()
	configInterface, err := container.ResolveType(configType)
	if err != nil {
		return nil, err
	}

	config, ok := configInterface.(interfaces.ConfigProviderInterface)
	if !ok {
		return nil, interfaces.NewDatabaseRepositoryError(
			interfaces.RepositoryErrorCodeInvalidDependencies,
			"Resolved config is not of correct type",
			nil,
		)
	}

	// Resolve logger (optional)
	loggerType := reflect.TypeOf((*interfaces.LoggerInterface)(nil)).Elem()
	loggerInterface, err := container.ResolveType(loggerType)
	var logger interfaces.LoggerInterface
	if err != nil {
		// Logger is optional - use default if not available
		logger = NewDefaultLogger()
	} else {
		logger, ok = loggerInterface.(interfaces.LoggerInterface)
		if !ok {
			logger = NewDefaultLogger()
		}
	}

	// Create dependencies struct
	deps := interfaces.RepositoryDependencies{
		DatabaseClient: dbClient,
		Config:         config,
		Logger:         logger,
		MicroServiceID: "default", // TODO: This should be injected or configured
	}

	// Create repository using the dependency injection constructor
	return NewEndorRepositoryWithDependencies(deps)
}

// NewRepositoryWithClient creates a repository with explicit dependency injection.
// This is the direct construction pattern for scenarios where container resolution is not needed.
//
// Acceptance Criteria 5: Repository Factory Patterns support both NewRepositoryWithClient()
// direct construction and NewRepositoryFromContainer() for DI container resolution.
func NewRepositoryWithClient(client interfaces.DatabaseClientInterface, config interfaces.ConfigProviderInterface, microServiceID string) (interfaces.RepositoryInterface, error) {
	deps := interfaces.RepositoryDependencies{
		DatabaseClient: client,
		Config:         config,
		Logger:         NewDefaultLogger(),
		MicroServiceID: microServiceID,
	}

	return NewEndorRepositoryWithDependencies(deps)
}

// NewEndorRepositoryWithDependencies creates a repository implementation that satisfies RepositoryInterface.
// This serves as an adapter between the EndorServiceRepository and the RepositoryInterface contract.
//
// Acceptance Criteria 2: Repository implementations satisfy RepositoryInterface from
// interfaces package and can be mocked in tests.
func NewEndorRepositoryWithDependencies(deps interfaces.RepositoryDependencies) (interfaces.RepositoryInterface, error) {
	// Create the concrete EndorServiceRepository with dependencies
	repo, err := NewEndorServiceRepositoryWithDependencies(deps, nil, nil)
	if err != nil {
		return nil, err
	}

	// Return adapter that implements RepositoryInterface
	return &EndorRepositoryAdapter{
		endorRepo:    repo,
		dependencies: deps,
	}, nil
}

// EndorRepositoryAdapter adapts EndorServiceRepository to implement RepositoryInterface.
// This enables the EndorServiceRepository to be used through the standard repository interface.
type EndorRepositoryAdapter struct {
	endorRepo    *EndorServiceRepository
	dependencies interfaces.RepositoryDependencies
}

// Create implements the RepositoryInterface.Create method.
func (r *EndorRepositoryAdapter) Create(ctx context.Context, resource any) error {
	// Convert to the format expected by EndorServiceRepository
	if dto, ok := resource.(CreateDTO[Resource]); ok {
		return r.endorRepo.Create(dto)
	}

	return interfaces.NewDatabaseRepositoryError(
		interfaces.RepositoryErrorCodeOperationFailed,
		"Unsupported resource type for Create operation",
		nil,
	)
}

// Read implements the RepositoryInterface.Read method.
func (r *EndorRepositoryAdapter) Read(ctx context.Context, id string, result any) error {
	dto := ReadInstanceDTO{Id: id}
	resource, err := r.endorRepo.Instance(dto)
	if err != nil {
		return err
	}

	// TODO: Implement proper result mapping
	// This would need to properly marshal/unmarshal the resource into the result
	_ = resource
	return nil
}

// Update implements the RepositoryInterface.Update method.
func (r *EndorRepositoryAdapter) Update(ctx context.Context, resource any) error {
	if dto, ok := resource.(UpdateByIdDTO[Resource]); ok {
		_, err := r.endorRepo.UpdateOne(dto)
		return err
	}

	return interfaces.NewDatabaseRepositoryError(
		interfaces.RepositoryErrorCodeOperationFailed,
		"Unsupported resource type for Update operation",
		nil,
	)
}

// Delete implements the RepositoryInterface.Delete method.
func (r *EndorRepositoryAdapter) Delete(ctx context.Context, id string) error {
	dto := ReadInstanceDTO{Id: id}
	return r.endorRepo.DeleteOne(dto)
} // List implements the RepositoryInterface.List method.
func (r *EndorRepositoryAdapter) List(ctx context.Context, filter map[string]any, results any) error {
	// TODO: Implement proper list operation
	// This would need to convert the filter to the appropriate format
	// and properly handle the results mapping
	return interfaces.NewDatabaseRepositoryError(
		interfaces.RepositoryErrorCodeOperationFailed,
		"List operation not yet implemented in adapter",
		nil,
	)
}
