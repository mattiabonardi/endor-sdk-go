# Endor SDK Go - API Contracts

## Overview

The Endor SDK provides two distinct API contract patterns: **Static Service APIs** with manually defined endpoints and **Dynamic Hybrid Service APIs** with automatic CRUD operations and schema-driven endpoints.

## Core API Patterns

### Authentication

All API endpoints expect authentication via header:

```
Header: Session: {session-token}
```

The session information is automatically parsed and injected into the `EndorContext` for each request.

### Response Format

All endpoints follow a standardized response structure:

```go
type Response[T any] struct {
    Messages []Message   `json:"messages"`
    Data     *T          `json:"data"`
    Schema   *RootSchema `json:"schema"`
}

type Message struct {
    Gravity MessageGravity `json:"gravity"` // Info, Warning, Error, Fatal
    Value   string         `json:"value"`
}
```

## Static Service APIs (EndorService)

Static services define custom endpoints with manual registration:

### Service Definition Pattern

```go
type EndorService struct {
    Resource    string                           // Resource path prefix
    Description string                           // Service description
    Methods     map[string]EndorServiceAction    // Action definitions
    Priority    *int                            // Registration priority
    Version     string                          // API version
}
```

### Action Definition

```go
func NewAction[T any, R any](
    handler EndorHandlerFunc[T, R], 
    description string
) EndorServiceAction

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)
```

### Example Static Service Implementation

```go
func NewService1() EndorService {
    return EndorService{
        Resource:    "service1",
        Description: "Example static service",
        Methods: map[string]EndorServiceAction{
            "customAction": NewAction(handleCustomAction, "Custom business logic"),
            "process": NewConfigurableAction(EndorServiceActionOptions{
                Description:     "Process data",
                Public:          true,
                ValidatePayload: true,
            }, handleProcess),
        },
    }
}
```

### Generated Endpoints

For a static service with resource `"service1"`:

- `POST /service1/customAction` - Custom action endpoint
- `POST /service1/process` - Process action endpoint

## Dynamic Hybrid Service APIs (EndorHybridService)

Hybrid services provide automatic CRUD operations with MongoDB integration:

### Standard CRUD Endpoints

For each hybrid service resource, the following endpoints are automatically generated:

#### Schema Endpoint
```
GET /{resource}/schema
```

**Response:**
```json
{
    "messages": [],
    "data": {
        "type": "object",
        "properties": {...},
        "required": [...]
    },
    "schema": null
}
```

#### List Resources
```
GET /{resource}/list
```

**Query Parameters:**
- `limit`: Maximum number of results
- `skip`: Number of results to skip
- `sort`: Sort field and direction

**Response:**
```json
{
    "messages": [],
    "data": [
        {
            "This": {...},
            "Metadata": {...}
        }
    ],
    "schema": {...}
}
```

#### Get Single Resource
```
GET /{resource}/instance/{id}
```

**Response:**
```json
{
    "messages": [],
    "data": {
        "This": {...},
        "Metadata": {...}
    },
    "schema": {...}
}
```

#### Create Resource
```
POST /{resource}/instance
```

**Request Body:**
```json
{
    "This": {...},
    "Metadata": {...}
}
```

#### Update Resource
```
PUT /{resource}/instance/{id}
```

**Request Body:**
```json
{
    "This": {...},
    "Metadata": {...}
}
```

#### Delete Resource
```
DELETE /{resource}/instance/{id}
```

### Category-Based APIs

For services with category specialization:

#### Category-Specific CRUD
```
GET /{resource}/categories/{categoryId}/schema
GET /{resource}/categories/{categoryId}/list
GET /{resource}/categories/{categoryId}/instance/{id}
POST /{resource}/categories/{categoryId}/instance
PUT /{resource}/categories/{categoryId}/instance/{id}
DELETE /{resource}/categories/{categoryId}/instance/{id}
```

### Custom Actions in Hybrid Services

Hybrid services can register additional custom actions:

```go
func NewService2() EndorHybridService {
    return NewHybridService("products", "Product management").
        WithActions(func(getSchema func() RootSchema) map[string]EndorServiceAction {
            return map[string]EndorServiceAction{
                "validate": NewAction(validateProduct, "Validate product data"),
                "publish": NewAction(publishProduct, "Publish product to catalog"),
            }
        })
}
```

Generated endpoints:
- `POST /products/validate`
- `POST /products/publish`

## Built-in System Endpoints

### Health Checks

```
GET /readyz
GET /livez
```

**Response:**
```json
{
    "status": "ok"
}
```

### Metrics

```
GET /metrics
```

Returns Prometheus-formatted metrics.

### API Documentation

```
GET /swagger
```

Serves interactive Swagger UI for all registered endpoints.

## Error Responses

### Error Format

```json
{
    "messages": [
        {
            "gravity": "Error",
            "value": "Validation failed: required field missing"
        }
    ],
    "data": null,
    "schema": null
}
```

### HTTP Status Codes

| Status | Usage |
|--------|-------|
| 200 | Successful operation |
| 201 | Resource created |
| 400 | Bad request / validation error |
| 401 | Authentication required |
| 403 | Forbidden / authorization failed |
| 404 | Resource not found |
| 500 | Internal server error |

### Common Error Types

#### Validation Errors
```json
{
    "messages": [
        {
            "gravity": "Error", 
            "value": "Validation failed for field 'email': invalid format"
        }
    ]
}
```

#### Authorization Errors
```json
{
    "messages": [
        {
            "gravity": "Error",
            "value": "Insufficient permissions for this operation"
        }
    ]
}
```

#### Not Found Errors
```json
{
    "messages": [
        {
            "gravity": "Error",
            "value": "Resource with ID '12345' not found"
        }
    ]
}
```

## Request Context

### EndorContext Structure

All handlers receive a typed context:

```go
type EndorContext[T any] struct {
    MicroServiceId         string      // Service identifier
    Session                Session     // Authentication session
    Payload                T           // Request payload (typed)
    ResourceMetadataSchema RootSchema  // Schema for validation
    CategoryID             *string     // Category ID (if applicable)
    GinContext             *gin.Context // Underlying Gin context
}
```

### Session Information

```go
type Session struct {
    UserID       string
    Permissions  []string
    TokenExpiry  time.Time
    // Additional session data
}
```

## Schema System

### Automatic Schema Generation

Schemas are automatically generated from Go struct types using reflection:

```go
type ProductRequest struct {
    Name        string  `json:"name" validate:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" validate:"min=0"`
    CategoryID  string  `json:"categoryId" validate:"required"`
}
```

Generated schema:
```json
{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "description": {"type": "string"},
        "price": {"type": "number", "minimum": 0},
        "categoryId": {"type": "string"}
    },
    "required": ["name", "categoryId"]
}
```

### Schema Validation

- Automatic payload validation before handler execution
- JSON Schema Draft 7 compliance
- Custom validation tags support
- Error messages with field-specific details

## API Gateway Integration

### Traefik Configuration

The SDK automatically generates API gateway configuration:

```yaml
http:
  services:
    endor-service:
      loadBalancer:
        servers:
          - url: "http://service:8080"
  routers:
    endor-service:
      rule: "PathPrefix(`/api/v1/service`)"
      service: endor-service
      middlewares:
        - auth-middleware
```

### Service Discovery

Services are registered with the API gateway including:
- Health check endpoints
- Authentication requirements
- Rate limiting configuration
- Load balancing strategy

---

*This API contract documentation is automatically maintained and reflects the current state of the SDK's endpoint generation and routing capabilities.*