package sdk_resource

import (
	"context"
	"fmt"
	"maps"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorHybridSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface] struct {
	ID          string
	Description string
	ActionFn    func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface
}

func (h *EndorHybridSpecializedServiceCategory[T]) GetID() string {
	return h.ID
}

// GetActions implements sdk.EndorHybridSpecializedServiceCategoryInterface.
func (h *EndorHybridSpecializedServiceCategory[T]) GetActions() func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface {
	return h.ActionFn
}

func (h *EndorHybridSpecializedServiceCategory[T]) WithActions(actionFn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface) sdk.EndorHybridSpecializedServiceCategoryInterface {
	h.ActionFn = actionFn
	return h
}

func (h *EndorHybridSpecializedServiceCategory[T]) CreateDefaultActions(resource string, resourceDescription string, metadataSchema sdk.Schema, categoryMetadataSchema sdk.Schema) map[string]sdk.EndorServiceActionInterface {
	rootSchemWithCategory := getCategorySchemaWithMetadata[T](metadataSchema, categoryMetadataSchema)
	return getDefaultActionsForCategory[T](resource, *rootSchemWithCategory, resourceDescription, h.ID)
}

func NewEndorHybridSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface](categoryID string, categoryDescription string) sdk.EndorHybridSpecializedServiceCategoryInterface {
	return &EndorHybridSpecializedServiceCategory[T]{
		ID:          categoryID,
		Description: categoryDescription,
	}
}

type EndorHybridSpecializedService[T sdk.ResourceInstanceSpecializedInterface] struct {
	Resource            string
	ResourceDescription string
	Priority            *int
	methodsFn           func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface
	categories          map[string]sdk.EndorHybridSpecializedServiceCategoryInterface
}

func (h EndorHybridSpecializedService[T]) GetResource() string {
	return h.Resource
}

func (h EndorHybridSpecializedService[T]) GetResourceDescription() string {
	return h.ResourceDescription
}

func (h EndorHybridSpecializedService[T]) GetPriority() *int {
	return h.Priority
}

func NewHybridSpecializedService[T sdk.ResourceInstanceSpecializedInterface](resource, resourceDescription string) sdk.EndorHybridSpecializedServiceInterface {
	return EndorHybridSpecializedService[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
	}
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridSpecializedService[T]) WithActions(
	fn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface,
) sdk.EndorHybridSpecializedServiceInterface {
	h.methodsFn = fn
	return h
}

func (h EndorHybridSpecializedService[T]) WithCategories(categories []sdk.EndorHybridSpecializedServiceCategoryInterface) sdk.EndorHybridSpecializedServiceInterface {
	if h.categories == nil {
		h.categories = make(map[string]sdk.EndorHybridSpecializedServiceCategoryInterface)
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
	}
	return h
}

// create endor service instance
func (h EndorHybridSpecializedService[T]) ToEndorService(metadataSchema sdk.Schema, categoriesMetadataShema map[string]sdk.Schema) sdk.EndorService {
	var methods = make(map[string]sdk.EndorServiceActionInterface)

	// schema
	rootSchemaWithMetadata := getRootSchemaWithMetadata[T](metadataSchema)
	getSchemaCallback := func() sdk.RootSchema { return *rootSchemaWithMetadata }

	// add default CRUD methods
	methods = getDefaultActions[T](h.Resource, *rootSchemaWithMetadata, h.ResourceDescription)
	// remove delete and update
	delete(methods, "create")
	delete(methods, "update")
	// add custom methods
	if h.methodsFn != nil {
		maps.Copy(methods, h.methodsFn(getSchemaCallback))
	}

	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for categoryID, category := range h.categories {
			// add default CRUD methods specified for category
			categoryMethods := category.CreateDefaultActions(h.Resource, h.ResourceDescription, metadataSchema, categoriesMetadataShema[categoryID])
			maps.Copy(methods, categoryMethods)
		}
	}

	return sdk.EndorService{
		Resource:            h.Resource,
		ResourceDescription: h.ResourceDescription,
		Priority:            h.Priority,
		Actions:             methods,
	}
}

