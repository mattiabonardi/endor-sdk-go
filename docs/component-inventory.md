# Endor SDK Go - Component Inventory

## Overview

The Endor SDK Go is composed of several key components organized into distinct functional areas. This inventory provides a comprehensive overview of all components, their relationships, and their purposes within the SDK architecture.

## Core Components

### 1. Service Framework Components

#### EndorService (Static Services)
**Location**: `sdk/endor_service.go`  
**Type**: Core Abstraction  
**Purpose**: Provides framework for static services with predefined endpoints

**Key Features**:
- Manual action registration
- Custom HTTP handler creation  
- Direct business logic control
- Version management support

**Interfaces**:
```go
type EndorService struct {
    Resource         string
    Description      string  
    Methods          map[string]EndorServiceAction
    Priority         *int
    ResourceMetadata bool
    Version          string
}
```

#### EndorHybridService (Dynamic Services)
**Location**: `sdk/endor_hybrid_service.go`  
**Type**: Core Abstraction  
**Purpose**: Provides framework for dynamic services with automatic CRUD operations

**Key Features**:
- Automatic CRUD endpoint generation
- MongoDB integration
- Category-based specialization
- Schema-driven development

**Interfaces**:
```go
type EndorHybridService interface {
    GetResource() string
    GetResourceDescription() string
    GetPriority() *int
    WithCategories(categories []EndorHybridServiceCategory) EndorHybridService
    WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridService
    ToEndorService(metadataSchema Schema) EndorService
}
```

#### EndorServiceAction
**Location**: `sdk/endor_service.go`  
**Type**: Action Handler  
**Purpose**: Defines individual service actions with configuration

**Key Features**:
- Type-safe handler functions
- Configurable validation
- Public/private access control
- Input schema generation

### 2. Server and Initialization Components

#### Endor Server
**Location**: `sdk/server.go`  
**Type**: Core Framework  
**Purpose**: Main server initialization and HTTP routing

**Key Features**:
- Builder pattern initialization
- Gin framework integration
- Service registration
- Middleware pipeline setup

**Components**:
- `EndorInitializer`: Fluent builder for service configuration
- `Endor`: Main server instance
- Built-in endpoints (health, metrics, swagger)

#### API Gateway Integration
**Location**: `sdk/api_gateway.go`  
**Type**: Integration Component  
**Purpose**: Automatic API gateway configuration generation

**Key Features**:
- Traefik configuration generation
- Service discovery support
- Load balancer configuration
- Authentication middleware integration

### 3. Context and Request Processing Components

#### EndorContext
**Location**: `sdk/context.go`  
**Type**: Request Context  
**Purpose**: Type-safe request context with session and payload information

**Key Features**:
- Generic payload typing
- Session information access
- Resource metadata access
- Gin context integration

**Structure**:
```go
type EndorContext[T any] struct {
    MicroServiceId         string
    Session                Session
    Payload                T
    ResourceMetadataSchema RootSchema
    CategoryID             *string
    GinContext             *gin.Context
}
```

#### Response Framework
**Location**: `sdk/response.go`  
**Type**: Response Formatting  
**Purpose**: Standardized API response structure

**Key Features**:
- Type-safe response data
- Message system with severity levels
- Schema inclusion
- Error response helpers

### 4. Data Access Components

#### Repository Interfaces
**Location**: `sdk/resource_instance_repository.go`  
**Type**: Data Access Interface  
**Purpose**: Generic repository pattern for data operations

**Key Components**:
- `ResourceInstanceRepositoryInterface[T]`: Standard CRUD operations
- `ResourceInstanceSpecializedRepositoryInterface[T, C]`: Category-based operations
- `StaticResourceInstanceRepositoryInterface[T]`: Read-only operations

#### MongoDB Repository Implementations
**Locations**: 
- `sdk/mongo_resource_instance_repository.go`
- `sdk/mongo_static_resource_instance_repository.go`

**Type**: Data Access Implementation  
**Purpose**: MongoDB-specific repository implementations

**Key Features**:
- Generic type support
- Automatic ID generation
- Transaction support
- Connection pooling

#### MongoDB Connection Manager
**Location**: `sdk/mongo.go`  
**Type**: Database Connection  
**Purpose**: Singleton MongoDB connection management

**Key Features**:
- Automatic connection initialization
- Connection pooling
- Error handling
- Configuration-based setup

### 5. Schema and Validation Components

#### Schema System
**Location**: `sdk/schema.go`  
**Type**: Schema Management  
**Purpose**: JSON schema generation from Go structs

**Key Features**:
- Automatic reflection-based generation
- Custom format support
- Validation tag integration
- UI hint support

**Core Types**:
```go
type RootSchema struct {
    Type        string
    Properties  map[string]PropertySchema
    Required    []string
    Title       string
    Description string
}
```

#### DSL Support
**Location**: `sdk/dsl.go`  
**Type**: Domain-Specific Language  
**Purpose**: YAML-based schema definitions for dynamic resources

**Key Features**:
- YAML schema parsing
- Dynamic resource creation
- Schema validation
- Type mapping

### 6. CRUD and Resource Management Components

#### Generic CRUD Operations
**Location**: `sdk/crud.go`  
**Type**: Business Logic  
**Purpose**: Generic CRUD operations for hybrid services

**Key Features**:
- Type-safe operations
- Automatic validation
- Error handling
- Response formatting

#### Resource Abstractions
**Location**: `sdk/resource.go`  
**Type**: Core Abstractions  
**Purpose**: Base resource definitions and interfaces

