package sdk_resource

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorHybridService[T sdk.ResourceInstanceInterface] struct {
	Resource            string
	ResourceDescription string
	Priority            *int
	methodsFn           func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction
}

func (h EndorHybridService[T]) GetResource() string {
	return h.Resource
}

func (h EndorHybridService[T]) GetResourceDescription() string {
	return h.ResourceDescription
}

func (h EndorHybridService[T]) GetPriority() *int {
	return h.Priority
}

func NewHybridService[T sdk.ResourceInstanceInterface](resource, resourceDescription string) sdk.EndorHybridServiceInterface {
	return EndorHybridService[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
	}
}

// define methods. The params getSchema allow to inject the dynamic schema
func (h EndorHybridService[T]) WithActions(
	fn func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction,
) sdk.EndorHybridServiceInterface {
	h.methodsFn = fn
	return h
}

// create endor service instance
func (h EndorHybridService[T]) ToEndorService(metadataSchema sdk.Schema) sdk.EndorService {
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

func getDefaultActions[T sdk.ResourceInstanceInterface](resource string, schema sdk.RootSchema, resourceDescription string) map[string]sdk.EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := repository.NewResourceInstanceRepository[T](resource, repository.ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]sdk.EndorServiceAction{
		"schema": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
				return defaultSchema[T](c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", resource, resourceDescription),
		),
		"instance": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[*sdk.ResourceInstance[T]], error) {
				return defaultInstance(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", resource, resourceDescription),
		),
		"list": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadDTO]) (*sdk.Response[[]sdk.ResourceInstance[T]], error) {
				return defaultList(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", resource, resourceDescription),
		),
		"create": sdk.NewConfigurableAction(
			sdk.EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", resource, resourceDescription),
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
			func(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstance[T]]]) (*sdk.Response[sdk.ResourceInstance[T]], error) {
				return defaultCreate(c, schema, repository, resource)
			},
		),
		"update": sdk.NewConfigurableAction(
			sdk.EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", resource, resourceDescription),
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
			func(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInstance[T]]]) (*sdk.Response[sdk.ResourceInstance[T]], error) {
				return defaultUpdate(c, schema, repository, resource)
			},
		),
		"delete": sdk.NewAction(
			func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[any], error) {
				return defaultDelete(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", resource, resourceDescription),
		),
	}
}

func defaultSchema[T sdk.ResourceInstanceInterface](_ *sdk.EndorContext[sdk.NoPayload], schema sdk.RootSchema) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance[T sdk.ResourceInstanceInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T]) (*sdk.Response[*sdk.ResourceInstance[T]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[*sdk.ResourceInstance[T]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultList[T sdk.ResourceInstanceInterface](c *sdk.EndorContext[sdk.ReadDTO], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T]) (*sdk.Response[[]sdk.ResourceInstance[T]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.ResourceInstance[T]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreate[T sdk.ResourceInstanceInterface](c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInstance[T]]], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T], resource string) (*sdk.Response[sdk.ResourceInstance[T]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInstance[T]]().AddData(created).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s created", resource))).Build(), nil
}

func defaultUpdate[T sdk.ResourceInstanceInterface](c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInstance[T]]], schema sdk.RootSchema, repository *repository.ResourceInstanceRepository[T], resource string) (*sdk.Response[sdk.ResourceInstance[T]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInstance[T]]().AddData(updated).AddSchema(&schema).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s updated", resource))).Build(), nil
}

func defaultDelete[T sdk.ResourceInstanceInterface](c *sdk.EndorContext[sdk.ReadInstanceDTO], repository *repository.ResourceInstanceRepository[T], resource string) (*sdk.Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("%s deleted", resource))).Build(), nil
}
