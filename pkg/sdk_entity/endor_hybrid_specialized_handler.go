package sdk_entity

import (
	"context"
	"fmt"
	"maps"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

type EndorHybridSpecializedHandlerCategory[T sdk.EntityInstanceSpecializedInterface] struct {
	ID                string
	Description       string
	ActionFn          func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface
	repositoryFactory sdk.RepositoryFactory
}

func (h *EndorHybridSpecializedHandlerCategory[T]) GetID() string {
	return h.ID
}

func (h *EndorHybridSpecializedHandlerCategory[T]) GetDescription() string {
	return h.Description
}

func (h *EndorHybridSpecializedHandlerCategory[T]) GetSchema() string {
	schema, _ := getRootSchema[T]().ToYAML()
	return schema
}

// GetActions implements sdk.EndorHybridSpecializedServiceCategoryInterface.
func (h *EndorHybridSpecializedHandlerCategory[T]) GetActions() func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface {
	return h.ActionFn
}

func (h *EndorHybridSpecializedHandlerCategory[T]) GetRepository() sdk.RepositoryFactory {
	return h.repositoryFactory
}

func (h *EndorHybridSpecializedHandlerCategory[T]) WithActions(actionFn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface) sdk.EndorHybridSpecializedHandlerCategoryInterface {
	h.ActionFn = actionFn
	return h
}

func (h *EndorHybridSpecializedHandlerCategory[T]) WithRepository(
	fn sdk.RepositoryFactory,
) sdk.EndorHybridSpecializedHandlerCategoryInterface {
	h.repositoryFactory = fn
	return h
}

func (h *EndorHybridSpecializedHandlerCategory[T]) CreateDefaultActions(entity string, entityDescription string, metadataSchema sdk.RootSchema, categoryMetadataSchema sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface {
	rootSchemWithCategory := getCategorySchemaWithMetadata[T](metadataSchema, categoryMetadataSchema)
	return getDefaultActionsForCategory[T](entity, *rootSchemWithCategory, entityDescription, h.ID)
}

func NewEndorHybridSpecializedHandlerCategory[T sdk.EntityInstanceSpecializedInterface](categoryID string, categoryDescription string) sdk.EndorHybridSpecializedHandlerCategoryInterface {
	return &EndorHybridSpecializedHandlerCategory[T]{
		ID:          categoryID,
		Description: categoryDescription,
	}
}

type EndorHybridSpecializedHandler[T sdk.EntityInstanceSpecializedInterface] struct {
	Entity              string
	EntityDescription   string
	Priority            *int
	methodsFn           func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface
	staticCategories    []string
	categories          map[string]sdk.EndorHybridSpecializedHandlerCategoryInterface
	repositoryFactories map[string]sdk.RepositoryFactory
}

func (h EndorHybridSpecializedHandler[T]) GetEntity() string {
	return h.Entity
}

func (h EndorHybridSpecializedHandler[T]) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorHybridSpecializedHandler[T]) GetPriority() *int {
	return h.Priority
}

func (h EndorHybridSpecializedHandler[T]) GetSchema() *sdk.RootSchema {
	return getRootSchema[T]()
}

func NewEndorHybridSpecializedHandler[T sdk.EntityInstanceSpecializedInterface](entity, entityDescription string) sdk.EndorHybridSpecializedHandlerInterface {
	return EndorHybridSpecializedHandler[T]{
		Entity:            entity,
		EntityDescription: entityDescription,
	}
}

func (h EndorHybridSpecializedHandler[T]) WithPriority(
	priority int,
) sdk.EndorHybridSpecializedHandlerInterface {
	h.Priority = &priority
	return h
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridSpecializedHandler[T]) WithActions(
	fn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface,
) sdk.EndorHybridSpecializedHandlerInterface {
	h.methodsFn = fn
	return h
}

func (h EndorHybridSpecializedHandler[T]) WithHybridCategories(categories []sdk.EndorHybridSpecializedHandlerCategoryInterface) sdk.EndorHybridSpecializedHandlerInterface {
	if h.categories == nil {
		h.categories = make(map[string]sdk.EndorHybridSpecializedHandlerCategoryInterface)
	}
	if h.staticCategories == nil {
		h.staticCategories = []string{}
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
		h.staticCategories = append(h.staticCategories, category.GetID())
	}
	return h
}

func (h EndorHybridSpecializedHandler[T]) GetHybridCategories() []sdk.HybridCategory {
	staticCategories := []sdk.HybridCategory{}
	for _, categoryID := range h.staticCategories {
		staticCategories = append(staticCategories, sdk.HybridCategory{
			ID:          h.categories[categoryID].GetID(),
			Description: h.categories[categoryID].GetDescription(),
			Schema:      h.categories[categoryID].GetSchema(),
		})
	}
	return staticCategories
}

// create endor service instance
func (h EndorHybridSpecializedHandler[T]) ToEndorHandler(metadataSchema sdk.RootSchema, categoriesMetadataSchema map[string]sdk.RootSchema, additionalCategories []sdk.DynamicCategory) sdk.EndorHandler {
	var methods = make(map[string]sdk.EndorHandlerActionInterface)

	if h.repositoryFactories == nil {
		h.repositoryFactories = map[string]sdk.RepositoryFactory{}
	}

	// merge additional categories
	for _, additionalCategory := range additionalCategories {
		h.categories[additionalCategory.ID] = NewEndorHybridSpecializedHandlerCategory[T](additionalCategory.ID, additionalCategory.Description)
		additionalCategorySchema, _ := additionalCategory.UnmarshalAdditionalAttributes()
		categoriesMetadataSchema[additionalCategory.ID] = *additionalCategorySchema
	}

	// schema
	rootSchemaWithMetadata := getRootSchemaWithMetadata[T](metadataSchema)
	getSchemaCallback := func() sdk.RootSchema { return *rootSchemaWithMetadata }

	// add default CRUD methods
	methods = getDefaultActions[T](h.Entity, *rootSchemaWithMetadata, h.EntityDescription)
	// remove delete and update
	delete(methods, "create")
	delete(methods, "update")
	// add custom methods
	if h.methodsFn != nil {
		maps.Copy(methods, h.methodsFn(getSchemaCallback))
	}

	masterRepositoryFactory := func(session sdk.Session, container sdk.EndorDIContainerInterface) sdk.EndorRepositoryInterface {
		autogenerateID := true
		return NewEntityInstanceRepository[T](h.Entity, *rootSchemaWithMetadata, sdk.EntityInstanceRepositoryOptions{
			AutoGenerateID: &autogenerateID,
		}, session, container)
	}
	h.repositoryFactories[h.Entity] = masterRepositoryFactory

	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for categoryID, category := range h.categories {
			categoryRepositoryFactory := func(session sdk.Session, container sdk.EndorDIContainerInterface) sdk.EndorRepositoryInterface {
				autogenerateID := true
				return NewEntityInstanceRepository[T](h.Entity, *rootSchemaWithMetadata, sdk.EntityInstanceRepositoryOptions{
					AutoGenerateID: &autogenerateID,
				}, session, container)
			}
			h.repositoryFactories[h.Entity+"/"+categoryID] = categoryRepositoryFactory
			// add default CRUD methods specified for category
			categoryMethods := category.CreateDefaultActions(h.Entity, h.EntityDescription, metadataSchema, categoriesMetadataSchema[categoryID])
			maps.Copy(methods, categoryMethods)
		}
	}

	return sdk.EndorHandler{
		Entity:              h.Entity,
		EntityDescription:   h.EntityDescription,
		Priority:            h.Priority,
		Actions:             methods,
		RepositoryFactories: h.repositoryFactories,
	}
}

