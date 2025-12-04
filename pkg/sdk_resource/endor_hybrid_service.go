package sdk_resource

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorHybridServiceCategoryImpl[T sdk.ResourceInstanceInterface, C sdk.ResourceInstanceSpecializedInterface] struct {
	Category sdk.Category
}

func (h *EndorHybridServiceCategoryImpl[T, C]) GetID() string {
	return h.Category.ID
}

func (h *EndorHybridServiceCategoryImpl[T, C]) CreateDefaultActions(resource string, resourceDescription string, metadataSchema sdk.Schema) map[string]sdk.EndorServiceAction {
	rootSchemWithCategory := getCategorySchemaWithMetadata[T, C](metadataSchema, h.Category)
	return getDefaultActionsForCategory[T, C](resource, *rootSchemWithCategory, resourceDescription, h.Category.ID)
}

func NewEndorHybridServiceCategory[T sdk.ResourceInstanceInterface, R sdk.ResourceInstanceSpecializedInterface](category sdk.Category) sdk.EndorHybridServiceCategory {
	return &EndorHybridServiceCategoryImpl[T, R]{
		Category: category,
	}
}

type EndorHybridServiceImpl[T sdk.ResourceInstanceInterface] struct {
	Resource            string
	ResourceDescription string
	Priority            *int
	methodsFn           func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction
	categories          map[string]sdk.EndorHybridServiceCategory
}

func (h EndorHybridServiceImpl[T]) GetResource() string {
	return h.Resource
}

func (h EndorHybridServiceImpl[T]) GetResourceDescription() string {
	return h.ResourceDescription
}

func (h EndorHybridServiceImpl[T]) GetPriority() *int {
	return h.Priority
}

func NewHybridService[T sdk.ResourceInstanceInterface](resource, resourceDescription string) sdk.EndorHybridService {
	return EndorHybridServiceImpl[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
	}
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridServiceImpl[T]) WithActions(
	fn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction,
) sdk.EndorHybridService {
	h.methodsFn = fn
	return h
}

func (h EndorHybridServiceImpl[T]) WithCategories(categories []sdk.EndorHybridServiceCategory) sdk.EndorHybridService {
	if h.categories == nil {
		h.categories = make(map[string]sdk.EndorHybridServiceCategory)
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
	}
	return h
}

// create endor service instance
func (h EndorHybridServiceImpl[T]) ToEndorService(metadataSchema sdk.Schema) sdk.EndorService {
	var methods = make(map[string]sdk.EndorServiceAction)

	// schema
	rootSchemWithMetadata := getRootSchemaWithMetadata[T](metadataSchema)
	getSchemaCallback := func() sdk.RootSchema { return *rootSchemWithMetadata }

	// add default CRUD methods
	methods = getDefaultActions[T](h.Resource, *rootSchemWithMetadata, h.ResourceDescription)
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
		Resource:    h.Resource,
		Description: h.ResourceDescription,
		Priority:    h.Priority,
		Methods:     methods,
	}
}

func getRootSchemaWithMetadata[T sdk.ResourceInstanceInterface](metadataSchema sdk.Schema) *sdk.RootSchema {
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	if metadataSchema.Properties != nil {
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	return rootSchema
}

func getCategorySchemaWithMetadata[T sdk.ResourceInstanceInterface, C sdk.ResourceInstanceSpecializedInterface](metadataSchema sdk.Schema, category sdk.Category) *sdk.RootSchema {
	// create root schema
	rootSchema := getRootSchemaWithMetadata[T](metadataSchema)

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

func getDefaultActions[T sdk.ResourceInstanceInterface](resource string, schema RootSchema, resourceDescription string) map[string]EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := NewResourceInstanceRepository[T](resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", resource, resourceDescription),
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[T]], error) {
				return defaultInstance(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", resource, resourceDescription),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[T]], error) {
				return defaultList(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", resource, resourceDescription),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", resource, resourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return defaultCreate(c, schema, repository, resource)
			},
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", resource, resourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]]) (*Response[ResourceInstance[T]], error) {
				return defaultUpdate(c, schema, repository, resource)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDelete(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", resource, resourceDescription),
		),
	}
}

func defaultSchema[T ResourceInstanceInterface](_ *EndorContext[NoPayload], schema RootSchema) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance[T ResourceInstanceInterface](c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceRepository[T]) (*Response[*ResourceInstance[T]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[T]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultList[T ResourceInstanceInterface](c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceRepository[T]) (*Response[[]ResourceInstance[T]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[T]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreate[T ResourceInstanceInterface](c *EndorContext[CreateDTO[ResourceInstance[T]]], schema RootSchema, repository *ResourceInstanceRepository[T], resource string) (*Response[ResourceInstance[T]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", resource))).Build(), nil
}

func defaultUpdate[T ResourceInstanceInterface](c *EndorContext[UpdateByIdDTO[ResourceInstance[T]]], schema RootSchema, repository *ResourceInstanceRepository[T], resource string) (*Response[ResourceInstance[T]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[T]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", resource))).Build(), nil
}

func defaultDelete[T ResourceInstanceInterface](c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceRepository[T], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", resource))).Build(), nil
}

func getDefaultActionsForCategory[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](resource string, schema RootSchema, resourceDescription string, categoryID string) map[string]EndorServiceAction {
	autogenerateID := true

	repository := NewResourceInstanceSpecializedRepository[T, C](
		resource,
		ResourceInstanceRepositoryOptions{AutoGenerateID: &autogenerateID},
	)

	return map[string]EndorServiceAction{
		categoryID + "/schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstanceSpecialized[T, C]], error) {
				return defaultListSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstanceSpecialized[T, C]]]) (*Response[ResourceInstanceSpecialized[T, C]], error) {
				return defaultCreateSpecialized(c, schema, repository, resource)
			},
		),
		categoryID + "/instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstanceSpecialized[T, C]], error) {
				return defaultInstanceSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		categoryID + "/update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[T, C]]]) (*Response[ResourceInstanceSpecialized[T, C]], error) {
				return defaultUpdateSpecialized(c, schema, repository, resource)
			},
		),
		categoryID + "/delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDeleteSpecialized(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
	}
}

func defaultListSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C]) (*Response[[]ResourceInstanceSpecialized[T, C]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstanceSpecialized[T, C]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreateSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[CreateDTO[ResourceInstanceSpecialized[T, C]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C], resource string) (*Response[ResourceInstanceSpecialized[T, C]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[T, C]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created (category)", resource))).Build(), nil
}

func defaultInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C]) (*Response[*ResourceInstanceSpecialized[T, C]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstanceSpecialized[T, C]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultUpdateSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[T, C]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[T, C], resource string) (*Response[ResourceInstanceSpecialized[T, C]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[T, C]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated (category)", resource))).Build(), nil
}

func defaultDeleteSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceSpecializedRepository[T, C], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted (category)", resource))).Build(), nil
}
