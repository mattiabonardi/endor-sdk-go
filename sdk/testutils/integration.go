package testutils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InMemoryRepository provides an in-memory implementation of repository interfaces
// for integration testing without external dependencies
type InMemoryRepository[T any] struct {
	mu    sync.RWMutex
	data  map[string]T
	idGen func() string
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository[T any]() *InMemoryRepository[T] {
	return &InMemoryRepository[T]{
		data: make(map[string]T),
		idGen: func() string {
			return primitive.NewObjectID().Hex()
		},
	}
}

// Create implements the repository Create method
func (r *InMemoryRepository[T]) Create(ctx context.Context, item T) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.idGen()
	r.data[id] = item
	return id, nil
}

// FindByID implements the repository FindByID method
func (r *InMemoryRepository[T]) FindByID(ctx context.Context, id string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var zero T
	item, exists := r.data[id]
	if !exists {
		return zero, fmt.Errorf("item with id %s not found", id)
	}
	return item, nil
}

// FindAll implements the repository FindAll method
func (r *InMemoryRepository[T]) FindAll(ctx context.Context, filter bson.M) ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []T
	for _, item := range r.data {
		// Simple filter matching for testing - in real implementation this would be more sophisticated
		if r.matchesFilter(item, filter) {
			results = append(results, item)
		}
	}
	return results, nil
}

// Update implements the repository Update method
func (r *InMemoryRepository[T]) Update(ctx context.Context, id string, item T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return fmt.Errorf("item with id %s not found", id)
	}
	r.data[id] = item
	return nil
}

// Delete implements the repository Delete method
func (r *InMemoryRepository[T]) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return fmt.Errorf("item with id %s not found", id)
	}
	delete(r.data, id)
	return nil
}

// Count returns the number of items in the repository
func (r *InMemoryRepository[T]) Count(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return int64(len(r.data)), nil
}

// Clear removes all items from the repository
func (r *InMemoryRepository[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data = make(map[string]T)
}

// matchesFilter performs basic filter matching for testing
func (r *InMemoryRepository[T]) matchesFilter(item T, filter bson.M) bool {
	// For testing purposes, we'll match if filter is empty
	// In a real implementation, this would perform actual filter matching
	return len(filter) == 0
}

// IntegrationTestDatabase provides utilities for MongoDB integration testing
type IntegrationTestDatabase struct {
	ConnectionString string
	DatabaseName     string
	CollectionPrefix string
	CleanupAfter     bool
	collections      []string
	mu               sync.Mutex
}

// NewIntegrationTestDatabase creates a new integration test database helper
func NewIntegrationTestDatabase(databaseName string) *IntegrationTestDatabase {
	return &IntegrationTestDatabase{
		ConnectionString: "mongodb://localhost:27017",
		DatabaseName:     databaseName,
		CollectionPrefix: "test_",
		CleanupAfter:     true,
		collections:      make([]string, 0),
	}
}

// CreateTestCollection creates a collection for testing
func (db *IntegrationTestDatabase) CreateTestCollection(name string) string {
	db.mu.Lock()
	defer db.mu.Unlock()

	collectionName := db.CollectionPrefix + name
	db.collections = append(db.collections, collectionName)
	return collectionName
}

// GetCollectionName returns the full collection name with prefix
func (db *IntegrationTestDatabase) GetCollectionName(name string) string {
	return db.CollectionPrefix + name
}

// Cleanup removes all test collections if CleanupAfter is true
func (db *IntegrationTestDatabase) Cleanup() error {
	if !db.CleanupAfter {
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// In a real implementation, this would connect to MongoDB and drop collections
	// For now, we'll just clear the collection list
	db.collections = make([]string, 0)
	return nil
}

// TestLifecycleManager handles the lifecycle of test services and resources
type TestLifecycleManager struct {
	services       map[string]interfaces.EndorServiceInterface
	hybridServices map[string]interfaces.EndorHybridServiceInterface
	repositories   map[string]interface{}
	database       *IntegrationTestDatabase
	mu             sync.Mutex
}

// NewTestLifecycleManager creates a new test lifecycle manager
func NewTestLifecycleManager() *TestLifecycleManager {
	return &TestLifecycleManager{
		services:       make(map[string]interfaces.EndorServiceInterface),
		hybridServices: make(map[string]interfaces.EndorHybridServiceInterface),
		repositories:   make(map[string]interface{}),
	}
}

// RegisterService registers a service for lifecycle management
func (lm *TestLifecycleManager) RegisterService(name string, service interfaces.EndorServiceInterface) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.services[name] = service
}

// RegisterHybridService registers a hybrid service for lifecycle management
func (lm *TestLifecycleManager) RegisterHybridService(name string, service interfaces.EndorHybridServiceInterface) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.hybridServices[name] = service
}

