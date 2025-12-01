# Comprehensive Testing Strategies for endor-sdk-go

This guide provides complete testing strategies for the endor-sdk-go framework, covering unit testing with mock dependencies, integration testing patterns, service composition testing, and lifecycle testing patterns developed across the framework's evolution.

---

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Unit Testing with Mocks](#unit-testing-with-mocks)
3. [Integration Testing Patterns](#integration-testing-patterns)
4. [Service Composition Testing](#service-composition-testing)
5. [Lifecycle Testing Patterns](#lifecycle-testing-patterns)
6. [Performance and Benchmark Testing](#performance-and-benchmark-testing)
7. [Test Organization and Build Tags](#test-organization-and-build-tags)
8. [Continuous Integration Patterns](#continuous-integration-patterns)

---

## Testing Philosophy

The endor-sdk-go framework embraces a **layered testing strategy** that mirrors the dependency injection and service composition architecture:

### Testing Pyramid

```
    ┌─────────────────┐
    │  E2E Tests      │ ← Few, realistic scenarios
    │     (5%)        │
    ├─────────────────┤
    │ Integration     │ ← Component interactions
    │ Tests (25%)     │
    ├─────────────────┤
    │  Unit Tests     │ ← Business logic isolated
    │     (70%)       │
    └─────────────────┘
```

### Core Testing Principles

1. **Fast Feedback**: Unit tests run in milliseconds using mocked dependencies
2. **Realistic Integration**: Integration tests use real implementations with test infrastructure
3. **Composition Validation**: Service composition tests verify hierarchical service behavior
4. **Performance Regression**: Benchmark tests prevent performance degradation
5. **Lifecycle Coverage**: Lifecycle tests validate startup, runtime, and shutdown behavior

---

## Unit Testing with Mocks

Unit tests isolate business logic by mocking all external dependencies using the `sdk/testutils/` package.

### Basic Service Unit Testing

```go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
)

//go:build unit

func TestUserService_CreateUser_Success(t *testing.T) {
    // Arrange: Set up mocked dependencies
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    mockValidator := testutils.NewMockValidator()
    
    user := User{
        ID:    "user-123",
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // Configure mock expectations
    mockValidator.On("ValidateUser", user).Return(nil)
    mockRepo.On("Create", mock.Any, user).Return(nil)
    mockLogger.On("Info", "Creating user", mock.MatchedBy(func(fields map[string]interface{}) bool {
        return fields["user_id"] == "user-123"
    })).Return()
    mockLogger.On("Info", "User created successfully", mock.Any).Return()
    
    service := NewUserService(mockRepo, mockLogger, mockValidator)
    
    // Act: Execute the business logic
    err := service.CreateUser(context.Background(), user)
    
    // Assert: Verify behavior and mock interactions
    assert.NoError(t, err)
    mockValidator.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}

func TestUserService_CreateUser_ValidationError(t *testing.T) {
    // Test error handling with mocked dependencies
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    mockValidator := testutils.NewMockValidator()
    
    user := User{Name: "John Doe", Email: "invalid-email"} // Invalid email
    validationError := errors.New("invalid email format")
    
    // Mock validation failure
    mockValidator.On("ValidateUser", user).Return(validationError)
    mockLogger.On("Error", "User validation failed", validationError, mock.Any).Return()
    
    service := NewUserService(mockRepo, mockLogger, mockValidator)
    
    // Verify error handling
    err := service.CreateUser(context.Background(), user)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validation failed")
    mockValidator.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
    
    // Repository should not be called on validation failure
    mockRepo.AssertNotCalled(t, "Create")
}

func TestUserService_CreateUser_RepositoryError(t *testing.T) {
    // Test database errors with mocked dependencies
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    mockValidator := testutils.NewMockValidator()
    
    user := User{ID: "user-123", Name: "John Doe", Email: "john@example.com"}
    repositoryError := errors.New("database connection failed")
    
    mockValidator.On("ValidateUser", user).Return(nil)
    mockRepo.On("Create", mock.Any, user).Return(repositoryError)
    mockLogger.On("Info", "Creating user", mock.Any).Return()
    mockLogger.On("Error", "Failed to create user", repositoryError, mock.Any).Return()
    
    service := NewUserService(mockRepo, mockLogger, mockValidator)
    
    err := service.CreateUser(context.Background(), user)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to create user")
    assert.Contains(t, err.Error(), "database connection failed")
}
```

### Testing with Mock Builders

Use the test utilities builders for complex mock setups:

```go
func TestUserService_ComplexScenario(t *testing.T) {
    // Use builders for complex mock setup
    mockService := testutils.NewTestEndorServiceBuilder().
        WithResource("users").
        WithMockRepository(func(repo *testutils.MockRepository) {
            repo.On("Create", mock.Any, mock.Any).Return(nil)
            repo.On("FindByEmail", mock.Any, "john@example.com", mock.Any).Return(nil)
        }).
        WithMockLogger(func(logger *testutils.MockLogger) {
            logger.On("Info", mock.Any, mock.Any).Return()
        }).
        Build()
    
    // Test complex business logic
    result, err := mockService.ProcessUserRegistration(context.Background(), registrationData)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockService.AssertAllExpectations(t)
}
```

### Table-Driven Testing for Multiple Scenarios

```go
func TestUserService_CreateUser_MultipleScenarios(t *testing.T) {
    tests := []struct {
        name           string
        user           User
        validationErr  error
        repositoryErr  error
        expectedError  string
        shouldCallRepo bool
    }{
        {
            name: "Success",
            user: User{ID: "1", Name: "John", Email: "john@example.com"},
            expectedError: "",
            shouldCallRepo: true,
        },
        {
            name: "ValidationError", 
            user: User{Name: "John", Email: "invalid"},
            validationErr: errors.New("invalid email"),
            expectedError: "validation failed",
            shouldCallRepo: false,
        },
        {
            name: "RepositoryError",
            user: User{ID: "1", Name: "John", Email: "john@example.com"},
            repositoryErr: errors.New("database error"),
            expectedError: "failed to create user",
            shouldCallRepo: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set up mocks for each test case
            mockRepo := testutils.NewMockRepository()
            mockLogger := testutils.NewMockLogger()
            mockValidator := testutils.NewMockValidator()
            
            mockValidator.On("ValidateUser", tt.user).Return(tt.validationErr)
            mockLogger.On("Info", mock.Any, mock.Any).Return()
            mockLogger.On("Error", mock.Any, mock.Any, mock.Any).Return()
            
            if tt.shouldCallRepo {
                mockRepo.On("Create", mock.Any, tt.user).Return(tt.repositoryErr)
            }
            
            service := NewUserService(mockRepo, mockLogger, mockValidator)
            
            err := service.CreateUser(context.Background(), tt.user)
            
            if tt.expectedError == "" {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            }
            
            if !tt.shouldCallRepo {
                mockRepo.AssertNotCalled(t, "Create")
            }
        })
    }
}
```

---

## Integration Testing Patterns

Integration tests validate component interactions using real implementations with test infrastructure.

### Database Integration Testing

```go
//go:build integration

func TestUserService_Integration_DatabaseOperations(t *testing.T) {
    // Set up test database
    testDB := setupTestDatabase(t)
    defer cleanupTestDatabase(testDB)
    
    // Use real implementations with test database
    mongoRepo := NewMongoRepository(testDB)
    testLogger := &TestLogger{t: t}
    prodValidator := &ProductionValidator{}
    
    service := NewUserService(mongoRepo, testLogger, prodValidator)
    
    // Test with real database operations
    user := User{
        ID:    uuid.New().String(),
        Name:  "Integration Test User",
        Email: "integration@test.com",
    }
    
    // Create user
    err := service.CreateUser(context.Background(), user)
    require.NoError(t, err)
    
    // Verify user was persisted
    retrievedUser, err := service.GetUser(context.Background(), user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Name, retrievedUser.Name)
    assert.Equal(t, user.Email, retrievedUser.Email)
    
    // Test update
    user.Name = "Updated Name"
    err = service.UpdateUser(context.Background(), user)
    require.NoError(t, err)
    
    updatedUser, err := service.GetUser(context.Background(), user.ID)
    require.NoError(t, err)
    assert.Equal(t, "Updated Name", updatedUser.Name)
    
    // Test deletion
    err = service.DeleteUser(context.Background(), user.ID)
    require.NoError(t, err)
    
    _, err = service.GetUser(context.Background(), user.ID)
    assert.Error(t, err) // Should not find deleted user
}

// Test helper functions for integration tests
func setupTestDatabase(t *testing.T) *mongo.Database {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(getTestMongoURI()))
    require.NoError(t, err)
    
    // Use unique database per test to avoid conflicts
    dbName := fmt.Sprintf("test_%s_%d", t.Name(), time.Now().Unix())
    db := client.Database(dbName)
    
    // Create indices and test data if needed
    return db
}

func cleanupTestDatabase(db *mongo.Database) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    db.Drop(ctx)
}
```

### HTTP Integration Testing

```go
//go:build integration

func TestUserHTTPHandler_Integration(t *testing.T) {
    // Set up test server with real dependencies
    container := di.NewContainer()
    setupIntegrationDependencies(container, t)
    
    userService, err := di.Resolve[interfaces.UserServiceInterface](container)
    require.NoError(t, err)
    
    // Create test server
    router := gin.New()
    userHandler := NewUserHandler(userService)
    router.POST("/users", userHandler.CreateUser)
    router.GET("/users/:id", userHandler.GetUser)
    
    server := httptest.NewServer(router)
    defer server.Close()
    
    // Test user creation via HTTP
    userData := `{"name":"HTTP Test User","email":"http@test.com"}`
    resp, err := http.Post(server.URL+"/users", "application/json", strings.NewReader(userData))
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    var createdUser User
    err = json.NewDecoder(resp.Body).Decode(&createdUser)
    require.NoError(t, err)
    assert.Equal(t, "HTTP Test User", createdUser.Name)
    assert.NotEmpty(t, createdUser.ID)
    
    // Test user retrieval via HTTP
    resp, err = http.Get(server.URL + "/users/" + createdUser.ID)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var retrievedUser User
    err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
    require.NoError(t, err)
    assert.Equal(t, createdUser.ID, retrievedUser.ID)
    assert.Equal(t, createdUser.Name, retrievedUser.Name)
}
```

---

## Service Composition Testing

Test hierarchical service behavior and service embedding patterns developed in Epic 3.

### Testing ServiceChain Composition

```go
//go:build unit

func TestServiceChain_UserRegistrationWorkflow(t *testing.T) {
    // Set up mock services for the chain
    mockValidation := testutils.NewMockValidationService()
    mockUserService := testutils.NewMockUserService()
    mockNotification := testutils.NewMockNotificationService()
    
    registrationData := UserRegistrationData{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "secure123",
    }
    validatedData := ValidatedUserData{Name: "Test User", Email: "test@example.com"}
    createdUser := User{ID: "user-123", Name: "Test User", Email: "test@example.com"}
    
    // Set up chain expectations
    mockValidation.On("Execute", mock.Any, registrationData).Return(validatedData, nil)
    mockUserService.On("Execute", mock.Any, validatedData).Return(createdUser, nil)
    mockNotification.On("Execute", mock.Any, createdUser).Return(nil, nil)
    
    // Create and test service chain
    registrationChain := composition.ServiceChain(
        mockValidation,
        mockUserService,
        mockNotification,
    ).WithConfig(composition.CompositionConfig{
        Timeout:  30 * time.Second,
        FailFast: true,
    })
    
    result, err := registrationChain.Execute(context.Background(), registrationData)
    
    assert.NoError(t, err)
    assert.Equal(t, createdUser, result)
    mockValidation.AssertExpectations(t)
    mockUserService.AssertExpectations(t)
    mockNotification.AssertExpectations(t)
}

func TestServiceChain_ErrorHandling(t *testing.T) {
    // Test chain error propagation
    mockValidation := testutils.NewMockValidationService()
    mockUserService := testutils.NewMockUserService()
    mockNotification := testutils.NewMockNotificationService()
    
    registrationData := UserRegistrationData{Email: "invalid-email"}
    validationError := errors.New("invalid email format")
    
    // Validation fails, chain should stop
    mockValidation.On("Execute", mock.Any, registrationData).Return(nil, validationError)
    
    chain := composition.ServiceChain(mockValidation, mockUserService, mockNotification)
    
    result, err := chain.Execute(context.Background(), registrationData)
    
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "invalid email format")
    
    // Subsequent services should not be called
    mockUserService.AssertNotCalled(t, "Execute")
    mockNotification.AssertNotCalled(t, "Execute")
}
```

### Testing ServiceBranch Parallel Execution

```go
func TestServiceBranch_ParallelNotifications(t *testing.T) {
    // Test parallel service execution
    mockEmail := testutils.NewMockEmailService()
    mockSMS := testutils.NewMockSMSService()
    mockPush := testutils.NewMockPushService()
    
    notification := NotificationData{
        UserID:  "user-123",
        Message: "Welcome to our service!",
    }
    
    // All services should execute in parallel
    mockEmail.On("Execute", mock.Any, notification).Return("email-sent", nil).After(10 * time.Millisecond)
    mockSMS.On("Execute", mock.Any, notification).Return("sms-sent", nil).After(15 * time.Millisecond)
    mockPush.On("Execute", mock.Any, notification).Return("push-sent", nil).After(5 * time.Millisecond)
    
    branch := composition.ServiceBranch(mockEmail, mockSMS, mockPush).
        WithConfig(composition.CompositionConfig{
            Timeout:    1 * time.Second,
            RequireAll: false, // Allow partial success
        })
    
    start := time.Now()
    results, err := branch.Execute(context.Background(), notification)
    duration := time.Since(start)
    
    assert.NoError(t, err)
    assert.Len(t, results, 3)
    
    // Should take roughly the time of the slowest service (15ms), not sum of all (30ms)
    assert.Less(t, duration, 25*time.Millisecond)
    assert.Greater(t, duration, 14*time.Millisecond)
    
    mockEmail.AssertExpectations(t)
    mockSMS.AssertExpectations(t)
    mockPush.AssertExpectations(t)
}
```

### Testing Service Embedding in EndorHybridService

```go
func TestEndorHybridService_ServiceEmbedding(t *testing.T) {
    // Test service embedding functionality from Epic 3.2
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    mockAuthService := testutils.NewMockEndorService()
    mockProfileService := testutils.NewMockEndorService()
    
    // Configure embedded service mocks
    mockAuthService.On("GetResource").Return("auth")
    mockAuthService.On("GetMethods").Return(map[string]sdk.EndorServiceAction{
        "login":  sdk.NewAction(mockLoginHandler, "User login"),
        "logout": sdk.NewAction(mockLogoutHandler, "User logout"),
    })
    
    mockProfileService.On("GetResource").Return("profiles")
    mockProfileService.On("GetMethods").Return(map[string]sdk.EndorServiceAction{
        "get":    sdk.NewAction(mockGetProfileHandler, "Get profile"),
        "update": sdk.NewAction(mockUpdateProfileHandler, "Update profile"),
    })
    
    // Create hybrid service with embedded services
    hybridService := sdk.NewHybridService[User]("users", "User management with embedded services").
        WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
            return map[string]sdk.EndorServiceAction{
                "bulk-import": sdk.NewAction(mockBulkImportHandler, "Bulk import"),
            }
        })
    
    // Test service embedding
    err := hybridService.EmbedService("auth", mockAuthService)
    assert.NoError(t, err)
    
    err = hybridService.EmbedService("profile", mockProfileService)
    assert.NoError(t, err)
    
    // Verify embedded services
    embeddedServices := hybridService.GetEmbeddedServices()
    assert.Len(t, embeddedServices, 2)
    assert.Equal(t, mockAuthService, embeddedServices["auth"])
    assert.Equal(t, mockProfileService, embeddedServices["profile"])
    
    // Convert to EndorService and verify method resolution
    endorService := hybridService.ToEndorService(testSchema)
    methods := endorService.GetMethods()
    
    // Should include CRUD methods + embedded methods + custom actions
    assert.Contains(t, methods, "create")        // Automatic CRUD
    assert.Contains(t, methods, "auth.login")    // Embedded auth service
    assert.Contains(t, methods, "profile.get")   // Embedded profile service  
    assert.Contains(t, methods, "bulk-import")   // Custom action
}
```

### Testing Hierarchical Service Dependencies

```go
func TestServiceHierarchy_DependencyInjection(t *testing.T) {
    // Test complex service hierarchy with DI
    hierarchy := testutils.NewTestServiceHierarchy()
    
    // Build hierarchy: OrderService -> UserService -> AuthService
    mockAuthService := testutils.NewMockAuthService()
    mockUserService := testutils.NewMockUserService()
    mockOrderService := testutils.NewMockOrderService()
    
    // Set up dependency chain
    mockAuthService.On("ValidateToken", "valid-token").Return(UserClaims{UserID: "user-123"}, nil)
    mockUserService.On("GetUser", "user-123").Return(User{ID: "user-123", Status: "active"}, nil)
    mockOrderService.On("CreateOrder", mock.Any).Return(Order{ID: "order-123"}, nil)
    
    hierarchy.AddService("auth", mockAuthService).
             AddService("users", mockUserService).
             AddService("orders", mockOrderService)
    
    // Test hierarchical service interaction
    orderRequest := CreateOrderRequest{
        Token: "valid-token",
        Items: []OrderItem{{ProductID: "prod-1", Quantity: 2}},
    }
    
    // Simulate service chain: auth -> user -> order
    claims, err := hierarchy.GetService("auth").ValidateToken(orderRequest.Token)
    require.NoError(t, err)
    
    user, err := hierarchy.GetService("users").GetUser(claims.UserID)
    require.NoError(t, err)
    
    order, err := hierarchy.GetService("orders").CreateOrder(orderRequest)
    require.NoError(t, err)
    
    assert.Equal(t, "order-123", order.ID)
    hierarchy.AssertAllServicesCalledInOrder([]string{"auth", "users", "orders"})
}
```

---

## Lifecycle Testing Patterns

Test service lifecycle management patterns from Epic 3.5, including startup, runtime, and shutdown behavior.

### Testing Service Lifecycle Events

```go
//go:build unit

func TestServiceLifecycle_StartupSequence(t *testing.T) {
    // Test service startup lifecycle from Epic 3.5
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    lifecycleManager := lifecycle.NewManager()
    
    // Create service with lifecycle support
    service := &UserServiceWithLifecycle{
        repository: mockRepo,
        logger:     mockLogger,
        lifecycle:  lifecycleManager,
    }
    
    // Mock lifecycle events
    mockRepo.On("Initialize").Return(nil)
    mockLogger.On("Info", "Service starting up", mock.Any).Return()
    mockLogger.On("Info", "Service ready", mock.Any).Return()
    
    // Test startup sequence
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := service.Startup(ctx)
    assert.NoError(t, err)
    
    // Verify lifecycle state
    assert.Equal(t, lifecycle.StateRunning, service.GetState())
    
    // Verify initialization calls
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}

func TestServiceLifecycle_GracefulShutdown(t *testing.T) {
    // Test graceful shutdown patterns
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    lifecycleManager := lifecycle.NewManager()
    
    service := &UserServiceWithLifecycle{
        repository: mockRepo,
        logger:     mockLogger,
        lifecycle:  lifecycleManager,
    }
    
    // Start service first
    service.Startup(context.Background())
    
    // Mock shutdown sequence
    mockLogger.On("Info", "Service shutting down", mock.Any).Return()
    mockRepo.On("Close").Return(nil)
    mockLogger.On("Info", "Service stopped", mock.Any).Return()
    
    // Test graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := service.Shutdown(ctx)
    assert.NoError(t, err)
    assert.Equal(t, lifecycle.StateStopped, service.GetState())
    
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}

func TestServiceLifecycle_HealthMonitoring(t *testing.T) {
    // Test service health monitoring during lifecycle
    mockHealthChecker := testutils.NewMockHealthChecker()
    service := &UserServiceWithLifecycle{
        healthChecker: mockHealthChecker,
    }
    
    // Mock health check responses
    mockHealthChecker.On("CheckHealth").Return(lifecycle.HealthStatus{
        Status: "healthy",
        Checks: map[string]bool{
            "database":    true,
            "cache":       true,
            "external-api": true,
        },
    }, nil)
    
    service.Startup(context.Background())
    
    // Test health monitoring
    health, err := service.GetHealth()
    assert.NoError(t, err)
    assert.Equal(t, "healthy", health.Status)
    assert.True(t, health.Checks["database"])
    
    mockHealthChecker.AssertExpectations(t)
}
```

### Testing Service Lifecycle Integration

```go
//go:build integration

func TestServiceLifecycle_Integration_CompleteLifecycle(t *testing.T) {
    // Integration test for complete service lifecycle
    container := di.NewContainer()
    setupLifecycleIntegrationDependencies(container, t)
    
    // Create service with real lifecycle dependencies
    userService, err := di.Resolve[interfaces.UserServiceInterface](container)
    require.NoError(t, err)
    
    lifecycleService := userService.(lifecycle.LifecycleInterface)
    
    // Test startup phase
    startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err = lifecycleService.Startup(startupCtx)
    require.NoError(t, err)
    
    // Verify service is ready and healthy
    health, err := lifecycleService.GetHealth()
    require.NoError(t, err)
    assert.Equal(t, "healthy", health.Status)
    
    // Test service functionality during runtime
    user := User{Name: "Lifecycle Test User", Email: "lifecycle@test.com"}
    err = userService.CreateUser(context.Background(), user)
    require.NoError(t, err)
    
    // Test graceful shutdown
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err = lifecycleService.Shutdown(shutdownCtx)
    require.NoError(t, err)
    assert.Equal(t, lifecycle.StateStopped, lifecycleService.GetState())
}
```

---

## Performance and Benchmark Testing

Validate performance characteristics and detect regressions.

### Service Performance Benchmarks

```go
func BenchmarkUserService_CreateUser(b *testing.B) {
    // Benchmark service performance with mocked dependencies
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    mockValidator := testutils.NewMockValidator()
    
    user := User{ID: "bench-user", Name: "Benchmark User", Email: "bench@test.com"}
    
    mockValidator.On("ValidateUser", mock.Any).Return(nil)
    mockRepo.On("Create", mock.Any, mock.Any).Return(nil)
    mockLogger.On("Info", mock.Any, mock.Any).Return()
    
    service := NewUserService(mockRepo, mockLogger, mockValidator)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := service.CreateUser(context.Background(), user)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDI_ServiceResolution(b *testing.B) {
    // Benchmark dependency injection resolution
    container := di.NewContainer()
    
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    
    di.Register[interfaces.RepositoryInterface](container, mockRepo, di.Singleton)
    di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
    
    di.RegisterFactory[interfaces.UserServiceInterface](container, func(c di.Container) (interfaces.UserServiceInterface, error) {
        repo, _ := di.Resolve[interfaces.RepositoryInterface](c)
        logger, _ := di.Resolve[interfaces.LoggerInterface](c)
        return NewUserService(repo, logger, nil), nil
    }, di.Singleton)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := di.Resolve[interfaces.UserServiceInterface](container)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkServiceComposition_Chain(b *testing.B) {
    // Benchmark service composition performance
    service1 := &MockService{processTime: 100 * time.Microsecond}
    service2 := &MockService{processTime: 100 * time.Microsecond}
    service3 := &MockService{processTime: 100 * time.Microsecond}
    
    chain := composition.ServiceChain(service1, service2, service3)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := chain.Execute(context.Background(), "test-data")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Memory Usage Testing

```go
func TestMemoryUsage_ServiceCreation(t *testing.T) {
    // Test memory usage patterns
    var m1, m2 runtime.MemStats
    
    // Baseline memory
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Create 1000 services with DI
    container := di.NewContainer()
    services := make([]interfaces.UserServiceInterface, 1000)
    
    for i := 0; i < 1000; i++ {
        mockRepo := testutils.NewMockRepository()
        mockLogger := testutils.NewMockLogger()
        services[i] = NewUserService(mockRepo, mockLogger, nil)
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    memoryPerService := (m2.HeapAlloc - m1.HeapAlloc) / 1000
    
    // Verify memory usage is reasonable (less than 10KB per service)
    assert.Less(t, memoryPerService, uint64(10*1024), 
        "Memory usage per service should be less than 10KB, got %d bytes", memoryPerService)
}
```

---

## Test Organization and Build Tags

Organize tests using Go build tags for selective execution.

### Build Tag Structure

```go
// Unit tests - fast, use mocks
//go:build unit

// Integration tests - slower, use real dependencies
//go:build integration

// End-to-end tests - slowest, use complete system
//go:build e2e

// Performance tests - benchmarks and performance validation
//go:build performance
```

### Test Directory Structure

```
test/
├── unit/                  # Unit tests with mocks
│   ├── service_test.go
│   ├── repository_test.go
│   └── handler_test.go
├── integration/           # Integration tests with real dependencies
│   ├── database_test.go
│   ├── http_test.go
│   └── lifecycle_test.go
├── e2e/                  # End-to-end tests
│   ├── user_workflow_test.go
│   └── service_composition_test.go
├── performance/          # Performance and benchmark tests
│   ├── benchmark_test.go
│   └── memory_test.go
└── testutils/            # Test utilities and helpers
    ├── mocks.go
    ├── builders.go
    └── helpers.go
```

### Running Tests by Category

```bash
# Run only unit tests (fast)
go test -tags=unit ./...

# Run only integration tests 
go test -tags=integration ./...

# Run all tests except e2e
go test -tags="unit integration" ./...

# Run performance benchmarks
go test -tags=performance -bench=. ./...

# Run specific test with coverage
go test -tags=unit -coverprofile=coverage.out -covermode=atomic ./...
```

---

## Continuous Integration Patterns

Configure CI/CD pipelines for optimal test execution.

### GitHub Actions Test Pipeline

```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run Unit Tests
      run: go test -tags=unit -v -coverprofile=coverage.out ./...
      
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      mongodb:
        image: mongo:6.0
        ports:
          - 27017:27017
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run Integration Tests
      run: go test -tags=integration -v ./...
      env:
        MONGO_URI: mongodb://localhost:27017/test

  performance-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run Benchmarks
      run: go test -tags=performance -bench=. -benchmem ./...
    
    - name: Performance Regression Check
      run: |
        # Compare against baseline benchmarks
        ./scripts/check-performance-regression.sh
```

### Makefile Test Targets

```makefile
# Makefile test targets
.PHONY: test test-unit test-integration test-e2e test-performance

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -tags=unit -v -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	docker-compose up -d mongodb
	go test -tags=integration -v ./...
	docker-compose down

test-e2e:
	@echo "Running e2e tests..."
	docker-compose up -d
	go test -tags=e2e -v ./...
	docker-compose down

test-performance:
	@echo "Running performance tests..."
	go test -tags=performance -bench=. -benchmem ./...

test-coverage:
	go test -tags="unit integration" -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-watch:
	# Watch for changes and run unit tests
	find . -name "*.go" | entr -c make test-unit
```

---

## Testing Best Practices Summary

### ✅ **Unit Testing Guidelines**

1. **Mock All External Dependencies**: Use `sdk/testutils/` mocks for all interfaces
2. **Test Business Logic Only**: Focus on your service's behavior, not framework behavior  
3. **Use Table-Driven Tests**: Cover multiple scenarios efficiently
4. **Assert Mock Interactions**: Verify dependencies are called correctly
5. **Keep Tests Fast**: Unit tests should run in milliseconds

### ✅ **Integration Testing Guidelines**

1. **Use Real Implementations**: Test with actual databases and external services
2. **Isolate Test Data**: Use unique test databases per test
3. **Test Component Boundaries**: Verify services work together correctly
4. **Clean Up Resources**: Always clean up test data and connections
5. **Test Error Scenarios**: Include network failures and timeouts

### ✅ **Service Composition Testing Guidelines**

1. **Test Service Chains**: Verify sequential processing works correctly
2. **Test Parallel Branches**: Validate concurrent execution patterns
3. **Test Service Embedding**: Verify hierarchical service behavior
4. **Mock Service Dependencies**: Use service-level mocks for composition tests
5. **Test Error Propagation**: Ensure errors bubble up correctly

### ✅ **Performance Testing Guidelines**

1. **Benchmark Critical Paths**: Focus on hot code paths
2. **Set Performance Baselines**: Detect regressions automatically
3. **Test Memory Usage**: Monitor for memory leaks
4. **Test Under Load**: Validate performance under realistic conditions
5. **Profile Bottlenecks**: Use go tool pprof for optimization

This comprehensive testing strategy ensures the endor-sdk-go framework maintains high quality, performance, and reliability while providing excellent developer experience through fast, reliable tests.