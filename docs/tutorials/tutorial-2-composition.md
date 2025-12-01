# Tutorial 2: Advanced Composition - Multi-Service Hierarchies and Patterns

This tutorial demonstrates advanced service composition patterns using endor-sdk-go, including service hierarchies, embedding, and complex workflow orchestration.

---

## What You'll Build

By the end of this tutorial, you'll have created an **E-commerce Order System** that demonstrates:

- ✅ Service composition with ServiceChain and ServiceBranch
- ✅ EndorHybridService embedding multiple EndorServices  
- ✅ Complex workflow orchestration with error handling
- ✅ Service hierarchies with shared dependencies
- ✅ Testing composed services with realistic scenarios

**Estimated Time:** 45 minutes

---

## Prerequisites

- Completed [Tutorial 1: Building Your First Service](tutorial-1-first-service.md)
- Understanding of dependency injection concepts
- Familiarity with Go interfaces and composition

---

## Architecture Overview

We'll build a complete order processing system with these services:

```
OrderHybridService (Main Service)
├── AuthService (Embedded: "auth" prefix)
├── InventoryService (Embedded: "inventory" prefix)  
├── PaymentService (Embedded: "payment" prefix)
└── NotificationService (Embedded: "notifications" prefix)

Workflows:
OrderProcessingChain: Validation → Inventory → Payment → Fulfillment
NotificationBranch: Email + SMS + Push (parallel)
```

---

## Step 1: Define Domain Models

Create `internal/order/models.go`:

```go
package order

import (
    "time"
    "github.com/shopspring/decimal"
)

// Order represents a customer order
type Order struct {
    ID           string          `json:"id" bson:"_id"`
    CustomerID   string          `json:"customer_id" bson:"customer_id"`
    Items        []OrderItem     `json:"items" bson:"items"`
    Total        decimal.Decimal `json:"total" bson:"total"`
    Status       OrderStatus     `json:"status" bson:"status"`
    PaymentID    string          `json:"payment_id,omitempty" bson:"payment_id,omitempty"`
    ShippingAddr Address         `json:"shipping_address" bson:"shipping_address"`
    CreatedAt    time.Time       `json:"created_at" bson:"created_at"`
    UpdatedAt    time.Time       `json:"updated_at" bson:"updated_at"`
}

type OrderItem struct {
    ProductID   string          `json:"product_id" bson:"product_id"`
    ProductName string          `json:"product_name" bson:"product_name"`
    Quantity    int             `json:"quantity" bson:"quantity"`
    UnitPrice   decimal.Decimal `json:"unit_price" bson:"unit_price"`
    Subtotal    decimal.Decimal `json:"subtotal" bson:"subtotal"`
}

type Address struct {
    Street  string `json:"street" bson:"street"`
    City    string `json:"city" bson:"city"`
    State   string `json:"state" bson:"state"`
    ZipCode string `json:"zip_code" bson:"zip_code"`
    Country string `json:"country" bson:"country"`
}

type OrderStatus string

const (
    OrderStatusPending    OrderStatus = "pending"
    OrderStatusValidated  OrderStatus = "validated"
    OrderStatusPaid       OrderStatus = "paid"
    OrderStatusFulfilled  OrderStatus = "fulfilled"
    OrderStatusShipped    OrderStatus = "shipped"
    OrderStatusDelivered  OrderStatus = "delivered"
    OrderStatusCancelled  OrderStatus = "cancelled"
)

// Request/Response models
type CreateOrderRequest struct {
    CustomerID   string      `json:"customer_id" validate:"required"`
    Items        []OrderItem `json:"items" validate:"required,min=1"`
    ShippingAddr Address     `json:"shipping_address" validate:"required"`
}

type OrderProcessingResult struct {
    Order         Order                    `json:"order"`
    PaymentResult PaymentProcessingResult  `json:"payment_result"`
    Inventory     InventoryReservationResult `json:"inventory"`
    Notifications []NotificationResult     `json:"notifications"`
}

// Supporting models for embedded services
type Customer struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    IsActive bool   `json:"is_active"`
}

type PaymentProcessingResult struct {
    PaymentID     string          `json:"payment_id"`
    Status        string          `json:"status"`
    Amount        decimal.Decimal `json:"amount"`
    TransactionID string          `json:"transaction_id"`
    ProcessedAt   time.Time       `json:"processed_at"`
}

type InventoryReservationResult struct {
    ReservationID string               `json:"reservation_id"`
    Items         []InventoryReservation `json:"items"`
    ExpiresAt     time.Time            `json:"expires_at"`
}

type InventoryReservation struct {
    ProductID        string `json:"product_id"`
    RequestedQty     int    `json:"requested_qty"`
    ReservedQty      int    `json:"reserved_qty"`
    AvailableQty     int    `json:"available_qty"`
    PartiallyFulfilled bool `json:"partially_fulfilled"`
}

type NotificationResult struct {
    Channel string    `json:"channel"` // "email", "sms", "push"
    Status  string    `json:"status"`  // "sent", "failed", "pending"
    SentAt  time.Time `json:"sent_at"`
    Error   string    `json:"error,omitempty"`
}
```

---

## Step 2: Define Service Interfaces

Create `internal/order/interfaces.go`:

