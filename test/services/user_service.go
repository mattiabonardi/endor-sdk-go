package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// User represents the base user model
type User struct {
	ID       string `json:"id" bson:"_id" schema:"title=User ID,readOnly=true"`
	Email    string `json:"email" bson:"email" schema:"title=Email,format=email"`
	Name     string `json:"name" bson:"name" schema:"title=Full Name"`
	Status   string `json:"status" bson:"status" schema:"title=Status,enum=active|inactive|suspended"`
	CreateAt string `json:"createAt" bson:"createAt" schema:"title=Created At,readOnly=true"`
}

func (u *User) GetID() *string {
	return &u.ID
}

func (u *User) SetID(id string) {
	u.ID = id
}

// AdminUser represents admin-specific attributes (category specialization)
type AdminUser struct {
	CategoryType    string   `json:"categoryType" bson:"categoryType" schema:"title=Category Type,readOnly=true"`
	AdminLevel      int      `json:"adminLevel" bson:"adminLevel" schema:"title=Admin Level,minimum=1,maximum=5"`
	Permissions     []string `json:"permissions" bson:"permissions" schema:"title=Permissions"`
	LastAdminAction string   `json:"lastAdminAction" bson:"lastAdminAction" schema:"title=Last Admin Action,readOnly=true"`
}

func (a *AdminUser) GetCategoryType() *string {
	return &a.CategoryType
}

func (a *AdminUser) SetCategoryType(categoryType string) {
	a.CategoryType = categoryType
}

// PremiumUser represents premium user attributes (category specialization)
type PremiumUser struct {
	CategoryType     string   `json:"categoryType" bson:"categoryType" schema:"title=Category Type,readOnly=true"`
	SubscriptionType string   `json:"subscriptionType" bson:"subscriptionType" schema:"title=Subscription Type,enum=monthly|yearly"`
	ExpiryDate       string   `json:"expiryDate" bson:"expiryDate" schema:"title=Subscription Expiry,format=date"`
	PremiumFeatures  []string `json:"premiumFeatures" bson:"premiumFeatures" schema:"title=Premium Features"`
}

func (p *PremiumUser) GetCategoryType() *string {
	return &p.CategoryType
}

func (p *PremiumUser) SetCategoryType(categoryType string) {
	p.CategoryType = categoryType
}

// Additional schemas for dynamic attributes (stored in MongoDB)
type AdminAdditionalSchema struct {
	Department    string   `json:"department" schema:"title=Department"`
	SecurityLevel string   `json:"securityLevel" schema:"title=Security Level,enum=low|medium|high"`
	AccessRegions []string `json:"accessRegions" schema:"title=Access Regions"`
}

type PremiumAdditionalSchema struct {
	BillingAddress string `json:"billingAddress" schema:"title=Billing Address"`
	PaymentMethod  string `json:"paymentMethod" schema:"title=Payment Method,enum=card|paypal|bank"`
	ReferralCode   string `json:"referralCode" schema:"title=Referral Code"`
}

// Custom action payloads
type PromoteUserPayload struct {
	UserID     string `json:"userId" schema:"title=User ID"`
	AdminLevel int    `json:"adminLevel" schema:"title=Admin Level,minimum=1,maximum=5"`
}

type SendNotificationPayload struct {
	UserID  string `json:"userId" schema:"title=User ID"`
	Message string `json:"message" schema:"title=Message"`
	Type    string `json:"type" schema:"title=Type,enum=info|warning|urgent"`
}

type UserBulkOperationPayload struct {
	UserIDs   []string `json:"userIds" schema:"title=User IDs"`
	Operation string   `json:"operation" schema:"title=Operation,enum=activate|deactivate|delete"`
}

// UserService implements the service logic
type UserService struct {
	// You can add service-specific dependencies here if needed
	// These would be injected through the DI container in production
}

// Custom action implementations
func (s *UserService) promoteUser(c *sdk.EndorContext[PromoteUserPayload]) (*sdk.Response[any], error) {
	// In a real implementation, you would:
	// 1. Validate the user exists
	// 2. Check permissions
	// 3. Create admin category instance
	// 4. Update user status

	var result any = map[string]interface{}{
		"userId":     c.Payload.UserID,
		"promoted":   true,
		"adminLevel": c.Payload.AdminLevel,
		"message":    "User successfully promoted to admin",
	}
	return sdk.NewResponseBuilder[any]().
		AddData(&result).
		AddMessage(sdk.NewMessage(sdk.Info, "User promoted successfully")).
		Build(), nil
}

func (s *UserService) sendNotification(c *sdk.EndorContext[SendNotificationPayload]) (*sdk.Response[any], error) {
	// In a real implementation, you would:
	// 1. Validate user exists
	// 2. Check notification preferences
	// 3. Send notification via appropriate channel
	// 4. Log notification

	var result any = map[string]interface{}{
		"userId":           c.Payload.UserID,
		"notificationSent": true,
		"type":             c.Payload.Type,
		"timestamp":        "2025-12-01T10:00:00Z",
	}
	return sdk.NewResponseBuilder[any]().
		AddData(&result).
		AddMessage(sdk.NewMessage(sdk.Info, "Notification sent successfully")).
		Build(), nil
}

func (s *UserService) bulkOperations(c *sdk.EndorContext[UserBulkOperationPayload]) (*sdk.Response[any], error) {
	// In a real implementation, you would:
	// 1. Validate all user IDs
	// 2. Check permissions for bulk operations
	// 3. Execute operation in transaction
	// 4. Return operation results

	processedCount := len(c.Payload.UserIDs)

	var result any = map[string]interface{}{
		"operation":      c.Payload.Operation,
		"processedCount": processedCount,
		"userIds":        c.Payload.UserIDs,
		"timestamp":      "2025-12-01T10:00:00Z",
	}
	return sdk.NewResponseBuilder[any]().
		AddData(&result).
		AddMessage(sdk.NewMessage(sdk.Info, "Bulk operation completed successfully")).
		Build(), nil
}

