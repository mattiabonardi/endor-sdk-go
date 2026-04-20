package sdk_entity

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

type EndorHybridHandler[T sdk.EntityInstanceInterface] struct {
	Entity            string
	EntityDescription string
	Priority          *int
	methodsFn         func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface
}

func (h EndorHybridHandler[T]) GetEntity() string {
	return h.Entity
}

func (h EndorHybridHandler[T]) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorHybridHandler[T]) GetPriority() *int {
	return h.Priority
}

func (h EndorHybridHandler[T]) GetSchema() *sdk.RootSchema {
	return getRootSchema[T]()
}

func NewEndorHybridHandler[T sdk.EntityInstanceInterface](entity, entityDescription string) sdk.EndorHybridHandlerInterface {
	return EndorHybridHandler[T]{
		Entity:            entity,
		EntityDescription: entityDescription,
	}
}

func (h EndorHybridHandler[T]) WithPriority(
	priority int,
) sdk.EndorHybridHandlerInterface {
	h.Priority = &priority
	return h
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridHandler[T]) WithActions(
	fn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface,
) sdk.EndorHybridHandlerInterface {
	h.methodsFn = fn
	return h
}

// create endor service instance
func (h EndorHybridHandler[T]) ToEndorHandler(metadataSchema sdk.RootSchema) sdk.EndorHandler {
	var methods = make(map[string]sdk.EndorHandlerActionInterface)

	// schema
	rootSchemWithMetadata := getRootSchemaWithMetadata[T](metadataSchema)
	getSchemaCallback := func() sdk.RootSchema { return *rootSchemWithMetadata }

	// add default CRUD methods
	methods = getDefaultActions[T](h.Entity, *rootSchemWithMetadata, h.EntityDescription)
	// add custom methods
	if h.methodsFn != nil {
		for methodName, method := range h.methodsFn(getSchemaCallback) {
			methods[methodName] = method
		}
	}

	// define repository factory
	repositoryFactory := func(container sdk.EndorDIContainer) sdk.EndorRepositoryInterface {
		autogenerateID := true
		return NewEntityInstanceRepository[T](h.Entity, *rootSchemWithMetadata, sdk.EntityInstanceRepositoryOptions{
			AutoGenerateID: &autogenerateID,
		}, container)
	}

	return sdk.EndorHandler{
		Entity:              h.Entity,
		EntityDescription:   h.EntityDescription,
		Priority:            h.Priority,
		Actions:             methods,
		EntitySchema:        *rootSchemWithMetadata,
		RepositoryFactories: []sdk.RepositoryFactory{repositoryFactory},
	}
}

func getRootSchemaWithMetadata[T sdk.EntityInstanceInterface](metadataSchema sdk.RootSchema) *sdk.RootSchema {
	rootSchema := getRootSchema[T]()
	if metadataSchema.Schema.Properties != nil {
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	// merge ui schema
	if metadataSchema.UISchema != nil {
		rootSchema.UISchema = metadataSchema.UISchema
	}
	return rootSchema
}

func getDefaultActions[T sdk.EntityInstanceInterface](entity string, schema sdk.RootSchema, entityDescription string) map[string]sdk.EndorHandlerActionInterface {
	return map[string]sdk.EndorHandlerActionInterface{
		"schema": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", entity, entityDescription),
		),
		"instance": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[*sdk.EntityInstance[T]], error) {
				return defaultInstance[T](c, schema, entity)
			},
			fmt.Sprintf("Get the instance of %s (%s)", entity, entityDescription),
		),
		"list": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadDTO]) (*sdk.Response[[]sdk.EntityInstance[T]], error) {
				return defaultList[T](c, schema, entity)
			},
			fmt.Sprintf("Search for available list of %s (%s)", entity, entityDescription),
		),
		"create": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description: fmt.Sprintf("Create the instance of %s (%s)", entity, entityDescription),
				InputSchema: &sdk.RootSchema{
					Schema: sdk.Schema{
						Type: sdk.SchemaTypeObject,
						Properties: &map[string]sdk.Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInstance[T]]]) (*sdk.Response[sdk.EntityInstance[T]], error) {
				return defaultCreate(c, schema, entity)
			},
		),
		"update": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description: fmt.Sprintf("Update the existing instance of %s (%s)", entity, entityDescription),
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
				return defaultUpdate(c, schema, entity)
			},
		),
		"delete": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[any], error) {
				return defaultDelete[T](c, entity)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", entity, entityDescription),
		),
	}
}

func defaultSchema[T sdk.EntityInstanceInterface](_ *sdk.EndorContext[sdk.NoPayload], schema sdk.RootSchema) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], schema sdk.RootSchema, entity string) (*sdk.Response[*sdk.EntityInstance[T]], error) {
	instance, references, err := sdk.GetDynamicRepository[T](c.DIContainer, entity).InstanceWithReferences(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[*sdk.EntityInstance[T]]().AddData(&instance).AddSchema(&schema).AddReferences(references).Build(), nil
}

func defaultList[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReadDTO], schema sdk.RootSchema, entity string) (*sdk.Response[[]sdk.EntityInstance[T]], error) {
	list, references, err := sdk.GetDynamicRepository[T](c.DIContainer, entity).ListWithReferences(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.EntityInstance[T]]().AddData(&list).AddSchema(&schema).AddReferences(references).Build(), nil
}

func defaultCreate[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInstance[T]]], schema sdk.RootSchema, entity string) (*sdk.Response[sdk.EntityInstance[T]], error) {
	created, err := sdk.GetDynamicRepository[T](c.DIContainer, entity).Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, sdk_i18n.T(c.Locale, "entities.entity.created", map[string]any{"id": entity}))).Build(), nil
}

func defaultUpdate[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]], schema sdk.RootSchema, entity string) (*sdk.Response[sdk.EntityInstance[T]], error) {
	updated, err := sdk.GetDynamicRepository[T](c.DIContainer, entity).Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInstance[T]]().AddData(updated).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, sdk_i18n.T(c.Locale, "entities.entity.updated", map[string]any{"id": entity}))).Build(), nil
}

func defaultDelete[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], entity string) (*sdk.Response[any], error) {
	err := sdk.GetDynamicRepository[T](c.DIContainer, entity).Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, sdk_i18n.T(c.Locale, "entities.entity.deleted", map[string]any{"id": entity}))).Build(), nil
}