func getCategorySchemaWithMetadata[T sdk.EntityInstanceSpecializedInterface](metadataSchema sdk.RootSchema, categoryMetadataSchema sdk.RootSchema) *sdk.RootSchema {
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

func getDefaultActionsForCategory[T sdk.EntityInstanceSpecializedInterface](entity string, schema sdk.RootSchema, entityDescription string, categoryID string) map[string]sdk.EndorHandlerActionInterface {
	entityPath := entity + "/" + categoryID

	return map[string]sdk.EndorHandlerActionInterface{
		categoryID + "/schema": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", entity, entityDescription, categoryID),
		),
		categoryID + "/list": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadDTO]) (*sdk.Response[[]sdk.EntityInstance[T]], error) {
				return defaultListSpecialized[T](c, schema, entityPath)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", entity, entityDescription, categoryID),
		),
		categoryID + "/create": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description: fmt.Sprintf("Create the instance of %s (%s) for category %s", entity, entityDescription, categoryID),
				InputSchema: &sdk.RootSchema{
					Schema: sdk.Schema{
						Type: sdk.SchemaTypeObject,
						Properties: &map[string]sdk.Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInstanceSpecialized[T]]]) (*sdk.Response[sdk.EntityInstance[T]], error) {
				return defaultCreateSpecialized(c, schema, entityPath)
			},
		),
		categoryID + "/instance": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[*sdk.EntityInstance[T]], error) {
				return defaultInstanceSpecialized[T](c, schema, entityPath)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", entity, entityDescription, categoryID),
		),
		categoryID + "/update": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description: fmt.Sprintf("Update the existing instance of %s (%s) for category %s", entity, entityDescription, categoryID),
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
			func(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]]) (*sdk.Response[sdk.EntityInstance[T]], error) {
				return defaultUpdateSpecialized(c, schema, entityPath)
			},
		),
	}
}