func getCategorySchemaWithMetadata[T sdk.ResourceInstanceSpecializedInterface](metadataSchema sdk.Schema, categoryMetadataSchema sdk.Schema) *sdk.RootSchema {
	// create root schema
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	if metadataSchema.Properties != nil {
		maps.Copy((*rootSchema.Properties), *metadataSchema.Properties)
	}

	// add category metadata schema (dynamic)
	if categoryMetadataSchema.Properties != nil {
		maps.Copy((*rootSchema.Properties), *categoryMetadataSchema.Properties)
	}

	return rootSchema
}

func getDefaultActionsForCategory[T sdk.ResourceInstanceSpecializedInterface](resource string, schema sdk.RootSchema, resourceDescription string, categoryID string) map[string]sdk.EndorServiceActionInterface {
	autogenerateID := true

	repository := repository.NewResourceInstanceRepository[T](
		resource,
		repository.ResourceInstanceRepositoryOptions{AutoGenerateID: &autogenerateID},
	)

	return map[string]sdk.EndorServiceActionInterface{
		categoryID + "/schema": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/list": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadDTO]) (*sdk.Response[[]sdk.ResourceInstance[T]], error) {
				return defaultListSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/create": sdk.NewConfigurableAction(
			sdk.EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &sdk.RootSchema{
					Schema: sdk.Schema{
						Type: sdk.SchemaTypeObject,
						Properties: &map[string]sdk.Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T]]]) (*sdk.Response[sdk.ResourceInstance[T]], error) {
				return defaultCreateSpecialized(c, schema, repository, resource)
			},
		),
		categoryID + "/instance": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[*sdk.ResourceInstance[T]], error) {
				return defaultInstanceSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/update": sdk.NewConfigurableAction(
			sdk.EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &sdk.RootSchema{
					Schema: sdk.Schema{
						Type: sdk.SchemaTypeObject,
						Properties: &map[string]sdk.Schema{
							"id": {
								Type: sdk.SchemaTypeString,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T]]]) (*sdk.Response[sdk.ResourceInstance[T]], error) {
				return defaultUpdateSpecialized(c, schema, repository, resource)
			},
		),
	}
}

func defaultListSpecialized[T sdk.ResourceInstanceSpecializedInterface](c *sdk.EndorContext[sdk.ReadDTO], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T]) (*sdk.Response[[]sdk.ResourceInstance[T]], error) {
	categoryFilter := map[string]interface{}{"type": c.CategoryType}
	if len(c.Payload.Filter) > 0 {
		c.Payload.Filter = map[string]interface{}{
			"$and": []interface{}{
				categoryFilter,
				c.Payload.Filter,
			},
		}
	} else {
		c.Payload.Filter = categoryFilter
	}
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.ResourceInstance[T]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreateSpecialized[T sdk.ResourceInstanceSpecializedInterface](c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T]]], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T], resource string) (*sdk.Response[sdk.ResourceInstance[T]], error) {
	c.Payload.Data.SetCategoryType(c.CategoryType)
	if resourceInstance, ok := any(c.Payload.Data).(sdk.ResourceInstance[T]); ok {
		created, err := repository.Create(context.TODO(), sdk.CreateDTO[sdk.ResourceInstance[T]]{
			Data: resourceInstance,
		})
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[sdk.ResourceInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s %s created", resource, *created.GetID()))).Build(), nil
	}
	return nil, fmt.Errorf("invalid type assertion")
}

func defaultInstanceSpecialized[T sdk.ResourceInstanceSpecializedInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T]) (*sdk.Response[*sdk.ResourceInstance[T]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[*sdk.ResourceInstance[T]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultUpdateSpecialized[T sdk.ResourceInstanceSpecializedInterface](c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T]]], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T], resource string) (*sdk.Response[sdk.ResourceInstance[T]], error) {
	c.Payload.Data.SetCategoryType(c.CategoryType)
	if resourceInstance, ok := any(c.Payload.Data).(sdk.ResourceInstance[T]); ok {
		updated, err := repository.Update(context.TODO(), sdk.UpdateByIdDTO[sdk.ResourceInstance[T]]{
			Id:   c.Payload.Id,
			Data: resourceInstance,
		})
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[sdk.ResourceInstance[T]]().AddData(updated).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s updated (category)", resource))).Build(), nil
	}
	return nil, fmt.Errorf("invalid type assertion")
}
