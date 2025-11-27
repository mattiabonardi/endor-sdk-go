# Endor SDK Go - Development Guide

## Prerequisites

Before you begin developing with the Endor SDK, ensure you have the following installed:

### Required Software

| Tool | Minimum Version | Purpose |
|------|----------------|---------|
| **Go** | 1.21.4+ | Core development language |
| **MongoDB** | 4.4+ | Database backend for hybrid services |
| **Git** | 2.0+ | Version control |

### Recommended Tools

| Tool | Purpose |
|------|---------|
| **Docker** | Local MongoDB instance |
| **VS Code** | Development environment with Go extension |
| **Postman/Insomnia** | API testing |
| **MongoDB Compass** | Database visualization |

## Environment Setup

### 1. Clone and Initialize Project

```bash
# Clone the repository
git clone https://github.com/mattiabonardi/endor-sdk-go.git
cd endor-sdk-go

# Install dependencies
go mod tidy

# Verify installation
go version
```

### 2. MongoDB Setup

#### Option A: Local MongoDB Installation
```bash
# Install MongoDB (macOS with Homebrew)
brew install mongodb-community
brew services start mongodb-community

# Install MongoDB (Ubuntu)
sudo apt-get install mongodb
sudo systemctl start mongodb
```

#### Option B: Docker MongoDB
```bash
# Run MongoDB in Docker
docker run --name endor-mongodb -p 27017:27017 -d mongo:latest

# Verify connection
docker logs endor-mongodb
```

### 3. Environment Configuration

Create a `.env` file in the project root:

```bash
# .env file
PORT=8080
DOCUMENT_DB_URI=mongodb://localhost:27017
HYBRID_RESOURCES_ENABLED=true
DYNAMIC_RESOURCES_ENABLED=true
DYNAMIC_RESOURCE_DOCUMENT_DB_NAME=endor_dynamic
```

### Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `DOCUMENT_DB_URI` | mongodb://localhost:27017 | MongoDB connection string |
| `HYBRID_RESOURCES_ENABLED` | false | Enable hybrid service features |
| `DYNAMIC_RESOURCES_ENABLED` | false | Enable runtime resource creation |
| `DYNAMIC_RESOURCE_DOCUMENT_DB_NAME` | endor_dynamic | Database name for dynamic resources |

## Local Development Commands

### Build and Run

```bash
# Build the project
go build -o bin/endor-sdk-demo ./main.go

# Run the example application
go run main.go

# Run with specific environment
PORT=9090 go run main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test ./sdk

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Development Tools

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Generate documentation
godoc -http=:6060

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/endor-sdk-linux ./main.go
```

## Development Workflow

### 1. Creating a New Service

#### Static Service Example

```go
// Create a new static service
func NewOrderService() sdk.EndorService {
    return sdk.EndorService{
        Resource:    "orders",
        Description: "Order management service",
        Priority:    intPtr(1),
        Methods: map[string]sdk.EndorServiceAction{
            "validate": sdk.NewAction(validateOrder, "Validate order data"),
            "process":  sdk.NewAction(processOrder, "Process order payment"),
            "cancel":   sdk.NewAction(cancelOrder, "Cancel existing order"),
        },
    }
}

// Handler implementation
func validateOrder(ctx *sdk.EndorContext[OrderValidationRequest]) (*sdk.Response[OrderValidationResponse], error) {
    // Access request payload
    order := ctx.Payload
    
    // Access session information
    userID := ctx.Session.UserID
    
    // Business logic
    if order.Amount <= 0 {
        return sdk.NewErrorResponse[OrderValidationResponse](
            sdk.CreateBadRequestError("Invalid order amount")
        ), nil
    }
    
    // Return success response
    return sdk.NewSuccessResponse(OrderValidationResponse{
        Valid:   true,
        OrderID: generateOrderID(),
    }), nil
}
```

#### Hybrid Service Example

```go
// Create a hybrid service with automatic CRUD
func NewProductService() sdk.EndorHybridService {
    return sdk.NewHybridService("products", "Product catalog management").
        WithCategories([]sdk.EndorHybridServiceCategory{
            {
                CategoryId:          "electronics",
                CategoryDescription: "Electronic products",
                CategorySchema:      generateElectronicsSchema(),
            },
            {
                CategoryId:          "clothing",
                CategoryDescription: "Clothing items",
                CategorySchema:      generateClothingSchema(),
            },
        }).
        WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
            return map[string]sdk.EndorServiceAction{
                "search":    sdk.NewAction(searchProducts, "Search products by criteria"),
                "recommend": sdk.NewAction(getRecommendations, "Get product recommendations"),
            }
        })
}
```

