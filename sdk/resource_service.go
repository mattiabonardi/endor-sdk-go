package sdk

import (
	"context"
	"errors"
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
				ValidationHandler,
				resourceService.list,
			),
			"instance": NewMethod(
				ValidationHandler,
				resourceService.instance,
			),
			"create": NewMethod(
				ValidationHandler,
				resourceService.create,
			),
			"update": NewMethod(
				ValidationHandler,
				resourceService.update,
			),
			"delete": NewMethod(
				ValidationHandler,
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

func (h *ResourceService) list(c *EndorContext[NoPayload]) {
	resources, err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).List()
	if err != nil {
		c.InternalServerError(err)
		return
	}
	c.End(NewResponseBuilder[[]Resource]().AddData(&resources).AddSchema(NewSchema(&Resource{})).Build())
}

func (h *ResourceService) instance(c *EndorContext[ReadInstanceDTO]) {
	resource, err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).Instance(c.Payload)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.NotFound(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).Build())
}

func (h *ResourceService) create(c *EndorContext[CreateDTO[Resource]]) {
	err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).Create(c.Payload)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			c.Conflict(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[Resource]().AddData(&c.Payload.Data).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, fmt.Sprintf("resource %s created", c.Payload.Data.ID))).Build())
}

func (h *ResourceService) update(c *EndorContext[UpdateByIdDTO[Resource]]) {
	resource, err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).UpdateOne(c.Payload)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.NotFound(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, "resource updated")).Build())
}

func (h *ResourceService) delete(c *EndorContext[DeleteByIdDTO]) {
	err := NewResourceRepository(h.microServiceId, h.services, h.mongoClient, h.context, h.databaseName).DeleteOne(c.Payload)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	c.End(NewResponseBuilder[Resource]().AddMessage(NewMessage(Info, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build())
}
