package sdk_entity

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorHybridService[T sdk.EntityInstanceInterface] struct {
	Entity            string
	EntityDescription string
	Priority          *int
	methodsFn         func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface
}

func (h EndorHybridService[T]) GetEntity() string {
	return h.Entity
}

func (h EndorHybridService[T]) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorHybridService[T]) GetPriority() *int {
	return h.Priority
}

func (h EndorHybridService[T]) GetSchema() *sdk.RootSchema {
	return getRootSchema[T]()
}

func NewEndorHybridService[T sdk.EntityInstanceInterface](entity, entityDescription string) sdk.EndorHybridServiceInterface {
	return EndorHybridService[T]{
		Entity:            entity,
		EntityDescription: entityDescription,
	}
}

func (h EndorHybridService[T]) WithPriority(
	priority int,
) sdk.EndorHybridServiceInterface {
	h.Priority = &priority
	return h
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridService[T]) WithActions(
	fn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface,
) sdk.EndorHybridServiceInterface {
	h.methodsFn = fn
	return h
}

// create endor service instance
func (h EndorHybridService[T]) ToEndorService(metadataSchema sdk.RootSchema) sdk.EndorService {
	var methods = make(map[string]sdk.EndorServiceActionInterface)

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

	return sdk.EndorService{
		Entity:            h.Entity,
		EntityDescription: h.EntityDescription,
		Priority:          h.Priority,
		Actions:           methods,
		EntitySchema:      *rootSchemWithMetadata,
	}
}

func getRootSchemaWithMetadata[T sdk.EntityInstanceInterface](metadataSchema sdk.RootSchema) *sdk.RootSchema {
	rootSchema := getRootSchema[T]()
	if metadataSchema.Schema.Properties != nil {
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	return rootSchema
}

func getDefaultActions[T sdk.EntityInstanceInterface](entity string, schema sdk.RootSchema, entityDescription string) map[string]sdk.EndorServiceActionInterface {
	// Crea repository usando DynamicEntity come default (per ora)
	autogenerateID := true
	repository := NewEntityInstanceRepository[T](entity, sdk.EntityInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", entity, entityDescription),
		),
		"instance": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[*sdk.EntityInstance[T]], error) {
				return defaultInstance(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", entity, entityDescription),
		),
		"list": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadDTO]) (*sdk.Response[[]sdk.EntityInstance[T]], error) {
				return defaultList(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", entity, entityDescription),
		),
		"create": sdk.NewConfigurableAction(
			sdk.EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", entity, entityDescription),
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
			func(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInstance[T]]]) (*sdk.Response[sdk.EntityInstance[T]], error) {
				return defaultCreate(c, schema, repository, entity)
			},
		),
		"update": sdk.NewConfigurableAction(
			sdk.EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", entity, entityDescription),
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
			func(c *sdk.EndorContext[sdk.ReplaceByIdDTO[sdk.EntityInstance[T]]]) (*sdk.Response[sdk.EntityInstance[T]], error) {
				return defaultReplace(c, schema, repository, entity)
			},
		),
		"delete": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[any], error) {
				return defaultDelete(c, repository, entity)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", entity, entityDescription),
		),
	}
}

func defaultSchema[T sdk.EntityInstanceInterface](_ *sdk.EndorContext[sdk.NoPayload], schema sdk.RootSchema) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], schema sdk.RootSchema, repository *EntityInstanceRepository[T]) (*sdk.Response[*sdk.EntityInstance[T]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[*sdk.EntityInstance[T]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultList[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReadDTO], schema sdk.RootSchema, repository *EntityInstanceRepository[T]) (*sdk.Response[[]sdk.EntityInstance[T]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.EntityInstance[T]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreate[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInstance[T]]], schema sdk.RootSchema, repository *EntityInstanceRepository[T], entity string) (*sdk.Response[sdk.EntityInstance[T]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s created", entity))).Build(), nil
}

func defaultReplace[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReplaceByIdDTO[sdk.EntityInstance[T]]], schema sdk.RootSchema, repository *EntityInstanceRepository[T], entity string) (*sdk.Response[sdk.EntityInstance[T]], error) {
	replaced, err := repository.Replace(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInstance[T]]().AddData(replaced).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s replaced", entity))).Build(), nil
}

func defaultDelete[T sdk.EntityInstanceInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], repository *EntityInstanceRepository[T], entity string) (*sdk.Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s deleted", entity))).Build(), nil
}