```go
package order

import (
    "context"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// Core order service interface
type OrderServiceInterface interface {
    CreateOrder(ctx context.Context, request CreateOrderRequest) (OrderProcessingResult, error)
    GetOrder(ctx context.Context, orderID string) (Order, error)
    UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error
    CancelOrder(ctx context.Context, orderID string) error
    ListOrders(ctx context.Context, customerID string) ([]Order, error)
}

// Authentication service interface
type AuthServiceInterface interface {
    ValidateCustomer(ctx context.Context, customerID string) (Customer, error)
    CheckPermissions(ctx context.Context, customerID string, action string) error
}

// Inventory service interface  
type InventoryServiceInterface interface {
    CheckAvailability(ctx context.Context, items []OrderItem) (bool, error)
    ReserveInventory(ctx context.Context, items []OrderItem) (InventoryReservationResult, error)
    ReleaseReservation(ctx context.Context, reservationID string) error
    ConfirmReservation(ctx context.Context, reservationID string) error
}

// Payment service interface
type PaymentServiceInterface interface {
    ProcessPayment(ctx context.Context, customerID string, amount decimal.Decimal, orderID string) (PaymentProcessingResult, error)
    RefundPayment(ctx context.Context, paymentID string, amount decimal.Decimal) error
    GetPaymentStatus(ctx context.Context, paymentID string) (string, error)
}

// Notification service interface  
type NotificationServiceInterface interface {
    SendOrderConfirmation(ctx context.Context, order Order, customer Customer) ([]NotificationResult, error)
    SendShippingUpdate(ctx context.Context, order Order, trackingNumber string) error
    SendPaymentReceipt(ctx context.Context, order Order, payment PaymentProcessingResult) error
}

// Order validation service interface
type OrderValidatorInterface interface {
    ValidateCreateRequest(request CreateOrderRequest) error
    ValidateOrderItems(items []OrderItem) error
    ValidateAddress(address Address) error
}
```

---

## Step 3: Implement Individual Services

### Authentication Service

Create `internal/auth/service.go`:

```go
package auth

import (
    "context"
    "fmt"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
    "github.com/yourname/ecommerce-service/internal/order"
)

type AuthService struct {
    repository interfaces.RepositoryInterface
    logger     interfaces.LoggerInterface
}

func NewAuthService(
    repository interfaces.RepositoryInterface,
    logger interfaces.LoggerInterface,
) order.AuthServiceInterface {
    return &AuthService{
        repository: repository,
        logger:     logger,
    }
}

// Implement EndorServiceInterface for embedding
func (s *AuthService) GetResource() string {
    return "auth"
}

func (s *AuthService) GetDescription() string {
    return "Customer authentication and authorization service"
}

func (s *AuthService) GetMethods() map[string]sdk.EndorServiceAction {
    return map[string]sdk.EndorServiceAction{
        "validate": sdk.NewAction(s.handleValidateCustomer, "Validate customer"),
        "check-permissions": sdk.NewAction(s.handleCheckPermissions, "Check permissions"),
    }
}

func (s *AuthService) ValidateCustomer(ctx context.Context, customerID string) (order.Customer, error) {
    s.logger.Debug("Validating customer", map[string]interface{}{
        "customer_id": customerID,
    })
    
    var customer order.Customer
    err := s.repository.FindByID(ctx, customerID, &customer)
    if err != nil {
        s.logger.Error("Customer validation failed", err, map[string]interface{}{
            "customer_id": customerID,
        })
        return order.Customer{}, fmt.Errorf("customer not found: %w", err)
    }
    
    if !customer.IsActive {
        return order.Customer{}, fmt.Errorf("customer account is inactive")
    }
    
    return customer, nil
}

func (s *AuthService) CheckPermissions(ctx context.Context, customerID string, action string) error {
    s.logger.Debug("Checking permissions", map[string]interface{}{
        "customer_id": customerID,
        "action":      action,
    })
    
    customer, err := s.ValidateCustomer(ctx, customerID)
    if err != nil {
        return err
    }
    
    // Simple permission check - in reality this would be more complex
    switch action {
    case "create_order", "view_order", "cancel_order":
        return nil // All active customers can perform these actions
    default:
        return fmt.Errorf("permission denied for action: %s", action)
    }
}

// HTTP handlers for EndorService interface
func (s *AuthService) handleValidateCustomer(c *gin.Context) {
    customerID := c.Param("customer_id")
    customer, err := s.ValidateCustomer(c.Request.Context(), customerID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, customer)
}

func (s *AuthService) handleCheckPermissions(c *gin.Context) {
    customerID := c.Param("customer_id")
    action := c.Query("action")
    
    err := s.CheckPermissions(c.Request.Context(), customerID, action)
    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"allowed": true})
}
```

### Inventory Service

Create `internal/inventory/service.go`:

```go
package inventory

import (
    "context"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
    "github.com/yourname/ecommerce-service/internal/order"
)

type InventoryService struct {
    repository interfaces.RepositoryInterface
    logger     interfaces.LoggerInterface
}

func NewInventoryService(
    repository interfaces.RepositoryInterface,
    logger interfaces.LoggerInterface,
) order.InventoryServiceInterface {
    return &InventoryService{
        repository: repository,
        logger:     logger,
    }
}

func (s *InventoryService) GetResource() string {
    return "inventory"
}

func (s *InventoryService) GetDescription() string {
    return "Product inventory management service"
}

func (s *InventoryService) GetMethods() map[string]sdk.EndorServiceAction {
    return map[string]sdk.EndorServiceAction{
        "check-availability": sdk.NewAction(s.handleCheckAvailability, "Check product availability"),
        "reserve": sdk.NewAction(s.handleReserveInventory, "Reserve inventory"),
        "release": sdk.NewAction(s.handleReleaseReservation, "Release reservation"),
        "confirm": sdk.NewAction(s.handleConfirmReservation, "Confirm reservation"),
    }
}

func (s *InventoryService) CheckAvailability(ctx context.Context, items []order.OrderItem) (bool, error) {
    s.logger.Debug("Checking inventory availability", map[string]interface{}{
        "item_count": len(items),
    })
    
    for _, item := range items {
        var product Product
        err := s.repository.FindByID(ctx, item.ProductID, &product)
        if err != nil {
            s.logger.Error("Product not found", err, map[string]interface{}{
                "product_id": item.ProductID,
            })
            return false, fmt.Errorf("product %s not found: %w", item.ProductID, err)
        }
        
        if product.AvailableQuantity < item.Quantity {
            s.logger.Info("Insufficient inventory", map[string]interface{}{
                "product_id": item.ProductID,
                "requested": item.Quantity,
                "available": product.AvailableQuantity,
            })
            return false, nil
        }
    }
    
    return true, nil
}

func (s *InventoryService) ReserveInventory(ctx context.Context, items []order.OrderItem) (order.InventoryReservationResult, error) {
    s.logger.Info("Reserving inventory", map[string]interface{}{
        "item_count": len(items),
    })
    
    reservationID := uuid.New().String()
    reservations := make([]order.InventoryReservation, len(items))
    
    for i, item := range items {
        var product Product
        err := s.repository.FindByID(ctx, item.ProductID, &product)
        if err != nil {
            return order.InventoryReservationResult{}, fmt.Errorf("product %s not found: %w", item.ProductID, err)
        }
        
        reservedQty := item.Quantity
        if product.AvailableQuantity < item.Quantity {
            reservedQty = product.AvailableQuantity
        }
        
        reservations[i] = order.InventoryReservation{
            ProductID:          item.ProductID,
            RequestedQty:       item.Quantity,
            ReservedQty:        reservedQty,
            AvailableQty:       product.AvailableQuantity,
            PartiallyFulfilled: reservedQty < item.Quantity,
        }
        
        // Update product availability
        product.AvailableQuantity -= reservedQty
        product.ReservedQuantity += reservedQty
        
        if err := s.repository.Update(ctx, product); err != nil {
            return order.InventoryReservationResult{}, fmt.Errorf("failed to update product inventory: %w", err)
        }
        
        // Create reservation record
        reservation := Reservation{
            ID:            reservationID,
            ProductID:     item.ProductID,
            ReservedQty:   reservedQty,
            Status:        "active",
            ExpiresAt:     time.Now().Add(15 * time.Minute), // 15 minute expiry
            CreatedAt:     time.Now(),
        }
        
        if err := s.repository.Create(ctx, reservation); err != nil {
            return order.InventoryReservationResult{}, fmt.Errorf("failed to create reservation: %w", err)
        }
    }
    
    result := order.InventoryReservationResult{
        ReservationID: reservationID,
        Items:         reservations,
        ExpiresAt:     time.Now().Add(15 * time.Minute),
    }
    
    s.logger.Info("Inventory reserved successfully", map[string]interface{}{
        "reservation_id": reservationID,
        "item_count":     len(items),
    })
    
    return result, nil
}

func (s *InventoryService) ReleaseReservation(ctx context.Context, reservationID string) error {
    s.logger.Info("Releasing inventory reservation", map[string]interface{}{
        "reservation_id": reservationID,
    })
    
    // Implementation details for releasing reservations...
    // This would restore inventory quantities and mark reservation as released
    
    return nil
}

func (s *InventoryService) ConfirmReservation(ctx context.Context, reservationID string) error {
    s.logger.Info("Confirming inventory reservation", map[string]interface{}{
        "reservation_id": reservationID,
    })
    
    // Implementation details for confirming reservations...
    // This would permanently deduct inventory and mark reservation as confirmed
    
    return nil
}

// Supporting models
type Product struct {
    ID                string `bson:"_id" json:"id"`
    Name              string `bson:"name" json:"name"`
    AvailableQuantity int    `bson:"available_quantity" json:"available_quantity"`
    ReservedQuantity  int    `bson:"reserved_quantity" json:"reserved_quantity"`
    TotalQuantity     int    `bson:"total_quantity" json:"total_quantity"`
}

type Reservation struct {
    ID          string    `bson:"_id" json:"id"`
    ProductID   string    `bson:"product_id" json:"product_id"`
    ReservedQty int       `bson:"reserved_qty" json:"reserved_qty"`
    Status      string    `bson:"status" json:"status"`
    ExpiresAt   time.Time `bson:"expires_at" json:"expires_at"`
    CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}
```

---

## Step 4: Implement Service Composition

### Order Processing Chain

Create `internal/order/workflow.go`:

```go
package order

import (
    "context"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/mattiabonardi/endor-sdk-go/sdk/composition"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// OrderProcessingWorkflow orchestrates the complete order processing flow
type OrderProcessingWorkflow struct {
    authService         AuthServiceInterface
    inventoryService    InventoryServiceInterface
    paymentService      PaymentServiceInterface
    notificationService NotificationServiceInterface
    repository          interfaces.RepositoryInterface
    logger              interfaces.LoggerInterface
}

func NewOrderProcessingWorkflow(
    authService AuthServiceInterface,
    inventoryService InventoryServiceInterface,
    paymentService PaymentServiceInterface,
    notificationService NotificationServiceInterface,
    repository interfaces.RepositoryInterface,
    logger interfaces.LoggerInterface,
) *OrderProcessingWorkflow {
    return &OrderProcessingWorkflow{
        authService:         authService,
        inventoryService:    inventoryService,
        paymentService:      paymentService,
        notificationService: notificationService,
        repository:          repository,
        logger:              logger,
    }
}

func (w *OrderProcessingWorkflow) ProcessOrder(ctx context.Context, request CreateOrderRequest) (OrderProcessingResult, error) {
    w.logger.Info("Starting order processing workflow", map[string]interface{}{
        "customer_id": request.CustomerID,
        "item_count":  len(request.Items),
    })
    
    // Create sequential processing chain
    processingChain := composition.ServiceChain(
        w.createAuthValidationStep(request.CustomerID),
        w.createInventoryReservationStep(request.Items),
        w.createPaymentProcessingStep(request),
        w.createOrderPersistenceStep(request),
    ).WithConfig(composition.CompositionConfig{
        Timeout:  60 * time.Second,
        FailFast: true, // Stop on first error
    })
    
    // Execute the processing chain
    result, err := processingChain.Execute(ctx, request)
    if err != nil {
        w.logger.Error("Order processing failed", err, map[string]interface{}{
            "customer_id": request.CustomerID,
        })
        
        // Handle cleanup on failure
        if err := w.handleProcessingFailure(ctx, request); err != nil {
            w.logger.Error("Failed to cleanup after processing failure", err)
        }
        
        return OrderProcessingResult{}, fmt.Errorf("order processing failed: %w", err)
    }
    
    orderResult := result.(OrderProcessingResult)
    
    // Send notifications in parallel (don't block order completion)
    go w.sendNotificationsAsync(ctx, orderResult)
    
    w.logger.Info("Order processing completed successfully", map[string]interface{}{
        "order_id":    orderResult.Order.ID,
        "customer_id": request.CustomerID,
        "total":       orderResult.Order.Total,
    })
    
    return orderResult, nil
}

// Individual processing steps
func (w *OrderProcessingWorkflow) createAuthValidationStep(customerID string) interfaces.EndorServiceInterface {
    return &AuthValidationStep{
        authService: w.authService,
        logger:      w.logger,
        customerID:  customerID,
    }
}

func (w *OrderProcessingWorkflow) createInventoryReservationStep(items []OrderItem) interfaces.EndorServiceInterface {
    return &InventoryReservationStep{
        inventoryService: w.inventoryService,
        logger:          w.logger,
        items:           items,
    }
}

func (w *OrderProcessingWorkflow) createPaymentProcessingStep(request CreateOrderRequest) interfaces.EndorServiceInterface {
    return &PaymentProcessingStep{
        paymentService: w.paymentService,
        logger:         w.logger,
        request:        request,
    }
}

func (w *OrderProcessingWorkflow) createOrderPersistenceStep(request CreateOrderRequest) interfaces.EndorServiceInterface {
    return &OrderPersistenceStep{
        repository: w.repository,
        logger:     w.logger,
        request:    request,
    }
}

// Parallel notification processing
func (w *OrderProcessingWorkflow) sendNotificationsAsync(ctx context.Context, result OrderProcessingResult) {
    // Create parallel notification branch
    notificationBranch := composition.ServiceBranch(
        w.createEmailNotificationStep(result),
        w.createSMSNotificationStep(result),
        w.createPushNotificationStep(result),
    ).WithConfig(composition.CompositionConfig{
        Timeout:    30 * time.Second,
        RequireAll: false, // Allow partial notification success
    })
    
    notifications, err := notificationBranch.Execute(ctx, result)
    if err != nil {
        w.logger.Error("Notification processing failed", err, map[string]interface{}{
            "order_id": result.Order.ID,
        })
        return
    }
    
    w.logger.Info("Notifications sent successfully", map[string]interface{}{
        "order_id":         result.Order.ID,
        "notification_count": len(notifications.([]NotificationResult)),
    })
}

func (w *OrderProcessingWorkflow) handleProcessingFailure(ctx context.Context, request CreateOrderRequest) error {
    // Implement cleanup logic for failed orders
    // - Release inventory reservations
    // - Refund payments if processed
    // - Mark order as failed
    return nil
}

// Individual processing step implementations
type AuthValidationStep struct {
    authService interfaces.EndorServiceInterface
    logger      interfaces.LoggerInterface
    customerID  string
}

func (s *AuthValidationStep) Execute(ctx context.Context, data interface{}) (interface{}, error) {
    s.logger.Debug("Validating customer authentication", map[string]interface{}{
        "customer_id": s.customerID,
    })
    
    authSvc := s.authService.(AuthServiceInterface)
    customer, err := authSvc.ValidateCustomer(ctx, s.customerID)
    if err != nil {
        return nil, fmt.Errorf("customer validation failed: %w", err)
    }
    
    // Check permissions
    if err := authSvc.CheckPermissions(ctx, s.customerID, "create_order"); err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    
    // Pass through original data with customer info
    result := data.(CreateOrderRequest)
    return AuthValidationResult{
        Request:  result,
        Customer: customer,
    }, nil
}

type AuthValidationResult struct {
    Request  CreateOrderRequest
    Customer Customer
}

type InventoryReservationStep struct {
    inventoryService InventoryServiceInterface
    logger          interfaces.LoggerInterface
    items           []OrderItem
}

func (s *InventoryReservationStep) Execute(ctx context.Context, data interface{}) (interface{}, error) {
    authResult := data.(AuthValidationResult)
    
    s.logger.Debug("Reserving inventory", map[string]interface{}{
        "item_count": len(s.items),
    })
    
    // Check availability first
    available, err := s.inventoryService.CheckAvailability(ctx, s.items)
    if err != nil {
        return nil, fmt.Errorf("availability check failed: %w", err)
    }
    if !available {
        return nil, fmt.Errorf("insufficient inventory for requested items")
    }
    
    // Reserve inventory
    reservation, err := s.inventoryService.ReserveInventory(ctx, s.items)
    if err != nil {
        return nil, fmt.Errorf("inventory reservation failed: %w", err)
    }
    
    return InventoryReservationResult{
        AuthResult:  authResult,
        Reservation: reservation,
    }, nil
}

type InventoryReservationResult struct {
    AuthResult  AuthValidationResult
    Reservation InventoryReservationResult
}
```

---

## Step 5: Create EndorHybridService with Service Embedding

Create `internal/order/hybrid_service.go`:

```go
package order

import (
    "context"
    "github.com/mattiabonardi/endor-sdk-go/sdk"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// NewOrderHybridService creates a comprehensive order management service with embedded services
func NewOrderHybridService(
    repository interfaces.RepositoryInterface,
    authService interfaces.EndorServiceInterface,
    inventoryService interfaces.EndorServiceInterface,
    paymentService interfaces.EndorServiceInterface,
    notificationService interfaces.EndorServiceInterface,
    logger interfaces.LoggerInterface,
) interfaces.EndorHybridServiceInterface {
    
    // Create hybrid service with automatic CRUD for orders
    hybridService := sdk.NewHybridService[Order]("orders", "Comprehensive order management system").
        WithCategories([]sdk.EndorHybridServiceCategory{
            // Priority orders category for expedited processing
            sdk.NewEndorHybridServiceCategory[Order, PriorityOrder](priorityOrderCategory),
        }).
        WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
            return map[string]sdk.EndorServiceAction{
                "process": sdk.NewAction(handleOrderProcessing, "Process complete order workflow"),
                "bulk-import": sdk.NewAction(handleBulkOrderImport, "Bulk order import"),
                "analytics": sdk.NewAction(handleOrderAnalytics, "Order analytics and reporting"),
            }
        })
    
    // Embed individual services with appropriate prefixes
    if err := hybridService.EmbedService("auth", authService); err != nil {
        logger.Error("Failed to embed auth service", err)
    }
    
    if err := hybridService.EmbedService("inventory", inventoryService); err != nil {
        logger.Error("Failed to embed inventory service", err)
    }
    
    if err := hybridService.EmbedService("payment", paymentService); err != nil {
        logger.Error("Failed to embed payment service", err)
    }
    
    if err := hybridService.EmbedService("notifications", notificationService); err != nil {
        logger.Error("Failed to embed notification service", err)
    }
    
    logger.Info("Order hybrid service created with embedded services", map[string]interface{}{
        "embedded_services": []string{"auth", "inventory", "payment", "notifications"},
    })
    
    return hybridService
}

// Result: Comprehensive API with clear organization
//
// Automatic Order CRUD:
//   POST   /orders              - Create order (with full workflow)
//   GET    /orders              - List orders
//   GET    /orders/:id          - Get order by ID  
//   PUT    /orders/:id          - Update order
//   DELETE /orders/:id          - Cancel order
//   POST   /orders/process      - Process order workflow
//   POST   /orders/bulk-import  - Bulk import orders
//   GET    /orders/analytics    - Order analytics
//
// Embedded Authentication (with "auth" prefix):
//   GET    /orders/auth/validate/:customer_id    - Validate customer
//   GET    /orders/auth/check-permissions        - Check permissions
//
// Embedded Inventory (with "inventory" prefix):
//   GET    /orders/inventory/check-availability  - Check availability
//   POST   /orders/inventory/reserve            - Reserve inventory
//   DELETE /orders/inventory/reserve/:id        - Release reservation
//   POST   /orders/inventory/confirm/:id        - Confirm reservation
//
// Embedded Payment (with "payment" prefix):
//   POST   /orders/payment/process              - Process payment
//   POST   /orders/payment/refund               - Refund payment
//   GET    /orders/payment/status/:id           - Get payment status
//
// Embedded Notifications (with "notifications" prefix):
//   POST   /orders/notifications/order-confirmation  - Send confirmation
//   POST   /orders/notifications/shipping-update     - Send shipping update
//   POST   /orders/notifications/payment-receipt     - Send payment receipt

// Custom action handlers
func handleOrderProcessing(c *gin.Context) {
    // Implementation for complete order processing workflow
    var request CreateOrderRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Get workflow from context (injected via DI)
    workflow := c.MustGet("orderWorkflow").(*OrderProcessingWorkflow)
    
    result, err := workflow.ProcessOrder(c.Request.Context(), request)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, result)
}

func handleBulkOrderImport(c *gin.Context) {
    // Implementation for bulk order import
    var requests []CreateOrderRequest
    if err := c.ShouldBindJSON(&requests); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Process orders in batches using service composition
    // Implementation details...
    
    c.JSON(http.StatusOK, gin.H{
        "imported": len(requests),
        "status": "success",
    })
}

func handleOrderAnalytics(c *gin.Context) {
    // Implementation for order analytics
    customerID := c.Query("customer_id")
    dateRange := c.Query("date_range")
    
    // Generate analytics using composed services
    // Implementation details...
    
    analytics := OrderAnalytics{
        TotalOrders:    100,
        TotalRevenue:   decimal.NewFromFloat(50000.00),
        AverageOrder:   decimal.NewFromFloat(500.00),
        TopProducts:    []string{"product-1", "product-2"},
    }
    
    c.JSON(http.StatusOK, analytics)
}

// Supporting models
type PriorityOrder struct {
    Order
    PriorityLevel   string    `json:"priority_level" bson:"priority_level"`
    ExpectedDelivery time.Time `json:"expected_delivery" bson:"expected_delivery"`
    SpecialHandling  string    `json:"special_handling" bson:"special_handling"`
}

type OrderAnalytics struct {
    TotalOrders   int             `json:"total_orders"`
    TotalRevenue  decimal.Decimal `json:"total_revenue"`
    AverageOrder  decimal.Decimal `json:"average_order"`
    TopProducts   []string        `json:"top_products"`
}

var priorityOrderCategory = Category{
    ID:          "priority",
    Name:        "Priority Orders",
    Description: "High-priority orders with expedited processing",
}
```

---

## Step 6: Testing Service Composition

Create `internal/order/composition_test.go`:

```go
package order

import (
    "context"
    "testing"
    "time"
    "errors"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/shopspring/decimal"
    "github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
)

//go:build unit

func TestOrderProcessingWorkflow_Success(t *testing.T) {
    // Set up all mocked services
    mockAuth := testutils.NewMockAuthService()
    mockInventory := testutils.NewMockInventoryService() 
    mockPayment := testutils.NewMockPaymentService()
    mockNotification := testutils.NewMockNotificationService()
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    
    // Create test data
    request := CreateOrderRequest{
        CustomerID: "customer-123",
        Items: []OrderItem{
            {
                ProductID:   "product-1",
                ProductName: "Test Product",
                Quantity:    2,
                UnitPrice:   decimal.NewFromFloat(50.00),
                Subtotal:    decimal.NewFromFloat(100.00),
            },
        },
        ShippingAddr: Address{
            Street:  "123 Test St",
            City:    "Test City",
            State:   "TX",
            ZipCode: "12345",
            Country: "US",
        },
    }
    
    customer := Customer{
        ID:       "customer-123",
        Name:     "Test Customer",
        Email:    "test@example.com",
        IsActive: true,
    }
    
    reservation := InventoryReservationResult{
        ReservationID: "reservation-123",
        Items: []InventoryReservation{
            {
                ProductID:    "product-1",
                RequestedQty: 2,
                ReservedQty:  2,
                AvailableQty: 10,
            },
        },
        ExpiresAt: time.Now().Add(15 * time.Minute),
    }
    
    payment := PaymentProcessingResult{
        PaymentID:     "payment-123",
        Status:        "completed",
        Amount:        decimal.NewFromFloat(100.00),
        TransactionID: "txn-123",
        ProcessedAt:   time.Now(),
    }
    
    // Set up mock expectations in order of execution
    mockAuth.On("ValidateCustomer", mock.Any, "customer-123").Return(customer, nil)
    mockAuth.On("CheckPermissions", mock.Any, "customer-123", "create_order").Return(nil)
    
    mockInventory.On("CheckAvailability", mock.Any, request.Items).Return(true, nil)
    mockInventory.On("ReserveInventory", mock.Any, request.Items).Return(reservation, nil)
    
    mockPayment.On("ProcessPayment", mock.Any, "customer-123", decimal.NewFromFloat(100.00), mock.Any).Return(payment, nil)
    
    mockRepo.On("Create", mock.Any, mock.MatchedBy(func(order Order) bool {
        return order.CustomerID == "customer-123" && order.Status == OrderStatusPaid
    })).Return(nil)
    
    mockLogger.On("Info", mock.Any, mock.Any).Return()
    mockLogger.On("Debug", mock.Any, mock.Any).Return()
    
    // Create workflow and execute
    workflow := NewOrderProcessingWorkflow(
        mockAuth, mockInventory, mockPayment, mockNotification, mockRepo, mockLogger,
    )
    
    result, err := workflow.ProcessOrder(context.Background(), request)
    
    // Verify results
    assert.NoError(t, err)
    assert.Equal(t, "customer-123", result.Order.CustomerID)
    assert.Equal(t, OrderStatusPaid, result.Order.Status)
    assert.Equal(t, "payment-123", result.PaymentResult.PaymentID)
    assert.Equal(t, "reservation-123", result.Inventory.ReservationID)
    
    // Verify all services were called in correct order
    mockAuth.AssertExpectations(t)
    mockInventory.AssertExpectations(t)
    mockPayment.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
}

func TestOrderProcessingWorkflow_AuthFailure(t *testing.T) {
    // Test authentication failure stops the workflow
    mockAuth := testutils.NewMockAuthService()
    mockInventory := testutils.NewMockInventoryService()
    mockPayment := testutils.NewMockPaymentService()
    mockNotification := testutils.NewMockNotificationService()
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    
    request := CreateOrderRequest{
        CustomerID: "inactive-customer",
        Items:      []OrderItem{{ProductID: "product-1", Quantity: 1}},
    }
    
    authError := errors.New("customer account is inactive")
    mockAuth.On("ValidateCustomer", mock.Any, "inactive-customer").Return(Customer{}, authError)
    mockLogger.On("Info", mock.Any, mock.Any).Return()
    mockLogger.On("Error", mock.Any, mock.Any, mock.Any).Return()
    
    workflow := NewOrderProcessingWorkflow(
        mockAuth, mockInventory, mockPayment, mockNotification, mockRepo, mockLogger,
    )
    
    result, err := workflow.ProcessOrder(context.Background(), request)
    
    // Verify error handling
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "order processing failed")
    assert.Contains(t, err.Error(), "customer account is inactive")
    assert.Equal(t, OrderProcessingResult{}, result)
    
    // Verify downstream services were not called
    mockInventory.AssertNotCalled(t, "CheckAvailability")
    mockPayment.AssertNotCalled(t, "ProcessPayment")
    mockRepo.AssertNotCalled(t, "Create")
}

func TestOrderProcessingWorkflow_InventoryFailure(t *testing.T) {
    // Test inventory failure with cleanup
    mockAuth := testutils.NewMockAuthService()
    mockInventory := testutils.NewMockInventoryService()
    mockPayment := testutils.NewMockPaymentService()
    mockNotification := testutils.NewMockNotificationService()
    mockRepo := testutils.NewMockRepository()
    mockLogger := testutils.NewMockLogger()
    
    request := CreateOrderRequest{
        CustomerID: "customer-123",
        Items: []OrderItem{
            {ProductID: "out-of-stock", Quantity: 5},
        },
    }
    
    customer := Customer{ID: "customer-123", IsActive: true}
    
    // Auth succeeds
    mockAuth.On("ValidateCustomer", mock.Any, "customer-123").Return(customer, nil)
    mockAuth.On("CheckPermissions", mock.Any, "customer-123", "create_order").Return(nil)
    
    // Inventory fails
    mockInventory.On("CheckAvailability", mock.Any, request.Items).Return(false, nil)
    
    mockLogger.On("Info", mock.Any, mock.Any).Return()
    mockLogger.On("Debug", mock.Any, mock.Any).Return()
    mockLogger.On("Error", mock.Any, mock.Any, mock.Any).Return()
    
    workflow := NewOrderProcessingWorkflow(
        mockAuth, mockInventory, mockPayment, mockNotification, mockRepo, mockLogger,
    )
    
    result, err := workflow.ProcessOrder(context.Background(), request)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient inventory")
    assert.Equal(t, OrderProcessingResult{}, result)
    
    // Verify payment and persistence were not called
    mockPayment.AssertNotCalled(t, "ProcessPayment")
    mockRepo.AssertNotCalled(t, "Create")
}

func TestOrderHybridService_ServiceEmbedding(t *testing.T) {
    // Test service embedding in hybrid service
    mockRepo := testutils.NewMockRepository()
    mockAuth := testutils.NewMockEndorService()
    mockInventory := testutils.NewMockEndorService()
    mockPayment := testutils.NewMockEndorService()
    mockNotification := testutils.NewMockEndorService()
    mockLogger := testutils.NewMockLogger()
    
    // Configure embedded service mocks
    mockAuth.On("GetResource").Return("auth")
    mockAuth.On("GetMethods").Return(map[string]sdk.EndorServiceAction{
        "validate": sdk.NewAction(mockValidateHandler, "Validate customer"),
    })
    
    mockInventory.On("GetResource").Return("inventory")
    mockInventory.On("GetMethods").Return(map[string]sdk.EndorServiceAction{
        "check-availability": sdk.NewAction(mockCheckAvailabilityHandler, "Check availability"),
    })
    
    mockPayment.On("GetResource").Return("payment")
    mockNotification.On("GetResource").Return("notifications")
    
    mockLogger.On("Info", mock.Any, mock.Any).Return()
    mockLogger.On("Error", mock.Any, mock.Any, mock.Any).Return()
    
    // Create hybrid service with embedded services
    hybridService := NewOrderHybridService(
        mockRepo, mockAuth, mockInventory, mockPayment, mockNotification, mockLogger,
    )
    
    // Verify service embedding
    embeddedServices := hybridService.GetEmbeddedServices()
    
    assert.Len(t, embeddedServices, 4)
    assert.Equal(t, mockAuth, embeddedServices["auth"])
    assert.Equal(t, mockInventory, embeddedServices["inventory"])
    assert.Equal(t, mockPayment, embeddedServices["payment"])
    assert.Equal(t, mockNotification, embeddedServices["notifications"])
    
    // Convert to EndorService and verify method resolution
    endorService := hybridService.ToEndorService(testSchema)
    methods := endorService.GetMethods()
    
    // Should include automatic CRUD + embedded methods + custom actions
    assert.Contains(t, methods, "create")                        // Automatic CRUD
    assert.Contains(t, methods, "auth.validate")                 // Embedded auth
    assert.Contains(t, methods, "inventory.check-availability")  // Embedded inventory
    assert.Contains(t, methods, "process")                       // Custom action
    assert.Contains(t, methods, "analytics")                     // Custom action
}

func TestServiceComposition_ParallelNotifications(t *testing.T) {
    // Test parallel notification processing
    mockEmail := testutils.NewMockNotificationService()
    mockSMS := testutils.NewMockNotificationService()
    mockPush := testutils.NewMockNotificationService()
    
    order := Order{
        ID:         "order-123",
        CustomerID: "customer-123",
        Total:      decimal.NewFromFloat(100.00),
    }
    
    customer := Customer{
        ID:    "customer-123",
        Email: "test@example.com",
        Phone: "+1234567890",
    }
    
    emailResult := NotificationResult{Channel: "email", Status: "sent", SentAt: time.Now()}
    smsResult := NotificationResult{Channel: "sms", Status: "sent", SentAt: time.Now()}
    pushResult := NotificationResult{Channel: "push", Status: "sent", SentAt: time.Now()}
    
    // All notifications should execute in parallel
    mockEmail.On("SendOrderConfirmation", mock.Any, order, customer).
        Return([]NotificationResult{emailResult}, nil).
        After(10 * time.Millisecond)
        
    mockSMS.On("SendOrderConfirmation", mock.Any, order, customer).
        Return([]NotificationResult{smsResult}, nil).
        After(15 * time.Millisecond)
        
    mockPush.On("SendOrderConfirmation", mock.Any, order, customer).
        Return([]NotificationResult{pushResult}, nil).
        After(5 * time.Millisecond)
    
    // Create parallel notification branch
    notificationBranch := composition.ServiceBranch(mockEmail, mockSMS, mockPush).
        WithConfig(composition.CompositionConfig{
            Timeout:    1 * time.Second,
            RequireAll: false,
        })
    
    start := time.Now()
    results, err := notificationBranch.Execute(context.Background(), map[string]interface{}{
        "order":    order,
        "customer": customer,
    })
    duration := time.Since(start)
    
    assert.NoError(t, err)
    assert.NotNil(t, results)
    
    // Should complete in roughly 15ms (slowest service), not 30ms (sum of all)
    assert.Less(t, duration, 25*time.Millisecond)
    assert.Greater(t, duration, 14*time.Millisecond)
    
    mockEmail.AssertExpectations(t)
    mockSMS.AssertExpectations(t)
    mockPush.AssertExpectations(t)
}
```