### 2. Data Model Development

#### Define Resource Interface

```go
type Product struct {
    ID          string  `json:"id" bson:"_id"`
    Name        string  `json:"name" bson:"name" validate:"required"`
    Description string  `json:"description" bson:"description"`
    Price       float64 `json:"price" bson:"price" validate:"min=0"`
    CategoryID  string  `json:"categoryId" bson:"categoryId"`
    Active      bool    `json:"active" bson:"active"`
}

// Implement ResourceInstanceInterface
func (p *Product) GetId() string {
    return p.ID
}

func (p *Product) SetId(id string) {
    p.ID = id
}
```

#### Custom Repository (Optional)

```go
type ProductRepository struct {
    sdk.ResourceInstanceRepositoryInterface[Product]
    collection *mongo.Collection
}

func NewProductRepository() *ProductRepository {
    return &ProductRepository{
        ResourceInstanceRepositoryInterface: sdk.NewMongoResourceInstanceRepository[Product]("products"),
        collection: sdk.GetMongo().Database.Collection("products"),
    }
}

// Custom query methods
func (r *ProductRepository) FindByCategory(ctx context.Context, categoryID string) ([]sdk.ResourceInstance[Product], error) {
    filter := bson.M{"categoryId": categoryID}
    cursor, err := r.collection.Find(ctx, filter)
    // Implementation...
}
```

### 3. Testing Strategy

#### Unit Testing

```go
func TestProductService(t *testing.T) {
    // Setup test service
    service := NewProductService()
    
    // Test service configuration
    assert.Equal(t, "products", service.GetResource())
    assert.Equal(t, "Product catalog management", service.GetResourceDescription())
    
    // Test category registration
    categories := service.GetCategories()
    assert.Len(t, categories, 2)
}

func TestProductValidation(t *testing.T) {
    // Test request validation
    ctx := &sdk.EndorContext[ProductCreateRequest]{
        Payload: ProductCreateRequest{
            Name:  "", // Invalid: required field
            Price: -10, // Invalid: negative price
        },
    }
    
    response, err := createProduct(ctx)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validation failed")
}
```

#### Integration Testing

```go
func TestProductServiceIntegration(t *testing.T) {
    // Setup test database
    testDB := setupTestMongoDB(t)
    defer testDB.Cleanup()
    
    // Initialize SDK with test config
    endor := sdk.NewEndorInitializer().
        WithHybridServices(&[]sdk.EndorHybridService{
            NewProductService(),
        }).
        Build()
    
    // Start test server
    server := httptest.NewServer(endor.GetRouter())
    defer server.Close()
    
    // Test API endpoints
    response := testCreateProduct(t, server.URL, validProductData)
    assert.Equal(t, http.StatusCreated, response.StatusCode)
}
```

## API Development

### Request/Response Patterns

```go
// Request DTO
type CreateProductRequest struct {
    Name        string  `json:"name" validate:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" validate:"min=0"`
    CategoryID  string  `json:"categoryId" validate:"required"`
}

// Response DTO
type ProductResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Price       float64   `json:"price"`
    CreatedAt   time.Time `json:"createdAt"`
}