// RegisterRepository registers a repository for lifecycle management
func (lm *TestLifecycleManager) RegisterRepository(name string, repo interface{}) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.repositories[name] = repo
}

// SetDatabase sets the test database
func (lm *TestLifecycleManager) SetDatabase(db *IntegrationTestDatabase) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.database = db
}

// GetService retrieves a registered service
func (lm *TestLifecycleManager) GetService(name string) (interfaces.EndorServiceInterface, bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	service, exists := lm.services[name]
	return service, exists
}

// GetHybridService retrieves a registered hybrid service
func (lm *TestLifecycleManager) GetHybridService(name string) (interfaces.EndorHybridServiceInterface, bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	service, exists := lm.hybridServices[name]
	return service, exists
}

// GetRepository retrieves a registered repository
func (lm *TestLifecycleManager) GetRepository(name string) (interface{}, bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	repo, exists := lm.repositories[name]
	return repo, exists
}

// StartAll initializes all registered services
func (lm *TestLifecycleManager) StartAll() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Validate all services
	for name, service := range lm.services {
		if err := service.Validate(); err != nil {
			return fmt.Errorf("failed to validate service %s: %w", name, err)
		}
	}

	for name, service := range lm.hybridServices {
		if err := service.Validate(); err != nil {
			return fmt.Errorf("failed to validate hybrid service %s: %w", name, err)
		}
	}

	return nil
}

// StopAll cleans up all registered services and repositories
func (lm *TestLifecycleManager) StopAll() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Clear all in-memory repositories
	for _, repo := range lm.repositories {
		// Use type assertion to check if it's an InMemoryRepository
		if clearable, ok := repo.(interface{ Clear() }); ok {
			clearable.Clear()
		}
	}

	// Cleanup database if present
	if lm.database != nil {
		if err := lm.database.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup database: %w", err)
		}
	}

	return nil
}

// IntegrationTestSuite provides a complete test suite for integration testing
type IntegrationTestSuite struct {
	Manager       *TestLifecycleManager
	Database      *IntegrationTestDatabase
	Config        interfaces.ConfigProviderInterface
	Timeout       time.Duration
	SetupTasks    []func() error
	TeardownTasks []func() error
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(suiteName string) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		Manager:  NewTestLifecycleManager(),
		Database: NewIntegrationTestDatabase("test_" + suiteName),
		Config: NewTestConfigProvider().
			WithDocumentDBUri("mongodb://localhost:27017/test_" + suiteName).
			WithServerPort("0"). // Use random port for testing
			Build(),
		Timeout:       30 * time.Second,
		SetupTasks:    make([]func() error, 0),
		TeardownTasks: make([]func() error, 0),
	}
}

// AddSetupTask adds a task to be executed during suite setup
func (suite *IntegrationTestSuite) AddSetupTask(task func() error) {
	suite.SetupTasks = append(suite.SetupTasks, task)
}

// AddTeardownTask adds a task to be executed during suite teardown
func (suite *IntegrationTestSuite) AddTeardownTask(task func() error) {
	suite.TeardownTasks = append(suite.TeardownTasks, task)
}

// Setup initializes the test suite
func (suite *IntegrationTestSuite) Setup() error {
	// Set database in manager
	suite.Manager.SetDatabase(suite.Database)

	// Execute setup tasks
	for i, task := range suite.SetupTasks {
		if err := task(); err != nil {
			return fmt.Errorf("setup task %d failed: %w", i, err)
		}
	}

	// Start all services
	return suite.Manager.StartAll()
}

// Teardown cleans up the test suite
func (suite *IntegrationTestSuite) Teardown() error {
	var errors []error

	// Stop all services and repositories
	if err := suite.Manager.StopAll(); err != nil {
		errors = append(errors, err)
	}

	// Execute teardown tasks
	for i, task := range suite.TeardownTasks {
		if err := task(); err != nil {
			errors = append(errors, fmt.Errorf("teardown task %d failed: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("teardown completed with errors: %v", errors)
	}

	return nil
}

// WithTimeout sets the suite timeout
func (suite *IntegrationTestSuite) WithTimeout(timeout time.Duration) *IntegrationTestSuite {
	suite.Timeout = timeout
	return suite
}

// WithCleanupDisabled disables automatic cleanup (useful for debugging)
func (suite *IntegrationTestSuite) WithCleanupDisabled() *IntegrationTestSuite {
	suite.Database.CleanupAfter = false
	return suite
}
