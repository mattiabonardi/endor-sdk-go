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

func NewAbstractResourceService(resource string, description string, definition ResourceDefinition, mongoClient *mongo.Client, mongoDB string, context context.Context) EndorResource {
	service := AbstractResourceService{
		resource:    resource,
		definition:  definition,
		mongoClient: mongoClient,
		mongoDB:     mongoDB,
		context:     context,
	}
	return EndorResource{
		Resource:    resource,
		Description: description,
		Methods: map[string]EndorResourceAction{
			"schema": NewAction(
				service.schema,
				fmt.Sprintf("Get the schema of the %s (%s)", resource, description),
			),
			"list": NewAction(
				service.list,
				fmt.Sprintf("Search for available list of %s (%s)", resource, description),
			),
			"create": NewAction(
				service.create,
				fmt.Sprintf("Create the instance of %s (%s)", resource, description),
			),
			"instance": NewAction(
				service.instance,
				fmt.Sprintf("Get the instance of %s (%s)", resource, description),
			),
			"update": NewAction(
				service.update,
				fmt.Sprintf("Update the existing instance of %s (%s)", resource, description),
			),
			"delete": NewAction(
				service.delete,
				fmt.Sprintf("Delete the existing instance of %s (%s)", resource, description),
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
