# User Service Example - New Architecture

This example demonstrates how to create services using the new refactored endor-sdk-go architecture.

## Files

- `user_service.go` - Complete UserService implementation
- `user_service_test.go` - Comprehensive tests for the service

## What this demonstrates

### 1. **EndorHybridService Pattern**
The new architecture introduces `EndorHybridService` which combines:
- **Automatic CRUD operations** (like the old EndorService)
- **Category-based specialization** (new feature)
- **Custom actions** (enhanced)
- **Full testability** (new feature)

### 2. **Category-Based Specialization**
```go
// Users can have different categories with specialized schemas
userService.WithCategories([]sdk.EndorHybridServiceCategory{
    // Admin users with admin-specific fields
    sdk.NewEndorHybridServiceCategory[*User, *AdminUser](adminCategory),
    // Premium users with subscription fields  
    sdk.NewEndorHybridServiceCategory[*User, *PremiumUser](premiumCategory),
})
```

### 3. **Generated API Endpoints**
The service automatically creates these endpoints:

#### Base Resource Endpoints:
- `GET /users/schema` - Get schema
- `GET /users/list` - List all users
- `GET /users/instance/{id}` - Get specific user
- `POST /users/create` - Create user
- `PUT /users/update/{id}` - Update user
- `DELETE /users/delete/{id}` - Delete user

#### Category-Specific Endpoints:
- `GET /users/admin/schema` - Admin user schema
- `GET /users/admin/list` - List admin users
- `POST /users/admin/create` - Create admin user
- `GET /users/premium/schema` - Premium user schema
- `GET /users/premium/list` - List premium users
- `POST /users/premium/create` - Create premium user

#### Custom Action Endpoints:
- `POST /users/promote-user` - Promote user to admin
- `POST /users/send-notification` - Send notification
- `POST /users/bulk-operations` - Bulk operations

### 4. **Dependency Injection Ready**
```go
// Production usage with dependency injection
func NewUserServiceWithDependencies(
    repository interfaces.RepositoryPattern,
    config interfaces.ConfigProviderInterface,
    logger interfaces.LoggerInterface,
) sdk.EndorHybridService {
    // Full DI pattern - ready for production use
}
```

### 5. **Full Testing Support**
The new architecture enables comprehensive testing:
- Unit tests with mocked dependencies
- Integration tests with real databases
- Category specialization testing
- Custom action testing

## Key Differences from Old Architecture

### Before (EndorService only):
```go
// Old way - limited, hard to test
service := sdk.NewEndorService("users", "User service", methods)
// - No categories
// - No automatic CRUD
// - Hard to test (tight coupling)
// - Limited customization
```

### After (EndorHybridService):
```go
// New way - powerful, testable, composable
userService := sdk.NewHybridService[*User]("users", "User Management")
    .WithCategories(categories)      // Category specialization
    .WithActions(customActions)      // Custom business logic
    .WithMiddleware(middleware)      // Cross-cutting concerns

// Convert to EndorService when needed
endorService := userService.ToEndorService(metadataSchema)
```

## Running the Example

1. **Run tests:**
   ```bash
   go test ./test/services/ -v
   ```

2. **Use in your application:**
   ```go
   import "your-project/test/services"
   
   userService := services_test.NewUserService()
   endorService := userService.ToEndorService(sdk.Schema{})
   
   // Register with your HTTP router
   // The service provides all the endpoints automatically
   ```

## Benefits of the New Architecture

1. **🔧 Dependency Injection** - Easy testing and production flexibility
2. **🎯 Category Specialization** - Different user types with specialized schemas
3. **⚡ Automatic CRUD** - No boilerplate for basic operations
4. **🔍 Full Testability** - Mock everything, test everything
5. **📦 Service Composition** - Services can embed other services
6. **🎨 Type Safety** - Go generics ensure compile-time correctness
7. **📖 Auto Documentation** - Schemas generated automatically from Go structs