---

## Step 7: Integration Testing

Create `internal/order/integration_test.go`:

```go
//go:build integration

package order

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/shopspring/decimal"
    "github.com/mattiabonardi/endor-sdk-go/sdk/di"
)

func TestOrderProcessingWorkflow_Integration(t *testing.T) {
    // Set up integration test container with real services
    container := di.NewContainer()
    setupOrderIntegrationDependencies(container, t)
    
    // Resolve workflow with real implementations
    workflow, err := di.Resolve[*OrderProcessingWorkflow](container)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Test complete order workflow with real services
    request := CreateOrderRequest{
        CustomerID: "integration-customer-123",
        Items: []OrderItem{
            {
                ProductID:   "integration-product-1",
                ProductName: "Integration Test Product",
                Quantity:    3,
                UnitPrice:   decimal.NewFromFloat(25.99),
                Subtotal:    decimal.NewFromFloat(77.97),
            },
        },
        ShippingAddr: Address{
            Street:  "123 Integration Test Ave",
            City:    "Test City",
            State:   "CA",
            ZipCode: "90210",
            Country: "US",
        },
    }
    
    // Execute order processing workflow
    result, err := workflow.ProcessOrder(ctx, request)
    require.NoError(t, err)
    
    // Verify order was processed successfully
    assert.Equal(t, request.CustomerID, result.Order.CustomerID)
    assert.Equal(t, OrderStatusPaid, result.Order.Status)
    assert.NotEmpty(t, result.Order.ID)
    assert.Equal(t, decimal.NewFromFloat(77.97), result.Order.Total)
    
    // Verify payment was processed
    assert.Equal(t, "completed", result.PaymentResult.Status)
    assert.Equal(t, decimal.NewFromFloat(77.97), result.PaymentResult.Amount)
    assert.NotEmpty(t, result.PaymentResult.PaymentID)
    
    // Verify inventory was reserved
    assert.NotEmpty(t, result.Inventory.ReservationID)
    assert.Len(t, result.Inventory.Items, 1)
    assert.Equal(t, 3, result.Inventory.Items[0].ReservedQty)
    
    // Verify notifications were sent (may be asynchronous)
    time.Sleep(2 * time.Second) // Allow time for async notifications
    
    // Retrieve order to verify persistence
    retrievedOrder, err := workflow.GetOrder(ctx, result.Order.ID)
    require.NoError(t, err)
    assert.Equal(t, result.Order.ID, retrievedOrder.ID)
    assert.Equal(t, result.Order.CustomerID, retrievedOrder.CustomerID)
}

func TestOrderHybridService_Integration_CompleteAPI(t *testing.T) {
    // Integration test for hybrid service with embedded services
    container := di.NewContainer()
    setupOrderIntegrationDependencies(container, t)
    
    // Resolve hybrid service
    hybridService, err := di.Resolve[interfaces.EndorHybridServiceInterface](container)
    require.NoError(t, err)
    
    // Convert to EndorService for API testing
    endorService := hybridService.ToEndorService(testSchema)
    
    // Set up test HTTP server
    router := gin.New()
    handler := NewOrderHandler(endorService)
    
    v1 := router.Group("/api/v1")
    orders := v1.Group("/orders")
    
    // Register all routes including embedded services
    orders.POST("", handler.CreateOrder)                           // Main CRUD
    orders.GET("", handler.ListOrders)                             // Main CRUD
    orders.GET("/:id", handler.GetOrder)                           // Main CRUD
    orders.POST("/process", handler.ProcessOrder)                   // Custom action
    orders.GET("/analytics", handler.GetAnalytics)                 // Custom action
    
    orders.GET("/auth/validate/:customer_id", handler.ValidateCustomer)      // Embedded auth
    orders.GET("/inventory/check-availability", handler.CheckAvailability)   // Embedded inventory
    orders.POST("/payment/process", handler.ProcessPayment)                  // Embedded payment
    orders.POST("/notifications/order-confirmation", handler.SendNotification) // Embedded notifications
    
    server := httptest.NewServer(router)
    defer server.Close()
    
    // Test complete order processing via HTTP API
    orderJSON := `{
        "customer_id": "integration-customer-456",
        "items": [
            {
                "product_id": "integration-product-2",
                "product_name": "API Test Product",
                "quantity": 2,
                "unit_price": 49.99,
                "subtotal": 99.98
            }
        ],
        "shipping_address": {
            "street": "456 API Test Blvd",
            "city": "API City",
            "state": "NY",
            "zip_code": "10001",
            "country": "US"
        }
    }`
    
    // 1. Test order processing endpoint
    resp, err := http.Post(server.URL+"/api/v1/orders/process", "application/json", strings.NewReader(orderJSON))
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    var result OrderProcessingResult
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    assert.Equal(t, "integration-customer-456", result.Order.CustomerID)
    assert.Equal(t, OrderStatusPaid, result.Order.Status)
    
    // 2. Test embedded auth service endpoint
    resp, err = http.Get(server.URL + "/api/v1/orders/auth/validate/integration-customer-456")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var customer Customer
    err = json.NewDecoder(resp.Body).Decode(&customer)
    require.NoError(t, err)
    assert.Equal(t, "integration-customer-456", customer.ID)
    
    // 3. Test embedded inventory service endpoint
    checkJSON := `{
        "items": [
            {
                "product_id": "integration-product-2",
                "quantity": 5
            }
        ]
    }`
    
    resp, err = http.Post(server.URL+"/api/v1/orders/inventory/check-availability", "application/json", strings.NewReader(checkJSON))
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var availability struct {
        Available bool `json:"available"`
    }
    err = json.NewDecoder(resp.Body).Decode(&availability)
    require.NoError(t, err)
    assert.True(t, availability.Available)
    
    // 4. Test order retrieval via main CRUD endpoint
    resp, err = http.Get(server.URL + "/api/v1/orders/" + result.Order.ID)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var retrievedOrder Order
    err = json.NewDecoder(resp.Body).Decode(&retrievedOrder)
    require.NoError(t, err)
    assert.Equal(t, result.Order.ID, retrievedOrder.ID)
}
```

