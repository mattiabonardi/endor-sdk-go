# Endor SDK Go - Documentation Index

## Project Overview

**Endor SDK Go** is a sophisticated backend framework designed for building microservices within the Endor ecosystem. It provides a complete abstraction layer for creating RESTful APIs with automatic CRUD operations, dynamic schema generation, MongoDB integration, and API gateway configuration.

- **Type:** Backend Library/SDK (Monolith)
- **Primary Language:** Go (v1.21.4) 
- **Architecture:** Service Framework Pattern
- **Framework:** Gin HTTP framework with MongoDB backend

## Quick Reference

### Technology Stack
- **Runtime:** Go 1.21.4 with Gin framework
- **Database:** MongoDB with automatic schema management  
- **Monitoring:** Prometheus metrics integration
- **Documentation:** Swagger/OpenAPI 3.1 automatic generation
- **Configuration:** Environment-based with .env support

### Architecture Pattern
**Service Framework Pattern** - Dual service architecture supporting both static services (manual endpoints) and dynamic hybrid services (automatic CRUD with MongoDB integration)

### Entry Point
```go
sdk.NewEndorInitializer().
    WithEndorServices(&[]sdk.EndorService{...}).
    WithHybridServices(&[]sdk.EndorHybridService{...}).
    Build().
    Init("microservice-name")
```

## Generated Documentation

### Core Documentation
- [**Project Overview**](./project-overview.md) - Executive summary and technology stack
- [**Architecture**](./architecture.md) - Detailed architectural design and patterns
- [**API Contracts**](./api-contracts.md) - Complete API endpoint documentation
- [**Data Models**](./data-models.md) - Database schema and persistence layer
- [**Development Guide**](./development-guide.md) - Setup, workflow, and best practices
- [**Source Tree Analysis**](./source-tree-analysis.md) - Codebase structure and organization
- [**Component Inventory**](./component-inventory.md) - Comprehensive component catalog

## Existing Documentation

### Project Management
- [**BMM Workflow Status**](./bmm-workflow-status.yaml) - BMM methodology progress tracking
- [**Sprint Artifacts**](./sprint-artifacts/) - Sprint planning and management materials  
- [**Project Brief**](../user-data/brief.md) - High-level project description

## Getting Started

### Quick Start
1. **Clone Repository**: `git clone https://github.com/mattiabonardi/endor-sdk-go.git`
2. **Install Dependencies**: `go mod tidy`
3. **Setup MongoDB**: Local instance or Docker container
4. **Configure Environment**: Create `.env` file with database URI
5. **Run Example**: `go run main.go`
6. **Access Documentation**: Visit `http://localhost:8080/swagger`

### Key Directories
- **`/sdk`** - Core framework implementation (24 Go files)
- **`/test`** - Unit tests and example service implementations  
- **`/docs`** - Generated documentation (this directory)
- **`main.go`** - Example application demonstrating SDK usage

### Essential Configuration
```bash
PORT=8080
DOCUMENT_DB_URI=mongodb://localhost:27017
HYBRID_RESOURCES_ENABLED=true
DYNAMIC_RESOURCES_ENABLED=false
```

## Development Workflows

### Creating Static Services
1. Define resource model implementing `ResourceInstanceInterface`
2. Create `EndorService` with custom actions
3. Implement typed handler functions
4. Register service with initializer

### Creating Hybrid Services  
1. Define resource model with automatic CRUD
2. Create `EndorHybridService` with optional categories
3. Add custom actions as needed
4. MongoDB integration handles persistence automatically

### API Development
- **Type Safety**: Generic handlers with compile-time validation
- **Auto-Documentation**: Swagger UI automatically generated
- **Schema Validation**: JSON schema generation from Go structs
- **Error Handling**: Standardized error responses with severity levels

## Key Features

### Service Framework
- **Dual Service Types**: Static (manual) and Hybrid (automatic CRUD)
- **Type Safety**: Extensive use of Go generics
- **Schema-Driven**: Dynamic JSON schema generation
- **Category System**: Resource specialization support

### Data Management
- **Repository Pattern**: Clean separation of business logic and data access
- **MongoDB Integration**: Automatic connection management and schema handling
- **Generic CRUD**: Type-safe operations with automatic validation
- **Transaction Support**: ACID compliance for complex operations

### Integration & Deployment
- **API Gateway Ready**: Automatic Traefik configuration generation
- **Health Monitoring**: Built-in readiness/liveness checks and Prometheus metrics
- **Authentication**: Header-based session management
- **Documentation**: Interactive Swagger UI with auto-generated specs

## Architecture Highlights

### Service Layer
```
EndorService (Static) ←→ EndorHybridService (Dynamic)
        ↓                           ↓
   Custom Actions            Automatic CRUD + Custom Actions
        ↓                           ↓  
   Manual Endpoints          Schema-Driven Endpoints
```

### Data Layer
```
Repository Interface → MongoDB Implementation → Database Collections
        ↓                      ↓                      ↓
   Generic CRUD          Connection Pooling    Document Storage
```

### Request Pipeline
```
HTTP → Authentication → Validation → Context → Handler → Repository → Response
```

## Next Steps

### For New Developers
1. Read [Architecture](./architecture.md) for design understanding
2. Follow [Development Guide](./development-guide.md) for setup  
3. Review [API Contracts](./api-contracts.md) for endpoint patterns
4. Examine test services in `/test/services/` for examples

### For API Integration
1. Check [API Contracts](./api-contracts.md) for endpoint specifications
2. Review authentication and response formats
3. Use Swagger UI at `/swagger` for interactive testing
4. Monitor health via `/readyz` and `/livez` endpoints

### For Data Modeling
1. Study [Data Models](./data-models.md) for persistence patterns  
2. Understand ResourceInstance wrapper pattern
3. Review MongoDB integration and repository interfaces
4. Consider category system for specialized resources

---

**Generated on 2025-11-27 by BMM Document Project workflow**

*This documentation provides comprehensive coverage of the Endor SDK Go framework, enabling efficient development of scalable microservices with consistent patterns and robust integrations.*