package sdk

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func NewResourceService(microServiceId string, services []EndorService, client *mongo.Client, context context.Context, databaseName string) EndorService {
	resourceService := ResourceService{
		microServiceId: microServiceId,
		services:       services,
		mongoClient:    client,
		context:        context,
		databaseName:   databaseName,
	}
	return EndorService{
		Resource:    "resource",
		Description: "Resource",
		Methods: map[string]EndorServiceMethod{
			"list": NewMethod(
				resourceService.list,
			),
			"instance": NewMethod(
				resourceService.instance,
			),
			"create": NewMethod(
				resourceService.create,
			),
			"update": NewMethod(
				resourceService.update,
			),
			"delete": NewMethod(
				resourceService.delete,
			),
		},
	}
}

type ResourceService struct {
	microServiceId string
	services       []EndorService
	mongoClient    *mongo.Client
	context        context.Context
	databaseName   string
}

func (h *ResourceService) list(c *EndorContext[NoPayload]) (*Response[[]Resource], error) {
	resources, err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).List()
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]Resource]().AddData(&resources).AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[Resource], error) {
	resource, err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) create(c *EndorContext[CreateDTO[Resource]]) (*Response[Resource], error) {
	err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(&c.Payload.Data).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, fmt.Sprintf("resource %s created", c.Payload.Data.ID))).Build(), nil
}

func (h *ResourceService) update(c *EndorContext[UpdateByIdDTO[Resource]]) (*Response[Resource], error) {
	resource, err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, "resource updated")).Build(), nil
}

func (h *ResourceService) delete(c *EndorContext[DeleteByIdDTO]) (*Response[Resource], error) {
	err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).DeleteOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddMessage(NewMessage(Info, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build(), nil
}