---

## What You've Accomplished

✅ **Built a complex multi-service system** with clear service boundaries  
✅ **Implemented ServiceChain** for sequential order processing workflow  
✅ **Used ServiceBranch** for parallel notification processing  
✅ **Created EndorHybridService** with embedded services for comprehensive API  
✅ **Tested service composition** with realistic scenarios and error handling  
✅ **Demonstrated service hierarchies** with shared dependencies  

### Key Composition Patterns Mastered

1. **Sequential Processing**: ServiceChain for order workflow steps
2. **Parallel Processing**: ServiceBranch for independent notifications  
3. **Service Embedding**: EndorHybridService embedding multiple EndorServices
4. **Error Propagation**: Proper error handling across service boundaries
5. **Dependency Sharing**: Multiple services sharing common dependencies

### Benefits Achieved

| Aspect | Before Composition | After Composition | Improvement |
|--------|------------------|------------------|-------------|
| **API Organization** | Scattered endpoints | Hierarchical with namespaces | 🚀 Clear structure |
| **Code Reusability** | Duplicated logic | Shared service implementations | 🚀 High reuse |
| **Testing Isolation** | Monolithic tests | Independent service tests | 🚀 Focused testing |
| **Error Handling** | Mixed contexts | Clear error boundaries | 🚀 Better debugging |
| **Team Ownership** | Unclear boundaries | Clear service ownership | 🚀 Better collaboration |

---

## Next Steps

Ready to explore more advanced patterns?

1. **[Tutorial 3: Testing Strategies](testing-strategies.md)** - Advanced testing for composed services
2. **[Tutorial 4: Performance Optimization](tutorial-4-performance.md)** - Optimize service composition  
3. **[Migration Guide](migration-guide.md)** - Migrate existing services to composition patterns

### Advanced Patterns to Explore

- **Service Proxy**: Add caching and circuit breakers
- **Event-Driven Composition**: Use events between services
- **Distributed Transactions**: Handle complex multi-service transactions
- **Service Discovery**: Dynamic service registration and discovery
- **Load Balancing**: Distribute load across service instances

Congratulations on mastering advanced service composition! 🎉