// Handler with type safety
func createProduct(ctx *sdk.EndorContext[CreateProductRequest]) (*sdk.Response[ProductResponse], error) {
    req := ctx.Payload
    
    // Validation (automatic via struct tags)
    // Business logic
    product := Product{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        CategoryID:  req.CategoryID,
        Active:      true,
    }
    
    // Repository operation
    repository := sdk.NewMongoResourceInstanceRepository[Product]("products")
    created, err := repository.Create(context.Background(), sdk.CreateDTO[sdk.ResourceInstance[Product]]{
        Value: sdk.ResourceInstance[Product]{This: product},
    })
    if err != nil {
        return nil, sdk.CreateInternalServerError("Failed to create product")
    }
    
    // Response
    return sdk.NewSuccessResponse(ProductResponse{
        ID:        created.This.GetId(),
        Name:      created.This.Name,
        Price:     created.This.Price,
        CreatedAt: time.Now(),
    }), nil
}
```

### Error Handling Best Practices

```go
// Custom error types
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Error response helper
func handleValidationError(errors []ValidationError) (*sdk.Response[interface{}], error) {
    messages := make([]sdk.Message, len(errors))
    for i, err := range errors {
        messages[i] = sdk.Message{
            Gravity: sdk.MessageGravityError,
            Value:   fmt.Sprintf("%s: %s", err.Field, err.Message),
        }
    }
    
    return &sdk.Response[interface{}]{
        Messages: messages,
        Data:     nil,
    }, nil
}
```

## Debugging and Development Tools

### Logging

```go
import "log"

func debugHandler(ctx *sdk.EndorContext[RequestType]) (*sdk.Response[ResponseType], error) {
    log.Printf("Request: %+v", ctx.Payload)
    log.Printf("Session: %+v", ctx.Session)
    log.Printf("Resource Schema: %+v", ctx.ResourceMetadataSchema)
    
    // Handler logic...
}
```

### Swagger Documentation

Access interactive API documentation:

```bash
# Start the server
go run main.go

# Open browser
open http://localhost:8080/swagger
```

### Health Checks

Monitor service health:

```bash
# Readiness check
curl http://localhost:8080/readyz

# Liveness check  
curl http://localhost:8080/livez

# Prometheus metrics
curl http://localhost:8080/metrics
```

### Database Inspection

```bash
# Connect to MongoDB
mongo

# Switch to your database
use endor_dynamic

# List collections
show collections

# Query documents
db.products.find().pretty()

# Check indexes
db.products.getIndexes()
```

## Common Development Tasks

### Adding Custom Validation

```go
import "github.com/go-playground/validator/v10"

type CustomValidatedRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Phone    string `json:"phone" validate:"required,e164"`
    Website  string `json:"website" validate:"omitempty,url"`
    Age      int    `json:"age" validate:"min=18,max=120"`
}
```

### Implementing Middleware

```go
func authenticationMiddleware(userRepository UserRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Missing authorization header"})
            return
        }
        
        // Validate token and set session
        session, err := validateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        c.Set("session", session)
        c.Next()
    }
}
```

### Custom Schema Generation

```go
func generateCustomSchema() sdk.RootSchema {
    return sdk.RootSchema{
        Type: "object",
        Properties: map[string]sdk.PropertySchema{
            "customField": {
                Type:        "string",
                Format:      "custom-format",
                Description: "Custom field with special validation",
                Pattern:     "^[A-Z]{2,4}[0-9]{4}$",
            },
        },
        Required: []string{"customField"},
    }
}
```

## Performance Optimization

### Connection Pooling

MongoDB connection pooling is configured automatically, but you can customize it:

```go
// Custom MongoDB configuration
clientOptions := options.Client().
    ApplyURI(config.DocumentDBUri).
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second).
    SetMaxConnecting(10)
```

### Caching Strategies

```go
import "sync"

type CacheService struct {
    cache map[string]interface{}
    mutex sync.RWMutex
}

func (c *CacheService) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    value, exists := c.cache[key]
    return value, exists
}

func (c *CacheService) Set(key string, value interface{}) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.cache[key] = value
}
```

## Deployment Preparation

### Building for Production

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/endor-service ./main.go

# Create Docker image
docker build -t endor-service:latest .
```

### Configuration for Production

```bash
# Production environment variables
export PORT=8080
export DOCUMENT_DB_URI=mongodb://prod-mongo-cluster:27017/endor
export HYBRID_RESOURCES_ENABLED=true
export DYNAMIC_RESOURCES_ENABLED=false  # Disable for security
```

### Health Check Configuration

Ensure health endpoints are accessible for orchestration platforms:

```go
func configureHealthChecks(router *gin.Engine) {
    router.GET("/readyz", func(c *gin.Context) {
        // Check database connectivity
        if err := sdk.GetMongo().Client.Ping(context.Background(), nil); err != nil {
            c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"status": "ready"})
    })
}
```

---

*This development guide provides the foundation for building robust microservices with the Endor SDK. Refer to the API contracts and architecture documentation for detailed technical specifications.*