func defaultListSpecialized[T sdk.EntityInstanceSpecializedInterface](c *sdk.EndorContext[sdk.ReadDTO], schema sdk.RootSchema, entityPath string) (*sdk.Response[[]sdk.EntityInstance[T]], error) {
	repo, err := sdk.GetDynamicRepository[T](c.DIContainer, entityPath)
	if err != nil {
		return nil, err
	}
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
	list, references, err := repo.ListWithReferences(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.EntityInstance[T]]().AddData(&list).AddSchema(&schema).AddReferences(references).Build(), nil
}

func defaultCreateSpecialized[T sdk.EntityInstanceSpecializedInterface](c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInstanceSpecialized[T]]], schema sdk.RootSchema, entityPath string) (*sdk.Response[sdk.EntityInstance[T]], error) {
	repo, err := sdk.GetDynamicRepository[T](c.DIContainer, entityPath)
	if err != nil {
		return nil, err
	}
	c.Payload.Data.SetCategoryType(c.CategoryType)
	created, err := repo.Create(context.TODO(), sdk.CreateDTO[sdk.EntityInstance[T]]{
		Data: c.Payload.Data.EntityInstance,
	})
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, sdk_i18n.T(c.Locale, "entities.entity.created_specialized", map[string]any{"name": entityPath, "id": created.GetID()}))).Build(), nil
}

func defaultInstanceSpecialized[T sdk.EntityInstanceSpecializedInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], schema sdk.RootSchema, entityPath string) (*sdk.Response[*sdk.EntityInstance[T]], error) {
	repo, err := sdk.GetDynamicRepository[T](c.DIContainer, entityPath)
	if err != nil {
		return nil, err
	}
	instance, references, err := repo.InstanceWithReferences(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[*sdk.EntityInstance[T]]().AddData(&instance).AddSchema(&schema).AddReferences(references).Build(), nil
}

func defaultUpdateSpecialized[T sdk.EntityInstanceSpecializedInterface](c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]], schema sdk.RootSchema, entityPath string) (*sdk.Response[sdk.EntityInstance[T]], error) {
	repo, err := sdk.GetDynamicRepository[T](c.DIContainer, entityPath)
	if err != nil {
		return nil, err
	}
	updated, err := repo.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInstance[T]]().AddData(updated).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, sdk_i18n.T(c.Locale, "entities.entity.updated_category", map[string]any{"name": entityPath}))).Build(), nil
}