// NewUserService creates a new hybrid user service with categories and custom actions
func NewUserService() sdk.EndorHybridService {
	service := &UserService{}

	// Convert additional schemas to YAML for MongoDB storage
	adminAdditionalSchema, _ := sdk.NewSchema(AdminAdditionalSchema{}).ToYAML()
	premiumAdditionalSchema, _ := sdk.NewSchema(PremiumAdditionalSchema{}).ToYAML()

	return sdk.NewHybridService[*User]("users", "User Management System").
		WithCategories([]sdk.EndorHybridServiceCategory{
			// Admin category - for administrative users
			sdk.NewEndorHybridServiceCategory[*User, *AdminUser](sdk.Category{
				ID:                   "admin",
				Description:          "Administrative Users",
				AdditionalAttributes: adminAdditionalSchema,
			}),
			// Premium category - for premium subscribers
			sdk.NewEndorHybridServiceCategory[*User, *PremiumUser](sdk.Category{
				ID:                   "premium",
				Description:          "Premium Users",
				AdditionalAttributes: premiumAdditionalSchema,
			}),
		}).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
			return map[string]sdk.EndorServiceAction{
				// User promotion action
				"promote-user": sdk.NewAction(
					service.promoteUser,
					"Promote a user to admin with specified level",
				),
				// Notification action
				"send-notification": sdk.NewAction(
					service.sendNotification,
					"Send notification to a specific user",
				),
				// Bulk operations action
				"bulk-operations": sdk.NewAction(
					service.bulkOperations,
					"Perform bulk operations on multiple users",
				),
			}
		})
}

// Example of creating the service with dependency injection (production usage)
func NewUserServiceWithDependencies(
	repository interfaces.RepositoryPattern,
	config interfaces.ConfigProviderInterface,
	logger interfaces.LoggerInterface,
) sdk.EndorHybridService {
	// Create service with explicit dependencies using the DI pattern
	deps := sdk.EndorHybridServiceDependencies{
		Repository: repository,
		Config:     config,
		Logger:     logger,
	}

	// Create the base service with dependencies
	baseService, err := sdk.NewEndorHybridServiceWithDeps[*User](
		"users",
		"User Management System with Dependency Injection",
		deps,
	)
	if err != nil {
		// In production, you'd handle this error appropriately
		// For this example, fall back to simple version
		return NewUserService()
	}

	// Add categories and custom actions to the DI-enabled service
	adminAdditionalSchema, _ := sdk.NewSchema(AdminAdditionalSchema{}).ToYAML()
	premiumAdditionalSchema, _ := sdk.NewSchema(PremiumAdditionalSchema{}).ToYAML()

	service := &UserService{}

	return baseService.WithCategories([]sdk.EndorHybridServiceCategory{
		sdk.NewEndorHybridServiceCategory[*User, *AdminUser](sdk.Category{
			ID:                   "admin",
			Description:          "Administrative Users",
			AdditionalAttributes: adminAdditionalSchema,
		}),
		sdk.NewEndorHybridServiceCategory[*User, *PremiumUser](sdk.Category{
			ID:                   "premium",
			Description:          "Premium Users",
			AdditionalAttributes: premiumAdditionalSchema,
		}),
	}).WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
		return map[string]sdk.EndorServiceAction{
			"promote-user": sdk.NewAction(
				service.promoteUser,
				"Promote a user to admin with specified level",
			),
			"send-notification": sdk.NewAction(
				service.sendNotification,
				"Send notification to a specific user",
			),
			"bulk-operations": sdk.NewAction(
				service.bulkOperations,
				"Perform bulk operations on multiple users",
			),
		}
	})
}

// Example of creating the service using DI Container (modern production pattern)
func NewUserServiceFromContainer(container di.Container) (sdk.EndorHybridService, error) {
	// Let the container automatically resolve all dependencies
	baseService, err := sdk.NewEndorHybridServiceFromContainer[*User](
		container,
		"users",
		"User Management System with DI Container",
	)
	if err != nil {
		return nil, err
	}

	// Add categories and custom actions
	adminAdditionalSchema, _ := sdk.NewSchema(AdminAdditionalSchema{}).ToYAML()
	premiumAdditionalSchema, _ := sdk.NewSchema(PremiumAdditionalSchema{}).ToYAML()

	service := &UserService{}

	return baseService.WithCategories([]sdk.EndorHybridServiceCategory{
		sdk.NewEndorHybridServiceCategory[*User, *AdminUser](sdk.Category{
			ID:                   "admin",
			Description:          "Administrative Users",
			AdditionalAttributes: adminAdditionalSchema,
		}),
		sdk.NewEndorHybridServiceCategory[*User, *PremiumUser](sdk.Category{
			ID:                   "premium",
			Description:          "Premium Users",
			AdditionalAttributes: premiumAdditionalSchema,
		}),
	}).WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
		return map[string]sdk.EndorServiceAction{
			"promote-user": sdk.NewAction(
				service.promoteUser,
				"Promote a user to admin with specified level",
			),
			"send-notification": sdk.NewAction(
				service.sendNotification,
				"Send notification to a specific user",
			),
			"bulk-operations": sdk.NewAction(
				service.bulkOperations,
				"Perform bulk operations on multiple users",
			),
		}
	}), nil
}
