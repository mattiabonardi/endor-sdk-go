package sdk

import (
	"context"
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
			"schema": NewMethod(
				service.schema,
			),
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

func (h *AbstractResourceService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(h.createSchema()).Build(), nil
}

func (h *AbstractResourceService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		return nil, err
	}
	instance, err := repo.Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddData(&instance).AddSchema(h.createSchema()).Build(), nil
}

func (h *AbstractResourceService) list(c *EndorContext[NoPayload]) (*Response[[]any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		return nil, err
	}
	list, err := repo.List()
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]any]().AddData(&list).AddSchema(h.createSchema()).Build(), nil
}

func (h *AbstractResourceService) create(c *EndorContext[CreateDTO[any]]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		return nil, err
	}
	err = repo.Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddData(&c.Payload.Data).AddSchema(h.createSchema()).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build(), nil
}

func (h *AbstractResourceService) update(c *EndorContext[UpdateByIdDTO[any]]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		return nil, err
	}
	updated, err := repo.Update(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddData(&updated).AddSchema(h.createSchema()).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.resource))).Build(), nil
}

func (h *AbstractResourceService) delete(c *EndorContext[DeleteByIdDTO]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB, h.context)
	if err != nil {
		return nil, err
	}
	err = repo.Delete(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.resource))).Build(), nil
}

func (h *AbstractResourceService) createSchema() *RootSchema {
	schema := h.definition.Schema
	// id
	if h.definition.Id != "" {
		schema.UISchema = &UISchema{
			Id: &h.definition.Id,
		}
	}
	return &schema
}
