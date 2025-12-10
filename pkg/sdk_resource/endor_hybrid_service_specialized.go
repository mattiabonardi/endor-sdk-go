package sdk_resource

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorHybridSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface, C any] struct {
	Category sdk.HybridCategory
}

func (h *EndorHybridSpecializedServiceCategory[T, C]) GetID() string {
	return h.Category.ID
}

func (h *EndorHybridSpecializedServiceCategory[T, C]) CreateDefaultActions(resource string, resourceDescription string, metadataSchema sdk.Schema) map[string]sdk.EndorServiceActionInterface {
	rootSchemWithCategory := getCategorySchemaWithMetadata[T, C](metadataSchema, h.Category)
	return getDefaultActionsForCategory[T, C](resource, *rootSchemWithCategory, resourceDescription, h.Category.ID)
}

func NewEndorHybridSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface, C any](category sdk.HybridCategory) sdk.EndorHybridSpecializedServiceCategoryInterface {
	return &EndorHybridSpecializedServiceCategory[T, C]{
		Category: category,
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
func (h EndorHybridSpecializedService[T]) ToEndorService(metadataSchema sdk.Schema) sdk.EndorService {
	var methods = make(map[string]sdk.EndorServiceActionInterface)

	// schema
	rootSchemWithMetadata := getRootSchemaWithMetadata[T](metadataSchema)
	getSchemaCallback := func() sdk.RootSchema { return *rootSchemWithMetadata }

	// add default CRUD methods
	methods = getDefaultActions[T](h.Resource, *rootSchemWithMetadata, h.ResourceDescription)
	// remove delete and update
	delete(methods, "create")
	delete(methods, "update")
	// add custom methods
	if h.methodsFn != nil {
		for methodName, method := range h.methodsFn(getSchemaCallback) {
			methods[methodName] = method
		}
	}

	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for _, category := range h.categories {
			// add default CRUD methods specified for category
			categoryMethods := category.CreateDefaultActions(h.Resource, h.ResourceDescription, metadataSchema)
			for methodName, method := range categoryMethods {
				methods[methodName] = method
			}
		}
	}

	return sdk.EndorService{
		Resource:            h.Resource,
		ResourceDescription: h.ResourceDescription,
		Priority:            h.Priority,
		Actions:             methods,
	}
}

func getCategorySchemaWithMetadata[T sdk.ResourceInstanceSpecializedInterface, C any](metadataSchema sdk.Schema, category sdk.HybridCategory) *sdk.RootSchema {
	// create root schema
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	if metadataSchema.Properties != nil {
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}

	// add category base schema (hardcoded)
	var categoryBaseModel C
	categoryBaseSchema := sdk.NewSchema(categoryBaseModel)
	if categoryBaseSchema.Properties != nil {
		for k, v := range *categoryBaseSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}

	// add category metadata schema (dynamic)
	categoryMetadataSchema, _ := category.UnmarshalAdditionalAttributes()
	if categoryMetadataSchema.Properties != nil {
		for k, v := range *categoryMetadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}

	return rootSchema
}

func getDefaultActionsForCategory[T sdk.ResourceInstanceSpecializedInterface, C any](resource string, schema sdk.RootSchema, resourceDescription string, categoryID string) map[string]sdk.EndorServiceActionInterface {
	autogenerateID := true

	repository := repository.NewResourceInstanceSpecializedRepository[T, C](
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
			func(c *sdk.EndorContext[sdk.ReadDTO]) (*sdk.Response[[]sdk.ResourceInstanceSpecialized[T, C]], error) {
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
			func(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T, C]]]) (*sdk.Response[sdk.ResourceInstanceSpecialized[T, C]], error) {
				return defaultCreateSpecialized(c, schema, repository, resource)
			},
		),
		categoryID + "/instance": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[*sdk.ResourceInstanceSpecialized[T, C]], error) {
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
			func(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T, C]]]) (*sdk.Response[sdk.ResourceInstanceSpecialized[T, C]], error) {
				return defaultUpdateSpecialized(c, schema, repository, resource)
			},
		),
	}
}

func defaultListSpecialized[T sdk.ResourceInstanceSpecializedInterface, C any](c *sdk.EndorContext[sdk.ReadDTO], schema sdk.RootSchema, repository *repository.ResourceInstanceSpecializedRepository[T, C]) (*sdk.Response[[]sdk.ResourceInstanceSpecialized[T, C]], error) {
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
	return sdk.NewResponseBuilder[[]sdk.ResourceInstanceSpecialized[T, C]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreateSpecialized[T sdk.ResourceInstanceSpecializedInterface, C any](c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T, C]]], schema sdk.RootSchema, repository *repository.ResourceInstanceSpecializedRepository[T, C], resource string) (*sdk.Response[sdk.ResourceInstanceSpecialized[T, C]], error) {
	c.Payload.Data.SetCategoryType(c.CategoryType)
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInstanceSpecialized[T, C]]().AddData(created).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s %s created", resource, *created.GetID()))).Build(), nil
}

func defaultInstanceSpecialized[T sdk.ResourceInstanceSpecializedInterface, C any](c *sdk.EndorContext[sdk.ReadInstanceDTO], schema sdk.RootSchema, repository *repository.ResourceInstanceSpecializedRepository[T, C]) (*sdk.Response[*sdk.ResourceInstanceSpecialized[T, C]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[*sdk.ResourceInstanceSpecialized[T, C]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultUpdateSpecialized[T sdk.ResourceInstanceSpecializedInterface, C any](c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T, C]]], schema sdk.RootSchema, repository *repository.ResourceInstanceSpecializedRepository[T, C], resource string) (*sdk.Response[sdk.ResourceInstanceSpecialized[T, C]], error) {
	c.Payload.Data.SetCategoryType(c.CategoryType)
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInstanceSpecialized[T, C]]().AddData(updated).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s updated (category)", resource))).Build(), nil
}
