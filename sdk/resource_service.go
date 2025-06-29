package sdk

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func NewResourceService(microServiceId string, services *[]EndorResource, client *mongo.Client, context context.Context, databaseName string) *EndorResource {
	resourceService := ResourceService{
		microServiceId: microServiceId,
		services:       services,
		mongoClient:    client,
		context:        context,
		databaseName:   databaseName,
	}
	return &EndorResource{
		Resource:    "resource",
		Description: "Resource",
		Methods: map[string]EndorResourceAction{
			"schema": NewAction(
				resourceService.schema,
				"Get the schema of the resource",
			),
			"list": NewAction(
				resourceService.list,
				"Search for available resources",
			),
			"instance": NewAction(
				resourceService.instance,
				"Get the specified instance of resources",
			),
			"create": NewAction(
				resourceService.create,
				"Create a new resource",
			),
			"update": NewAction(
				resourceService.update,
				"Update an existing resource",
			),
			"delete": NewAction(
				resourceService.delete,
				"Delete an existing resource",
			),
		},
	}
}

type ResourceService struct {
	microServiceId string
	services       *[]EndorResource
	mongoClient    *mongo.Client
	context        context.Context
	databaseName   string
}

func (h *ResourceService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) list(c *EndorContext[NoPayload]) (*Response[[]Resource], error) {
	resources, err := NewEndorResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).ResourceList()
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]Resource]().AddData(&resources).AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[Resource], error) {
	resource, err := NewEndorResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(&resource.resource).AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) create(c *EndorContext[CreateDTO[Resource]]) (*Response[Resource], error) {
	err := NewEndorResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(&c.Payload.Data).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, fmt.Sprintf("resource %s created", c.Payload.Data.ID))).Build(), nil
}

func (h *ResourceService) update(c *EndorContext[UpdateByIdDTO[Resource]]) (*Response[Resource], error) {
	resource, err := NewEndorResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, "resource updated")).Build(), nil
}

func (h *ResourceService) delete(c *EndorContext[DeleteByIdDTO]) (*Response[Resource], error) {
	err := NewEndorResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).DeleteOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddMessage(NewMessage(Info, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build(), nil
}
