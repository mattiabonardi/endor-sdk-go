package sdk

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type AbstractResourceService struct {
	resource    string
	definition  ResourceDefinition
	mongoClient *mongo.Client
	mongoDB     string
	context     context.Context
}

func NewAbstractResourceService(resource string, description string, definition ResourceDefinition, mongoClient *mongo.Client, mongoDB string, context context.Context) EndorService {
	service := AbstractResourceService{
		resource:    resource,
		definition:  definition,
		mongoClient: mongoClient,
		mongoDB:     mongoDB,
		context:     context,
	}
	return EndorService{
		Resource:    resource,
		Description: description,
		Methods: map[string]EndorServiceMethod{
			"list": NewMethod(
				service.list,
			),
			"create": NewMethod(
				service.create,
			),
			"instance": NewMethod(
				service.instance,
			),
			"update": NewMethod(
				service.update,
			),
			"delete": NewMethod(
				service.delete,
			),
		},
	}
}

func (h *AbstractResourceService) instance(c *EndorContext[ReadInstanceDTO]) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	instance, err := repo.Instance(c.Payload)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.NotFound(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[any]().AddData(&instance).AddSchema(&h.definition.Schema).Build())
}

func (h *AbstractResourceService) list(c *EndorContext[NoPayload]) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	list, err := repo.List()
	if err != nil {
		c.InternalServerError(err)
		return
	}
	c.End(NewResponseBuilder[[]any]().AddData(&list).AddSchema(&h.definition.Schema).Build())
}

func (h *AbstractResourceService) create(c *EndorContext[CreateDTO[any]]) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	err = repo.Create(c.Payload)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			c.Conflict(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[any]().AddData(&c.Payload.Data).AddSchema(&h.definition.Schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build())
}

func (h *AbstractResourceService) update(c *EndorContext[UpdateByIdDTO[any]]) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	updated, err := repo.Update(c.Payload)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.NotFound(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[any]().AddData(&updated).AddSchema(&h.definition.Schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.resource))).Build())
}

func (h *AbstractResourceService) delete(c *EndorContext[DeleteByIdDTO]) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	err = repo.Delete(c.Payload)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.NotFound(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[any]().AddSchema(&h.definition.Schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.resource))).Build())
}
