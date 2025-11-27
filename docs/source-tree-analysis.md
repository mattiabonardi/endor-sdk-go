# Endor SDK Go - Source Tree Analysis

## Project Structure Overview

The endor-sdk-go project follows a clean, modular architecture with clear separation of concerns. The codebase is organized as a single Go module providing SDK functionality for building Endor microservices.

```
endor-sdk-go/                    # Root directory
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── main.go                      # Example usage/demo application
├── docs/                        # 📁 Generated documentation
│   ├── bmm-workflow-status.yaml # BMM methodology tracking
│   └── sprint-artifacts/        # Sprint planning artifacts
├── sdk/                         # 🎯 Core SDK implementation
│   ├── *.go                     # SDK source files (24 files)
│   ├── swagger/                 # 📄 Swagger UI assets
│   └── utils/                   # 🔧 Utility functions
├── test/                        # 🧪 Test implementation examples
│   ├── *_test.go               # Unit tests
│   └── services/               # Example service implementations
└── user-data/                  # 📋 Project documentation
    └── brief.md                # Project brief
```

## Critical Directories

### `/sdk` - Core Framework Implementation

The heart of the Endor SDK, containing all framework functionality:

**Purpose**: Provides the complete framework for building Endor microservices  
**Entry Points**: 
- `server.go` - Main server initialization and routing
- `endor_service.go` - Static service definitions  
- `endor_hybrid_service.go` - Dynamic service implementations

**Key Components**:

```
sdk/
├── server.go                    # 🚀 HTTP server and initialization
├── endor_service.go             # 📋 Static service abstractions
├── endor_hybrid_service.go      # 🔄 Dynamic hybrid services
├── api_gateway.go               # 🌐 API gateway integration
├── configuration.go             # ⚙️ Environment configuration
├── context.go                   # 📦 Request context management
├── crud.go                      # 🗃️ Generic CRUD operations
├── schema.go                    # 📄 JSON schema generation
├── types.go                     # 🏷️ Core type definitions
├── response.go                  # 📤 Response formatting
├── errors.go                    # ❌ Error handling
├── resource.go                  # 📚 Resource abstractions
├── resource_service.go          # 🔧 Resource service logic
├── resource_action_service.go   # ⚡ Action handling
├── mongo.go                     # 🍃 MongoDB connection
├── mongo_resource_instance_repository.go           # 🗄️ Standard MongoDB repo
├── mongo_static_resource_instance_repository.go    # 📖 Static MongoDB repo  
├── resource_instance_repository.go                 # 🔌 Repository interfaces
├── static_resource_instance_repository.go          # 🔍 Static repo interfaces
├── endor_resource_repository.go # 🏪 Resource repository abstractions
├── constants.go                 # 📊 Framework constants
├── dsl.go                      # 🎨 Domain-specific language support
├── swagger.go                  # 📚 Swagger/OpenAPI integration
├── swagger/                    # 📄 Static Swagger UI assets
│   ├── index.html              # Swagger UI entry point
│   ├── *.js                    # Swagger UI JavaScript
│   └── *.css                   # Swagger UI styling
└── utils/                      # 🔧 Utility functions
    └── array.go                # Array manipulation utilities
```

### `/test` - Test Examples and Validation

**Purpose**: Contains example implementations and unit tests  
**Entry Points**: `*_test.go` files for unit testing

```
test/
├── endor_hybrid_service_test.go # 🧪 Hybrid service tests
├── resource_test.go             # 🧪 Resource handling tests  
├── schema_test.go               # 🧪 Schema generation tests
├── swagger_test.go              # 🧪 Swagger integration tests
└── services/                    # 💼 Example service implementations
    ├── service1.go              # Static service example
    └── service2.go              # Hybrid service example
```

### `/docs` - Project Documentation

**Purpose**: Contains generated documentation and project management artifacts

```
docs/
├── bmm-workflow-status.yaml     # 📊 BMM methodology progress
└── sprint-artifacts/            # 🗓️ Sprint planning materials
```

### `/user-data` - Project Artifacts

**Purpose**: Contains project briefing and planning documents

```
user-data/
└── brief.md                     # 📋 Project overview document
```

## Entry Points and Bootstrapping

### Primary Entry Point: `main.go`