**Key Components**:
- `ResourceInstanceInterface`: Base resource interface
- `ResourceInstance[T]`: Wrapper with metadata
- `ResourceInstanceSpecialized[T, C]`: Category-specialized wrapper

#### Resource Service Logic
**Location**: `sdk/resource_service.go`  
**Type**: Service Logic  
**Purpose**: High-level resource management operations

**Key Features**:
- Service orchestration
- Business rule enforcement
- Repository coordination
- Transaction management

#### Resource Action Service
**Location**: `sdk/resource_action_service.go`  
**Type**: Action Handler  
**Purpose**: Coordination of resource actions and operations

**Key Features**:
- Action routing
- Parameter validation
- Response coordination
- Error handling

### 7. Configuration and Error Handling Components

#### Configuration Management
**Location**: `sdk/configuration.go`  
**Type**: Configuration  
**Purpose**: Environment-based configuration management

**Key Features**:
- Environment variable loading
- Default value support
- Type-safe configuration struct
- Singleton pattern

**Configuration Structure**:
```go
type ServerConfig struct {
    ServerPort                    string
    DocumentDBUri                 string
    HybridResourcesEnabled        bool
    DynamicResourcesEnabled       bool
    DynamicResourceDocumentDBName string
}
```

#### Error Handling System
**Location**: `sdk/errors.go`  
**Type**: Error Management  
**Purpose**: Standardized error handling and HTTP status mapping

**Key Features**:
- Custom error types
- HTTP status code mapping
- Error response formatting
- Stack trace support

#### Constants and Types
**Locations**: 
- `sdk/constants.go`
- `sdk/types.go`

**Type**: Core Definitions  
**Purpose**: Framework constants and type definitions

### 8. Documentation and API Discovery Components

#### Swagger Integration
**Location**: `sdk/swagger.go`  
**Type**: API Documentation  
**Purpose**: Automatic OpenAPI 3.1 specification generation

**Key Features**:
- Automatic endpoint discovery
- Schema documentation
- Authentication documentation
- Interactive UI serving

#### Swagger UI Assets
**Location**: `sdk/swagger/`  
**Type**: Static Assets  
**Purpose**: Embedded Swagger UI for interactive API documentation

**Components**:
- `index.html`: Swagger UI entry point
- `*.js`: JavaScript assets
- `*.css`: Styling assets
- `swagger-initializer.js`: Configuration

### 9. Utility Components

#### Array Utilities
**Location**: `sdk/utils/array.go`  
**Type**: Utility Functions  
**Purpose**: Generic array manipulation functions

**Key Features**:
- Generic array operations
- Functional programming helpers
- Type-safe transformations
- Common array algorithms

## Test Components

### Unit Test Suite
**Locations**: 
- `test/endor_hybrid_service_test.go`
- `test/resource_test.go`
- `test/schema_test.go`
- `test/swagger_test.go`

**Type**: Test Components  
**Purpose**: Unit tests for core SDK functionality

**Coverage Areas**:
- Service initialization
- Schema generation
- Repository operations
- API endpoint generation

### Example Service Implementations
**Locations**:
- `test/services/service1.go` - Static service example
- `test/services/service2.go` - Hybrid service example

**Type**: Reference Implementations  
**Purpose**: Concrete examples of service patterns

**Key Features**:
- Complete service implementations
- Best practice demonstrations
- Integration testing support
- Development reference

## Component Relationships

### Dependency Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  ┌─────────────────┐  ┌─────────────────────────────────┐  │
│  │  Static         │  │  Hybrid Services                │  │
│  │  Services       │  │  (with CRUD)                    │  │
│  └─────────────────┘  └─────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Service Framework                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐ │
│  │  Context &      │  │  Actions &      │  │  Response     │ │
│  │  Routing        │  │  Handlers       │  │  Formatting   │ │
│  └─────────────────┘  └─────────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Data Access Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐ │
│  │  Repository     │  │  MongoDB        │  │  Schema       │ │
│  │  Interfaces     │  │  Implementation │  │  Generation   │ │
│  └─────────────────┘  └─────────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐ │
│  │  HTTP Server    │  │  Configuration  │  │  Error        │ │
│  │  (Gin)          │  │  Management     │  │  Handling     │ │
│  └─────────────────┘  └─────────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Component Interactions

1. **Service Registration Flow**:
   `EndorInitializer` → `EndorService`/`EndorHybridService` → `Router Registration` → `HTTP Endpoints`

2. **Request Processing Flow**:
   `HTTP Request` → `Gin Middleware` → `EndorContext` → `Action Handler` → `Repository` → `Response`

3. **Data Access Flow**:
   `Service Logic` → `Repository Interface` → `MongoDB Implementation` → `Database`

4. **Schema Flow**:
   `Go Structs` → `Reflection` → `JSON Schema` → `Validation` → `Documentation`

## Reusable vs. Specific Components

### Reusable Components
- Repository interfaces and implementations
- Schema generation system
- Response formatting
- Error handling
- Context management
- Configuration system

### Framework-Specific Components
- Endor service abstractions
- API gateway integration
- Swagger configuration
- MongoDB connection management
- Hybrid service CRUD operations

### Application-Specific Components
- Service implementations (in test/services/)
- Custom action handlers
- Business logic validators
- Domain-specific schemas

---

*This component inventory provides a comprehensive map of the Endor SDK's architecture, enabling developers to understand the framework's structure and extend it effectively for their specific use cases.*