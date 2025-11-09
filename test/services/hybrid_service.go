package services_test

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

// Example of a hybrid resource - User model with static fields
type User struct {
	ID       string `json:"id" bson:"_id" schema:"title=ID,readOnly=true"`
	Username string `json:"username" bson:"username" schema:"title=Username,required=true"`
	Email    string `json:"email" bson:"email" schema:"title=Email,format=email,required=true"`
	Name     string `json:"name" bson:"name" schema:"title=Full Name"`
}

// Implement ResourceInstanceInterface
func (u *User) GetID() *string {
	return &u.ID
}

func (u *User) SetID(id string) {
	u.ID = id
}

func NewHybridService() sdk.EndorService {
	return sdk.CreateHybridEndorService("hybrid-test", "Hybrid Test", func(service *sdk.EndorHybridService[*User]) error {
		service.SetCreateHandler(func(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstance[*User]]]) (*sdk.Response[sdk.ResourceInstance[*User]], error) {
			// Custom business logic for user creation
			fmt.Printf("Creating user: %+v\n", c.Payload.Data.This)

			// Validation
			if c.Payload.Data.This.Email == "" {
				return nil, sdk.NewBadRequestError(fmt.Errorf("email is required"))
			}

			// Call repository directly or use the default logic
			created, err := service.GetRepository().Create(context.TODO(), c.Payload)
			if err != nil {
				return nil, err
			}

			fmt.Printf("User created successfully: %s\n", *created.GetID())

			return sdk.NewResponseBuilder[sdk.ResourceInstance[*User]]().
				AddData(created).
				AddSchema(service.GetRootSchema()).
				AddMessage(sdk.NewMessage(sdk.Info, "User created successfully")).
				Build(), nil
		})

		// Add custom action
		service.AddCustomAction("activate", sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInstance[*User]], error) {
				// Get user instance
				instance, err := service.GetRepository().Instance(context.TODO(), c.Payload)
				if err != nil {
					return nil, err
				}

				// Add activation logic
				if instance.Metadata == nil {
					instance.Metadata = make(map[string]any)
				}
				instance.Metadata["activated"] = true
				instance.Metadata["activatedAt"] = "2024-01-01T00:00:00Z"

				// Update instance
				updateDTO := sdk.UpdateByIdDTO[sdk.ResourceInstance[*User]]{
					Id:   *instance.GetID(),
					Data: *instance,
				}

				updated, err := service.GetRepository().Update(context.TODO(), updateDTO)
				if err != nil {
					return nil, err
				}

				return sdk.NewResponseBuilder[sdk.ResourceInstance[*User]]().
					AddData(updated).
					AddSchema(service.GetRootSchema()).
					AddMessage(sdk.NewMessage(sdk.Info, "User activated successfully")).
					Build(), nil
			},
			"Activate user account",
		))

		return nil
	})
}