```go
func main() {
    sdk.NewEndorInitializer().
        WithEndorServices(&[]sdk.EndorService{
            services_test.NewService1(),
        }).
        WithHybridServices(&[]sdk.EndorHybridService{
            services_test.NewService2(),
        }).
        Build().
        Init("endor-sdk-service")
}
```

**Purpose**: Demonstrates SDK usage and serves as example application

### SDK Entry Point: `sdk/server.go`

```go
func (h *Endor) Init(microserviceId string) {
    // Core initialization logic
    // - Load configuration
    // - Setup MongoDB connection
    // - Register services
    // - Start HTTP server
}
```

**Purpose**: Main SDK initialization and HTTP server startup

## Service Definition Patterns

### Static Services (`services/service1.go`)

```go
func NewService1() sdk.EndorService {
    return sdk.EndorService{
        Resource:    "service1",
        Description: "Example static service",
        Methods: map[string]sdk.EndorServiceAction{
            "action1": sdk.NewAction(handleAction1, "Handle action 1"),
        },
    }
}
```

### Hybrid Services (`services/service2.go`)

```go  
func NewService2() sdk.EndorHybridService {
    return sdk.NewHybridService("service2", "Example hybrid service").
        WithCategories([]sdk.EndorHybridServiceCategory{
            {CategoryId: "category1", CategoryDescription: "Category 1"},
        })
}
```

## Data Flow Architecture

### Request Processing Pipeline

```
HTTP Request → Authentication → Validation → Context Creation → Handler → Response
     ↓              ↓              ↓              ↓            ↓         ↓
  gin.Context → Session → Payload → EndorContext → Business → Response[T]
```

### Repository Pattern Flow

```
Service Layer → Repository Interface → MongoDB Implementation → Database
     ↓                    ↓                      ↓               ↓
Business Logic → Generic CRUD → Collection Operations → Documents
```

## Configuration Management

### Environment-Based Configuration

Configuration is managed through environment variables with fallback defaults:

```
.env file → os.Getenv() → ServerConfig struct → Singleton pattern
```

**Key Configuration Files**:
- **Environment variables**: Runtime configuration
- **go.mod**: Dependency management  
- **MongoDB connection**: Dynamic database configuration

## Integration Points

### MongoDB Integration

**Connection Flow**:
```
server.go → configuration.go → mongo.go → MongoDB Client → Collections
```

**Repository Registration**:
```
HybridService → Repository Factory → MongoDB Repository → Collection
```

### API Gateway Integration

**Configuration Generation**:
```
Services → api_gateway.go → Traefik Config → YAML Output
```

### Swagger Documentation

**Documentation Flow**:
```
Services → Schema Generation → OpenAPI Spec → Swagger UI
```

## Testing Structure

### Unit Test Organization

```
Component Tests:
├── endor_hybrid_service_test.go # Hybrid service functionality
├── resource_test.go             # Resource management
├── schema_test.go               # Schema generation
└── swagger_test.go              # API documentation

Example Implementations:
├── services/service1.go         # Static service pattern
└── services/service2.go         # Hybrid service pattern
```

### Test Data Flow

```
Test Setup → Mock Dependencies → Service Creation → HTTP Testing → Validation
```

## Development Workflow

### Local Development Setup

1. **Environment**: Configure `.env` file or environment variables
2. **Database**: Start MongoDB instance (local or container)
3. **Dependencies**: Run `go mod tidy` to install dependencies
4. **Run**: Execute `go run main.go` to start example application
5. **Test**: Run `go test ./...` to execute test suite

### Service Development Pattern

1. **Define Model**: Implement `ResourceInstanceInterface`
2. **Create Service**: Choose `EndorService` or `EndorHybridService`
3. **Register Actions**: Define custom business logic actions
4. **Test**: Write unit tests and integration tests
5. **Deploy**: Use generated API gateway configuration

## Dependencies and External Integration

### Core Dependencies

- **gin-gonic/gin**: HTTP framework and routing
- **go.mongodb.org/mongo-driver**: MongoDB client and operations
- **prometheus/client_golang**: Metrics collection and monitoring
- **joho/godotenv**: Environment configuration management

### Integration Capabilities

- **API Gateway**: Traefik configuration generation
- **Monitoring**: Prometheus metrics endpoint
- **Documentation**: Automatic Swagger UI generation
- **Authentication**: Header-based session management

---

*This source tree structure enables rapid development of consistent microservices while maintaining clear separation between framework code, examples, tests, and documentation.*