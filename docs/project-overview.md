# Endor SDK Go - Project Overview

## Executive Summary

The **Endor SDK Go** is a sophisticated backend framework designed for building microservices within the Endor ecosystem. It provides a complete abstraction layer for creating RESTful APIs with automatic CRUD operations, dynamic schema generation, MongoDB integration, and API gateway configuration.

## Project Structure

- **Type:** Backend Library/SDK
- **Architecture:** Service-based SDK for building Endor microservices  
- **Primary Language:** Go (v1.21.4)
- **Framework:** Gin HTTP framework
- **Database:** MongoDB with automatic schema management
- **Repository Type:** Monolith (single cohesive SDK)

## Technology Stack

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| Language | Go | 1.21.4 | Core development language |
| HTTP Framework | Gin | v1.10.0 | REST API routing and middleware |
| Database | MongoDB | v1.17.3 | Document storage with dynamic schemas |
| Metrics | Prometheus | v1.21.0 | Application monitoring |
| Configuration | godotenv | v1.5.1 | Environment configuration |
| Documentation | Swagger/OpenAPI | 3.1 | Automatic API documentation |

## Architecture Type

**Service Framework Pattern** - The SDK implements a service framework architecture that enables developers to build consistent microservices by:

1. **Service Abstractions**: Two service types (EndorService for static, EndorHybridService for dynamic)
2. **Repository Pattern**: Generic data access layer with MongoDB backend
3. **Schema System**: Dynamic JSON schema generation from Go structs
4. **API Gateway Integration**: Automatic route configuration for distributed systems

## Core Components

### Service Layer
- **EndorService**: Traditional static services with predefined endpoints
- **EndorHybridService**: Dynamic services that adapt to MongoDB schema definitions
- **Category System**: Resource specialization through configurable categories

### Data Layer  
- Generic repository interfaces with MongoDB implementations
- Support for both wrapped (ResourceInstance) and direct model access
- Specialized repositories for category-based resources

### API Management
- Automatic CRUD endpoint generation
- Dynamic API gateway configuration (Traefik format) 
- Built-in authentication middleware integration
- Swagger/OpenAPI 3.1 documentation generation

## Key Features

- **Type Safety**: Extensive use of Go generics for compile-time validation
- **Convention over Configuration**: Sensible defaults with customization points
- **Schema-Driven Development**: Dynamic schema generation and validation
- **Microservice-Ready**: Built-in API gateway integration and health checks
- **Developer Experience**: Automatic documentation and tooling integration

## Getting Started

The SDK follows a builder pattern for initialization:

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

## Documentation Index

### Core Documentation
- [Architecture Details](./architecture.md)
- [API Contracts](./api-contracts.md) 
- [Data Models](./data-models.md)
- [Development Guide](./development-guide.md)
- [Source Tree Analysis](./source-tree-analysis.md)

### Supporting Documentation
- [Component Inventory](./component-inventory.md)
- [Configuration Reference](./configuration-reference.md)
- [Testing Strategy](./testing-guide.md)

### Existing Documentation
- [Workflow Status](./bmm-workflow-status.yaml) - BMM methodology progress tracking

---

*Generated on 2025-11-27 by BMM Document Project workflow*