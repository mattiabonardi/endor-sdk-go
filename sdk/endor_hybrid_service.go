package sdk

import (
	"context"
	"fmt"
)

type EndorHybridService struct {
	Resource    string
	Description string
	Priority    *int
	methodsFn   func(getSchema func() RootSchema) map[string]EndorServiceAction
	baseModel   ResourceInstanceInterface // Modello base per ResourceInstanceRepository
	categories  map[string]Category       // Cache delle categorie disponibili
}

func NewHybridService(resource, description string) EndorHybridService {
	return EndorHybridService{
		Resource:    resource,
		Description: description,
		baseModel:   &DynamicResource{},
	}
}

// WithBaseModel permette di specificare un modello base personalizzato
func (h EndorHybridService) WithBaseModel(model ResourceInstanceInterface) EndorHybridService {
	h.baseModel = model
	return h
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridService) WithActions(
	fn func(getSchema func() RootSchema) map[string]EndorServiceAction,
) EndorHybridService {
	h.methodsFn = fn
	return h
}

// WithCategories permette di configurare le categorie disponibili
func (h EndorHybridService) WithCategories(categories []Category) EndorHybridService {
	if h.categories == nil {
		h.categories = make(map[string]Category)
	}
	for _, category := range categories {
		if category.BaseModel == nil {
			category.BaseModel = &DynamicResourceSpecialized{}
		}
		h.categories[category.ID] = category
	}
	return h
}

// create endor service instance
func (h EndorHybridService) ToEndorService(metadataSchema Schema) EndorService {
	var methods = make(map[string]EndorServiceAction)

	// schema
	rootSchemWithMetadata := h.getRootSchemaWithMetadata(metadataSchema)
	getSchemaCallback := func() RootSchema { return *rootSchemWithMetadata }

	// add default CRUD methods
	methods = getDefaultActions(h.Resource, *rootSchemWithMetadata, h.Description)
	// add custom methods
	if h.methodsFn != nil {
		for methodName, method := range h.methodsFn(getSchemaCallback) {
			methods[methodName] = method
		}
	}

	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for categoryID, category := range h.categories {
			// copy values for closure issues
			currentCategory := category
			currentCategoryID := categoryID

			// schema
			rootSchemWithCategory := h.getCategorySchemaWithMetadata(metadataSchema, currentCategory)
			getSchemaWithCategoryCallback := func() RootSchema { return *rootSchemWithCategory }

			// add default CRUD methods specified for category
			categoryMethods := getDefaultActionsForCategory(h.Resource, *rootSchemWithCategory, h.Description, currentCategoryID)
			for methodName, method := range categoryMethods {
				// add category prefix <category>/<action>
				categoryMethodName := currentCategoryID + "/" + methodName
				methods[categoryMethodName] = method
			}

			for methodName, method := range h.methodsFn(getSchemaWithCategoryCallback) {
				// add category prefix <category>/<action>
				categoryMethodName := currentCategoryID + "/" + methodName
				methods[categoryMethodName] = method
			}
		}
	}

	return EndorService{
		Resource:    h.Resource,
		Description: h.Description,
		Priority:    h.Priority,
		Methods:     methods,
	}
}

// combine baseModel schema with metadata schema
func (h EndorHybridService) getRootSchemaWithMetadata(metadataSchema Schema) *RootSchema {
	rootSchema := NewSchema(h.baseModel)
	if metadataSchema.Properties != nil {
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	return rootSchema
}

// combine root schema with category schema
func (h EndorHybridService) getCategorySchemaWithMetadata(metadataSchema Schema, category Category) *RootSchema {
	// create root schema
	rootSchema := h.getRootSchemaWithMetadata(metadataSchema)

	// add category base schema (hardcoded)
	categoryBaseSchema := NewSchema(category.BaseModel)
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

func getDefaultActions(resource string, schema RootSchema, resourceDescription string) map[string]EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := NewResourceInstanceRepository[ResourceInstanceInterface](resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema(c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", resource, resourceDescription),
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[ResourceInstanceInterface]], error) {
				return defaultInstance(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", resource, resourceDescription),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[ResourceInstanceInterface]], error) {
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
			func(c *EndorContext[CreateDTO[ResourceInstance[ResourceInstanceInterface]]]) (*Response[ResourceInstance[ResourceInstanceInterface]], error) {
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
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[ResourceInstanceInterface]]]) (*Response[ResourceInstance[ResourceInstanceInterface]], error) {
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

func defaultSchema(_ *EndorContext[NoPayload], schema RootSchema) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance(c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceRepository[ResourceInstanceInterface]) (*Response[*ResourceInstance[ResourceInstanceInterface]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[ResourceInstanceInterface]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultList(c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceRepository[ResourceInstanceInterface]) (*Response[[]ResourceInstance[ResourceInstanceInterface]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[ResourceInstanceInterface]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreate(c *EndorContext[CreateDTO[ResourceInstance[ResourceInstanceInterface]]], schema RootSchema, repository *ResourceInstanceRepository[ResourceInstanceInterface], resource string) (*Response[ResourceInstance[ResourceInstanceInterface]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[ResourceInstanceInterface]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", resource))).Build(), nil
}

func defaultUpdate(c *EndorContext[UpdateByIdDTO[ResourceInstance[ResourceInstanceInterface]]], schema RootSchema, repository *ResourceInstanceRepository[ResourceInstanceInterface], resource string) (*Response[ResourceInstance[ResourceInstanceInterface]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[ResourceInstanceInterface]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", resource))).Build(), nil
}

func defaultDelete(c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceRepository[ResourceInstanceInterface], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", resource))).Build(), nil
}

func getDefaultActionsForCategory(resource string, schema RootSchema, resourceDescription string, categoryID string) map[string]EndorServiceAction {
	// Per ora usa DynamicResource come base model e struct vuota come category model
	// TODO: Implementare logica per detectare il modello corretto della categoria
	autogenerateID := true

	repository := NewResourceInstanceSpecializedRepository[ResourceInstanceInterface, ResourceInstanceSpecializedInterface](
		resource,
		ResourceInstanceRepositoryOptions{AutoGenerateID: &autogenerateID},
	)

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema(c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
				return defaultListSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		"create": NewConfigurableAction(
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
			func(c *EndorContext[CreateDTO[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]]) (*Response[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
				return defaultCreateSpecialized(c, schema, repository, resource)
			},
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
				return defaultInstanceSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		"update": NewConfigurableAction(
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
			func(c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]]) (*Response[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
				return defaultUpdateSpecialized(c, schema, repository, resource)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDeleteSpecialized(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
	}
}

// Metodi specializzati per repository specializzati (analoghe ai metodi base)
func defaultListSpecialized(c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]) (*Response[[]ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreateSpecialized(c *EndorContext[CreateDTO[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[ResourceInstanceInterface, ResourceInstanceSpecializedInterface], resource string) (*Response[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created (category)", resource))).Build(), nil
}

func defaultInstanceSpecialized(c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]) (*Response[*ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultUpdateSpecialized(c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[ResourceInstanceInterface, ResourceInstanceSpecializedInterface], resource string) (*Response[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[ResourceInstanceInterface, ResourceInstanceSpecializedInterface]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated (category)", resource))).Build(), nil
}

func defaultDeleteSpecialized(c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceSpecializedRepository[ResourceInstanceInterface, ResourceInstanceSpecializedInterface], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted (category)", resource))).Build(), nil
}
